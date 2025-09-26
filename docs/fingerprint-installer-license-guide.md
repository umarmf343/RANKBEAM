# RankBeam Build + Licensing Walkthrough

This guide condenses everything you need to ship the RankBeam desktop app with a
machine-fingerprint license flow. It covers three areas:

1. Building the two Windows executables (`rankbeam.exe` and `fingerprint-helper.exe`).
2. Packaging them into an installer with Inno Setup.
3. Standing up the Go-based license server that the installer and app call into.

If you prefer deeper context, the original reference material lives in
[`docs/BUILD_WINDOWS.md`](BUILD_WINDOWS.md) and
[`docs/license-system-guide.md`](license-system-guide.md). This document stitches the
critical steps together so you can go from source code to a licensed installer quickly.

---

## 1. Build the Windows Executables

### 1.1 Prerequisites

| Component | Notes |
| --- | --- |
| Go toolchain | Install Go **1.23** or newer (the module pins Go 1.24.3 via the `toolchain` directive). |
| Cross compiler | On Linux/macOS install `mingw-w64` (`x86_64-w64-mingw32-gcc`) so Fyne and the Windows linker work. |
| Fyne CLI | `go install fyne.io/fyne/v2/cmd/fyne@latest` (optional unless you plan to bundle resources). |
| Git & make | Helpful for automation but not required. |

Confirm your Go environment is 64-bit:

```bash
go env GOARCH
```

The output should be `amd64`. If you see `386`, install the 64-bit Go toolchain before proceeding.

### 1.2 Fetch dependencies

From the repository root run:

```bash
go mod download
```

This primes the module cache so subsequent builds are reproducible.

### 1.3 Compile the binaries

```bash
# Build the desktop application (RankBeam UI)
GOOS=windows GOARCH=amd64 go build -o bin/rankbeam.exe ./cmd/app

# Build the helper used during installation to compute the machine fingerprint
GOOS=windows GOARCH=amd64 go build -o bin/fingerprint-helper.exe ./cmd/fingerprint-helper
```

Both executables land in `bin/`. Keep the relative paths because the Inno Setup script
expects them there. If you encounter linker errors about `-lgdi32` or `-lopengl32`, install
the 64-bit MinGW toolchain and retry.

> Optional: Package a resource-rich build with `fyne package -os windows -icon assets/app.png -name "RankBeam" -appID com.rankbeam.app -executable bin/rankbeam.exe -release`.

---

## 2. Package with Inno Setup

1. Install [Inno Setup 6+](https://jrsoftware.org/isinfo.php) on a Windows machine.
2. Copy `bin/rankbeam.exe` and `bin/fingerprint-helper.exe` to that machine (or build on Windows directly).
3. Open `installer/rankbeam.iss` in the Inno Setup IDE.
4. Review the `[Setup]` section to confirm `AppName`, `AppVersion`, and `OutputDir` values match your release.
5. Verify the `[Files]` entries point at the binaries you built:
   ```pascal
   Source: "..\bin\rankbeam.exe"; DestDir: "{app}"; Flags: ignoreversion
   Source: "..\bin\fingerprint-helper.exe"; DestDir: "{tmp}"; Flags: deleteafterinstall
   ```
6. Scroll to the `[Code]` section and update the API endpoint and installer token constants so the installer can talk to your license server.
7. Click **Build** → **Compile**. The IDE produces `dist\rankbeam.exe` (installer) or whatever filename you set via `OutputBaseFilename`.
8. Test the installer on a clean Windows VM. Confirm it:
   - prompts for customer information,
   - calls the license API, and
   - launches RankBeam with the acquired license key.

---

## 3. Deploy the License Server

The repository includes a Go service that issues and validates per-machine licenses. It
lives under the `server/` directory.

### 3.1 Configure environment

Set the following variables (or provide the equivalent CLI flags) before running the server:

| Variable | Purpose | Default |
| --- | --- | --- |
| `LICENSE_BIND_ADDR` | Listen address | `:8080` |
| `LICENSE_DB_PATH` | SQLite database path | `data/licenses.db` |
| `LICENSE_API_TOKEN` | Shared secret expected from installers/apps | *(disabled when empty)* |

### 3.2 Run migrations (first launch)

The server auto-creates the SQLite database. Ensure the `data/` directory exists and is
writable by the process:

```bash
mkdir -p data
```

### 3.3 Launch locally

```bash
go run ./server -token your-shared-secret
```

Visit `http://localhost:8080/healthz` to verify the service responds with `ok`.

### 3.4 API endpoints

- `POST /api/v1/licenses` creates or re-issues a license bound to a fingerprint.
- `POST /api/v1/licenses/validate` verifies an existing license + fingerprint pair.

Both endpoints expect JSON payloads and (optionally) the `X-Installer-Token` header that
matches `LICENSE_API_TOKEN`.

### 3.5 Production considerations

- Front the server with HTTPS (e.g., Caddy, Nginx, or a managed load balancer).
- Enable backups for `data/licenses.db` or switch to PostgreSQL/MySQL for multi-instance setups.
- Add rate limiting around license creation to prevent abuse.
- Monitor issuance/validation metrics and set up alerts for spikes in failures.

For a deeper dive (including PascalScript snippets and application runtime checks), follow
the comprehensive walkthrough in [`docs/license-system-guide.md`](license-system-guide.md).

---

## 4. End-to-End Checklist

1. Build `bin/rankbeam.exe` and `bin/fingerprint-helper.exe`.
2. Configure and deploy the license server with your chosen token and HTTPS endpoint.
3. Update `installer/rankbeam.iss` with the server URL/token, then compile the installer in Inno Setup.
4. Run the installer on Windows, confirm it fingerprints the machine, requests a license, and persists the key.
5. Launch RankBeam; the bundled runtime code validates the license on start-up.

Following the sequence above yields a production-ready installer that mimics the
Publisher Rocket experience with machine-tied licensing.
