package screenshot

import (
	"context"
	"time"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/gober/consensus"
	scheduletasks "github.com/cantara/gober/scheduletasks"
	"github.com/cantara/gober/stream"
	"github.com/cantara/wamper/sites"
)

type Service interface {
	Set(task Task) error
	Tasks() []scheduletasks.TaskMetadata
}

type service struct {
	schedule scheduletasks.Tasks[sites.Site]
	store    Store
}

func Init(s stream.Stream, consBuild consensus.ConsBuilderFunc, st Store, cryptoKey log.RedactedString, ctx context.Context) (out Service, err error) {
	ser := service{
		store: st,
	}
	//t, err := tasks.Init[sites.Site](s, "screenshot_task", "1.0.0", cryptKeyProvider(cryptoKey), ctx)
	tas, err := scheduletasks.Init(s, consBuild, "screenshot_schedule_task", "1.0.0", stream.StaticProvider(cryptoKey), ser.executeTask, time.Second*10, true, 1, ctx)
	if err != nil {
		return
	}
	ser.schedule = tas
	out = &ser
	return
}

func (s *service) Set(t Task) error {
	return s.schedule.Create(t.Site.Name, t.Time, t.Interval, t.Site)
}

func (s *service) Tasks() []scheduletasks.TaskMetadata {
	return s.schedule.Tasks()
}

func (s *service) executeTask(st sites.Site) bool {
	var scr Screenshot
	var err error
	scr, err = GetScreenshot(st)
	if err != nil {
		log.WithError(err).Error("while taking screenshot during scheduled task execution")
		return false
	}
	//Could add a check to see if the new picture differs a lot from the previous one. Thus, it would be possible to add a new task to post the screenshot.
	err = s.store.Set(scr)
	if err != nil {
		log.WithError(err).Error("while storing screenshot during scheduled task execution")
		return false
	}
	log.Info("took screenshot", "site", st, "url", scr.Url, "name", scr.Name, "created_at", scr.CreatedAt, "type", scr.Type)
	return true
}
