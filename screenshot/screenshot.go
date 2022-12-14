package screenshot

import (
	"context"
	"fmt"
	"github.com/cantara/wamper/sites"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"net/url"
	"time"
)

type Screenshot struct {
	Name      string    `json:"name"`
	Url       url.URL   `json:"url"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Buf       []byte    `json:"buf"`
}

type Task struct {
	Site     sites.Site    `json:"site"`
	Time     time.Time     `json:"time"`
	Interval time.Duration `json:"interval"`
}

func (s Screenshot) Id() string {
	return fmt.Sprintf("%s_%s", s.Name, s.CreatedAt.Format("2006-01-02_15:04:05"))
}
func GetScreenshotJenkins(site sites.Site) (s Screenshot, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1600, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx) //, chromedp.WithDebugf(log.Printf))
	defer cancel()

	err = chromedp.Run(ctx, fullScreenshotJenkins(site.Url, site.Username, site.Password, 90, &s.Buf))
	if err != nil {
		return
	}

	s.Name = site.Name
	s.Url = site.Url
	s.Type = "png"
	s.CreatedAt = time.Now()
	return
}

func GetScreenshot(site sites.Site) (s Screenshot, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	err = chromedp.Run(ctx, fullScreenshot(site.Url, 90, &s.Buf))
	if err != nil {
		return
	}

	s.Name = site.Name
	s.Url = site.Url
	s.Type = "png"
	s.CreatedAt = time.Now()
	return
}

type waiter struct {
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

func (w waiter) Do(ctx context.Context) error {
	return sleepContext(ctx, 5*time.Second)
}

func fullScreenshotJenkins(url url.URL, username, password string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url.String()),
		chromedp.WaitVisible(`j_username`, chromedp.ByID),
		chromedp.SetValue(`j_username`, username, chromedp.ByID),
		chromedp.SendKeys(`j_username`, kb.Tab+password+kb.Enter, chromedp.ByID),
		waiter{},
		chromedp.FullScreenshot(res, quality),
	}
}
func fullScreenshot(url url.URL, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url.String()),
		waiter{},
		chromedp.FullScreenshot(res, quality),
	}
}
