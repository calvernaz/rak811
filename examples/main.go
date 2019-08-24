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

	resp, err = lora.SetConfig(fmt.Sprintf("dev_eui:%v", resp[2:]))
	if err != nil {
		log.Fatal("deveui err: ", err)
	}
	fmt.Printf("set devui: %v\n", resp)

	resp, err = lora.SetConfig(fmt.Sprintf("app_eui:%v", "0102030405060708"))
	if err != nil {
		log.Fatal("appeui err: ", err)
	}
	fmt.Printf("set appeui: %v\n", resp)

	resp, err = lora.SetConfig(fmt.Sprintf("app_key:%v", "xxx2801aa3adf8add26e149xxxxxxxxx"))
	if err != nil {
		log.Fatal("appkey err: ", err)
	}
	fmt.Printf("set appkey: %v\n", resp)

	resp, err = lora.JoinOTAA()
	if err != nil {
		log.Fatal("failed to join: ", err)
	}
	fmt.Printf("Join: %s\n", resp)

	data := fmt.Sprintf("%x","hello world")
	resp, err = lora.Send(fmt.Sprintf("1,1,%s", data))
	if err != nil {
		log.Fatal("failed to send: ", err)
	}
	fmt.Printf("Send: %s\n", resp)

}
