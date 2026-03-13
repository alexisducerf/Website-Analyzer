# Website Analyzer - SEO Crawler

https://img.shields.io/github/v/release/alexisducerf/Website-Analyzer?label=Latest%20Release&style=for-the-badge

[![Latest Release](https://img.shields.io/github/v/release/alexisducerf/Website-Analyzer?label=Latest%20Release&style=for-the-badge&cacheSeconds=0)](https://github.com/alexisducerf/web-seo/releases/latest&cacheSeconds=0) [![License](https://img.shields.io/github/license/alexisducerf/Website-Analyzer?style=for-the-badge&cacheSeconds=0)](LICENSE)

A high-performance, desktop-based SEO crawler and auditing tool built with **Go (Wails)** and **Svelte**. It provides real-time crawling, detailed on-page analysis, and comprehensive reporting.

## 🚀 Features

### Core Crawling
-   **High Performance**: Concurrent crawling with adjustable threads (1-20).
-   **Rate Limiting**: Configurable request delay (0-2000ms) to generate polite traffic.
-   **Real-time Feedback**: Visual progress bar and live-updating results table.
-   **Search & Filtering**: Instantly filter results by status code (200, 404, etc.) or search by URL/Title.

### Deep SEO Analysis
Click on any crawled URL to see a detailed report:
-   **Overview**: Meta Title, Description, Word Count, Canonical, Robots directives.
-   **Issues Tab**: Auto-detected SEO problems (e.g., Missing H1, Images without Alt text, Title length issues).
-   **Headers**: Visual hierarchy of specific H1-H6 tags.
-   **Images**: Gallery view of all page images with Alt text validation.
-   **Links**: Scrollable lists of all Internal and External links found on the page.

### Export & Reporting
-   **CSV Export**: One-click export of the entire crawl data.
-   **Detailed Data**: Includes all standard SEO tags plus full lists of headers and links (pipe-delimited).
-   **Smart Filenaming**: Auto-generates filenames like `domain.com_2023-10-27.csv`.

---

## 🛠️ Getting Started

### Prerequisites
-   **Go**: v1.21 or higher
-   **Node.js**: v18 or higher
-   **Wails**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Running Locally
1.  Clone the repository:
    ```bash
    git clone https://github.com/alexisducerf/web-seo.git
    cd web-seo
    ```
2.  Start the development server:
    ```bash
    wails dev
    ```
3.  The app will launch in a native window.

### Building for Production
To create a standalone executable for your OS:
```bash
wails build
```
Binaries will be placed in the `build/bin` directory.

---

## 📦 CI/CD & Releases

This project uses **GitHub Actions** for automated cross-platform builds.
-   **Trigger**: Pushing a tag (e.g., `v1.0.0`) triggers the workflow.
-   **Platforms**: Builds specifically for **Windows** (`.exe`), **macOS** (`.app`), and **Linux**.
-   **Releases**: Automatically creates a GitHub Release and uploads zipped assets.

### macOS Note
If you download the `.app` directly, you might see a "damaged" or "unverified" warning. This is standard macOS Gatekeeper behavior for unsigned apps.
**Fix**: Run `xattr -cr /path/to/web-seo.app` in your terminal to allow it to run.
