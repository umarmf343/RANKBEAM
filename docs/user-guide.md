# RankBeam User Guide

Welcome to **RankBeam**, the desktop companion for Amazon KDP authors and product researchers. This guide walks you through installing the application, activating your license, and mastering each workflow so you can surface winning ideas faster.

---

## 1. System Requirements

- **Operating system:** Windows 10 or 11 (64-bit). RankBeam is built with the Fyne toolkit and ships 64-bit binaries only.
- **Hardware:** 4 GB RAM (8 GB recommended) and at least 500 MB of free disk space for cache and exports.
- **Network:** Stable broadband connection. All research workflows query Amazon endpoints in real time.
- **Accounts:** A valid RankBeam license key issued by your administrator or vendor.

> **Tip:** If you self-build the app, ensure the Go toolchain is configured for `GOARCH=amd64` before compiling.

## 2. Installation

You can install RankBeam either from a provided Windows installer or by building from source.

### 2.1 Using the Windows installer

1. Download the latest `RankBeam-Setup.exe` packaged with Inno Setup.
2. Right-click the installer and choose **Run as administrator**.
3. Follow the wizard, accepting the license agreement and default install location (usually `C:\Program Files\RankBeam`).
4. When prompted, paste your license key. The installer stores it securely for first launch.
5. Finish the wizard. A desktop shortcut and Start Menu entry will be created.

### 2.2 Building from source

1. Install Go 1.21+ and ensure the environment includes required Fyne build dependencies (e.g., MSYS2 or Visual Studio Build Tools).
2. Clone the repository and install JavaScript dependencies if you plan to run the backend service: `npm install`.
3. Build the Windows executable:
   ```bash
   go env -w GOOS=windows GOARCH=amd64
   go build -o rankbeam.exe ./cmd/app
   ```
4. Optionally bundle an installer following [`BUILD_WINDOWS.md`](./BUILD_WINDOWS.md).
5. Copy `rankbeam.exe` to your preferred location and create a shortcut.

## 3. Launching RankBeam

- Double-click the desktop shortcut or run `rankbeam.exe` from a terminal.
- On first start, RankBeam validates the locally stored license key. If the key is missing or invalid, the activation screen appears.
- Subsequent launches reuse the cached key, allowing the intelligence suite to open immediately.

> **Trouble activating?** Confirm the machine has internet access and that the license server URL and token were configured during installation. Contact support if you receive an "invalid or expired" message.

## 4. Activating Your License

When the activation view opens:

1. Review the status banner for any environment misconfiguration (e.g., missing license server URL).
2. Click **Get Your License** if you still need a key—this opens the RankBeam storefront in your browser.
3. Paste your license key into the text box.
4. Press **Activate & Launch** to validate the key against the licensing service.
5. Once activated, RankBeam automatically loads the main application and updates the window title with a redacted license suffix.

If activation fails, the status label displays the exact error (empty key, unauthorized token, expired key). Resolve the issue and retry.

## 5. Navigating the Interface

RankBeam opens with a top navigation bar and four main tabs:

1. **Product Lookup** – Fetch detailed information for a single ASIN in any Amazon marketplace.
2. **Keyword Research** – Generate keyword suggestions, category insights, and bestseller snapshots for a seed term.
3. **Competitive Analysis** – Combine reverse ASIN lookups with an Amazon Ads campaign builder.
4. **International** – Expand a seed keyword into localized suggestions across multiple marketplaces.

Use the **Tutorial** button in the top-right corner to open community-made walkthroughs on YouTube.

## 6. Workflow Deep Dive

### 6.1 Product Lookup

1. Enter the target ASIN in the **ASIN** field.
2. Pick the destination **Marketplace** (country) from the drop-down list.
3. Click **Fetch Product**.
4. RankBeam retrieves pricing, availability, review metrics, bestseller ranks, and metadata. Results are shown in the scrollable panel below the form.

**Tips:**
- Use the country selector to compare listings between regions.
- Copy the displayed summary directly into your research notes or spreadsheets.

### 6.2 Keyword Research Toolkit

This tab combines three complementary actions:

- **Fetch Keyword Suggestions** – Returns search suggestions with estimated volume, competition, and title density.
- **Category Insights** – Highlights categories that align with your seed term.
- **Bestseller Snapshot** – Analyzes the top-ranking ASINs, including BSR, pricing, and indie-only filtering.

**Steps:**
1. Enter a seed keyword (e.g., "children's book about space").
2. Select the target marketplace.
3. Optional: refine filters
   - **Min Search Volume** – Ignore low-volume ideas.
   - **Max Competition** – Cap the acceptable competition score.
   - **Max Title Density** – Screen out saturated titles.
   - **Max BSR** and **Indie authors only** – Focus bestseller analysis on manageable competition.
4. Run each action via its dedicated button. Outputs populate the corresponding labeled sections below.

### 6.3 Competitive Analysis

This workspace blends reverse ASIN reconnaissance with campaign planning:

- **Reverse ASIN Intelligence**
  1. Enter a competitor ASIN and marketplace.
  2. Tune the keyword filters (volume, competition, title density).
  3. Click **Run Reverse ASIN** to reveal the highest-leverage keywords that drive the listing.

- **Amazon Ads Planner**
  1. Provide your product title and a concise description.
  2. Paste competitor keywords or phrases separated by commas.
  3. Click **Generate Campaign Keywords** to receive a curated list formatted for AMS targeting.

Each result panel scrolls independently so you can review long reports without losing form context.

### 6.4 International Keyword Expansion

1. Enter a base keyword.
2. Tick the marketplaces you want to evaluate (a recommended default selection is pre-filled).
3. Click **Generate Suggestions** to produce localized keyword lists across the selected countries, enabling multi-market launch planning.

## 7. Exporting & Sharing Results

- Use the mouse to highlight any report and press **Ctrl+C** to copy it to the clipboard.
- Paste the formatted text into Excel, Google Sheets, or your favorite note-taking app.
- For automation, pair RankBeam with the command-line tools or the backend service included in this repository to export JSON/CSV data.

## 8. Best Practices

- **Throttle requests**: Space out back-to-back queries to avoid Amazon rate limits.
- **Validate ideas**: Cross-reference RankBeam keyword scores with live Amazon searches before publishing.
- **Stay compliant**: Watch for flagged terms in the Amazon Ads planner to prevent policy violations.
- **Keep updated**: Periodically check for RankBeam updates that include new heuristics and bug fixes.

## 9. Troubleshooting

| Symptom | Suggested Fix |
| --- | --- |
| "Set LICENSE_API_URL" message on launch | Re-run the installer or set the environment variable so the app can reach your license server. |
| "License key not found" | Paste your key on the activation screen or reinstall to restore the stored key. |
| Empty results after a fetch | Verify your internet connection and confirm the marketplace selection is valid. |
| Frequent timeouts | Increase your network timeout by relaunching after ensuring no VPN/firewall blocks Amazon. |

If issues persist, gather screenshots, note the action performed, and forward the details to your support contact alongside the sanitized license ID shown in the window title.

## 10. Getting Help

- **Video tutorials:** Click the in-app **Tutorial** button.
- **Documentation:** Explore the other guides in the `docs/` folder for licensing, installer creation, and system architecture.
- **Support:** Reach out to your distributor or administrator with your license summary and error messages.

---

Empower your publishing strategy with RankBeam—combining intuitive workflows and deep data to help you launch, optimize, and scale on Amazon.
