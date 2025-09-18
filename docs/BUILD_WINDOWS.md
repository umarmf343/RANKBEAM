# Building the Amazon Product Intelligence Suite for Windows

This repository now contains a Go/Fyne desktop application that wraps the scraping logic exposed by the original amazon-product-api project. Follow the steps below to compile a distributable Windows executable and package it with an installer.

## 1. Prerequisites

- Go **1.21** or newer installed locally.
- The [Fyne prerequisites](https://docs.fyne.io/started/) for your target platform. For Windows cross-compilation from Linux you also need a MinGW toolchain (`x86_64-w64-mingw32-gcc`).
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

## 4. Cross-compile for Windows 64-bit

```bash
GOOS=windows GOARCH=amd64 fyne package -os windows -icon assets/app.png -name "Amazon Product Intelligence" -appID com.amazon.intelligence
# Alternatively, use the standard go build command:
GOOS=windows GOARCH=amd64 go build -o amazon-product-scraper.exe ./cmd/app
```

The `fyne package` command embeds the required resources and produces an `.exe` file plus metadata. If you only need the executable, the `go build` command is sufficient.

> **Note:** Fyne requires a C compiler. When cross-compiling from Linux you may have to install `mingw-w64` and set `CC=x86_64-w64-mingw32-gcc` before building.

## 5. Package with Inno Setup

1. Copy the generated `amazon-product-scraper.exe` into the project root (next to this repository's `README.md`).
2. Open `installer/amazon-product-scraper.iss` with Inno Setup.
3. Adjust the optional icon path if you have a custom icon.
4. Build the installer to produce `amazon-product-intelligence-setup.exe`.

## 6. Verify the build

- Execute the `.exe` file on a Windows machine.
- Ensure the Product Lookup, Keyword Research, Competitive Analysis, and International tabs perform network requests successfully.
- If requests are throttled by Amazon, increase the timeout or provide cookies inside the source code before rebuilding.

## 7. Ship it

Bundle the generated installer and the product description from `docs/product-description.md` when publishing the tool.
