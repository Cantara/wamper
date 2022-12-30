package screenshot

import (
	"context"
	"github.com/cantara/gober/persistenteventmap"
	"github.com/cantara/gober/stream"
)

type Store interface {
	Set(Screenshot) error
	Get(string) (Screenshot, error)
}

type store struct {
	screenshots persistenteventmap.EventMap[Screenshot]
}

func InitStore(s stream.Stream, cryptoKey string, ctx context.Context) (out Store, err error) {
	screenshots, err := persistenteventmap.Init[Screenshot](s, "screenshot", "0.1.0", func(key string) string {
		return cryptoKey
	}, func(s Screenshot) string {
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
	err = s.screenshots.Set(scr)
	return
}

func (s *store) Get(name string) (scr Screenshot, err error) {
	return s.screenshots.Get(name)
}
