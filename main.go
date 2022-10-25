package main

import (
	"context"
	"github.com/cantara/wamper/atomic"
	"github.com/cantara/wamper/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/cantara/bragi"
	"github.com/cantara/wamper/slack"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
}

func main() {
	loadEnv()
	logDir := os.Getenv("log.dir")
	if logDir != "" {
		log.SetPrefix("wamper")
		cloaser := log.SetOutputFolder(logDir)
		if cloaser == nil {
			log.Fatal("Unable to sett logdir")
		}
		defer cloaser()
		done := make(chan func())
		log.StartRotate(done)
		defer close(done)
	}
	postTime, err := strconv.Atoi(os.Getenv("post_time"))
	if err != nil {
		log.AddError(err).Fatal("missing integer for post time")
	}
	serv := web.Init()
	slack.NewClient(os.Getenv("slack.token"))

	jenkinsimg, err := GetScreenshotJenkins()
	if err != nil {
		log.AddError(err).Fatal("while taking screenshot")
	}

	jenk := atomic.NewValue[[]byte](jenkinsimg)
	refreshJenkins := time.NewTicker(5 * time.Minute)
	defer refreshJenkins.Stop()
	go func() {
		for range refreshJenkins.C {
			jenkinsimg, err = GetScreenshotJenkins()
			if err != nil {
				log.AddError(err).Error("while taking screenshot")
				continue
			}
			jenk.Store(jenkinsimg)
		}
	}()

	visualeimg, err := GetScreenshotVisuale()
	if err != nil {
		log.AddError(err).Fatal("while taking screenshot")
	}

	visuale := atomic.NewValue[[]byte](visualeimg)
	refreshVisuale := time.NewTicker(5 * time.Minute)
	defer refreshVisuale.Stop()
	go func() {
		if os.Getenv("visuale.pass") == "" {
			return
		}
		for range refreshVisuale.C {
			visualeimg, err = GetScreenshotVisuale()
			if err != nil {
				log.AddError(err).Error("while taking screenshot")
				continue
			}
			visuale.Store(visualeimg)
		}
	}()

	go func() {
		nextDay := time.Now().UTC()
		//Changing to next day if we are passed the post time with a buffer of 10 sec.
		if nextDay.Hour() >= postTime || nextDay.Hour() == postTime-1 && nextDay.Minute() == 59 && nextDay.Second() > 50 {
			nextDay = nextDay.AddDate(0, 0, 1)
		}
		nextDay = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), postTime, 0, 0, 1, time.UTC)
		nextDayIn := nextDay.Sub(time.Now().UTC())
		dailyTicker := time.NewTicker(nextDayIn)
		firstDay := true
		for range dailyTicker.C {
			if firstDay {
				dailyTicker.Reset(24 * time.Hour)
				firstDay = false
			}
			log.Println(slack.SendFile(os.Getenv("slack.channel"), "Today's Jenkins build status!", jenk.Load()))
		}
	}()
	serv.API.GET("/image", func(c *gin.Context) {
		url, ok := c.GetQuery("url")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "url query not provided",
			})
			return
		}
		img, err := GetScreenshot(url)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "png", img)
	})
	serv.API.GET("/jenkins", func(c *gin.Context) {
		c.Data(http.StatusOK, "png", jenk.Load())
	})
	serv.API.GET("/visuale", func(c *gin.Context) {
		c.Data(http.StatusOK, "png", visuale.Load())
	})

	serv.Run()
}

func GetScreenshotJenkins() (buf []byte, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1600, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx) //, chromedp.WithDebugf(log.Printf))
	defer cancel()

	url := "https://jenkins." + os.Getenv("domain") + "." + os.Getenv("tld") +
		"/view/" + os.Getenv("view") + "/"
	err = chromedp.Run(ctx, fullScreenshotWAuth(url, 90, &buf))
	return
}

func GetScreenshotVisuale() (buf []byte, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	url := "https://visuale." + os.Getenv("domain") + "." + os.Getenv("tld") +
		"/?accessToken=" + os.Getenv("visuale.pass") + "&servicetype=true&ui_extension=groupByTag"
	err = chromedp.Run(ctx, fullScreenshot(url, 90, &buf))
	return
}

func GetScreenshot(url string) (buf []byte, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1600, 1200),
		chromedp.DisableGPU,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	err = chromedp.Run(ctx, fullScreenshot(url, 90, &buf))
	return
}

type waiter struct {
}

func (w waiter) Do(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	return nil
}

func fullScreenshotWAuth(url string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`j_username`, chromedp.ByID),
		chromedp.SetValue(`j_username`, os.Getenv("user"), chromedp.ByID),
		chromedp.SendKeys(`j_username`, kb.Tab+os.Getenv("pass")+kb.Enter, chromedp.ByID),
		//chromedp.WaitReady("settings-toggle", chromedp.ByID),
		waiter{},
		chromedp.FullScreenshot(res, quality),
	}
}
func fullScreenshot(url string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		waiter{},
		chromedp.FullScreenshot(res, quality),
	}
}
