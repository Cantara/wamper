package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
)

var (
	width  = 1600
	height = 1200
)

func init() {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}
}

func main() {
	scr, err := ScreenshotWAuth("https://jenkins.cantara.no/view/Build%20Monitor/", "viewer", "NxP3Kscez7KrTAQa9rM4F")
	if err != nil {
		log.Fatalf("unable to take screenshot: %v", err)
	}

	out, err := os.OpenFile("tmp/example.png", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Writer the body to file
	_, err = io.Copy(out, bytes.NewBuffer(scr))
	if err != nil {
		log.Fatal(err)
	}
}

func ScreenshotWAuth(path, username, password string) (out []byte, err error) {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch playwright: %v", err)
	}
	browser, err := pw.WebKit.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage(playwright.BrowserNewContextOptions{
		Screen: &playwright.BrowserNewContextOptionsScreen{
			Width:  &width,
			Height: &height,
		},
		Viewport: &playwright.BrowserNewContextOptionsViewport{
			Width:  &width,
			Height: &height,
		},
	})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto(path, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	err = page.Fill("input[type=\"text\"]", username)
	if err != nil {
		return
	}
	err = page.Fill("input[type=\"password\"]", password)
	if err != nil {
		return
	}
	err = page.Click("button[type=\"submit\"]")
	if err != nil {
		return
	}
	page.WaitForLoadState(string(*playwright.WaitUntilStateDomcontentloaded))
	time.Sleep(time.Second * 5)
	fullscreen := true
	out, err = page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: &fullscreen,
		//Path:     playwright.String("example.png"),
		Type: playwright.ScreenshotTypePng,
	})
	if err != nil {
		log.Fatalf("could not create screenshot: %v", err)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
	return
}
