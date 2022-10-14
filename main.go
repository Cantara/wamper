package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"log"
	"os"
	"time"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1600, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	base := "cantara.no"
	url := "https://jenkins." + base + "/view/Build%20Monitor/"
	filename := "jenkins." + base + ".png"

	var imageBuf []byte
	if err := chromedp.Run(ctx, fullScreenshot(url, 90, &imageBuf)); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(filename, imageBuf, 0o644); err != nil {
		log.Fatal(err)
	}
}

type waiter struct {
}

func (w waiter) Do(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	return nil
}
func fullScreenshot(url string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`j_username`, chromedp.ByID),
		chromedp.SetValue(`j_username`, os.Getenv("user"), chromedp.ByID),
		chromedp.SendKeys(`j_username`, kb.Tab+os.Getenv("pass")+kb.Enter, chromedp.ByID),
		chromedp.WaitReady("settings-toggle", chromedp.ByID),
		waiter{},
		chromedp.FullScreenshot(res, quality),
	}
}
