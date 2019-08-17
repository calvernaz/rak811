package main

import (
	"fmt"
	"github.com/tarm/serial"
	"log"

	"github.com/calvernaz/rak811"
)

func main() {
	cfg := &serial.Config{
		Name:        "/dev/ttyAMA0",
	}

	lora, err := rak811.New(cfg)
	if err != nil {
		log.Fatal(err, "failed to create rak811 instance")
	}

	resp, err := lora.Version()
	if err != nil {
		log.Fatalf("failed to get version: ", err)
	}

	fmt.Println(resp)
}
