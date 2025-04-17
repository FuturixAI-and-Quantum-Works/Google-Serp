// Package browser provides browser automation functionality
package browser

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Pool manages a pool of browser contexts for reuse
type Pool struct {
	contexts      chan context.Context
	cancelFuncs   map[context.Context]context.CancelFunc
	initOnce      sync.Once
	minSize       int
	maxSize       int
	currentSize   int
	mu            sync.Mutex
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	initialized   bool
	scaleUpCount  int       // Tracks consecutive scale up events
	scaleDownTime time.Time // Last time we scaled down
	waitQueue     int       // Count of waiting requests
}

// New creates a new browser pool with the specified minimum and maximum sizes
func New(minSize, maxSize int) *Pool {
	return &Pool{
		minSize:     minSize,
		maxSize:     maxSize,
		currentSize: 0,
		contexts:    make(chan context.Context, maxSize),
		cancelFuncs: make(map[context.Context]context.CancelFunc),
	}
}

// DefaultPool is a global browser pool with auto-scaling
var DefaultPool = New(10, 30)

// Initialize creates the browser pool
func (pool *Pool) Initialize() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.initialized {
		return
	}

	// Create a shared allocator context with options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-site-isolation-trials", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("enable-javascript", true), // Allow JavaScript execution
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"),
	)

	pool.allocCtx, pool.allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)

	// Create initial set of browser contexts
	pool.scaleUp(pool.minSize)

	// Start the auto-scaler
	go pool.autoScaler()

	pool.initialized = true
	fmt.Printf("Browser pool initialized with %d browsers (min: %d, max: %d)\n",
		pool.currentSize, pool.minSize, pool.maxSize)
}

// scaleUp adds n browser instances to the pool
func (pool *Pool) scaleUp(n int) {
	for i := 0; i < n; i++ {
		ctx, cancel := chromedp.NewContext(pool.allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
			// Silent logging
		}))

		// Add event listeners to handle the events that need handling
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			// Silently handle the EventFrameStartedNavigating event
			switch ev.(type) {
			case *page.EventFrameStartedNavigating:
				// Silent handling
			}
		})

		// Initialize the browser in advance
		if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
			fmt.Printf("Error initializing browser: %v\n", err)
			cancel()
			continue
		}

		pool.contexts <- ctx
		pool.cancelFuncs[ctx] = cancel
		pool.currentSize++
	}

	if n > 0 {
		fmt.Printf("Scaled up pool by %d browsers to %d total\n", n, pool.currentSize)
	}
}

// scaleDown removes n browser instances from the pool
func (pool *Pool) scaleDown(n int) {
	if pool.currentSize <= pool.minSize {
		return // Don't scale below minimum
	}

	// Limit scale down to not go below minSize
	if pool.currentSize-n < pool.minSize {
		n = pool.currentSize - pool.minSize
	}

	if n <= 0 {
		return
	}

	for i := 0; i < n; i++ {
		select {
		case ctx := <-pool.contexts:
			// Get the cancel function
			if cancel, exists := pool.cancelFuncs[ctx]; exists {
				cancel()
				delete(pool.cancelFuncs, ctx)
				pool.currentSize--
			}
		default:
			// No contexts available to scale down
			return
		}
	}

	pool.scaleDownTime = time.Now()
	fmt.Printf("Scaled down pool by %d browsers to %d total\n", n, pool.currentSize)
}

// autoScaler periodically checks if the pool needs to be resized
func (pool *Pool) autoScaler() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		pool.mu.Lock()

		// Get metrics
		poolSize := pool.currentSize
		availableBrowsers := len(pool.contexts)
		waitingRequests := pool.waitQueue

		// Calculate utilization percentage (0-100)
		utilization := 0
		if poolSize > 0 {
			utilization = 100 * (poolSize - availableBrowsers) / poolSize
		}

		// Scale up logic
		if (utilization > 80 || waitingRequests > 0) && poolSize < pool.maxSize {
			// Calculate how many browsers to add
			toAdd := 1
			if waitingRequests > 1 {
				// Add more browsers if there are multiple waiting requests
				toAdd = min(waitingRequests, pool.maxSize-poolSize)
			}

			pool.scaleUpCount++
			// If we've needed to scale up multiple times in succession, add more browsers
			if pool.scaleUpCount > 3 {
				toAdd = min(toAdd*2, pool.maxSize-poolSize)
			}

			pool.scaleUp(toAdd)
		} else {
			pool.scaleUpCount = 0
		}

		// Scale down logic - only if utilization is low and we haven't scaled down recently
		cooldownPeriod := 2 * time.Minute
		if utilization < 30 && poolSize > pool.minSize && time.Since(pool.scaleDownTime) > cooldownPeriod {
			// Calculate how many browsers to remove
			excessBrowsers := availableBrowsers - max(1, poolSize/5) // Keep at least 20% capacity as buffer
			toRemove := min(excessBrowsers, poolSize-pool.minSize)

			if toRemove > 0 {
				pool.scaleDown(toRemove)
			}
		}

		pool.mu.Unlock()
	}
}

// GetContext gets a browser context from the pool
func (pool *Pool) GetContext() (context.Context, context.CancelFunc, error) {
	pool.initOnce.Do(func() {
		pool.Initialize()
	})

	pool.mu.Lock()
	pool.waitQueue++
	pool.mu.Unlock()

	// Try to get a context immediately or wait up to 500ms
	select {
	case ctx := <-pool.contexts:
		pool.mu.Lock()
		pool.waitQueue--
		pool.mu.Unlock()

		// Create a return function that puts the context back in the pool
		returnCtx := func() {
			// Refresh the browser before returning to pool
			refreshCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			// Navigate to blank page to clear state and reduce memory
			_ = chromedp.Run(refreshCtx,
				network.ClearBrowserCookies(),
				chromedp.Navigate("about:blank"),
			)

			select {
			case pool.contexts <- ctx:
			default:
				// Pool channel is full, context will be GC'd
				if cancel, exists := pool.cancelFuncs[ctx]; exists {
					pool.mu.Lock()
					cancel()
					delete(pool.cancelFuncs, ctx)
					pool.currentSize--
					pool.mu.Unlock()
				}
			}
		}

		return ctx, returnCtx, nil

	case <-time.After(500 * time.Millisecond):
		// If we waited more than 500ms, try to create a new browser instance
		pool.mu.Lock()

		// Only create a new instance if we're below max capacity
		if pool.currentSize < pool.maxSize {
			ctx, cancel := chromedp.NewContext(pool.allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
				// Silent logging
			}))

			// Try to initialize the browser quickly
			initCtx, initCancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := chromedp.Run(initCtx, chromedp.Navigate("about:blank"))
			initCancel()

			if err != nil {
				cancel()
				pool.waitQueue--
				pool.mu.Unlock()
				return nil, nil, fmt.Errorf("failed to create new browser instance: %v", err)
			}

			// Add to pool management
			pool.cancelFuncs[ctx] = cancel
			pool.currentSize++
			pool.waitQueue--
			pool.mu.Unlock()

			// Create return function
			returnCtx := func() {
				// Refresh before returning to pool
				refreshCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()

				_ = chromedp.Run(refreshCtx,
					network.ClearBrowserCookies(),
					chromedp.Navigate("about:blank"),
				)

				// Put context back in the pool
				select {
				case pool.contexts <- ctx:
					// Successfully returned to pool
				default:
					// Pool is full, this extra instance will be closed
					if cancel, exists := pool.cancelFuncs[ctx]; exists {
						pool.mu.Lock()
						cancel()
						delete(pool.cancelFuncs, ctx)
						pool.currentSize--
						pool.mu.Unlock()
					}
				}
			}

			return ctx, returnCtx, nil
		}

		// If we couldn't create a new instance, wait longer for an existing one
		pool.mu.Unlock()

		// Wait up to 3 more seconds for a context to become available
		select {
		case ctx := <-pool.contexts:
			pool.mu.Lock()
			pool.waitQueue--
			pool.mu.Unlock()

			returnCtx := func() {
				select {
				case pool.contexts <- ctx:
					// Successfully returned to pool
				default:
					// Pool is full
					fmt.Println("Browser pool is full, cannot return context")
				}
			}

			return ctx, returnCtx, nil
		case <-time.After(3 * time.Second):
			pool.mu.Lock()
			pool.waitQueue--
			pool.mu.Unlock()
			return nil, nil, fmt.Errorf("timeout getting browser context from pool")
		}
	}
}

// Shutdown closes all browser instances
func (pool *Pool) Shutdown() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.initialized {
		return
	}

	// Cancel all contexts
	for ctx, cancel := range pool.cancelFuncs {
		cancel()
		delete(pool.cancelFuncs, ctx)
	}

	// Cancel the allocator
	if pool.allocCancel != nil {
		pool.allocCancel()
	}

	// Clear the channel
	for len(pool.contexts) > 0 {
		<-pool.contexts
	}

	pool.currentSize = 0
	pool.initialized = false
	fmt.Println("Browser pool shut down")
}

// FetchURL navigates to a URL and returns the HTML content
func (pool *Pool) FetchURL(url string, timeout time.Duration) (string, error) {
	ctx, returnCtx, err := pool.GetContext()
	if err != nil {
		return "", fmt.Errorf("failed to get browser context: %v", err)
	}
	defer returnCtx() // Return the context to the pool when done

	var htmlContent string

	// Add a timeout for this specific operation
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Navigate to the URL and scrape the content
	err = chromedp.Run(timeoutCtx,
		// Navigate to the search URL
		chromedp.Navigate(url),
		// Wait for a moment
		chromedp.Sleep(1000*time.Millisecond),
		// Extract the full HTML of the page
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL content: %v", err)
	}

	return htmlContent, nil
}

// initialize the pool at package initialization time
func init() {
	// Initialize the browser pool in a background goroutine
	go DefaultPool.Initialize()
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
