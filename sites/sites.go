package sites

import (
	"context"
	"github.com/cantara/gober/persistenteventmap"
	"github.com/cantara/gober/stream"
	"net/url"
)

type Sites interface {
	Set(Site) error
	Range(func(data Site) error)
	Get(string) (Site, error)
}

type storeService struct {
	sites persistenteventmap.EventMap[Site]
}

var cryptKey = "MdgKIHmlbRszXjLbS7pXnSBdvl+SR1bSejtpFTQXxro="

func Init(st stream.Stream, ctx context.Context) (s Sites, err error) {
	siteMap, err := persistenteventmap.Init[Site](st, "site", "0.1.0", func(key string) string {
		return cryptKey
	}, func(s Site) string {
		return s.Id()
	}, ctx)
	if err != nil {
		return
	}
	s = &storeService{
		sites: siteMap,
	}
	return
}

func (s *storeService) Set(site Site) (err error) {
	err = s.sites.Set(site)
	if err != nil {
		return
	}
	return
}

func (s *storeService) Range(f func(data Site) error) {
	s.sites.Range(func(_ string, data Site) error {
		return f(data)
	})
}

func (s *storeService) Get(id string) (o Site, err error) {
	return s.sites.Get(id)
}

type Site struct {
	Name     string  `json:"name"`
	Url      url.URL `json:"url"`
	Jenkins  bool    `json:"jenkins"`
	Username string  `json:"username"`
	Password string  `json:"password"`
}

func (s Site) Id() (out string) {
	/* Dynamic single site id
	out = s.Url.Host + s.Url.Path
	if s.Url.RawQuery != "" {
		out += "?" + s.Url.RawQuery
	}
	*/
	return s.Name
}
