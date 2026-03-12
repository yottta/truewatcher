package client

import "strings"

type Filter interface {
	Allowed(Application) bool
}

type FilterChain []Filter

func (m FilterChain) Allowed(a Application) bool {
	for _, f := range m {
		if !f.Allowed(a) {
			return false
		}
	}
	return true
}

type FilterFunc func(Application) bool

func (f FilterFunc) Allowed(wr Application) bool {
	return f(wr)
}

func WhitelistFilter(allowed []string) FilterFunc {
	if len(allowed) == 0 {
		return func(a Application) bool {
			return true // Allow everything
		}
	}
	mapped := make(map[string]struct{}, len(allowed))
	for _, s := range allowed {
		mapped[strings.ToLower(s)] = struct{}{}
	}
	return func(a Application) bool {
		_, ok := mapped[strings.ToLower(a.Name)]
		return ok
	}
}

func BlacklistFilter(allowed []string) FilterFunc {
	mapped := make(map[string]struct{}, len(allowed))
	for _, s := range allowed {
		mapped[strings.ToLower(s)] = struct{}{}
	}
	return func(a Application) bool {
		_, ok := mapped[strings.ToLower(a.Name)]
		return !ok
	}
}
