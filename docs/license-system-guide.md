# Comprehensive License System Guide (Inno Setup + Go Server)

This guide walks you through replicating a "Publisher Rocket" style licensing flow for the Amazon Product Intelligence Suite. The solution has two cooperating halves:

1. **Client-side installer automation** authored with **Inno Setup**. During setup it fingerprints the machine, calls your licensing API, returns a license key to the customer, and stores it for future validations.
2. **Server-side APIs** written in **Go** that mint, store, and validate licenses that are bound to hardware fingerprints.

The document is self-contained so that you can integrate licensing without needing additional references.

---

## 0. Prerequisites

| Area | Requirement |
| --- | --- |
| Installer tooling | [Inno Setup 6+](https://jrsoftware.org/isinfo.php) with the built-in script editor or Inno Script Studio. |
| Desktop binary | Windows build of your Go/Fyne application (e.g. `GOOS=windows GOARCH=amd64 go build ./cmd/app`). |
| Backend | Go 1.21+, SQLite 3 (or preferred DB), and ability to expose an HTTPS endpoint. |
| Networking | Outbound HTTPS connectivity from installers and clients to your API domain. |
| Security | Access to Windows Credential Locker or DPAPI if you plan to encrypt stored keys. |

---

## 1. Client Side – Inno Setup Integration

The installer will: (a) collect a customer identifier, (b) derive a machine fingerprint, (c) request a license key from the API, (d) persist the key locally, and (e) surface the key to the user. All of this runs automatically as part of installation.

### 1.1 Installer Script Layout

Create `installer/amazon-product-suite.iss` (or adapt your existing script) with the following skeleton:

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
Source: "..\bin\fingerprint-helper.exe"; DestDir: "{tmp}"; Flags: deleteafterinstall

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
  if GeneratedKey = '' then begin
    MsgBox('Unable to obtain license key. Please contact support.', mbError, MB_OK);
    Result := '';
  end else begin
    PersistLicenseKey(GeneratedKey);
    ShowLicenseKey(GeneratedKey);
    Result := GeneratedKey;
  end;
end;
```

Key setup pages to add:

- **Customer information**: use the `UserInfo` support in Inno Setup or build a custom wizard page to collect name/email. Keep the entered value in a global variable so that `RequestLicenseFromServer` can include it.
- **Firewall exception (optional)**: ensure the installer can make outbound HTTPS calls by asking Windows Firewall for temporary access if required.

### 1.2 Machine Fingerprinting

For predictability, ship a helper binary that calculates the fingerprint (e.g. a Go CLI that you also use inside the desktop app). You can still use WMI directly if you prefer PascalScript.

#### Option A: WMI in PascalScript

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

#### Option B: Helper Executable

1. Build `fingerprint-helper.exe` from the repository's helper command so it mirrors the desktop app's fingerprint logic:

   ```bash
   GOOS=windows GOARCH=amd64 go build -o bin/fingerprint-helper.exe ./cmd/fingerprint-helper
   ```

2. Run it from the installer and capture the generated fingerprint into a temporary file:

```pascal
function GetMachineFingerprint(): string;
var
  ResultCode: Integer;
  Output: AnsiString;
begin
  if not Exec(ExpandConstant('{tmp}') + '\\fingerprint-helper.exe', '--output "' + ExpandConstant('{tmp}') + '\\fingerprint.out"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
    raise Exception.Create('Fingerprint helper failed');
  LoadStringFromFile(ExpandConstant('{tmp}') + '\\fingerprint.out', Output);
  Result := Trim(Output);
end;
```

The helper also prints the fingerprint to stdout, so you can debug by running it manually without the `--output` parameter. Store the helper output in a temp file before deleting it in the `[Run]` cleanup step.

### 1.3 Talking to the License API

Use WinHTTP (built into Windows) to make HTTPS requests. If you need TLS 1.2+, explicitly set the option.

```pascal
function RequestLicenseFromServer(const Fingerprint: string): string;
var
  WinHttpReq: Variant;
  Url, Payload: string;
  CustomerEmail: string;
begin
  CustomerEmail := WizardForm.UserInfoPage.Values[1]; // adapt to your data capture
  Url := 'https://licensing.yourdomain.com/api/v1/licenses';
  Payload := '{"customerId":"' + CustomerEmail + '","fingerprint":"' + Fingerprint + '"}';

  WinHttpReq := CreateOleObject('WinHttp.WinHttpRequest.5.1');
  WinHttpReq.Open('POST', Url, False);
  WinHttpReq.Option[9] := 128; // WINHTTP_OPTION_SECURE_PROTOCOLS -> TLS1.2
  WinHttpReq.SetRequestHeader('Content-Type', 'application/json');
  WinHttpReq.SetRequestHeader('X-Installer-Token', '{#LicenseApiToken}');
  WinHttpReq.Send(Payload);

  if WinHttpReq.Status <> 201 then begin
    Log(Format('License request failed: %d %s', [WinHttpReq.Status, WinHttpReq.ResponseText]));
    Result := '';
    exit;
  end;

  Result := WinHttpReq.ResponseText;
end;
```

### 1.4 Parsing and Persisting the Key

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
  StoragePath := ExpandConstant('{localappdata}') + '\\AmazonProductSuite\\license.key';
  ForceDirectories(ExtractFilePath(StoragePath));
  if not SaveStringToFile(StoragePath, Key, False) then
    Log('Failed to store license key at ' + StoragePath);
end;

procedure ShowLicenseKey(const Key: string);
begin
  MsgBox('Installation complete! Your license key is:\n\n' + Key + '\n\nIt has been saved automatically. Please store it for your records.',
    mbInformation, MB_OK);
end;
```

### 1.5 Verifying on Application Launch

- When the application starts, recompute the fingerprint using the same helper logic.
- Load the cached license key from `license.key`.
- Call `POST /api/v1/licenses/validate` (documented below) to confirm the key matches the fingerprint.
- Cache the result for 12–24 hours to handle offline launches. When offline, fall back to the last successful validation timestamp.

---

## 2. Server Side – License Generation API (Go)

The API exposes two endpoints:

1. `POST /api/v1/licenses` – create a license bound to a fingerprint.
2. `POST /api/v1/licenses/validate` – confirm a license/fingerprint pair is valid.

The sample implementation uses SQLite and the standard `net/http` package for readability. Adjust to match your infrastructure.

### 2.1 Repository Layout

This repository now ships a working reference implementation in [`server/`](../server/):

```
server/
├── server.go              # HTTP bootstrap, logging middleware, graceful shutdown
├── db.go                  # SQLite-backed LicenseStore with create/validate helpers
├── license_gen.go         # Fingerprint hashing + license key formatting
├── license_handler.go     # `POST /api/v1/licenses` and `/api/v1/licenses/validate`
├── license_*_test.go      # Unit tests covering the store, handlers, and key generator
└── db_test.go             # Exercices create/validate lifecycle against a temp SQLite file
```

### 2.2 Database Helpers (`server/db.go`)

`LicenseStore` wraps a SQLite database (powered by [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)) and enforces one license per fingerprint. Fingerprints are SHA-256 hashed before persistence so the raw hardware identifiers never leave the installer.

```go
type License struct {
        Key             string
        FingerprintHash string
        CustomerID      string
        CreatedAt       time.Time
}

func (s *LicenseStore) CreateLicense(ctx context.Context, customerID, fingerprintHash string) (*License, bool, error) {
        if existing, err := s.FindByFingerprint(ctx, fingerprintHash); err == nil {
                return existing, false, nil
        }
        key, err := GenerateLicenseKey(customerID, fingerprintHash)
        if err != nil {
                return nil, false, err
        }
        // ...insert row, handling concurrent upserts...
        return &License{Key: key, FingerprintHash: fingerprintHash, CustomerID: customerID, CreatedAt: now}, true, nil
}

func (s *LicenseStore) ValidateLicense(ctx context.Context, key, fingerprintHash string) (*License, error) {
        lic, err := s.FindByKey(ctx, key)
        if err != nil {
                return nil, err
        }
        if lic.FingerprintHash != fingerprintHash {
                return nil, ErrFingerprintMismatch
        }
        return lic, nil
}
```

The store helper automatically creates the `licenses` table (`key` primary key, `fingerprint_hash` unique, timestamps) on startup.

### 2.3 Key Generation (`server/license_gen.go`)

`GenerateLicenseKey` normalises the customer identifier, samples two random base32 segments, and mixes in the first 12 characters of the fingerprint hash to produce keys such as `ACMECO-1A2B-3C4D-XYZ12-AB789`.

```go
func HashFingerprint(raw string) string {
        sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
        return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func GenerateLicenseKey(customerID, fingerprintHash string) (string, error) {
        sanitized := sanitizeCustomerID(customerID)
        rand1, _ := randomSegment(5)
        rand2, _ := randomSegment(5)
        return fmt.Sprintf("%s-%s-%s-%s-%s-%s", sanitized, fingerprintHash[0:4], fingerprintHash[4:8], fingerprintHash[8:12], rand1, rand2), nil
}
```

### 2.4 HTTP Handlers (`server/license_handler.go`)

Both endpoints share an optional `X-Installer-Token` header check. When the token is configured (via environment variable or CLI flag) requests without the matching secret receive `403 Forbidden`.

```go
func (h *LicenseHandler) CreateLicense(w http.ResponseWriter, r *http.Request) {
        if !h.authorize(w, r) {
                return
        }
        var payload struct {
                CustomerID  string `json:"customerId"`
                Fingerprint string `json:"fingerprint"`
        }
        // decode + validate JSON, hash fingerprint, then call store.CreateLicense
        license, created, err := h.store.CreateLicense(r.Context(), sanitizeCustomerID(payload.CustomerID), HashFingerprint(payload.Fingerprint))
        // respond with 201 for new licenses or 200 when the fingerprint already owns a key
}

func (h *LicenseHandler) ValidateLicense(w http.ResponseWriter, r *http.Request) {
        if !h.authorize(w, r) {
                return
        }
        // decode JSON, hash fingerprint, call store.ValidateLicense
        // respond with {"status":"valid"} or 401 on mismatch/not-found
}
```

### 2.5 Server Entrypoint (`server/server.go`)

`server.go` wires everything together, exposes `/healthz`, logs requests, and honours three environment variables (also configurable via flags):

| Variable | Purpose | Default |
| --- | --- | --- |
| `LICENSE_BIND_ADDR` | HTTP listen address | `:8080` |
| `LICENSE_DB_PATH` | SQLite database path | `data/licenses.db` |
| `LICENSE_API_TOKEN` | Shared secret for installer/app requests | unset (token disabled) |

Run the API locally with:

```bash
go run ./server -token your-shared-secret
```

The bundled tests (`go test ./server/...`) cover create/validate flows, handler token enforcement, and key formatting.

### 2.6 API Usage

**Request license**

```http
POST /api/v1/licenses
X-Installer-Token: <shared secret>
Content-Type: application/json

{
  "customerId": "customer@example.com",
  "fingerprint": "A1B2C3D4E5F6..."
}
```

**Response**

```json
{
  "licenseKey": "CUSTOMER-1A2B3C-ABCDE-FGHIJ-KLMNO-PQRST"
}
```

**Validate license**

```http
POST /api/v1/licenses/validate
Content-Type: application/json

{
  "licenseKey": "CUSTOMER-1A2B3C-ABCDE-FGHIJ-KLMNO-PQRST",
  "fingerprint": "A1B2C3D4E5F6..."
}
```

**Successful response**

```json
{"status":"valid"}
```

### 2.7 Operational Checklist

- **Secrets**: store `LICENSE_API_TOKEN`, TLS cert paths, and DB credentials in environment variables or a secrets manager.
- **Rate limiting**: guard `/licenses` with IP-based throttling to block abuse (e.g., [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)).
- **Monitoring**: expose Prometheus metrics or log counts for license generation/validation.
- **Backups**: snapshot the SQLite database or, for larger scale, migrate to PostgreSQL/MySQL with automated backups.
- **Data privacy**: hash or encrypt stored fingerprints before persisting if you handle sensitive identifiers.

---

## 3. Runtime License Checks in the Desktop App

1. **Startup**: compute the same fingerprint used by the installer.
2. **Load**: read the cached license key from `%LOCALAPPDATA%\AmazonProductSuite\license.key` (or Credential Locker).
3. **Validate**: call the `/validate` endpoint. If successful, cache the timestamp; if not, show a blocking dialog.
4. **Offline grace**: allow a small grace period (e.g., 3 launches within 48 hours) when the last validation was successful.
5. **Transfer flow**: provide a support channel or self-service portal to revoke an old fingerprint and issue a replacement key when a user upgrades hardware.

> **Desktop configuration:** The Fyne UI integrates `internal/license` to read `%LOCALAPPDATA%\AmazonProductSuite\license.key`, compute the local fingerprint, and call the validation endpoint on every launch. Set `LICENSE_API_URL` (e.g., `https://licensing.yourdomain.com`) and `LICENSE_API_TOKEN` before starting the app so it can reach your server. Missing variables or validation failures surface a blocking dialog that explains how to recover.

---

## 4. Security & Hardening Checklist

- **HTTPS everywhere**: require TLS for all installer and runtime calls. Reject plain HTTP at the server.
- **API authentication**: the installer uses an `X-Installer-Token`; runtime validations can additionally use HMAC signatures or OAuth tokens.
- **Code integrity**: sign the installer with Authenticode, and consider binary obfuscation to slow down reverse engineering.
- **Tamper resistance**: store an HMAC alongside `license.key` so that any edits can be detected before contacting the server.
- **Auditing**: log every license issuance and validation with request metadata. Alert on spikes or repeated failures.
- **Privacy**: avoid storing raw hardware identifiers—hash them (e.g., SHA-256) before persistence.

---

## 5. End-to-End Workflow Recap

1. **Build** the desktop executable and fingerprint helper.
2. **Deploy** the Go license server with HTTPS, database, rate limiting, and monitoring in place.
3. **Compile** the Inno Setup installer containing the helper and custom script.
4. **Distribute** the installer. On each installation the server issues a unique key tied to the machine fingerprint and customer ID.
5. **Verify** the license on every launch, granting access only when the fingerprint and key remain valid.

This architecture enforces one-license-per-machine while delivering a polished onboarding flow comparable to Publisher Rocket.
