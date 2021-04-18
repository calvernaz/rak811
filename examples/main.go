package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/calvernaz/rak811"
	"periph.io/x/conn/v3/uart"
	"periph.io/x/conn/v3/uart/uartreg"
)

func main() {
	cfg := &rak811.Config{}

	lora, err := rak811.New(cfg)
	if err != nil {
		log.Fatal("failed to create rak811 instance: ", err)
	}

	fmt.Print("UART ports available:\n")
	for _, ref := range uartreg.All() {
		fmt.Printf("- %s\n", ref.Name)
		if ref.Number != -1 {
			fmt.Printf("  %d\n", ref.Number)
		}
		if len(ref.Aliases) != 0 {
			fmt.Printf("  %s\n", strings.Join(ref.Aliases, " "))
		}

		b, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v", err)
		}
		if p, ok := b.(uart.Pins); ok {
			fmt.Printf("  RX : %s", p.RX())
			fmt.Printf("  TX : %s", p.TX())
			fmt.Printf("  RTS: %s", p.RTS())
			fmt.Printf("  CTS: %s", p.CTS())
		}
		if err := b.Close(); err != nil {
			fmt.Printf("  Failed to close: %v", err)
		}
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

	resp, err = lora.SetConfig(fmt.Sprintf("app_key:%v", "4ca2801aa3adf8add26e149bf8a0d440"))
	if err != nil {
		log.Fatal("appkey err: ", err)
	}
	fmt.Printf("set appkey: %v\n", resp)

	resp, err = lora.JoinOTAA()
	if err != nil {
		log.Fatal("failed to join: ", err)
	}
	fmt.Printf("Join: %s\n", resp)

	// at+send=0,2,010203040506 /*APP port:2, unconfirmed message*/
	// at+recv=2,0,0
	resp, err = lora.Send("0,2,010203040506")
	if err != nil {
		log.Fatal("failed to send: ", err)
	}
	fmt.Printf("Send tx success: %s\n", resp)

	// at+send=1,2,010203040506 /*APP port :2, confirmed message*/
	// at+recv=1,0,0
	resp, err = lora.Send("1,2,010203040506")
	if err != nil {
		log.Fatal("failed to send: ", err)
	}
	fmt.Printf("Send acknowledge by gateway: %s\n", resp)

}
