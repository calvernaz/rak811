package main

import (
	"github.com/pkg/errors"
	"github.com/tarm/serial"

	"github.com/calvernaz/rak811"
)

func main() {
	cfg := &serial.Config{
		Name:        "/dev/ttyAMA0",
	}

	lora, err := rak811.New(cfg)
	if err != nil {
		errors.Wrap(err, "failed to create rak811 instance")
	}

	lora.Version()
}
