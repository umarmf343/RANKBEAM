# RankBeam License API

This service powers the Paystack-backed licensing workflow for RankBeam. It now supports multiple SQLite backends so that Windows users can install it without compiling native binaries.

## Installation

```bash
cd license-api
npm install
```

`better-sqlite3` is declared as an optional dependency. When prebuilt binaries are available for your platform, the API will use `better-sqlite3` for maximum performance. If the native module cannot be installed (for example on Windows without Visual Studio), the API automatically falls back to a pure JavaScript engine powered by [`sql.js`](https://github.com/sql-js/sql.js).

No extra configuration is required. The database file is stored at `data/licenses.db` by default and is created on first run regardless of which backend is active.

## Development

Start the API in watch mode:

```bash
npm run dev
```

## Environment variables

- `PORT` / `LICENSE_API_PORT` – Port to listen on (defaults to `8080`).
- `DATABASE_PATH` – Full path to the SQLite database file. Defaults to `data/licenses.db` next to this README.
- `LICENSE_API_TOKEN` – Shared secret used to protect the `/paystack/validate` and `/paystack/deactivate` routes.
- `PAYSTACK_SECRET_KEY` – Paystack secret key for API requests.
- `PAYSTACK_PLAN_CODE` – Paystack plan code used when starting subscriptions.
- `PAYSTACK_PUBLIC_KEY` – Optional Paystack public key (forwarded in webhook payloads).
- `PAYSTACK_WEBHOOK_IPS` – Comma-separated list of IPs allowed to call the webhook (defaults to Paystack's IPs).
- `DEBUG_SQLITE_FALLBACK` – Set to a truthy value to log the underlying error when `better-sqlite3` is unavailable.

## Database backends

You can inspect which backend is active by looking at the server logs during startup:

- `better-sqlite3` in use – no extra log entry.
- `sql.js` fallback – logs `better-sqlite3 unavailable, using sql.js fallback`.

When using the `sql.js` backend the database is saved to disk after each write so data remains durable across restarts.
