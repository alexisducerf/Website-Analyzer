<script>
  import { onMount } from "svelte";
  import { CrawlWebsite, StopCrawl } from "../wailsjs/go/main/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";

  let url = "https://wails.io";
  let concurrency = 5;
  let isCrawling = false;
  let pages = [];
  let expandedRows = new Set();
  let filter = "all";

  let stats = {
    discovered: 0,
    success: 0,
    error: 0,
    totalWords: 0,
  };

  onMount(() => {
    EventsOn("page-discovered", (page) => {
      pages = [page, ...pages];
      stats.discovered++;
      if (page.status === 200) {
        stats.success++;
        stats.totalWords += page.wordCount || 0;
      } else {
        stats.error++;
      }
    });

    EventsOn("crawl-complete", () => {
      isCrawling = false;
    });

    EventsOn("crawl-error", (err) => {
      isCrawling = false;
      alert("Error: " + err);
    });
  });

  function startCrawl() {
    if (!url) return;
    pages = [];
    expandedRows = new Set();
    stats = { discovered: 0, success: 0, error: 0, totalWords: 0 };
    isCrawling = true;
    CrawlWebsite(url, parseInt(concurrency));
  }

  function stopCrawl() {
    StopCrawl();
    isCrawling = false;
  }

  function toggleRow(pageUrl) {
    if (expandedRows.has(pageUrl)) {
      expandedRows.delete(pageUrl);
    } else {
      expandedRows.add(pageUrl);
    }
    expandedRows = expandedRows;
  }

  function exportToCSV() {
    if (pages.length === 0) return;
    const headers = [
      "URL",
      "Status",
      "Title",
      "Description",
      "H1",
      "Word Count",
    ];
    const rows = pages.map((p) => [
      p.url,
      p.status,
      `"${(p.title || "").replace(/"/g, '""')}"`,
      `"${(p.description || "").replace(/"/g, '""')}"`,
      `"${(p.h1 || "").replace(/"/g, '""')}"`,
      p.wordCount,
    ]);
    const csvContent = [headers, ...rows].map((e) => e.join(",")).join("\n");
    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const link = document.createElement("a");
    const downloadUrl = URL.createObjectURL(blob);
    link.setAttribute("href", downloadUrl);
    link.setAttribute(
      "download",
      `seo_audit_${new Date().toISOString().split("T")[0]}.csv`,
    );
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  $: filteredPages = pages.filter((p) => {
    if (filter === "all") return true;
    if (filter === "success") return p.status >= 200 && p.status < 300;
    if (filter === "error") return p.status >= 400 || p.status === 0;
    return true;
  });
</script>

<div class="app-container">
  <aside class="sidebar">
    <h1>SEO Spider</h1>

    <div class="sidebar-section">
      <h2>Configurations</h2>
      <div class="field-v2" style="margin-bottom: 1.5rem">
        <label class="field-v2-label" for="target-url">Target URL</label>
        <input
          id="target-url"
          class="input-field"
          type="text"
          bind:value={url}
          placeholder="https://example.com"
          disabled={isCrawling}
        />
      </div>

      <div class="field-v2" style="margin-bottom: 2rem">
        <label class="field-v2-label" for="concurrency"
          >Concurrency ({concurrency})</label
        >
        <input
          id="concurrency"
          type="range"
          bind:value={concurrency}
          min="1"
          max="20"
          disabled={isCrawling}
          style="width: 100%; accent-color: var(--primary);"
        />
      </div>

      {#if !isCrawling}
        <button class="btn btn-primary" on:click={startCrawl} disabled={!url}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"><polygon points="5 3 19 12 5 21 5 3" /></svg
          >
          Start Audit
        </button>
      {:else}
        <button class="btn btn-stop" on:click={stopCrawl}>
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            ><rect x="4" y="4" width="16" height="16" rx="2" ry="2" /></svg
          >
          Stop Crawl
        </button>
        <div class="progress-container">
          <div class="progress-bar" style="width: {concurrency * 5}%"></div>
        </div>
      {/if}
    </div>

    <div class="sidebar-section" style="margin-top: auto">
      <h2>Quick Actions</h2>
      <button
        class="btn btn-outline"
        on:click={exportToCSV}
        disabled={pages.length === 0}
      >
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          ><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline
            points="7 10 12 15 17 10"
          /><line x1="12" y1="15" x2="12" y2="3" /></svg
        >
        Export Audit
      </button>
    </div>
  </aside>

  <main class="main-content">
    <div class="stats-container">
      <div class="card stat-v2">
        <div class="stat-v2-label">Discovered</div>
        <div class="stat-v2-value">{stats.discovered}</div>
      </div>
      <div class="card stat-v2">
        <div class="stat-v2-label" style="color: var(--success)">
          Successful
        </div>
        <div class="stat-v2-value">{stats.success}</div>
      </div>
      <div class="card stat-v2">
        <div class="stat-v2-label" style="color: var(--error)">
          Bad Requests
        </div>
        <div class="stat-v2-value">{stats.error}</div>
      </div>
      <div class="card stat-v2">
        <div class="stat-v2-label" style="color: var(--secondary)">
          Avg Words
        </div>
        <div class="stat-v2-value">
          {stats.success > 0 ? Math.round(stats.totalWords / stats.success) : 0}
        </div>
      </div>
    </div>

    <div class="results-wrapper">
      <div class="tabs">
        <button
          class="tab {filter === 'all' ? 'active' : ''}"
          on:click={() => (filter = "all")}>All Resources</button
        >
        <button
          class="tab {filter === 'success' ? 'active' : ''}"
          on:click={() => (filter = "success")}>Success (2xx)</button
        >
        <button
          class="tab {filter === 'error' ? 'active' : ''}"
          on:click={() => (filter = "error")}>Errors & Issues</button
        >
      </div>

      <div class="scroll-v2">
        <table class="table-v2">
          <thead>
            <tr>
              <th style="width: 120px">Status</th>
              <th>Resource URL</th>
              <th>Meta Title</th>
              <th style="width: 48px"></th>
            </tr>
          </thead>
          <tbody>
            {#each filteredPages as page}
              <tr on:click={() => toggleRow(page.url)} style="cursor: pointer">
                <td>
                  <span
                    class="badge-v2 {page.status === 200
                      ? 'badge-200'
                      : 'badge-error'}"
                  >
                    {#if isCrawling && pages[0] === page}
                      <div class="crawl-loader"></div>
                    {/if}
                    {page.status || "ERR"}
                  </span>
                </td>
                <td>
                  <div
                    style="max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 500; color: var(--secondary)"
                  >
                    {page.url}
                  </div>
                </td>
                <td style="color: var(--text-muted)">{page.title || "-"}</td>
                <td>
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2.5"
                    style="transform: {expandedRows.has(page.url)
                      ? 'rotate(180deg)'
                      : 'none'}; transition: transform 0.2s"
                  >
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </td>
              </tr>
              {#if expandedRows.has(page.url)}
                <tr class="details-row">
                  <td colspan="4">
                    <div class="detail-pane">
                      <div class="field-v2">
                        <span class="field-v2-label">Page Title</span>
                        <span class="field-v2-value"
                          >{page.title || "No title found"}</span
                        >
                      </div>
                      <div class="field-v2">
                        <span class="field-v2-label">H1 Header</span>
                        <span class="field-v2-value"
                          >{page.h1 || "No H1 found"}</span
                        >
                      </div>
                      <div class="field-v2" style="grid-column: span 2">
                        <span class="field-v2-label">Meta Description</span>
                        <span class="field-v2-value"
                          >{page.description || "No description found"}</span
                        >
                      </div>
                      <div class="field-v2">
                        <span class="field-v2-label">Word Count</span>
                        <span class="field-v2-value"
                          >{page.wordCount} words</span
                        >
                      </div>
                      {#if page.error}
                        <div class="field-v2" style="grid-column: span 2">
                          <span
                            class="field-v2-label"
                            style="color: var(--error)">Error Report</span
                          >
                          <span
                            class="field-v2-value"
                            style="color: var(--error)">{page.error}</span
                          >
                        </div>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/if}
            {/each}
            {#if pages.length === 0 && !isCrawling}
              <tr>
                <td colspan="4">
                  <div
                    style="padding: 10rem 0; text-align: center; color: var(--text-dim)"
                  >
                    <svg
                      width="48"
                      height="48"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1"
                      style="margin-bottom: 1.5rem; opacity: 0.3"
                      ><circle cx="12" cy="12" r="10" /><line
                        x1="12"
                        y1="8"
                        x2="12"
                        y2="12"
                      /><line x1="12" y1="16" x2="12.01" y2="16" /></svg
                    >
                    <p style="font-size: 1.125rem; font-weight: 500">
                      Ready to audit your first resource.
                    </p>
                    <p style="font-size: 0.875rem">
                      Enter a URL in the sidebar and click Start Audit.
                    </p>
                  </div>
                </td>
              </tr>
            {/if}
          </tbody>
        </table>
      </div>
    </div>
  </main>
</div>

<style>
  .sidebar-section {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    margin-bottom: 3rem;
  }
</style>
