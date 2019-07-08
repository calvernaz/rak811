package rak811

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

// based on: https://github.com/PiSupply/rak811-python/blob/master/iotloranode.py

type config func(*serial.Config)

type command struct {
	operation string
}

type Lora struct {
	ch   <-chan *command

	port *serial.Port
}


func New(conf *serial.Config) (*Lora, error) {
	defaultConfig := &serial.Config{
		Name:        "/dev/serial0",
		Baud:        11500,
		ReadTimeout: 500 * time.Millisecond,
	}

	newConfig(conf)(defaultConfig)

	port, err := serial.OpenPort(defaultConfig)
	if err != nil {
		return nil, err
	}

	return &Lora{
		ch: make(chan *command, 10),
		port: port,
	}, nil
}

func (l *Lora) tx(cmd string) {
	command := fmt.Sprintf("at+%s\r\n", cmd)
	_, err := l.port.Write([]byte(command))
	if err != nil {
		log.Printf("failed to write command %s",command)
	}

	// read line
}

func newConfig(config *serial.Config) config {
	return func(defaultConfig *serial.Config) {
		if config.Baud > 0 {
			defaultConfig.Baud = config.Baud
		}
		if config.Name != "" {
			defaultConfig.Name = config.Name
		}
		if config.ReadTimeout > 0 {
			defaultConfig.ReadTimeout = config.ReadTimeout
		}
	}
}
