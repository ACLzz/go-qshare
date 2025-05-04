package goqshare

import (
	"math/rand/v2"
	"time"
)

// TODO: candidate for separate package

type clock interface {
	rand.Source
	Now() time.Time
}

type utc struct{}

func NewUTC() clock {
	return utc{}
}

func (utc) Now() time.Time {
	return time.Now().UTC()
}

func (c utc) Uint64() uint64 {
	return uint64(c.Now().Unix())
}

func NewStatic(t time.Time) clock {
	return static{
		t: t,
	}
}

type static struct {
	t time.Time
}

func (c static) Now() time.Time {
	return c.t
}

func (c static) Uint64() uint64 {
	return uint64(c.t.Unix())
}
