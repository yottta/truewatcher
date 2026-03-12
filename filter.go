package main

import "strings"

type filter interface {
	allowed(application) bool
}

type filterChain []filter

func (m filterChain) allowed(a application) bool {
	for _, f := range m {
		if !f.allowed(a) {
			return false
		}
	}
	return true
}

type filterFunc func(application) bool

func (f filterFunc) allowed(wr application) bool {
	return f(wr)
}

func whitelistFilter(allowed []string) filterFunc {
	if len(allowed) == 0 {
		return func(a application) bool {
			return true // Allow everything
		}
	}
	mapped := make(map[string]struct{}, len(allowed))
	for _, s := range allowed {
		mapped[strings.ToLower(s)] = struct{}{}
	}
	return func(a application) bool {
		_, ok := mapped[strings.ToLower(a.Name)]
		return ok
	}
}

func blacklistFilter(allowed []string) filterFunc {
	mapped := make(map[string]struct{}, len(allowed))
	for _, s := range allowed {
		mapped[strings.ToLower(s)] = struct{}{}
	}
	return func(a application) bool {
		_, ok := mapped[strings.ToLower(a.Name)]
		return !ok
	}
}
