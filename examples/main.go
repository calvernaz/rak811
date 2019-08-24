package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/tarm/serial"

	"github.com/krasi-georgiev/rak811"
)

func main() {
	cfg := &serial.Config{
		Name: "/dev/ttyAMA0",
	}

	lora, err := rak811.New(cfg)
	if err != nil {
		log.Fatal("failed to create rak811 instance: ", err)
	}

	resp, err := lora.HardReset()
	if err != nil {
		log.Fatal("failed to reset: ", err)
	}

	resp, err = lora.Version()
	if err != nil {
		log.Fatal("failed to get version: ", err)
	}
	fmt.Printf("Version: %s\n", resp[2:])

	resp, err = lora.GetConfig("dev_eui")
	if err != nil {
		log.Fatal("failed get config: ", err)
	}
	fmt.Printf("DevEUI: %s\n", resp[2:])

	resp, err = lora.SetConfig(fmt.Sprintf("dev_eui:%v", strings.ToLower(resp[2:])))
	if err != nil {
		log.Fatal("deveui err: ", err)
	}
	fmt.Printf("set devui: %v\n", resp)

	resp, err = lora.SetConfig(fmt.Sprintf("app_eui:%v", "0000010000000000"))
	if err != nil {
		log.Fatal("appeui err: ", err)
	}
	fmt.Printf("set appeui: %v\n", resp)

	resp, err = lora.SetConfig(fmt.Sprintf("app_key:%v", "09e714accc4450047ef24a6a59ac9a97"))
	if err != nil {
		log.Fatal("appkey err: ", err)
	}
	fmt.Printf("set appkey: %v\n", resp)

	resp, err = lora.JoinOTAA()
	if err != nil {
		log.Fatal("failed to join: ", err)
	}
	fmt.Printf("Join: %s\n", resp)
}
