# RankBeam Web Companion (Tailwind Prototype)

This folder contains a responsive web prototype for the RankBeam research workflows. It mirrors the desktop application's tabbed experience so teams that prefer the browser can explore the same product lookup, keyword research, and international expansion tools.

## Overview

- **Technology stack:** Tailwind CSS (via CDN) with lightweight vanilla JavaScript for tab switching and dialog toggles.
- **Feature parity:** Product lookup summaries, keyword/bestseller call-to-action panels, sortable localisation tables, and adjustable request throttling controls that match the desktop Settings dialog.
- **Usage model:** Static HTML that can be hosted from any CDN, S3 bucket, or documentation portal while backend integration is being planned.

## Quick start

1. Open the prototype locally:
   ```bash
   cd docs/web-companion
   python3 -m http.server 8080
   ```
2. Navigate to `http://localhost:8080/index.html` in your browser.
3. Switch tabs to preview the different RankBeam workflows and experiment with the Settings dialog to see how timeout/concurrency controls could be surfaced for browser users.

Because Tailwind is loaded via CDN there is no build step—refreshing the browser reflects any edits immediately.

## File structure

| File | Description |
| --- | --- |
| [`index.html`](index.html) | Main HTML prototype with Tailwind styling, tabbed navigation, and sample research panels. |

## Extending the prototype

- **Wire up data:** Replace the placeholder sections with AJAX calls to your preferred backend (e.g., the existing Go/Fyne service or a new API gateway) to hydrate tables and cards.
- **Add authentication:** Wrap the layout inside your SSO provider or add a lightweight login form before exposing it to analysts.
- **Deploy anywhere:** Because the bundle is static, it can be dropped into documentation portals, GitHub Pages, or internal knowledge bases to demo browser access.

## Parity checklist

| Desktop workflow | Web prototype status |
| --- | --- |
| Product lookup cards | ✅ Implemented as responsive summary tiles with ASIN, availability, and bestseller highlights. |
| Keyword research CTAs | ✅ Buttons and explainer cards map to the desktop actions. |
| International metrics | ✅ Rendered as sortable Tailwind tables (Country/Keyword/Volume) matching the new desktop table view. |
| Request throttling | ✅ Settings dialog exposes timeout and concurrency sliders mirroring the Go app configuration. |

Feel free to duplicate this prototype when collaborating with designers or sharing RankBeam with stakeholders who prefer the browser over the desktop build.
