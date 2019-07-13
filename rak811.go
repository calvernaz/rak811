package rak811

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
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

func New(conf *serial.Config) (*Lora, error) {
	defaultConfig := &serial.Config{
		Name:        "/dev/serial0",
		Baud:        115200,
		ReadTimeout: 1500 * time.Millisecond,
		Parity: serial.ParityNone,
		StopBits: serial.Stop1,
		Size: 8,
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

func (l *Lora) tx(cmd string) string {
	if _, err := l.port.Write(createCmd(cmd)); err != nil {
		log.Printf("failed to write command %s", cmd)
	}
	// read line
	//buf := bytes.Buffer{}
	//data := make([]byte, 32)
	//for {
	//	n, err := l.port.Read(data)
	//	if err != nil {
	//		if err == io.EOF {
	//			break
	//		}
	//	} else {
	//		buf.Write(data[:n])
	//	}
	//}
	//result := hex.EncodeToString(buf.Bytes())
	//fmt.Printf("response: %s\n", result)
	//return result
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

func (l *Lora) HardReset() {
	pin, err := rpi.OpenPin(rpi.GPIO17, gpio.ModeOutput)
	if err != nil {
		fmt.Printf("Error opening pin! %s\n", err)
		return
	}
	pin.Clear()
	time.Sleep(10 * time.Millisecond)
	pin.Set()
	time.Sleep(10 * time.Millisecond)
}

// Reload set LoRaWAN and LoraP2P configurations to default
func (l *Lora) Reload() {
	l.tx("reload")
}

// GetMode get module mode
func (l *Lora) GetMode() {
	l.tx("mode")
}

// SetMode set module to work for LoRaWAN or LoraP2P mode, defaults to 0
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

func (l *Lora) Close() {
	_ = l.port.Close()
}

//
// LoRaWAN commands
//

// SetConfig set LoRaWAN configurations
func (l *Lora) SetConfig(config string) {
	l.tx(config)
}

// Get LoRaWAN configuration
func (l *Lora) GetConfig(key string) {
	l.tx(fmt.Sprintf("get_config=%s", key))
}

// Get LoRaWAN band region
func (l *Lora) GetBand() {
	l.tx("band")
}

// Set LoRaWAN band region
func (l *Lora) SetBand(band string) {
	l.tx(fmt.Sprintf("band=%s", band))
}

// JoinOTAA join the configured network in OTAA mode
func (l *Lora) JoinOTAA() {
	l.tx("join=otaa")
}

// JoinABP join the configured network in ABP mode
func (l *Lora) JoinABP() {
	l.tx("join=abp")
}

// Signal check the radio rssi, snr, update by latest received radio packet
func (l *Lora) Signal() {
	l.tx("signal")
}

// GetDataRate get next send data rate
func (l *Lora) GetDataRate() {
	l.tx("dr")
}

// SetDataRate set next send data rate
func (l *Lora) SetDataRate(datarate string) {
	l.tx(fmt.Sprintf("dr=%s", datarate))
}

// GetLinkCnt get LoRaWAN uplink and downlink counter
func (l *Lora) GetLinkCnt() {
	l.tx("link_cnt")
}

// SetLinkCnt set LoRaWAN uplink and downlink counter
func (l *Lora) SetLinkCnt(uplinkCnt, downlinkCnt float32) {
	l.tx(fmt.Sprintf("link_cnt=%f,%f", uplinkCnt, downlinkCnt))
}

// GetABPInfo
func (l *Lora) GetABPInfo() {
	l.tx("abp_info")
}

// Send send data to LoRaWAN network
func (l *Lora) Send(data string) {
	l.tx(fmt.Sprintf("send=%s", data))
}

// Recv receive event and data from LoRaWAN or LoRaP2P network
func (l *Lora) Recv(data string) {
	l.tx(fmt.Sprintf("recv=%s", data))
}

// GetRfConfig get RF parameters
func (l *Lora) GetRfConfig() {
	l.tx("rf_config")
}

// SetRfConfig Set RF parameters
func (l *Lora) SetRfConfig(parameters string) {
	l.tx(fmt.Sprintf("rf_config=%s", parameters))
}

// Txc send LoraP2P message
func (l *Lora) Txc(parameters string) {
	l.tx(fmt.Sprintf("txc=%s", parameters))
}

// Rxc set module in LoraP2P receive mode
func (l *Lora) Rxc(enable int) {
	l.tx(fmt.Sprintf("rxc=%d", enable))
}

// TxStop stops LoraP2P TX
func (l *Lora) TxStop() {
	l.tx("tx_stop")
}

// Stop LoraP2P RX
func (l *Lora) RxStop() {
	l.tx("rx_stop")
}

//
// Radio commands
//

// GetStatus get radio statistics
func (l *Lora) GetRadioStatus() {
	l.tx("status")
}

// SetStatus clear radio statistics
func (l *Lora) ClearRadioStatus() {
	l.tx("status=0")
}

func createCmd(cmd string) []byte {
	command := fmt.Sprintf("at+%s\r\n", cmd)
	return []byte(command)
}

//
// Peripheral commands
//

// GetUART get UART configurations
func (l *Lora) GetUART() {
	l.tx("uart")
}

// SetUART set UART configurations
func (l *Lora) SetUART(configuration string) {
	l.tx(fmt.Sprintf("uart=%s", configuration))
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
