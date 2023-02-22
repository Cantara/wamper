package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/cantara/bragi/sbragi"
)

type slackMessage struct {
	SlackId string `json:"channel"`
	TS      string `json:"thread_ts"`
	Text    string `json:"text"`
	//	Username    string   `json:"username"`
	Pinned bool `json:"pinned"`
	//	Attachments []string `json:"attachments"`
}

type slackFile struct {
	Channels []string `json:"channels"`
	Text     string   `json:"initial_comment"`
	File     []byte   `json:"file"`
	Name     string   `json:"name"`
}

type slackRespons struct {
	Ok               bool        `json:"ok"`
	SlackId          string      `json:"channel"`
	TS               string      `json:"ts"`
	Message          Message     `json:"message"`
	Warning          string      `json:"warning"`
	ResponseMetadata interface{} `json:"response_metadata"`
}

type Message struct {
	BotId      string     `json:"bot_id"`
	Type       string     `json:"type"`
	Text       string     `json:"text"`
	User       string     `json:"user"`
	TS         string     `json:"ts"`
	Team       string     `json:"team"`
	BotProfile BotProfile `json:"bot_profile"`
	Deleted    bool       `json:"deleted"`
	Updated    int        `json:"updated"`
	TeamId     string     `json:"team_id"`
}

type BotProfile struct {
	/*
	   "id": "B02V186UMM5",
	   "app_id": "A02V959QU94",
	   "name": "Nerthus",
	   "icons": {
	     "image_36": "https:\\/\\/a.slack-edge.com\\/80588\\/img\\/plugins\\/app\\/bot_36.png",
	     "image_48": "https:\\/\\/a.slack-edge.com\\/80588\\/img\\/plugins\\/app\\/bot_48.png",
	     "image_72": "https:\\/\\/a.slack-edge.com\\/80588\\/img\\/plugins\\/app\\/service_72.png"
	*/
}

type client struct {
	baseurl string
	token   log.RedactedString
}

func NewClient(authToken log.RedactedString) (c client, err error) {
	c = client{
		baseurl: "https://slack.com",
		token:   authToken,
	}
	return c, nil
}

func (c *client) sendMessage(message, slackId, ts string) (resp slackRespons, err error) {
	return resp, c.PostAuth(c.baseurl+"/api/chat.postMessage", slackMessage{
		SlackId: slackId,
		TS:      ts,
		Text:    ":ghost:" + message,
		Pinned:  false,
	}, &resp)
}

func (c *client) SendFile(channel, message string, file []byte) (resp slackRespons, err error) {
	return resp, c.PostFormAuth(c.baseurl+"/api/files.upload", slackFile{
		Channels: []string{channel},
		Text:     message,
		File:     file,
		Name:     fmt.Sprintf("jenkins_cantara_%s.png", time.Now().UTC().Format("2006-01-02T15:04:05")),
	}, &resp)
}

func (c *client) PostAuth(uri string, data interface{}, out interface{}) (err error) {
	jsonValue, _ := json.Marshal(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	resp, err := client.Do(req)
	if err != nil || out == nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		log.WithError(err).Warning(fmt.Sprintf("%s\t%s", body, data))
	}
	return
}

func (c *client) PostFormAuth(uri string, file slackFile, out interface{}) (err error) {
	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	f, err := mp.CreateFormFile("file", file.Name)
	if err != nil {
		return
	}
	f.Write(file.File)
	mp.Close()
	data := url.Values{}
	data.Set("channels", strings.Join(file.Channels, ","))
	data.Set("initial_comment", file.Text)
	req, err := http.NewRequest("POST", uri, &b)
	req.Header.Set("Content-Type", mp.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.URL.RawQuery = data.Encode()
	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil || out == nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		log.WithError(err).Warning(fmt.Sprintf("%s\t%s", body, data))
	}
	return
}
