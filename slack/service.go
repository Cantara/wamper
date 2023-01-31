package slack

import (
	"context"
	log "github.com/cantara/bragi"
	scheduletasks "github.com/cantara/gober/scheduletasks"
	"github.com/cantara/gober/stream"
	"github.com/cantara/gober/webserver/health"
	"github.com/cantara/wamper/screenshot"
	"github.com/cantara/wamper/sites"
	"time"
)

func cryptKeyProvider(key string) func(_ string) string {
	return func(_ string) string {
		return key
		//os.Getenv("screenshot.task.key")
	}
}

type Task struct {
	Site         sites.Site    `json:"site"`
	Time         time.Time     `json:"time"`
	Interval     time.Duration `json:"interval"`
	SlackToken   string        `json:"slack_token"`
	SlackChannel string        `json:"slack_channel"`
}

type Service interface {
	Set(task Task) error
}

type service struct {
	schedule    scheduletasks.Tasks[Task]
	screenshots screenshot.Store
}

func Init(s stream.Stream, scr screenshot.Store, cryptoKey string, ctx context.Context) (out Service, err error) {
	ser := service{
		screenshots: scr,
	}
	//t, err := tasks.Init[Task](s, "slack_task", "", cryptKeyProvider(cryptoKey), ctx)
	tas, err := scheduletasks.Init[Task](s, "slack_task", "1.0.0", cryptKeyProvider(cryptoKey), ser.executeTask, ctx)
	if err != nil {
		return
	}
	ser.schedule = tas
	out = &ser
	return
}

func (s *service) Set(t Task) error {
	return s.schedule.Create(t.Time, t.Interval, t)
}

func (s *service) executeTask(t Task) bool {
	scr, err := s.screenshots.Get(t.Site.Id())
	if err != nil {
		log.AddError(err).Error("while getting screenshot during scheduled task execution")
		return false
	}
	slack, err := NewClient(t.SlackToken)
	if err != nil {
		log.AddError(err).Error("while creating slack client during scheduled task execution")
		return false
	}
	r, err := slack.SendFile(t.SlackChannel, "Today's Jenkins build status for "+t.Site.Name+"! From: "+health.GetOutboundIP().String(), scr.Buf)
	if err != nil {
		log.AddError(err).Error("while posting slack message during scheduled task execution")
		return false
	}
	log.Debug(r)
	return true
}
