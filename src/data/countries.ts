export type CountryConfig = {
  code: string;
  label: string;
  currency: string;
  host: string;
  marketplaceId: string;
};

export const COUNTRIES: CountryConfig[] = [
  { code: "US", label: "United States", currency: "USD", host: "www.amazon.com", marketplaceId: "ATVPDKIKX0DER" },
  { code: "CA", label: "Canada", currency: "CAD", host: "www.amazon.ca", marketplaceId: "A2EUQ1WTGCTBG2" },
  { code: "GB", label: "United Kingdom", currency: "GBP", host: "www.amazon.co.uk", marketplaceId: "A1F83G8C2ARO7P" },
  { code: "DE", label: "Germany", currency: "EUR", host: "www.amazon.de", marketplaceId: "A1PA6795UKMFR9" },
  { code: "FR", label: "France", currency: "EUR", host: "www.amazon.fr", marketplaceId: "A13V1IB3VIYZZH" },
  { code: "ES", label: "Spain", currency: "EUR", host: "www.amazon.es", marketplaceId: "A1RKKUPIHCS9HS" },
  { code: "IT", label: "Italy", currency: "EUR", host: "www.amazon.it", marketplaceId: "APJ6JRA9NG5V4" },
  { code: "IN", label: "India", currency: "INR", host: "www.amazon.in", marketplaceId: "A21TJRUUN4KGV" },
  { code: "JP", label: "Japan", currency: "JPY", host: "www.amazon.co.jp", marketplaceId: "A1VC38T7YXB528" },
  { code: "AU", label: "Australia", currency: "AUD", host: "www.amazon.com.au", marketplaceId: "A39IBJ37TRP1C6" },
  { code: "AT", label: "Austria", currency: "EUR", host: "www.amazon.at", marketplaceId: "A1PA6795UKMFR9" },
  { code: "BR", label: "Brazil", currency: "BRL", host: "www.amazon.com.br", marketplaceId: "A2Q3Y263D00KWC" },
  { code: "MX", label: "Mexico", currency: "MXN", host: "www.amazon.com.mx", marketplaceId: "A1AM78C64UM0Y8" },
  { code: "AE", label: "United Arab Emirates", currency: "AED", host: "www.amazon.ae", marketplaceId: "A2VIGQ35RCS4UG" },
  { code: "SG", label: "Singapore", currency: "SGD", host: "www.amazon.sg", marketplaceId: "A19VAU5U5O7RUS" }
];

export const DISPLAY_CODE_MAP: Record<string, string> = {
  GB: "UK"
};

export function resolveCountry(code: string): CountryConfig {
  const normalized = code.trim().toUpperCase();
  const canonical = Object.entries(DISPLAY_CODE_MAP).find(([, alias]) => alias === normalized)?.[0] ?? normalized;
  return COUNTRIES.find((c) => c.code === canonical) ?? COUNTRIES[0];
}
