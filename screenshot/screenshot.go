package screenshot

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/wamper/sites"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
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

var lock *sync.Mutex

func init() {
	lock = &sync.Mutex{}
}

func (s Screenshot) Id() string {
	return fmt.Sprintf("%s_%s", s.Name, s.CreatedAt.Format("2006-01-02_15:04:05"))
}

func GetScreenshot(site sites.Site) (s Screenshot, err error) {
	lock.Lock()
	defer lock.Unlock()
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	switch site.LoginType {
	case sites.Jenkins:
		err = chromedp.Run(ctx, fullScreenshotJenkins(site.Url, site.Username, string(site.Password), 90, &s.Buf))
	case sites.Github:
		err = chromedp.Run(ctx, fullScreenshotLoginField(site.Url, site.Username, string(site.Password), 90, &s.Buf))
	default:
		err = chromedp.Run(ctx, fullScreenshot(site.Url, 90, &s.Buf))
	}
	if err != nil {
		log.WithError(err).Error("while getting screenshot")
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
func fullScreenshotLoginField(url url.URL, username, password string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url.String()),
		chromedp.WaitVisible(`login_field`, chromedp.ByID),
		chromedp.SetValue(`login_field`, username, chromedp.ByID),
		chromedp.SendKeys(`login_field`, kb.Tab+password+kb.Enter, chromedp.ByID),
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
