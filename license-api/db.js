import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

let dbInstance;

export async function initDatabase(dbPath) {
  if (!dbPath) {
    throw new Error('DATABASE_PATH is required');
  }
  const dir = path.dirname(dbPath);
  fs.mkdirSync(dir, { recursive: true });

  const adapter = await createAdapter(dbPath);
  adapter.init();
  dbInstance = adapter;
  return dbInstance;
}

function getDb() {
  if (!dbInstance) {
    throw new Error('Database has not been initialised');
  }
  return dbInstance;
}

async function createAdapter(dbPath) {
  try {
    const betterSqlite3Module = await import('better-sqlite3');
    const BetterSqlite3 = betterSqlite3Module.default ?? betterSqlite3Module;
    return new BetterSqliteAdapter(new BetterSqlite3(dbPath));
  } catch (error) {
    if (process.env.DEBUG_SQLITE_FALLBACK) {
      console.warn('better-sqlite3 unavailable, falling back to sql.js', error);
    } else {
      console.warn('better-sqlite3 unavailable, using sql.js fallback');
    }
    return await createSqlJsAdapter(dbPath);
  }
}

class BetterSqliteAdapter {
  constructor(database) {
    this.db = database;
    this.db.pragma('journal_mode = WAL');
  }

  init() {
    this.db.exec(`
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
    this.db.exec('CREATE INDEX IF NOT EXISTS idx_licenses_email ON licenses(user_email);');
    this.db.exec('CREATE INDEX IF NOT EXISTS idx_licenses_status ON licenses(status);');
  }

  run(sql, params) {
    const statement = this.db.prepare(sql);
    if (Array.isArray(params)) {
      statement.run(...params);
    } else if (params && typeof params === 'object') {
      statement.run(params);
    } else if (params !== undefined) {
      statement.run(params);
    } else {
      statement.run();
    }
  }

  get(sql, params) {
    const statement = this.db.prepare(sql);
    if (Array.isArray(params)) {
      return statement.get(...params);
    }
    if (params && typeof params === 'object') {
      return statement.get(params);
    }
    if (params !== undefined) {
      return statement.get(params);
    }
    return statement.get();
  }

  ping() {
    this.db.prepare('SELECT 1').get();
  }
}

async function createSqlJsAdapter(dbPath) {
  const sqlJsModule = await import('sql.js');
  const initSqlJs = sqlJsModule.default ?? sqlJsModule;
  const sqlJs = await initSqlJs({
    locateFile: (file) => path.join(__dirname, 'node_modules', 'sql.js', 'dist', file),
  });
  const existing = fs.existsSync(dbPath) ? fs.readFileSync(dbPath) : null;
  const database = existing ? new sqlJs.Database(existing) : new sqlJs.Database();
  return new SqlJsAdapter(database, dbPath);
}

class SqlJsAdapter {
  constructor(database, dbPath) {
    this.db = database;
    this.dbPath = dbPath;
  }

  init() {
    this.db.exec(`
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
    this.db.exec('CREATE INDEX IF NOT EXISTS idx_licenses_email ON licenses(user_email);');
    this.db.exec('CREATE INDEX IF NOT EXISTS idx_licenses_status ON licenses(status);');
    this.persist();
  }

  run(sql, params) {
    const statement = this.db.prepare(sql);
    bindParams(statement, params);
    statement.step();
    statement.free();
    this.persist();
  }

  get(sql, params) {
    const statement = this.db.prepare(sql);
    bindParams(statement, params);
    let result = null;
    if (statement.step()) {
      result = statement.getAsObject();
    }
    statement.free();
    return result;
  }

  persist() {
    const data = this.db.export();
    const buffer = Buffer.from(data);
    fs.writeFileSync(this.dbPath, buffer);
  }

  ping() {
    this.db.exec('SELECT 1');
  }
}

function bindParams(statement, params) {
  if (params === undefined || params === null) {
    return;
  }
  if (Array.isArray(params)) {
    statement.bind(params.map((value) => (value === undefined ? null : value)));
    return;
  }
  if (typeof params !== 'object') {
    statement.bind([params === undefined ? null : params]);
    return;
  }

  const mapped = {};
  for (const [key, value] of Object.entries(params)) {
    const normalised = value === undefined ? null : value;
    mapped[`:${key}`] = normalised;
    mapped[`$${key}`] = normalised;
    mapped[`@${key}`] = normalised;
  }
  statement.bind(mapped);
}

export function savePendingSubscription({ email, licenseKey, fingerprint, reference }) {
  const db = getDb();
  const now = new Date().toISOString();
  db.run(
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
       updated_at=excluded.updated_at;`,
    { email, licenseKey, fingerprint, reference, now }
  );
}

export function databaseHealth() {
  try {
    const db = getDb();
    if (typeof db.ping === 'function') {
      db.ping();
    } else {
      db.get('SELECT 1 AS ok');
    }
    return { status: 'ok' };
  } catch (error) {
    return { status: 'error', error: error?.message || 'database unavailable' };
  }
}

export function activateLicense({ email, licenseKey, fingerprint, expiresAt, reference }) {
  const db = getDb();
  const now = new Date().toISOString();
  db.run(
    `INSERT INTO licenses (user_email, license_key, fingerprint, expires_at, paystack_reference, status, created_at, updated_at)
     VALUES (@email, @licenseKey, @fingerprint, @expiresAt, @reference, 'active', @now, @now)
     ON CONFLICT(license_key) DO UPDATE SET
       user_email=excluded.user_email,
       fingerprint=CASE
         WHEN licenses.fingerprint IS NULL OR licenses.fingerprint = '' THEN excluded.fingerprint
         WHEN excluded.fingerprint IS NULL OR excluded.fingerprint = '' THEN licenses.fingerprint
         ELSE licenses.fingerprint
       END,
       expires_at=excluded.expires_at,
       paystack_reference=COALESCE(excluded.paystack_reference, licenses.paystack_reference),
       status='active',
       updated_at=excluded.updated_at;`,
    { email, licenseKey, fingerprint, expiresAt, reference, now }
  );
}

export function getLicenseByKey(licenseKey) {
  const db = getDb();
  const row = db.get('SELECT * FROM licenses WHERE license_key = ?', [licenseKey]);
  if (!row) {
    return null;
  }
  if (row.expires_at) {
    const expiresAt = new Date(row.expires_at);
    if (Number.isFinite(expiresAt.valueOf()) && expiresAt < new Date()) {
      db.run("UPDATE licenses SET status = 'expired', updated_at = ? WHERE id = ?", [new Date().toISOString(), row.id]);
      row.status = 'expired';
    }
  }
  return row;
}

export function deactivateLicense(licenseKey) {
  const db = getDb();
  const now = new Date().toISOString();
  db.run(
    `UPDATE licenses
       SET fingerprint = NULL,
           status = 'deactivated',
           updated_at = @now
     WHERE license_key = @licenseKey`,
    { licenseKey, now }
  );
}

export function clearExpiredFingerprints() {
  const db = getDb();
  const nowIso = new Date().toISOString();
  db.run(
    `UPDATE licenses
       SET status = 'expired',
           updated_at = @nowIso
     WHERE status != 'expired' AND expires_at IS NOT NULL AND datetime(expires_at) <= datetime('now');`,
    { nowIso }
  );
}
