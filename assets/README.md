# Assets

Place optional artwork for the desktop build in this directory.

- `app.png` – primary source image for desktop packaging (used by Fyne and documentation).
- `app.ico` – Windows icon generated from `app.png`. The checked-in `.syso` resource in
  `cmd/app` embeds this file so that `go build` outputs an `.exe` with the correct
  Explorer icon. If you replace `app.png`, regenerate the `.ico` (any graphics editor or a
  short Pillow script will do) and rerun the resource command documented in
  `docs/BUILD_WINDOWS.md`.
