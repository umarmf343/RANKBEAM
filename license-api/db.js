import fs from 'fs';
import path from 'path';
import Database from 'better-sqlite3';

let dbInstance;

export function initDatabase(dbPath) {
  if (!dbPath) {
    throw new Error('DATABASE_PATH is required');
  }
  const dir = path.dirname(dbPath);
  fs.mkdirSync(dir, { recursive: true });

  dbInstance = new Database(dbPath);
  dbInstance.pragma('journal_mode = WAL');
  dbInstance.exec(`
    CREATE TABLE IF NOT EXISTS licenses (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      user_email TEXT NOT NULL,
      license_key TEXT NOT NULL UNIQUE,
      fingerprint TEXT,
      expires_at TEXT,
      paystack_reference TEXT,
      status TEXT NOT NULL DEFAULT 'pending',
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );
  `);
  dbInstance.exec('CREATE INDEX IF NOT EXISTS idx_licenses_email ON licenses(user_email);');
  dbInstance.exec('CREATE INDEX IF NOT EXISTS idx_licenses_status ON licenses(status);');
  return dbInstance;
}

function getDb() {
  if (!dbInstance) {
    throw new Error('Database has not been initialised');
  }
  return dbInstance;
}

export function savePendingSubscription({ email, licenseKey, fingerprint, reference }) {
  const db = getDb();
  const now = new Date().toISOString();
  db.prepare(
    `INSERT INTO licenses (user_email, license_key, fingerprint, expires_at, paystack_reference, status, created_at, updated_at)
     VALUES (@email, @licenseKey, @fingerprint, NULL, @reference, 'pending', @now, @now)
     ON CONFLICT(license_key) DO UPDATE SET
       user_email=excluded.user_email,
       fingerprint=CASE
         WHEN licenses.fingerprint IS NULL OR licenses.fingerprint = '' THEN excluded.fingerprint
         ELSE licenses.fingerprint
       END,
       paystack_reference=excluded.paystack_reference,
       status='pending',
       updated_at=excluded.updated_at;`
  ).run({ email, licenseKey, fingerprint, reference, now });
}

export function activateLicense({ email, licenseKey, fingerprint, expiresAt, reference }) {
  const db = getDb();
  const now = new Date().toISOString();
  db.prepare(
    `INSERT INTO licenses (user_email, license_key, fingerprint, expires_at, paystack_reference, status, created_at, updated_at)
     VALUES (@email, @licenseKey, @fingerprint, @expiresAt, @reference, 'active', @now, @now)
     ON CONFLICT(license_key) DO UPDATE SET
       user_email=excluded.user_email,
       fingerprint=CASE
         WHEN licenses.fingerprint IS NULL OR licenses.fingerprint = '' THEN excluded.fingerprint
         WHEN excluded.fingerprint IS NULL OR excluded.fingerprint = '' THEN licenses.fingerprint
         ELSE licenses.fingerprint
       END,
       expires_at=excluded.expiresAt,
       paystack_reference=COALESCE(excluded.reference, licenses.paystack_reference),
       status='active',
       updated_at=excluded.updated_at;`
  ).run({ email, licenseKey, fingerprint, expiresAt, reference, now });
}

export function getLicenseByKey(licenseKey) {
  const db = getDb();
  const row = db.prepare('SELECT * FROM licenses WHERE license_key = ?').get(licenseKey);
  if (!row) {
    return null;
  }
  if (row.expires_at) {
    const expiresAt = new Date(row.expires_at);
    if (Number.isFinite(expiresAt.valueOf()) && expiresAt < new Date()) {
      db.prepare('UPDATE licenses SET status = "expired", updated_at = ? WHERE id = ?').run(new Date().toISOString(), row.id);
      row.status = 'expired';
    }
  }
  return row;
}

export function deactivateLicense(licenseKey) {
  const db = getDb();
  const now = new Date().toISOString();
  db.prepare(
    `UPDATE licenses
       SET fingerprint = NULL,
           status = 'deactivated',
           updated_at = @now
     WHERE license_key = @licenseKey`
  ).run({ licenseKey, now });
}

export function clearExpiredFingerprints() {
  const db = getDb();
  const nowIso = new Date().toISOString();
  db.prepare(
    `UPDATE licenses
       SET status = 'expired',
           updated_at = @nowIso
     WHERE status != 'expired' AND expires_at IS NOT NULL AND datetime(expires_at) <= datetime('now');`
  ).run({ nowIso });
}
