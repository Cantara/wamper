package screenshot

import (
	"context"
	"github.com/cantara/gober/persistentbigdata"
	"github.com/cantara/gober/stream"
	"github.com/cantara/gober/webserver"
	"net/url"
	"time"
)

type Store interface {
	Set(Screenshot) error
	Get(string) (Screenshot, error)
}

type screenshotMeta struct {
	Name      string    `json:"name"`
	Url       url.URL   `json:"url"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type store struct {
	screenshots persistentbigmap.EventMap[Screenshot, screenshotMeta]
}

func InitStore(serv *webserver.Server, s stream.Stream, cryptoKey string, ctx context.Context) (out Store, err error) {
	screenshots, err := persistentbigmap.Init[Screenshot, screenshotMeta](serv, s, "screenshot", "0.1.0", func(key string) string {
		return cryptoKey
	}, func(s screenshotMeta) string {
		return s.Name //If a true id is used here. The screenshot database will just continuously increase in size without providing any real value
	}, ctx)
	if err != nil {
		return
	}
	out = &store{
		screenshots: screenshots,
	}
	return
}

func (s *store) Set(scr Screenshot) (err error) {
	err = s.screenshots.Set(scr, screenshotMeta{
		Name:      scr.Name,
		Url:       scr.Url,
		Type:      scr.Type,
		CreatedAt: scr.CreatedAt,
	})
	return
}

func (s *store) Get(name string) (scr Screenshot, err error) {
	return s.screenshots.Get(name)
}
