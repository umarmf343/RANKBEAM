import 'dotenv/config';
import crypto from 'crypto';
import express from 'express';
import bodyParser from 'body-parser';
import cors from 'cors';
import path from 'path';
import { fileURLToPath } from 'url';

import {
  activateLicense,
  clearExpiredFingerprints,
  deactivateLicense,
  getLicenseByKey,
  initDatabase,
  savePendingSubscription,
} from './db.js';
import { initializeTransaction, verifyWebhookSignature } from './paystack.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();

app.use(cors());
app.use(
  bodyParser.json({
    verify: (req, res, buf) => {
      req.rawBody = buf;
    },
  })
);

const PORT = Number(process.env.PORT || process.env.LICENSE_API_PORT || 8080);
const databasePath = process.env.DATABASE_PATH || path.join(__dirname, 'data', 'licenses.db');
initDatabase(databasePath);
clearExpiredFingerprints();

function normaliseEmail(email) {
  return String(email || '')
    .trim()
    .toLowerCase();
}

function ensureToken(req, res) {
  const expected = process.env.LICENSE_API_TOKEN;
  if (!expected) {
    return true;
  }
  const provided = req.get('X-License-Token') || req.get('X-Installer-Token');
  if (!provided || provided !== expected) {
    res.status(403).json({ error: 'Forbidden' });
    return false;
  }
  return true;
}

function createLicenseKey(email) {
  const normalized = normaliseEmail(email);
  const emailHash = crypto.createHash('sha256').update(normalized).digest('hex').slice(0, 12);
  const randomPart = crypto.randomBytes(6).toString('hex');
  const timePart = Date.now().toString(36).toUpperCase();
  return `${emailHash}-${randomPart}-${timePart}`.toUpperCase();
}

function calculateExpiry(baseDate) {
  const paidDate = baseDate ? new Date(baseDate) : new Date();
  if (!Number.isFinite(paidDate.valueOf())) {
    return new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString();
  }
  const expires = new Date(paidDate.getTime() + 30 * 24 * 60 * 60 * 1000);
  return expires.toISOString();
}

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.post('/paystack/subscribe', async (req, res) => {
  const email = normaliseEmail(req.body?.email);
  const fingerprint = String(req.body?.fingerprint || '').trim();
  const planCode = req.body?.planCode || process.env.PAYSTACK_PLAN_CODE;

  if (!email) {
    return res.status(400).json({ error: 'email is required' });
  }
  if (!fingerprint) {
    return res.status(400).json({ error: 'fingerprint is required' });
  }
  if (!planCode) {
    return res.status(500).json({ error: 'PAYSTACK_PLAN_CODE is not configured' });
  }

  try {
    const licenseKey = createLicenseKey(email);
    const reference = `RB-${Date.now()}-${crypto.randomBytes(4).toString('hex').toUpperCase()}`;

    savePendingSubscription({ email, licenseKey, fingerprint, reference });

    const metadata = {
      licenseKey,
      fingerprint,
    };

    const response = await initializeTransaction({
      email,
      planCode,
      reference,
      metadata,
    });

    return res.json({
      status: 'pending',
      authorizationUrl: response?.data?.authorization_url,
      accessCode: response?.data?.access_code,
      reference,
      licenseKey,
    });
  } catch (error) {
    console.error('subscribe error', error?.response?.data || error);
    const message = error?.response?.data?.message || error?.message || 'Unable to start subscription';
    return res.status(502).json({ error: message });
  }
});

app.post('/paystack/validate', (req, res) => {
  if (!ensureToken(req, res)) {
    return;
  }
  const email = normaliseEmail(req.body?.email);
  const licenseKey = String(req.body?.licenseKey || '').trim().toUpperCase();
  const fingerprint = String(req.body?.fingerprint || '').trim();

  if (!email || !licenseKey || !fingerprint) {
    return res.status(400).json({ error: 'licenseKey, email and fingerprint are required' });
  }

  const record = getLicenseByKey(licenseKey);
  if (!record) {
    return res.status(401).json({ error: 'license not found' });
  }
  if (normaliseEmail(record.user_email) !== email) {
    return res.status(401).json({ error: 'email mismatch' });
  }
  if (record.status === 'expired') {
    return res.status(401).json({ error: 'subscription expired', expiresAt: record.expires_at });
  }
  if (record.status === 'pending') {
    return res.status(409).json({ error: 'payment pending' });
  }
  if (record.status === 'deactivated') {
    return res.status(401).json({ error: 'license deactivated' });
  }
  if (!record.expires_at) {
    return res.status(401).json({ error: 'subscription inactive' });
  }
  if (record.fingerprint && record.fingerprint !== fingerprint) {
    return res.status(401).json({ error: 'fingerprint mismatch' });
  }

  if (!record.fingerprint) {
    activateLicense({
      email,
      licenseKey,
      fingerprint,
      expiresAt: record.expires_at,
      reference: record.paystack_reference,
    });
  }

  return res.json({ status: 'valid', expiresAt: record.expires_at, licenseKey: record.license_key });
});

app.post('/paystack/deactivate', (req, res) => {
  if (!ensureToken(req, res)) {
    return;
  }
  const licenseKey = String(req.body?.licenseKey || '').trim().toUpperCase();
  if (!licenseKey) {
    return res.status(400).json({ error: 'licenseKey is required' });
  }
  deactivateLicense(licenseKey);
  return res.json({ status: 'deactivated' });
});

app.post('/paystack/webhook', (req, res) => {
  const signature = req.get('x-paystack-signature');
  const rawPayload = req.rawBody?.toString('utf8') || '';

  let verified = false;
  try {
    verified = verifyWebhookSignature(rawPayload, signature);
  } catch (error) {
    console.error('webhook verification error', error.message);
    return res.status(500).json({ error: 'webhook verification error' });
  }

  if (!verified) {
    return res.status(400).json({ error: 'invalid signature' });
  }

  const payload = req.body || {};
  const event = payload.event;
  const supportedEvents = new Set([
    'subscription.create',
    'charge.success',
    'invoice.create',
    'subscription.renew',
  ]);
  if (event && !supportedEvents.has(event)) {
    console.info(`ignoring unsupported webhook event: ${event}`);
    return res.status(202).json({ status: 'ignored', reason: 'unsupported event' });
  }
  const data = payload.data || {};
  const paidAt = data.paid_at || data.paidAt || data.created_at || data.createdAt || new Date().toISOString();
  const metadata = data.metadata || {};
  const licenseKey = String(metadata.licenseKey || metadata.license_key || '').trim().toUpperCase();
  const fingerprint = String(metadata.fingerprint || metadata.device_id || '').trim();
  const email = normaliseEmail(data.customer?.email || payload.customer?.email || metadata.email);
  const reference = data.reference || data.subscription_code || data.subscription?.subscription_code || metadata.reference;

  if (!licenseKey || !email) {
    console.warn('webhook received without licenseKey/email', payload);
    return res.status(202).json({ status: 'ignored' });
  }

  const expiresAt = calculateExpiry(paidAt);
  activateLicense({ email, licenseKey, fingerprint, expiresAt, reference });

  return res.json({ status: 'processed', licenseKey, expiresAt });
});

app.use((err, req, res, next) => {
  console.error('Unhandled error', err);
  res.status(500).json({ error: 'internal server error' });
});

app.listen(PORT, () => {
  console.log(`RankBeam licensing API listening on port ${PORT}`);
});
