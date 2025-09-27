# Building RankBeam for Windows

This repository now contains a Go/Fyne desktop application that wraps the scraping logic exposed by the original amazon-product-api project. Follow the steps below to compile a distributable Windows executable and package it with an installer.

## 1. Prerequisites

- Go **1.23** or newer installed locally. The module enables the Go **1.24.3** toolchain via `go.mod`, so using an older Go release will trigger an automatic download of that toolchain as long as your base installation is 64-bit and recent enough to understand the [`toolchain`](https://go.dev/doc/toolchain) directive. Confirm `go env GOARCH` reports `amd64`; the GUI cannot be built with 32-bit toolchains.
- The [Fyne prerequisites](https://docs.fyne.io/started/) for your target platform. For Windows cross-compilation from Linux you also need a MinGW toolchain (`x86_64-w64-mingw32-gcc`).
- The Fyne CLI, installed with `go install fyne.io/fyne/v2/cmd/fyne@latest`. On Windows ensure `%USERPROFILE%\go\bin` is on your `PATH` (Command Prompt: `setx PATH "%PATH%;%USERPROFILE%\go\bin"`; PowerShell: `[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\go\bin", "User")`).
- Git and make (optional but convenient).
- Inno Setup 6 (for building the installer) if you are on Windows.

## 2. Fetch the dependencies

```bash
# inside the project root
go mod download
```

## 3. Run the application locally (Linux/macOS)

```bash
go run ./cmd/app
```

## 4. Build the Windows executables

```bash
# Application binary used by both the FYNE package and the installer
GOOS=windows GOARCH=amd64 go build -o bin/rankbeam.exe ./cmd/app

# Helper that calculates the machine fingerprint during setup
GOOS=windows GOARCH=amd64 go build -o bin/fingerprint-helper.exe ./cmd/fingerprint-helper

# Optional: produce a Fyne-packaged executable with embedded resources
fyne package -os windows -icon assets/app.png \
  -name "RankBeam" \
  -appID com.rankbeam.app \
  -executable bin/rankbeam.exe \
  -release
```

If you are running these commands directly inside a Windows Command Prompt, remember that environment variables use `set`/`setx` rather than the PowerShell `$env:` syntax. For example, enable CGO with `set CGO_ENABLED=1` before invoking `fyne package`. In PowerShell the equivalent command is `$env:CGO_ENABLED = "1"`.

The two `go build` commands place the artifacts where the installer expects them (`bin/`). The optional `fyne package` command generates a redistributable `.exe` with icons and metadata in the `dist/` directory; pass `-release` to strip debug information and reduce the binary size.

> **Note:** Fyne requires a C compiler. When cross-compiling from Linux you may have to install `mingw-w64` and set `CC=x86_64-w64-mingw32-gcc` before building. If you are developing on Windows, make sure the MinGW/MSYS2 environment you point to is the 64-bit `x86_64` variant. A 32-bit environment will fail during linking with messages such as `cannot find -lgdi32` and `cannot find -lopengl32`.

## 5. Package with Inno Setup

1. Ensure `bin/rankbeam.exe` and `bin/fingerprint-helper.exe` exist from the previous step.
2. Open `installer/rankbeam.iss` with Inno Setup.
3. Adjust the optional icon path if you have a custom icon.
4. Build the installer to produce `rankbeam-setup.exe`.

## 6. Verify the build

- Execute the `.exe` file on a Windows machine.
- Ensure the Product Lookup, Keyword Research, Competitive Analysis, and International tabs perform network requests successfully.
- If requests are throttled by Amazon, increase the timeout or provide cookies inside the source code before rebuilding.

## 7. Ship it

Bundle the generated installer and the product description from `docs/product-description.md` when publishing the tool.
