package main

import (
	"github.com/cantara/wamper/screenshot"
	"github.com/cantara/wamper/sites"
	"net/url"
	"os"
)

func main() {
	u, err := url.Parse("https://jenkins.exoreaction.com/view/Build%20Monitor/")
	if err != nil {
		panic(err)
	}
	s, err := screenshot.GetScreenshot(sites.Site{
		Name:      "test",
		Url:       *u,
		LoginType: sites.Github,
		Username:  "jenkins-dashboard",
		Password:  "",
	})
	if err != nil {
		panic(err)
	}
	os.WriteFile("test.jpg", s.Buf, 0644)
}
