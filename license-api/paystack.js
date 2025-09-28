import crypto from 'crypto';
import axios from 'axios';

const PAYSTACK_BASE_URL = 'https://api.paystack.co';

export async function initializeTransaction({ email, planCode, reference, metadata }) {
  const secretKey = process.env.PAYSTACK_SECRET_KEY;
  if (!secretKey) {
    throw new Error('PAYSTACK_SECRET_KEY must be configured');
  }
  if (!email) {
    throw new Error('email is required');
  }
  if (!planCode) {
    throw new Error('PAYSTACK_PLAN_CODE must be configured');
  }

  const response = await axios.post(
    `${PAYSTACK_BASE_URL}/transaction/initialize`,
    {
      email,
      plan: planCode,
      reference,
      metadata,
    },
    {
      headers: {
        Authorization: `Bearer ${secretKey}`,
        'Content-Type': 'application/json',
      },
      timeout: 15000,
    }
  );

  return response.data;
}

export function verifyWebhookSignature(payload, headerSignature) {
  const secret = process.env.PAYSTACK_WEBHOOK_SECRET;
  if (!secret) {
    throw new Error('PAYSTACK_WEBHOOK_SECRET must be configured');
  }
  if (!headerSignature) {
    return false;
  }
  const computed = crypto
    .createHmac('sha512', secret)
    .update(payload)
    .digest('hex');
  const providedBuffer = Buffer.from(headerSignature, 'hex');
  const computedBuffer = Buffer.from(computed, 'hex');
  if (providedBuffer.length !== computedBuffer.length) {
    return false;
  }
  return crypto.timingSafeEqual(computedBuffer, providedBuffer);
}
