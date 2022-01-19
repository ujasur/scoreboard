package main

import (
	"net/http"
	"time"
)

func queryKeySingular(r *http.Request, key string) string {
	keys, ok := r.URL.Query()[key]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func queryKey(r *http.Request, key string) []string {
	keys, ok := r.URL.Query()[key]
	if !ok || len(keys[0]) < 1 {
		return nil
	}
	return keys
}

type clock struct {
	offset time.Duration
}

func (c *clock) SetOffset(offset time.Duration) {
	c.offset = offset
}

func (c *clock) Now() time.Time {
	return time.Now().Add(c.offset).UTC()
}
