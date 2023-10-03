package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/cantara/gober/stream"
	"github.com/cantara/gober/stream/event/store/eventstore"
	"github.com/cantara/gober/stream/event/store/ondisk"
	"github.com/cantara/gober/webserver/health"
	"github.com/cantara/wamper/screenshot"
	"github.com/cantara/wamper/sites"
	"github.com/cantara/wamper/slack"
	"github.com/gin-gonic/gin"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/gober/webserver"
)

func init() {
	health.Name = "wamper"
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	portString := os.Getenv("webserver.port")
	port, err := strconv.Atoi(portString)
	if err != nil {
		log.WithError(err).Fatal("while getting webserver port")
	}
	serv, err := webserver.Init(uint16(port), false)
	if err != nil {
		log.WithError(err).Fatal("while initializing webserver")
	}
	var siteStream stream.Stream
	var scrStream stream.Stream
	var slackStream stream.Stream
	if esHost := os.Getenv("eventstore.host"); esHost != "" {
		es, err := eventstore.NewClient(esHost)
		if err != nil {
			panic(err)
		}
		siteStream, err = eventstore.NewStream(es, "sites", ctx)
		if err != nil {
			log.WithError(err).Fatal("while initializing site stream")
			return
		}
		scrStream, err = eventstore.NewStream(es, "screenshots", ctx)
		if err != nil {
			log.WithError(err).Fatal("while initializing site stream")
			return
		}
		slackStream, err = eventstore.NewStream(es, "slack", ctx)
		if err != nil {
			log.WithError(err).Fatal("while initializing site stream")
			return
		}
	} else {
		siteStream, err = ondisk.Init("sites", ctx)
		if err != nil {
			log.WithError(err).Fatal("while initializing site stream")
			return
		}
		scrStream, err = ondisk.Init("screenshots", ctx)
		if err != nil {
			log.WithError(err).Fatal("while initializing site stream")
			return
		}
		slackStream, err = ondisk.Init("slack", ctx)
		if err != nil {
			log.WithError(err).Fatal("while initializing site stream")
			return
		}
	}
	siteStore, err := sites.Init(siteStream, ctx)
	if err != nil {
		log.WithError(err).Fatal("while initializing sites store")
		return
	}
	scrStore, err := screenshot.InitStore(serv, scrStream, log.RedactedString(os.Getenv("screenshot.key")), ctx)
	if err != nil {
		log.WithError(err).Fatal("while initializing screenshot store")
		return
	}
	scrService, err := screenshot.Init(scrStream, scrStore, log.RedactedString(os.Getenv("screenshot.service.key")), ctx)
	if err != nil {
		log.WithError(err).Fatal("while initializing screenshot store")
		return
	}
	slackService, err := slack.Init(slackStream, scrStore, log.RedactedString(os.Getenv("slack.service.key")), ctx)
	if err != nil {
		log.WithError(err).Fatal("while initializing screenshot store")
		return
	}

	serv.API.PUT("/site", func(c *gin.Context) {
		auth := webserver.GetAuthHeader(c)
		if auth != os.Getenv("authkey") {
			webserver.ErrorResponse(c, "not authenticated", http.StatusForbidden)
			return
		}
		site, err := webserver.UnmarshalBody[Site](c)
		if err != nil {
			webserver.ErrorResponse(c, err.Error(), http.StatusBadRequest)
			return
		}
		if site.LoginType != "" && site.LoginType != string(sites.None) {
			if site.Username == "" {
				webserver.ErrorResponse(c, "username for jenkins is missing", http.StatusBadRequest)
				return
			}
			if site.Password == "" {
				webserver.ErrorResponse(c, "password for jenkins is missing", http.StatusBadRequest)
				return
			}
		}
		err = siteStore.Set(sites.Site{
			Name:      site.Name,
			Url:       *site.Url.Url(),
			LoginType: sites.LoginType(site.LoginType),
			Username:  site.Username,
			Password:  site.Password,
		})
		if err != nil {
			webserver.ErrorResponse(c, err.Error(), http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "site added"})
		return
	})
	serv.API.GET("/site", func(c *gin.Context) {
		name, ok := c.GetQuery("name")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "name query not provided",
			})
			return
		}
		site, err := siteStore.Get(name)
		if err != nil {
			log.WithError(err).Info("site not found during get request", name)
			webserver.ErrorResponse(c, "site not found", http.StatusNotFound)
			return
		}
		scr, err := scrStore.Get(site.Id())
		if err != nil { //Here we could add / do some check on weather it is a not found error or any other error
			log.WithError(err).Error("while getting screenshot during get request")
			webserver.ErrorResponse(c, err.Error(), http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "png", scr.Buf)
		return
	})
	serv.API.PUT("/screenshot/task", func(c *gin.Context) {
		auth := webserver.GetAuthHeader(c)
		if auth != os.Getenv("authkey") {
			webserver.ErrorResponse(c, "not authenticated", http.StatusForbidden)
			return
		}
		task, err := webserver.UnmarshalBody[ScreenshotTask](c)
		if err != nil {
			webserver.ErrorResponse(c, err.Error(), http.StatusBadRequest)
			return
		}
		log.Info("new task", "task", task)
		site, err := siteStore.Get(task.Site)
		if err != nil {
			log.WithError(err).Info("site not found during get request", task.Site)
			webserver.ErrorResponse(c, "site not found", http.StatusBadRequest) //Personally feel I could use not found, but that is technically wrong
			return
		}

		err = scrService.Set(screenshot.Task{
			Site:     site,
			Time:     task.Time,
			Interval: task.Interval,
		})
		if err != nil {
			webserver.ErrorResponse(c, err.Error(), http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "screenshot task added"})
		return
	})
	serv.API.PUT("/slack/task", func(c *gin.Context) {
		auth := webserver.GetAuthHeader(c)
		if auth != os.Getenv("authkey") {
			webserver.ErrorResponse(c, "not authenticated", http.StatusForbidden)
			return
		}
		task, err := webserver.UnmarshalBody[SlackTask](c)
		if err != nil {
			webserver.ErrorResponse(c, err.Error(), http.StatusBadRequest)
			return
		}
		site, err := siteStore.Get(task.Site)
		if err != nil {
			log.WithError(err).Info("site not found during get request", task.Site)
			webserver.ErrorResponse(c, "site not found", http.StatusBadRequest) //Personally feel I could use not found, but that is technically wrong
			return
		}

		err = slackService.Set(slack.Task{
			Site:         site,
			Time:         task.Time,
			Interval:     task.Interval,
			SlackToken:   task.SlackToken,
			SlackChannel: task.SlackChannel,
		})
		if err != nil {
			webserver.ErrorResponse(c, err.Error(), http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "slack task added"})
		return
	})

	serv.Run()
}

type ScreenshotTask struct {
	Site     string        `json:"site_name"`
	Time     time.Time     `json:"time"`
	Interval time.Duration `json:"interval"`
}

type SlackTask struct {
	Site         string             `json:"site_name"`
	Time         time.Time          `json:"time"`
	Interval     time.Duration      `json:"interval"`
	SlackToken   log.RedactedString `json:"slack_token"`
	SlackChannel string             `json:"slack_channel"`
}

type Site struct {
	Name      string             `json:"name"`
	Url       u                  `json:"url"`
	LoginType string             `json:"login_type"`
	Username  string             `json:"username"`
	Password  log.RedactedString `json:"password"`
}

type u url.URL

func (i *u) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	ur, err := url.Parse(s)
	if err != nil {
		return err
	}
	*i = (u)(*ur)
	return nil
}

func (i *u) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Url().String())
}

func (i *u) Url() *url.URL {
	return (*url.URL)(i)
}
