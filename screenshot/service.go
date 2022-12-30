package screenshot

import (
	"context"
	log "github.com/cantara/bragi"
	scheduletasks "github.com/cantara/gober/scheduletasks"
	"github.com/cantara/gober/stream"
	tasks "github.com/cantara/gober/taskssingle"
	"github.com/cantara/wamper/sites"
)

func cryptKeyProvider(key string) func(_ string) string {
	return func(_ string) string {
		return key
		//os.Getenv("screenshot.task.key")
	}
}

type Service interface {
	Set(task Task) error
}

type service struct {
	schedule scheduletasks.Tasks[sites.Site]
	store    Store
}

func Init(s stream.Stream, st Store, cryptoKey string, ctx context.Context) (out Service, err error) {
	ser := service{
		store: st,
	}
	t, err := tasks.Init[sites.Site](s, "screenshot_task", "", cryptKeyProvider(cryptoKey), ctx)
	tas, err := scheduletasks.Init[sites.Site](s, t, "screenshot_schedule_task", "1.0.0", cryptKeyProvider(cryptoKey), ser.executeTask, ctx)
	if err != nil {
		return
	}
	ser.schedule = tas
	out = &ser
	return
}

func (s *service) Set(t Task) error {
	return s.schedule.Create(t.Time, t.Interval, t.Site)
}

func (s *service) executeTask(st sites.Site) bool {
	var scr Screenshot
	var err error
	if st.Jenkins {
		scr, err = GetScreenshotJenkins(st)
		if err != nil {
			log.AddError(err).Error("while taking jenkins screenshot during scheduled task execution")
			return false
		}
	} else {
		scr, err = GetScreenshot(st)
		if err != nil {
			log.AddError(err).Error("while taking screenshot during scheduled task execution")
			return false
		}
	}
	//Could add a check to see if the new picture differs a lot from the previous one. Thus, it would be possible to add a new task to post the screenshot.
	err = s.store.Set(scr)
	if err != nil {
		log.AddError(err).Error("while storing screenshot during scheduled task execution")
		return false
	}
	return true
}
