import 'dotenv/config';
import crypto from 'crypto';
import express from 'express';
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
  databaseHealth,
} from './db.js';
import { initializeTransaction } from './paystack.js';

const truthyValues = new Set(['1', 'true', 'yes', 'on']);
const falsyValues = new Set(['0', 'false', 'no', 'off']);

function readBooleanEnv(name, defaultValue = false) {
  const raw = String(process.env[name] || '').trim().toLowerCase();
  if (!raw) {
    return defaultValue;
  }
  if (truthyValues.has(raw)) {
    return true;
  }
  if (falsyValues.has(raw)) {
    return false;
  }
  return defaultValue;
}

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const DEFAULT_INSTALLER_TOKEN =
  '7c9012993daa2abb40170bab55e1f88d2b24a9601afdec9958a302ce9ba9c43f';

const DEFAULT_PAYSTACK_WEBHOOK_IPS = ['52.31.139.75', '52.49.173.169', '52.214.14.220'];

const allowLocalWebhookIps = readBooleanEnv(
  'PAYSTACK_WEBHOOK_ALLOW_LOCAL',
  process.env.NODE_ENV !== 'production',
);
const allowLocalValidationRequests = readBooleanEnv(
  'LICENSE_API_ALLOW_LOCAL',
  process.env.NODE_ENV !== 'production',
);

const defaultWebhookIps = [...DEFAULT_PAYSTACK_WEBHOOK_IPS];
if (allowLocalWebhookIps) {
  defaultWebhookIps.push('127.0.0.1', '::1');
}

const webhookIpSource = (process.env.PAYSTACK_WEBHOOK_IPS || '').trim();

const configuredWebhookIps = (webhookIpSource || defaultWebhookIps.join(','))
  .split(',')
  .map((ip) => normaliseIpAddress(ip))
  .filter(Boolean);

const trustedPaystackIps = new Set(configuredWebhookIps);
const LOCAL_IPS = new Set(['127.0.0.1', '::1']);

function normaliseIpAddress(ip) {
  if (!ip) {
    return '';
  }
  const trimmed = String(ip).trim();
  if (!trimmed) {
    return '';
  }
  const withoutPrefix = trimmed.replace(/^::ffff:/, '');
  if (withoutPrefix === '::1') {
    return '127.0.0.1';
  }
  return withoutPrefix;
}

function isTrustedPaystackRequest(req) {
  if (!trustedPaystackIps.size) {
    return false;
  }
  for (const ip of collectCandidateIps(req)) {
    if (trustedPaystackIps.has(ip)) {
      return true;
    }
  }
  return false;
}

function isLocalRequest(req) {
  for (const ip of collectCandidateIps(req)) {
    if (LOCAL_IPS.has(ip)) {
      return true;
    }
  }
  return false;
}

function collectCandidateIps(req) {
  const candidates = new Set();
  const forwarded = req.headers['x-forwarded-for'];
  if (forwarded) {
    forwarded
      .split(',')
      .map((part) => normaliseIpAddress(part))
      .filter(Boolean)
      .forEach((ip) => candidates.add(ip));
  }
  const directIp = normaliseIpAddress(req.ip);
  const socketIp = normaliseIpAddress(req.socket?.remoteAddress);
  if (directIp) {
    candidates.add(directIp);
  }
  if (socketIp) {
    candidates.add(socketIp);
  }
  return candidates;
}

const app = express();

app.use(cors());
app.use(express.json());

const PORT = Number(process.env.PORT || process.env.LICENSE_API_PORT || 8080);
const databasePath = process.env.DATABASE_PATH || path.join(__dirname, 'data', 'licenses.db');
try {
  await initDatabase(databasePath);
  clearExpiredFingerprints();
} catch (error) {
  console.error('Failed to initialise database', error);
  process.exit(1);
}

function normaliseEmail(email) {
  return String(email || '')
    .trim()
    .toLowerCase();
}

function ensureToken(req, res) {
  if (allowLocalValidationRequests && isLocalRequest(req)) {
    return true;
  }
  const expected = process.env.LICENSE_API_TOKEN || DEFAULT_INSTALLER_TOKEN;
  if (!expected) {
    return true;
  }
  const provided =
    (req.get('X-License-Token') || req.get('X-Installer-Token') || '').trim();
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

const HEALTH_ENDPOINTS = ['/health', '/healthz', '/readyz'];

function healthSnapshot() {
  const dbStatus = databaseHealth();
  const healthy = dbStatus.status === 'ok';
  return {
    overallStatus: healthy ? 'ok' : 'degraded',
    httpStatus: healthy ? 200 : 503,
    payload: {
      status: healthy ? 'ok' : 'degraded',
      timestamp: new Date().toISOString(),
      uptimeSeconds: Math.round(process.uptime()),
      checks: {
        database: dbStatus,
      },
    },
  };
}

HEALTH_ENDPOINTS.forEach((path) => {
  app.get(path, (req, res) => {
    const snapshot = healthSnapshot();
    res.status(snapshot.httpStatus).json(snapshot.payload);
  });

  app.head(path, (req, res) => {
    const snapshot = healthSnapshot();
    res.sendStatus(snapshot.httpStatus === 200 ? 204 : snapshot.httpStatus);
  });
});

app.get('/paystack/subscribe', (req, res) => {
  res
    .status(405)
    .json({
      error: 'POST required',
      message:
        'Send a POST request with JSON body {"email": "user@example.com", "fingerprint": "DEVICE-ID"} to start a subscription.',
    });
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
  if (!isTrustedPaystackRequest(req)) {
    console.warn('webhook request from untrusted IP', {
      ip: req.ip,
      socketIp: req.socket?.remoteAddress,
      forwardedFor: req.headers['x-forwarded-for'],
    });
    return res.status(403).json({ error: 'untrusted source' });
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
