package main

import (
	"context"

	"github.com/tarm/serial"

	"github.com/calvernaz/rak811"
)

type Options struct {
	str string
	// for alternative data
	Context context.Context
}

type Option func(o *Options)

// NewConfig returns new config
func NewConfig(opts ...Option) *Config {
	return newConfig(opts...)
}

type Config struct {
	opts Options
}

func newConfig(opts ...Option) *Config {
	options := Options{
		str: "default",
		Context: nil,
	}

	for _, o := range opts {
		o(&options)
	}

	c := &Config{
		opts:     options,
	}

	return c
}

func WithSource(s context.Context) Option {
	return func(o *Options) {
		o.Context = s
	}
}


func WithString(s string) Option {
	return func(o *Options) {
		o.str = s
	}
}

func main() {
	cfg := &serial.Config{
		Name:        "/dev/ttyAMA0",
	}
	_, _ = rak811.New(cfg)

}
