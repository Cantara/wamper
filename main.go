package main

import (
	"context"
	"encoding/json"
	"github.com/cantara/gober/store/eventstore"
	"github.com/cantara/gober/store/inmemory"
	"github.com/cantara/gober/stream"
	"github.com/cantara/wamper/screenshot"
	"github.com/cantara/wamper/sites"
	"github.com/cantara/wamper/slack"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	log "github.com/cantara/bragi"
	"github.com/cantara/gober/webserver"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	portString := os.Getenv("webserver.port")
	port, err := strconv.Atoi(portString)
	if err != nil {
		log.AddError(err).Fatal("while getting webserver port")
	}
	serv, err := webserver.Init(uint16(port))
	if err != nil {
		log.AddError(err).Fatal("while initializing webserver")
	}
	var st stream.Persistence
	if os.Getenv("inmem") == "true" {
		var err error
		st, err = inmemory.Init()
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		st, err = eventstore.Init()
		if err != nil {
			panic(err)
		}
	}
	siteStream, err := stream.Init(st, "sites", ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing site stream")
		return
	}
	siteStore, err := sites.Init(siteStream, ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing sites store")
		return
	}
	scrStream, err := stream.Init(st, "screenshots", ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing site stream")
		return
	}
	scrStore, err := screenshot.InitStore(serv, scrStream, os.Getenv("screenshot.key"), ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing screenshot store")
		return
	}
	scrService, err := screenshot.Init(scrStream, scrStore, os.Getenv("screenshot.service.key"), ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing screenshot store")
		return
	}
	slackStream, err := stream.Init(st, "slack", ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing site stream")
		return
	}
	slackService, err := slack.Init(slackStream, scrStore, os.Getenv("slack.service.key"), ctx)
	if err != nil {
		log.AddError(err).Fatal("while initializing screenshot store")
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
		if site.Jenkins {
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
			Name:     site.Name,
			Url:      *site.Url.Url(),
			Jenkins:  site.Jenkins,
			Username: site.Username,
			Password: site.Password,
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
			log.AddError(err).Info("site not found during get request", name)
			webserver.ErrorResponse(c, "site not found", http.StatusNotFound)
			return
		}
		scr, err := scrStore.Get(site.Id())
		if err != nil { //Here we could add / do some check on weather it is a not found error or any other error
			log.AddError(err).Error("while getting screenshot during get request")
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
		log.Println(task)
		log.Println(time.Now())
		site, err := siteStore.Get(task.Site)
		if err != nil {
			log.AddError(err).Info("site not found during get request", task.Site)
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
			log.AddError(err).Info("site not found during get request", task.Site)
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
	Site         string        `json:"site_name"`
	Time         time.Time     `json:"time"`
	Interval     time.Duration `json:"interval"`
	SlackToken   string        `json:"slack_token"`
	SlackChannel string        `json:"slack_channel"`
}

type Site struct {
	Name     string `json:"name"`
	Url      u      `json:"url"`
	Jenkins  bool   `json:"jenkins"`
	Username string `json:"username"`
	Password string `json:"password"`
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
