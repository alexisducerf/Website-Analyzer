package main

import (
	"context"
	"fmt"
	"web-seo/internal/crawler"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CrawlWebsite starts the crawling process for the given URL
func (a *App) CrawlWebsite(startURL string, concurrency int) {
	if a.cancel != nil {
		a.cancel()
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel

	c, err := crawler.NewCrawler(startURL, concurrency)
	if err != nil {
		runtime.EventsEmit(a.ctx, "crawl-error", err.Error())
		return
	}

	go c.Start(ctx)

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case result, ok := <-c.Results:
				if !ok {
					runtime.EventsEmit(a.ctx, "crawl-complete", nil)
					return
				}
				runtime.EventsEmit(a.ctx, "page-discovered", result)
			}
		}
	}()
}

// StopCrawl stops the current crawling process
func (a *App) StopCrawl() {
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
