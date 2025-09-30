import axios from 'axios';

const PAYSTACK_BASE_URL = 'https://api.paystack.co';

const truthyValues = new Set(['1', 'true', 'yes', 'on']);
const falsyValues = new Set(['0', 'false', 'no', 'off']);

function isMockModeEnabled() {
  const configured = String(process.env.PAYSTACK_USE_MOCK || '').trim().toLowerCase();
  if (truthyValues.has(configured)) {
    return true;
  }
  if (falsyValues.has(configured)) {
    return false;
  }
  return process.env.NODE_ENV !== 'production';
}

function createMockResponse({ email, planCode, reference }) {
  const mockReference = reference || `MOCK-${Date.now()}`;
  return {
    status: true,
    message: 'Mock transaction initialised',
    data: {
      authorization_url: `https://paystack.mock/checkout/${encodeURIComponent(mockReference)}`,
      access_code: `MOCK-${mockReference}`,
      reference: mockReference,
      metadata: {
        email,
        planCode,
        mock: true,
      },
    },
    mock: true,
  };
}

function isIpRestrictionError(error) {
  const message = String(error?.response?.data?.message || error?.message || '').toLowerCase();
  return message.includes('ip address is not allowed');
}

export async function initializeTransaction({ email, planCode, reference, metadata }) {
  if (!email) {
    throw new Error('email is required');
  }
  if (!planCode) {
    throw new Error('PAYSTACK_PLAN_CODE must be configured');
  }

  const secretKey = process.env.PAYSTACK_SECRET_KEY;
  const mockEnabled = isMockModeEnabled();

  if (!secretKey && mockEnabled) {
    return createMockResponse({ email, planCode, reference });
  }
  if (!secretKey) {
    throw new Error('PAYSTACK_SECRET_KEY must be configured');
  }

  try {
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
  } catch (error) {
    if (mockEnabled && isIpRestrictionError(error)) {
      console.warn('Falling back to mock Paystack transaction due to IP restriction');
      return createMockResponse({ email, planCode, reference });
    }
    throw error;
  }
}

