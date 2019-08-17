package rak811

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
	"github.com/pkg/errors"
	"github.com/tarm/serial"
)

// ErrTimeout is returns when the serial port doesnt return any data
// within the test ReadTimeout
var ErrTimeout = errors.New("no response within the set timeout")

type config func(*serial.Config)

type Lora struct {
	port io.ReadWriteCloser
}

func New(conf *serial.Config) (*Lora, error) {
	defaultConfig := &serial.Config{
		Name:        "/dev/serial0",
		Baud:        115200,
		ReadTimeout: 1500 * time.Millisecond,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		Size:        8,
	}

	newConfig(conf)(defaultConfig)

	port, err := serial.OpenPort(defaultConfig)
	if err != nil {
		return nil, err
	}

	return &Lora{
		port: port,
	}, nil
}

func (l *Lora) tx(cmd string) (string, error) {
	if _, err := l.port.Write(createCmd(cmd)); err != nil {
		log.Printf("failed to write command %s", cmd)
	}

	// The response might contain more than a single line so read up to the OK string.
	scanner := bufio.NewScanner(bufio.NewReader(l.port))
	var resp string
	for scanner.Scan() {
		t := scanner.Text()
		resp += t
		if strings.HasPrefix(t, "OK") {
			return resp, nil
		}
	}

	if err := scanner.Err(); err != nil {
		if err == io.EOF {
			return "", ErrTimeout
		}
		log.Fatalf("reading response err:%v", err)
	}
	return "", fmt.Errorf("unexpected tx response:%v cmd:%v", resp, cmd)
}

//
// System Commands
//

// Version get module version
func (l *Lora) Version() (string, error) {
	return l.tx("version")
}

// Sleep enter sleep mode
func (l *Lora) Sleep() (string, error) {
	return l.tx("sleep")
}

// Reset module or LoRaWAN stack
// 0: reset and restart module
// 1: reset LoRaWAN stack and the module will reload
// LoRa configuration from EEPROM
func (l *Lora) Reset(mode int) (string, error) {
	return l.tx(fmt.Sprintf("reset=%d", mode))
}

// HardReset the module by reseting the hat pins.
func (l *Lora) HardReset() (string, error) {
	pin, err := rpi.OpenPin(rpi.GPIO17, gpio.ModeOutput)
	if err != nil {
		return "", fmt.Errorf("error opening pin err:%v", err)
	}
	pin.Clear()
	time.Sleep(10 * time.Millisecond)
	pin.Set()
	time.Sleep(10 * time.Millisecond)

	// After the reset need to drain the response from the serial port.
	reader := bufio.NewReader(l.port)
	var resp string
	for x := 0; x < 4; x++ {
		r, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return "", ErrTimeout
			}
			return resp, fmt.Errorf("read err:%v", err)
		}
		resp += r
	}
	// The module needs some time to be able to accept new commands.
	time.Sleep(1 * time.Second)
	return resp, nil
}

// Reload set LoRaWAN and LoraP2P configurations to default
func (l *Lora) Reload() (string, error) {
	return l.tx("reload")
}

// GetMode get module mode
func (l *Lora) GetMode() (string, error) {
	return l.tx("mode")
}

// SetMode set module to work for LoRaWAN or LoraP2P mode, defaults to 0
func (l *Lora) SetMode(mode int) (string, error) {
	return l.tx(fmt.Sprintf("mode=%d", mode))
}

// GetRecvEx get RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) GetRecvEx() (string, error) {
	return l.tx("recv_ex")
}

// SetRecvEx set RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) SetRecvEx(mode int) (string, error) {
	return l.tx(fmt.Sprintf("recv_ex=%d", mode))
}

// Close the serial port.
func (l *Lora) Close() {
	if err := l.port.Close(); err != nil {
		fmt.Printf("failed closing port, %v", err)
	}
}

//
// LoRaWAN commands
//

// SetConfig set LoRaWAN configurations
func (l *Lora) SetConfig(config string) (string, error) {
	return l.tx(fmt.Sprintf("set_config=%v", config))
}

// GetConfig LoRaWAN configuration
func (l *Lora) GetConfig(key string) (string, error) {
	return l.tx(fmt.Sprintf("get_config=%s", key))
}

// GetBand LoRaWAN band region
func (l *Lora) GetBand() (string, error) {
	return l.tx("band")
}

// SetBand LoRaWAN band region
func (l *Lora) SetBand(band string) (string, error) {
	return l.tx(fmt.Sprintf("band=%s", band))
}

const (
	// JoinSuccess successfull join.
	JoinSuccess = "at+recv=3,0,0"
	// JoinFail incorrect join config parameters.
	JoinFail = "at+recv=4,0,0"
	// JoinTimeout no response from a gateway.
	JoinTimeout = "at+recv=6,0,0"
)

// JoinOTAA join the configured network in OTAA mode.
// The module doesn't accept any other command before it returns a response.
// Response: JoinSuccess, JoinFail, JoinTimeout
// Returns an error when the timeout argument expires.
// This is as a fail safe so that the caller can reset the module
// if it doesn't return anything for a long period.
func (l *Lora) JoinOTAA(timeout time.Duration) (string, error) {
	if _, err := l.tx("join=otaa"); err != nil {
		return "", err
	}
	var (
		resp  string
		err   error
		start = time.Now()
	)
	for {
		resp, err = bufio.NewReader(l.port).ReadString('\n')
		resp = strings.TrimSuffix(strings.TrimSpace(resp), "\r")
		if err == nil && resp != "" {
			switch resp {
			case JoinSuccess:
				return JoinSuccess, nil
			case JoinFail:
				return JoinFail, nil
			case JoinTimeout:
				return JoinTimeout, nil
			default:
				return "", errors.Errorf("invalid join response resp:%v", resp)
			}
		}
		if time.Since(start) > timeout {
			return "", errors.New("no response after the set timeout")
		}
	}
}

// JoinABP join the configured network in ABP mode
func (l *Lora) JoinABP() (string, error) {
	return l.tx("join=abp")
}

// Signal check the radio rssi, snr, update by latest received radio packet
func (l *Lora) Signal() (string, error) {
	return l.tx("signal")
}

// GetDataRate get next send data rate
func (l *Lora) GetDataRate() (string, error) {
	return l.tx("dr")
}

// SetDataRate set next send data rate
func (l *Lora) SetDataRate(datarate string) (string, error) {
	return l.tx(fmt.Sprintf("dr=%s", datarate))
}

// GetLinkCnt get LoRaWAN uplink and downlink counter
func (l *Lora) GetLinkCnt() (string, error) {
	return l.tx("link_cnt")
}

// SetLinkCnt set LoRaWAN uplink and downlink counter
func (l *Lora) SetLinkCnt(uplinkCnt, downlinkCnt float32) (string, error) {
	return l.tx(fmt.Sprintf("link_cnt=%f,%f", uplinkCnt, downlinkCnt))
}

// GetABPInfo
func (l *Lora) GetABPInfo() (string, error) {
	return l.tx("abp_info")
}

// Send send data to LoRaWAN network
func (l *Lora) Send(data string) (string, error) {
	if _, err := l.tx(fmt.Sprintf("send=%s", data)); err != nil {
		return "", err
	}

	// Wait for the send success event.
	resp, err := bufio.NewReader(l.port).ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "", ErrTimeout
		}
		return "", fmt.Errorf("read err:%v", err)
	}

	resp = strings.TrimSpace(resp)
	if resp != "at+recv=2,0,0" {
		return "", fmt.Errorf("unexpected ack response:%v-", resp)
	}
	return resp, nil
}

// Recv receive event and data from LoRaWAN or LoRaP2P network
func (l *Lora) Recv(data string) (string, error) {
	return l.tx(fmt.Sprintf("recv=%s", data))
}

// GetRfConfig get RF parameters
func (l *Lora) GetRfConfig() (string, error) {
	return l.tx("rf_config")
}

// SetRfConfig Set RF parameters
func (l *Lora) SetRfConfig(parameters string) (string, error) {
	return l.tx(fmt.Sprintf("rf_config=%s", parameters))
}

// Txc send LoraP2P message
func (l *Lora) Txc(parameters string) (string, error) {
	return l.tx(fmt.Sprintf("txc=%s", parameters))
}

// Rxc set module in LoraP2P receive mode
func (l *Lora) Rxc(enable int) (string, error) {
	return l.tx(fmt.Sprintf("rxc=%d", enable))
}

// TxStop stops LoraP2P TX
func (l *Lora) TxStop() (string, error) {
	return l.tx("tx_stop")
}

// Stop LoraP2P RX
func (l *Lora) RxStop() (string, error) {
	return l.tx("rx_stop")
}

//
// Radio commands
//

// GetStatus get radio statistics
func (l *Lora) GetRadioStatus() (string, error) {
	return l.tx("status")
}

// SetStatus clear radio statistics
func (l *Lora) ClearRadioStatus() (string, error) {
	return l.tx("status=0")
}

func createCmd(cmd string) []byte {
	command := fmt.Sprintf("at+%s\r\n", cmd)
	return []byte(command)
}

//
// Peripheral commands
//

// GetUART get UART configurations
func (l *Lora) GetUART() (string, error) {
	return l.tx("uart")
}

// SetUART set UART configurations
func (l *Lora) SetUART(configuration string) (string, error) {
	return l.tx(fmt.Sprintf("uart=%s", configuration))
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
