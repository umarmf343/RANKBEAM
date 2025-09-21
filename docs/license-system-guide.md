# License System Implementation Guide

This document describes how to extend the Amazon Product Intelligence Suite desktop application with a machine-bound license system. The solution is split into two cooperating parts:

1. **Client-side installer enhancements** built with Inno Setup that generate a hardware fingerprint during installation, fetch the corresponding license key from a backend, and persist that key securely on the end-user's PC.
2. **Server-side APIs** implemented in Go that mint, store, and validate license keys for every fingerprinted device.

The repository now ships all moving pieces required to operate this workflow:

- [`cmd/license-server`](../cmd/license-server/main.go) – self-hosted REST API that issues and validates keys backed by SQLite.
- [`internal/licensing`](../internal/licensing) – shared database/service layer reused by the server.
- [`cmd/license-seeder`](../cmd/license-seeder/main.go) – helper used by the installer to bind a fresh key to the machine.
- [`internal/licenseclient`](../internal/licenseclient) – desktop app client for validating keys at launch and persisting them securely.
- [`internal/ui/license_gate.go`](../internal/ui/license_gate.go) – Fyne activation screen that prevents access until a key is confirmed.
- [`installer/amazon-product-scraper.iss`](../installer/amazon-product-scraper.iss) – updated script that prompts for customer identity, calls the seeder, and surfaces the resulting key.

Follow the instructions below to wire everything together from build to runtime.

> The walkthrough below is inspired by the licensing flow used by Publisher Rocket and is tailored to the stack that ships with this repository (Go/Fyne desktop client packaged for Windows by Inno Setup).

---

## 1. Prerequisites

| Area | Requirement |
| --- | --- |
| Installer authoring | [Inno Setup 6+](https://jrsoftware.org/isinfo.php) with the bundled ISTool/Inno Script Studio editor. |
| Backend | Go 1.21+, `go install` permissions, and SQLite 3 libraries (or equivalent DB). |
| Networking | Publicly reachable HTTPS endpoint for licensing APIs (self-hosted or cloud). |
| Build artifacts | Compiled desktop binary (e.g. `amazon-product-scraper.exe`) produced via `GOOS=windows GOARCH=amd64 go build ./cmd/app`. |
| Secure storage | Windows Credential Locker, DPAPI, or an application-specific encrypted file for persisting the returned license key. |

---

## 2. Quick start: build, issue, validate

The repository now contains runnable components so you can smoke-test the licensing loop locally before hardening it for production.

1. **Compile the desktop app and helper utilities (Windows targets).**

   ```bash
   GOOS=windows GOARCH=amd64 go build -o bin/amazon-product-scraper.exe ./cmd/app
   GOOS=windows GOARCH=amd64 go build -o bin/license-seeder.exe ./cmd/license-seeder
   ```

2. **Launch the licensing API.**

   ```bash
   go run ./cmd/license-server -addr :8080 -db licenses.db
   ```

   The server exposes `POST /api/v1/licenses` for minting keys and `POST /api/v1/licenses/validate` for runtime checks. Data is stored in `licenses.db` using SQLite.

3. **Simulate what the installer does.**

   ```bash
   # Customer identifier is usually the purchaser's email or order number
   APISUITE_LICENSE_URL=http://localhost:8080 \
     go run ./cmd/license-seeder --customer "demo@example.com" --output ./license-key.txt
   ```

   This command fingerprints the machine, requests a new key, writes it to the same config directory the desktop app uses, and echoes the key to stdout / `license-key.txt`.

4. **Run the desktop UI.**

   ```bash
   APISUITE_LICENSE_URL=http://localhost:8080 go run ./cmd/app
   ```

   On startup the activation screen reads the stored key, validates it against the server, and unlocks the full interface once the response is `valid`.

With the loop verified locally you can move on to installer automation and production deployment details below.

---

## 3. Installer Workflow (Inno Setup)

During the final stages of the installer, run a custom code section to:

1. Derive a machine fingerprint (CPU ID, BIOS serial, MAC address, or TPM UUID).
2. POST the fingerprint to your licensing API to mint a license key.
3. Display the key to the customer and store it locally for future validations.

### 3.1. Base Script Skeleton

Create a new script (or extend `installer/amazon-product-suite.iss`) with the following sections:

```pascal
[Setup]
AppName=Amazon Product Intelligence Suite
AppVersion=1.0.0
DefaultDirName={autopf}\Amazon Product Intelligence Suite
DefaultGroupName=Amazon Product Intelligence Suite
DisableProgramGroupPage=yes
OutputDir=dist
OutputBaseFilename=amazon-product-suite

[Files]
Source: "..\bin\amazon-product-scraper.exe"; DestDir: "{app}"; Flags: ignoreversion

[Run]
Filename: "{app}\amazon-product-scraper.exe"; Parameters: "/licensekey={code:GetLicenseKey}"; Flags: runhidden

[Code]
var
  GeneratedKey: string;

function GetLicenseKey(Param: string): string;
var
  Fingerprint, Response: string;
begin
  Fingerprint := GetMachineFingerprint();
  Response := RequestLicenseFromServer(Fingerprint);
  GeneratedKey := ParseLicenseKey(Response);
  if GeneratedKey = '' then
    MsgBox('Unable to obtain license key. Please contact support.', mbError, MB_OK)
  else
    PersistLicenseKey(GeneratedKey);
  Result := GeneratedKey;
end;
```

### 3.2. Hardware Fingerprinting

Use Windows Management Instrumentation (WMI) via `CreateOleObject('WbemScripting.SWbemLocator')` to gather unique identifiers. Combine multiple attributes to reduce collisions. Example helper:

```pascal
function GetMachineFingerprint(): string;
var
  Locator, Services, Items, Item: Variant;
  CpuID, BiosSerial, DiskSerial: string;
begin
  Locator := CreateOleObject('WbemScripting.SWbemLocator');
  Services := Locator.ConnectServer('.', 'root\\cimv2');

  Items := Services.ExecQuery('SELECT ProcessorId FROM Win32_Processor');
  for Item in Items do CpuID := Item.ProcessorId;

  Items := Services.ExecQuery('SELECT SerialNumber FROM Win32_BIOS');
  for Item in Items do BiosSerial := Item.SerialNumber;

  Items := Services.ExecQuery('SELECT SerialNumber FROM Win32_DiskDrive WHERE MediaType="Fixed hard disk media"');
  for Item in Items do begin
    DiskSerial := Item.SerialNumber;
    break;
  end;

  Result := GetMD5OfString(Uppercase(CpuID + '|' + BiosSerial + '|' + DiskSerial));
end;
```

> **Tip:** Ship a small helper executable (written in Go) that prints your canonical fingerprint to stdout, then call it via `Exec` to keep the script lean and to reuse logic inside the desktop app.

### 3.3. HTTPS Requests from Inno Setup

Leverage the built-in `THTTPSend` type (via InnoTools Downloader) or a lightweight helper executable to communicate with your API. The PascalScript snippet below uses WinHTTP:

```pascal
function RequestLicenseFromServer(const Fingerprint: string): string;
var
  WinHttpReq: Variant;
  Url, Payload: string;
begin
  Url := 'https://licensing.yourdomain.com/api/v1/licenses';
  Payload := '{"fingerprint":"' + Fingerprint + '"}';

  WinHttpReq := CreateOleObject('WinHttp.WinHttpRequest.5.1');
  WinHttpReq.Open('POST', Url, False);
  WinHttpReq.SetRequestHeader('Content-Type', 'application/json');
  WinHttpReq.Send(Payload);

  if WinHttpReq.Status <> 201 then begin
    Log(Format('License request failed: %d %s', [WinHttpReq.Status, WinHttpReq.ResponseText]));
    Result := '';
    exit;
  end;

  Result := WinHttpReq.ResponseText;
end;
```

### 3.4. Parsing and Persisting the Key

```pascal
function ParseLicenseKey(const Response: string): string;
var
  Json: Variant;
begin
  Json := CreateOleObject('Chilkat_9_5_0.JsonObject');
  if Json.Load(Response) then
    Result := Json.StringOf('licenseKey')
  else
    Result := '';
end;

procedure PersistLicenseKey(const Key: string);
var
  StoragePath: string;
begin
  StoragePath := ExpandConstant('{commonappdata}') + '\\AmazonProductSuite\\license.dat';
  ForceDirectories(ExtractFilePath(StoragePath));
  SaveStringToFile(StoragePath, Key, False);
end;
```

Display the key using `MsgBox` after `GeneratedKey` is set, and remind users to store it safely. For high security, encrypt the file contents with Windows DPAPI or rely on the Go application to store it inside the Windows Credential Locker on first launch.

---

## 4. Server-Side API (Go)

The server issues and validates license keys tied to machine fingerprints. A minimal but production-ready layout is:

```
cmd/
  licensing-server/
    main.go
internal/
  config/
    config.go
  http/
    middleware.go
    router.go
  license/
    generator.go
    handler.go
    repository.go
  storage/
    sqlite/
      migrations.sql
      repository.go
```

### 3.1. Database Layer

Use SQLite for a lightweight deployment or swap in PostgreSQL/MySQL by replacing the driver. Example initialization (`internal/storage/sqlite/repository.go`):

```go
package sqlite

import (
    "context"
    "database/sql"
    "embed"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

//go:embed migrations.sql
var migrations embed.FS

type Store struct {
    db *sql.DB
}

func New(path string) (*Store, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    if err := applyMigrations(db); err != nil {
        return nil, err
    }
    return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) InsertLicense(ctx context.Context, key, fingerprint, customerID string) error {
    _, err := s.db.ExecContext(ctx, `INSERT INTO licenses (key, fingerprint, customer_id) VALUES (?, ?, ?)`, key, fingerprint, customerID)
    return err
}

func (s *Store) LookupLicense(ctx context.Context, key string) (string, error) {
    var fingerprint string
    err := s.db.QueryRowContext(ctx, `SELECT fingerprint FROM licenses WHERE key = ?`, key).Scan(&fingerprint)
    return fingerprint, err
}
```

`migrations.sql` seeds the schema:

```sql
CREATE TABLE IF NOT EXISTS licenses (
    key TEXT PRIMARY KEY,
    fingerprint TEXT NOT NULL,
    customer_id TEXT,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3.2. License Generation Logic

`internal/license/generator.go`:

```go
package license

import (
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func GenerateKey(customerID, fingerprint string) (string, error) {
    raw := make([]byte, 16)
    if _, err := rand.Read(raw); err != nil {
        return "", fmt.Errorf("read entropy: %w", err)
    }

    payload := base32.NewEncoding(alphabet).WithPadding(base32.NoPadding).EncodeToString(raw)
    parts := []string{
        strings.ToUpper(customerID),
        fingerprint[:8],
        payload[:5], payload[5:10], payload[10:15], payload[15:20],
    }
    return strings.Join(parts, "-"), nil
}
```

### 3.3. HTTP Handlers

`internal/license/handler.go` wires the storage and generator together:

```go
package license

import (
    "crypto/subtle"
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yourorg/amazon-product-suite/internal/storage/sqlite"
)

type Handler struct {
    store *sqlite.Store
}

func NewHandler(store *sqlite.Store) *Handler { return &Handler{store: store} }

func (h *Handler) Register(r chi.Router) {
    r.Post("/licenses", h.generate)
    r.Post("/licenses/validate", h.validate)
}

func (h *Handler) generate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        CustomerID  string `json:"customerId"`
        Fingerprint string `json:"fingerprint"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid payload", http.StatusBadRequest)
        return
    }

    key, err := GenerateKey(req.CustomerID, req.Fingerprint)
    if err != nil {
        http.Error(w, "unable to create license", http.StatusInternalServerError)
        return
    }

    if err := h.store.InsertLicense(r.Context(), key, req.Fingerprint, req.CustomerID); err != nil {
        http.Error(w, "persist license", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"licenseKey": key})
}

func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        LicenseKey  string `json:"licenseKey"`
        Fingerprint string `json:"fingerprint"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid payload", http.StatusBadRequest)
        return
    }

    storedFingerprint, err := h.store.LookupLicense(r.Context(), req.LicenseKey)
    if err != nil {
        http.Error(w, "license not found", http.StatusUnauthorized)
        return
    }

    if subtle.ConstantTimeCompare([]byte(storedFingerprint), []byte(req.Fingerprint)) != 1 {
        http.Error(w, "fingerprint mismatch", http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
}
```

Wrap the handlers in a small `main.go` that loads configuration, enables HTTPS-only traffic (use Caddy, nginx, or Let’s Encrypt), and installs middleware for logging, rate limiting, and API authentication (e.g., HMAC or JWT).

### 3.4. Deployment Checklist

- **Environment variables**: `LICENSE_DB_PATH`, `LICENSE_API_KEY`, `LICENSE_ISSUER_DOMAIN`.
- **Rate limiting**: Apply IP-based throttling (e.g., with [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)).
- **Observability**: Emit structured logs and metrics for license generation/validation volume.
- **Backups**: Rotate encrypted database backups. SQLite can be replicated via `.backup` cron jobs.

---

## 5. Client Runtime Validation Flow

1. On each application launch, calculate the local machine fingerprint using the same algorithm as the installer (reuse the helper executable or shared Go package).
2. Load the stored license key from disk/Credential Locker.
3. Call `POST https://licensing.yourdomain.com/api/v1/licenses/validate` with `{"licenseKey": "...", "fingerprint": "..."}`.
4. Deny access if the server returns `401` or the status differs from `valid`. Cache successful validations for a short period (e.g., 24 hours) to tolerate transient network outages.
5. Offer a "Transfer license" flow that revokes the previous fingerprint via a support endpoint if the user upgrades hardware.

---

## 6. Security Best Practices

- **HTTPS everywhere**: Terminate TLS at your load balancer or reverse proxy. Reject plain HTTP requests in the handler.
- **API Authentication**: Require an installer API key or signed JWT when minting licenses to prevent abuse.
- **Least privilege**: Run the license server with a dedicated service account and minimal filesystem permissions.
- **Anti-tampering**: Obfuscate the desktop binary and verify the integrity of the stored license key using an HMAC that the server can check.
- **Monitoring**: Alert on repeated failed validation attempts or spikes in license issuance.
- **Data privacy**: Hash fingerprints before storing to protect user hardware identifiers.

---

## 7. Putting It All Together

1. **Build** the Go/Fyne application for Windows (`GOOS=windows GOARCH=amd64`).
2. **Launch** the licensing API server (Docker container or systemd service) with HTTPS enabled.
3. **Compile** the Inno Setup installer with the custom code that fingerprints hardware, requests a key, stores it, and displays it to the user.
4. **Distribute** the installer. Upon installation, each machine receives a unique license key tied to its fingerprint.
5. **Validate** the license on every app startup and surface clear messaging for invalid or expired keys.

By following this guide, you can control installations on a per-device basis while providing a professional onboarding flow that mirrors Publisher Rocket's licensing experience.
