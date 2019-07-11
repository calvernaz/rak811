package rak811

import (
	"fmt"
	"log"

	"github.com/tarm/serial"
)

// based on: https://github.com/PiSupply/rak811-python/blob/master/iotloranode.py
// base specification: https://github.com/RAKWireless/RAK811/blob/master/Software%20Development/RAK811%C2%A0Lora%C2%A0AT%C2%A0Command%C2%A0V1.4.pdf

type config func(*serial.Config)

type command struct {
	operation string
}

type Lora struct {
	ch <-chan *command

	port *serial.Port
}

func (l *Lora) tx(cmd string) {
	_, err := l.port.Write(createCmd(cmd))
	if err != nil {
		log.Printf("failed to write createCmd %s", cmd)
	}
	// read line

}

//
// System Commands
//

// Version get module version
func (l *Lora) Version() {
	l.tx("version")
}

// Sleep enter sleep mode
func (l *Lora) Sleep() {
	l.tx("sleep")
}

// Reset module or LoRaWAN stack
// 0: reset and restart module
// 1: reset LoRaWAN stack and the module will reload
// LoRa configuration from EEPROM
func (l *Lora) Reset(mode int) {
	l.tx(fmt.Sprintf("reset=%d", mode))
}

// Reload set LoRaWAN and LoraP2P configurations to default
func (l *Lora) Reload() {
	l.tx("reload")
}

// GetMode get module mode
func (l *Lora) GetMode() {
	l.tx("mode")
}

// SetMode set module to wprk for LoRaWAN or LoraP2P mode, defaults to 0
func (l *Lora) SetMode(mode int) {
	l.tx(fmt.Sprintf("mode=%d", mode))
}

// GetRecvEx get RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) GetRecvEx() {
	l.tx("recv_ex")
}

// SetRecvEx set RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) SetRecvEx(mode int) {
	l.tx(fmt.Sprintf("recv_ex=%d", mode))
}

// Join the configured network in OTAA mode
func (l *Lora) JoinOTAA() {
	l.tx("join=otaa")
}

// Join the configured network in ABP mode
func (l *Lora) JoinABP() {
	l.tx("join=abp")
}

func createCmd(cmd string) []byte {
	command := fmt.Sprintf("at+%s\r\n", cmd)
	return []byte(command)
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
