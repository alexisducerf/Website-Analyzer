<script>
  import { onMount } from "svelte";
  import {
    CrawlWebsite,
    StopCrawl,
    ExportCSV,
  } from "../wailsjs/go/main/App.js";
  import { EventsOn } from "../wailsjs/runtime/runtime.js";

  let url = "";
  let concurrency = 3;
  let delay = 1000;
  let isCrawling = false;
  let pages = [];
  let expandedRows = new Set();
  let filter = "all";
  let searchQuery = "";
  let activeDetailTab = "overview";

  function analyzeSEO(page) {
    const issues = [];
    if (!page.title) issues.push({ level: "error", message: "Missing Title" });
    else if (page.title.length < 10)
      issues.push({ level: "warning", message: "Title too short (<10 chars)" });
    else if (page.title.length > 60)
      issues.push({ level: "warning", message: "Title too long (>60 chars)" });

    if (!page.description)
      issues.push({ level: "error", message: "Missing Meta Description" });
    else if (page.description.length < 50)
      issues.push({
        level: "warning",
        message: "Description too short (<50 chars)",
      });
    else if (page.description.length > 160)
      issues.push({
        level: "warning",
        message: "Description too long (>160 chars)",
      });

    if (!page.h1) issues.push({ level: "error", message: "Missing H1 Header" });

    if (page.images) {
      const missingAlt = page.images.filter((img) => !img.alt).length;
      if (missingAlt > 0)
        issues.push({
          level: "warning",
          message: `${missingAlt} images missing Alt text`,
        });
    }
    return issues;
  }

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
    CrawlWebsite(url, parseInt(concurrency), parseInt(delay));
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
      "Canonical",
      "Robots",
      "Internal Links",
      "External Links",
      "Images",
      "All H1s",
      "All H2s",
      "All H3s",
      "Internal Links List",
      "External Links List",
      "SEO Errors",
      "SEO Warnings",
    ];
    const rows = pages.map((p) => {
      const issues = analyzeSEO(p);
      const errors = issues
        .filter((i) => i.level === "error")
        .map((i) => i.message)
        .join(" | ");
      const warnings = issues
        .filter((i) => i.level === "warning")
        .map((i) => i.message)
        .join(" | ");

      return [
        p.url,
        p.status,
        `"${(p.title || "").replace(/"/g, '""')}"`,
        `"${(p.description || "").replace(/"/g, '""')}"`,
        `"${(p.h1 || "").replace(/"/g, '""')}"`,
        p.wordCount,
        `"${(p.canonical || "").replace(/"/g, '""')}"`,
        `"${(p.robots || "").replace(/"/g, '""')}"`,
        p.linksInternal ? p.linksInternal.length : 0,
        p.linksExternal ? p.linksExternal.length : 0,
        p.images ? p.images.length : 0,
        `"${(p.headers && p.headers["h1"] ? p.headers["h1"].join(" | ") : "").replace(/"/g, '""')}"`,
        `"${(p.headers && p.headers["h2"] ? p.headers["h2"].join(" | ") : "").replace(/"/g, '""')}"`,
        `"${(p.headers && p.headers["h3"] ? p.headers["h3"].join(" | ") : "").replace(/"/g, '""')}"`,
        `"${(p.linksInternal ? p.linksInternal.join(" | ") : "").replace(/"/g, '""')}"`,
        `"${(p.linksExternal ? p.linksExternal.join(" | ") : "").replace(/"/g, '""')}"`,
        `"${errors.replace(/"/g, '""')}"`,
        `"${warnings.replace(/"/g, '""')}"`,
      ];
    });
    const csvContent = [headers, ...rows].map((e) => e.join(",")).join("\n");

    // Generate filename: domain_datetime.csv
    let domain = "scan";
    try {
      const u = new URL(url);
      domain = u.hostname.replace(/www\./, "");
    } catch (e) {}

    const now = new Date();
    const timestamp = now.toISOString().replace(/[:.]/g, "-").slice(0, 19);
    const filename = `${domain}_${timestamp}.csv`;

    ExportCSV(filename, csvContent).then((err) => {
      if (err) {
        alert("Failed to export: " + err);
      }
    });
  }

  let isSidebarOpen = true;

  function toggleSidebar() {
    isSidebarOpen = !isSidebarOpen;
  }

  // Helper to filter issues for the "SEO Errors" tab
  function getSeoErrors(page) {
    return analyzeSEO(page).filter((issue) => issue.level === "error");
  }

  $: filteredPages = pages.filter((p) => {
    // Filter by Search Query
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchUrl = p.url.toLowerCase().includes(q);
      const matchTitle = (p.title || "").toLowerCase().includes(q);
      if (!matchUrl && !matchTitle) return false;
    }

    // Filter by Status Tab
    if (filter === "all") return true;
    if (filter === "success") return p.status >= 200 && p.status < 300;
    if (filter === "error") return p.status >= 400 || p.status === 0;
    if (filter === "seo-issues") return analyzeSEO(p).length > 0;
    return true;
  });
</script>

<div class="app-container {isSidebarOpen ? '' : 'sidebar-closed'}">
  <aside class="sidebar">
    <div
      class="sidebar-header"
      style={!isSidebarOpen ? "justify-content: center;" : ""}
    >
      {#if isSidebarOpen}
        <h1 style="margin: 0; font-size: 1.5rem;">Website Analyzer</h1>
      {/if}
      <button
        class="btn-icon"
        on:click={toggleSidebar}
        title={isSidebarOpen ? "Collapse Sidebar" : "Expand Sidebar"}
      >
        {#if isSidebarOpen}
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            ><line x1="18" y1="6" x2="6" y2="18"></line><line
              x1="6"
              y1="6"
              x2="18"
              y2="18"
            ></line></svg
          >
        {:else}
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            ><line x1="3" y1="12" x2="21" y2="12"></line><line
              x1="3"
              y1="6"
              x2="21"
              y2="6"
            ></line><line x1="3" y1="18" x2="21" y2="18"></line></svg
          >
        {/if}
      </button>
    </div>

    {#if isSidebarOpen}
      <div class="sidebar-section">
        <h2>Configurations</h2>
        <div class="field-v2" style="margin-bottom: 1.5rem">
          <label class="field-v2-label" for="target-url">Target URL</label>
          <input
            id="target-url"
            class="input-field"
            type="text"
            bind:value={url}
            placeholder="https://website.tld"
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

        <div class="field-v2" style="margin-bottom: 2rem">
          <label class="field-v2-label" for="delay"
            >Request Delay ({delay}ms)</label
          >
          <input
            id="delay"
            type="range"
            bind:value={delay}
            min="0"
            max="2000"
            step="100"
            disabled={isCrawling}
            style="width: 100%; accent-color: var(--secondary);"
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
    {/if}

    <div
      class="sidebar-section"
      style={!isSidebarOpen
        ? "margin-top: auto; align-items: center;"
        : "margin-top: auto;"}
    >
      {#if isSidebarOpen}<h2>Quick Actions</h2>{/if}
      <button
        class="btn btn-outline {isSidebarOpen ? '' : 'btn-icon-only'}"
        on:click={exportToCSV}
        disabled={pages.length === 0}
        title="Export Audit"
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
        {#if isSidebarOpen}Export Audit{/if}
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
        <button
          class="tab {filter === 'seo-issues' ? 'active' : ''}"
          on:click={() => (filter = "seo-issues")}>SEO Issues</button
        >
      </div>

      <div class="search-bar-container" style="margin-bottom: 1rem;">
        <input
          class="input-field"
          type="text"
          placeholder="Search items..."
          bind:value={searchQuery}
          style="width: 100%;"
        />
      </div>

      <div class="scroll-v2">
        <table class="table-v2">
          <thead>
            <tr>
              <th style="width: 120px">Status</th>
              <th>Resource URL</th>
              <th>{filter === "seo-issues" ? "Issues" : "Meta Title"}</th>
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
                <td style="color: var(--text-muted)">
                  {#if filter === "seo-issues"}
                    <div style="display: flex; gap: 0.5rem; flex-wrap: wrap;">
                      {#each analyzeSEO(page) as issue}
                        <span
                          class="badge-v2 {issue.level === 'error'
                            ? 'badge-error'
                            : 'badge-warning'}"
                          style="font-size: 0.65rem;"
                        >
                          {issue.message}
                        </span>
                      {/each}
                    </div>
                  {:else}
                    {page.title || "-"}
                  {/if}
                </td>
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
                      <div
                        class="tabs"
                        style="margin-bottom: 1rem; border-color: rgba(255,255,255,0.1);"
                      >
                        <button
                          class="tab {activeDetailTab === 'overview'
                            ? 'active'
                            : ''}"
                          on:click={() => (activeDetailTab = "overview")}
                          >Overview</button
                        >
                        <button
                          class="tab {activeDetailTab === 'issues'
                            ? 'active'
                            : ''}"
                          on:click={() => (activeDetailTab = "issues")}
                          >Issues ({analyzeSEO(page).length})</button
                        >
                        <button
                          class="tab {activeDetailTab === 'seo-errors'
                            ? 'active'
                            : ''}"
                          on:click={() => (activeDetailTab = "seo-errors")}
                          >SEO Errors ({getSeoErrors(page).length})</button
                        >
                        <button
                          class="tab {activeDetailTab === 'headers'
                            ? 'active'
                            : ''}"
                          on:click={() => (activeDetailTab = "headers")}
                          >Headers</button
                        >
                        <button
                          class="tab {activeDetailTab === 'images'
                            ? 'active'
                            : ''}"
                          on:click={() => (activeDetailTab = "images")}
                          >Images</button
                        >
                        <button
                          class="tab {activeDetailTab === 'links'
                            ? 'active'
                            : ''}"
                          on:click={() => (activeDetailTab = "links")}
                          >Links</button
                        >
                      </div>

                      {#if activeDetailTab === "overview"}
                        <div
                          style="display: flex; flex-direction: column; gap: 0.5rem;"
                        >
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
                          <div class="field-v2">
                            <span class="field-v2-label">Meta Description</span>
                            <span class="field-v2-value"
                              >{page.description ||
                                "No description found"}</span
                            >
                          </div>
                          <div class="field-v2">
                            <span class="field-v2-label">Word Count</span>
                            <span class="field-v2-value"
                              >{page.wordCount} words</span
                            >
                          </div>
                          <div class="field-v2">
                            <span class="field-v2-label">Canonical</span>
                            <span
                              class="field-v2-value"
                              style="font-size: 0.75rem;"
                              >{page.canonical || "-"}</span
                            >
                          </div>
                          <div class="field-v2">
                            <span class="field-v2-label">Robots</span>
                            <span class="field-v2-value"
                              >{page.robots || "-"}</span
                            >
                          </div>
                        </div>
                      {/if}

                      {#if activeDetailTab === "issues"}
                        <div class="issues-list">
                          {#each analyzeSEO(page) as issue}
                            <div class="issue-item {issue.level}">
                              <span
                                class="badge-v2 {issue.level === 'error'
                                  ? 'badge-error'
                                  : 'badge-warning'}">{issue.level}</span
                              >
                              <span>{issue.message}</span>
                            </div>
                          {/each}
                          {#if analyzeSEO(page).length === 0}
                            <div
                              style="text-align: center; color: var(--success); padding: 1rem;"
                            >
                              No SEO issues found! 🎉
                            </div>
                          {/if}
                        </div>
                      {/if}

                      {#if activeDetailTab === "seo-errors"}
                        <div class="issues-list">
                          {#each getSeoErrors(page) as issue}
                            <div class="issue-item error">
                              <span class="badge-v2 badge-error">error</span>
                              <span>{issue.message}</span>
                            </div>
                          {/each}
                          {#if getSeoErrors(page).length === 0}
                            <div
                              style="text-align: center; color: var(--success); padding: 1rem;"
                            >
                              No SEO Critical Errors found! 🎉
                            </div>
                          {/if}
                        </div>
                      {/if}

                      {#if activeDetailTab === "headers"}
                        <div
                          style="display: flex; flex-direction: column; gap: 0.5rem; max-height: 300px; overflow-y: auto;"
                        >
                          {#if page.headers}
                            {#each ["h1", "h2", "h3", "h4", "h5", "h6"] as tag}
                              {#if page.headers[tag]}
                                {#each page.headers[tag] as text}
                                  <div class="header-item tag-{tag}">
                                    <span class="tag-badge"
                                      >{tag.toUpperCase()}</span
                                    >
                                    <span>{text}</span>
                                  </div>
                                {/each}
                              {/if}
                            {/each}
                          {:else}
                            <div style="color: var(--text-dim);">
                              No headers data available
                            </div>
                          {/if}
                        </div>
                      {/if}

                      {#if activeDetailTab === "images"}
                        <div style="max-height: 300px; overflow-y: auto;">
                          {#if page.images && page.images.length > 0}
                            <table class="table-v2" style="font-size: 0.8rem;">
                              <thead><th>Preview</th><th>Alt Text</th></thead>
                              <tbody>
                                {#each page.images as img}
                                  <tr>
                                    <td
                                      ><div
                                        style="width: 40px; height: 40px; overflow: hidden; border-radius: 4px; background: #000;"
                                      >
                                        <img
                                          src={img.src}
                                          alt="thumbnail"
                                          style="width: 100%; height: 100%; object-fit: cover;"
                                          onerror="this.src='data:image/svg+xml;base64,...'"
                                        />
                                      </div></td
                                    >
                                    <td
                                      style="color: {img.alt
                                        ? 'inherit'
                                        : 'var(--error)'}"
                                      >{img.alt || "MISSING ALT"}</td
                                    >
                                  </tr>
                                {/each}
                              </tbody>
                            </table>
                          {:else}
                            <div style="color: var(--text-dim);">
                              No images found
                            </div>
                          {/if}
                        </div>
                      {/if}

                      {#if activeDetailTab === "links"}
                        <div
                          style="display: flex; flex-direction: column; gap: 1rem;"
                        >
                          <div
                            class="card stat-v2 link-list-container"
                            style="background: rgba(0,0,0,0.2);"
                          >
                            <div class="stat-v2-label">
                              Internal Links ({page.linksInternal
                                ? page.linksInternal.length
                                : 0})
                            </div>
                            <div class="link-scroller">
                              {#if page.linksInternal}
                                {#each page.linksInternal as link}
                                  <div class="link-item">{link}</div>
                                {/each}
                              {/if}
                            </div>
                          </div>
                          <div
                            class="card stat-v2 link-list-container"
                            style="background: rgba(0,0,0,0.2);"
                          >
                            <div class="stat-v2-label">
                              External Links ({page.linksExternal
                                ? page.linksExternal.length
                                : 0})
                            </div>
                            <div class="link-scroller">
                              {#if page.linksExternal}
                                {#each page.linksExternal as link}
                                  <div class="link-item external">{link}</div>
                                {/each}
                              {/if}
                            </div>
                          </div>
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
  .app-container {
    display: grid;
    grid-template-columns: 320px 1fr;
    height: 100vh;
    overflow: hidden;
    background-color: var(--bg-main);
    transition: grid-template-columns 0.3s ease;
  }

  .app-container.sidebar-closed {
    grid-template-columns: 80px 1fr;
  }

  .sidebar {
    background-color: var(--bg-card);
    border-right: 1px solid var(--border-color);
    padding: 2rem 1.5rem;
    display: flex;
    flex-direction: column;
    overflow-y: hidden;
    overflow-x: hidden;
    white-space: nowrap;
  }

  .app-container.sidebar-closed .sidebar {
    padding: 2rem 0.5rem;
    align-items: center;
  }

  .sidebar-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    width: 100%;
  }

  .sidebar-section {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    margin-bottom: 3rem;
  }

  .btn-icon {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 8px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .btn-icon:hover {
    background-color: var(--bg-main);
    color: var(--text-main);
  }

  .btn-icon-only {
    padding: 0.75rem !important;
    justify-content: center !important;
  }
</style>
