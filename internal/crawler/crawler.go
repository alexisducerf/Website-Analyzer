package crawler

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/html"
)

// PageInfo represents data about a crawled page
type PageInfo struct {
	URL         string `json:"url"`
	Status      int    `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	H1          string `json:"h1"`
	WordCount   int    `json:"wordCount"`
	Error       string `json:"error,omitempty"`
}

// Crawler handles the asynchronous crawling of a website
type Crawler struct {
	BaseURL     *url.URL
	TargetHost  string
	Visited     map[string]bool
	visitedLock sync.Mutex
	Results     chan PageInfo
	Concurrency int
	Delay       time.Duration
}

// NewCrawler creates a new Crawler instance
func NewCrawler(startURL string, concurrency int, delayMs int) (*Crawler, error) {
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	return &Crawler{
		BaseURL:     parsedURL,
		TargetHost:  parsedURL.Host,
		Visited:     make(map[string]bool),
		Results:     make(chan PageInfo, 100),
		Concurrency: concurrency,
		Delay:       time.Duration(delayMs) * time.Millisecond,
	}, nil
}

// Start initiates the crawling process starting from the base URL
func (c *Crawler) Start(ctx context.Context) {
	queue := make(chan string, 10000)
	queue <- c.BaseURL.String()

	var wg sync.WaitGroup
	activeWorkers := int32(0)

	// Track active work
	wg.Add(1)
	atomic.AddInt32(&activeWorkers, 1)

	// Dispatcher
	go func() {
		for i := 0; i < c.Concurrency; i++ {
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case link, ok := <-queue:
						if !ok {
							return
						}
						c.crawlPage(ctx, link, queue, &wg, &activeWorkers)
					}
				}
			}()
		}
	}()

	// Wait for all work to be done in a separate goroutine
	go func() {
		wg.Wait()
		close(queue)
		close(c.Results)
	}()
}

func (c *Crawler) crawlPage(ctx context.Context, pageURL string, queue chan string, wg *sync.WaitGroup, activeWorkers *int32) {
	defer func() {
		atomic.AddInt32(activeWorkers, -1)
		wg.Done()
	}()

	c.visitedLock.Lock()
	if c.Visited[pageURL] {
		c.visitedLock.Unlock()
		return
	}
	c.Visited[pageURL] = true
	c.visitedLock.Unlock()

	// Rate limiting
	if c.Delay > 0 {
		time.Sleep(c.Delay)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		c.Results <- PageInfo{URL: pageURL, Error: err.Error()}
		return
	}

	// Set a modern browser User-Agent to avoid 403 errors
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		c.Results <- PageInfo{URL: pageURL, Error: err.Error()}
		return
	}
	defer resp.Body.Close()

	info := PageInfo{
		URL:    pageURL,
		Status: resp.StatusCode,
	}

	if resp.StatusCode == http.StatusOK {
		links, seoInfo := c.parse(resp)
		info.Title = seoInfo.Title
		info.Description = seoInfo.Description
		info.H1 = seoInfo.H1
		info.WordCount = seoInfo.WordCount

		for _, link := range links {
			if c.shouldCrawl(link) {
				c.visitedLock.Lock()
				alreadyVisited := c.Visited[link]
				c.visitedLock.Unlock()

				if !alreadyVisited {
					wg.Add(1)
					atomic.AddInt32(activeWorkers, 1)
					select {
					case queue <- link:
					default:
						// If queue is full, we should still decrement the counter we just added
						atomic.AddInt32(activeWorkers, -1)
						wg.Done()
					}
				}
			}
		}
	}

	c.Results <- info
}

func (c *Crawler) shouldCrawl(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Host == c.TargetHost
}

func (c *Crawler) parse(resp *http.Response) ([]string, PageInfo) {
	var links []string
	var info PageInfo
	var inTitle, inH1 bool

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links, info
		case html.TextToken:
			if inTitle {
				info.Title = strings.TrimSpace(z.Token().Data)
			}
			if inH1 && info.H1 == "" {
				info.H1 = strings.TrimSpace(z.Token().Data)
			}
			// Simple word count
			text := strings.TrimSpace(z.Token().Data)
			if text != "" {
				info.WordCount += len(strings.Fields(text))
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						val := strings.TrimSpace(a.Val)
						if val == "" || strings.HasPrefix(val, "#") || strings.HasPrefix(val, "javascript:") {
							continue
						}

						absURL := c.resolveURL(val)
						if absURL != "" {
							links = append(links, absURL)
						}
					}
				}
			}
			if t.Data == "title" {
				inTitle = true
			}
			if t.Data == "h1" {
				inH1 = true
			}
			if t.Data == "meta" {
				var name, content string
				for _, a := range t.Attr {
					if a.Key == "name" {
						name = strings.ToLower(a.Val)
					}
					if a.Key == "content" {
						content = a.Val
					}
				}
				if name == "description" {
					info.Description = content
				}
			}
		case html.EndTagToken:
			t := z.Token()
			if t.Data == "title" {
				inTitle = false
			}
			if t.Data == "h1" {
				inH1 = false
			}
		}
	}
}

func (c *Crawler) resolveURL(href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return c.BaseURL.ResolveReference(u).String()
}
