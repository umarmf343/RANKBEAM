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

