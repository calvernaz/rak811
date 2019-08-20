package rak811

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/davecheney/gpio"
	"github.com/davecheney/gpio/rpi"
	"github.com/tarm/serial"
)

const OK = "OK"

type config func(*serial.Config)

type Lora struct {
	port io.ReadWriteCloser
	// Returns an error when the global timeout argument expires.
	// This is as a fail safe so that the caller can reset the module
	// if it doesn't return anything for a long period.
	timeout time.Duration
}

// New sets the lora module configuration.
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
		port:    port,
		timeout: defaultConfig.ReadTimeout,
	}, nil
}

func (l *Lora) txr(cmd string, lines int) (string, error) {
	if _, err := l.port.Write(createCmd(cmd)); err != nil {
		return "", fmt.Errorf("failed to write command %q with: %s", cmd, err)
	}
	return l.tr(lines)
}

func (l *Lora) tr(lines int) (string, error) {
	reader := bufio.NewReader(l.port)
	var resp string
	start := time.Now()
	for line := 0; line < lines; line++ {
		for {
			r, err := reader.ReadString('\n')

			if err != nil {
				if err == io.EOF { // The serial port has a max timeout of 25sec so we rely on the l.timeout.
					continue
				}
				return "", fmt.Wrap(err, "failed read")
			}
			if strings.HasPrefix(r, "ERROR") {
				return "", fmt.Error(r)
			}
			resp += r
			if time.Since(start) > l.timeout {
				return "", fmt.Errorf("no response within:%v", l.timeout)
			}
			break
		}
	}
	return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(resp), "\r"), "\n"), nil
}

//
// System Commands
//

// Version get module version.
func (l *Lora) Version() (string, error) {
	return l.txr("version", 1)
}

// Sleep enter sleep mode.
func (l *Lora) Sleep() (string, error) {
	return l.txr("sleep", 1)
}

// Wakeup to wake up the module.
func (l *Lora) Wakeup() (string, error) {
	return l.txr("wake", 1) // Any command will wake it up.
}

// Reset module or LoRaWAN stack.
// 0: reset and restart module,
// 1: reset LoRaWAN stack and the module will reload
// LoRa configuration from EEPROM.
func (l *Lora) Reset(mode int) (string, error) {
	return l.txr(fmt.Sprintf("reset=%d", mode), 4)
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
	time.Sleep(2000 * time.Millisecond)
	pin.Close()

	return l.tr(4)
}

// Reload set LoRaWAN and LoraP2P configurations to default
func (l *Lora) Reload() (string, error) {
	return l.txr("reload", 4)
}

// GetMode get module mode
func (l *Lora) GetMode() (string, error) {
	return l.txr("mode", 1)
}

// SetMode set module to work for LoRaWAN or LoraP2P mode, defaults to 0
func (l *Lora) SetMode(mode int) (string, error) {
	resp, err := l.txr(fmt.Sprintf("mode=%d", mode), 4)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(resp, "OK") {
		return "", fmt.Errorf("unexpected response:", resp)
	}
	return resp, nil
}

// GetRecvEx get RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) GetRecvEx() (string, error) {
	return l.txr("recv_ex", 1)
}

// SetRecvEx set RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) SetRecvEx(mode int) (string, error) {
	return l.txr(fmt.Sprintf("recv_ex=%d", mode), 1)
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

// SetConfig set LoRaWAN configurations.
func (l *Lora) SetConfig(config string) (string, error) {
	return l.txr(fmt.Sprintf("set_config=%v", config), 1)
}

// GetConfig LoRaWAN configuration.
func (l *Lora) GetConfig(key string) (string, error) {
	return l.txr(fmt.Sprintf("get_config=%s", key), 1)
}

// GetBand LoRaWAN band region
func (l *Lora) GetBand() (string, error) {
	return l.txr("band", 1)
}

// SetBand LoRaWAN band region.
func (l *Lora) SetBand(band string) (string, error) {
	return l.txr(fmt.Sprintf("band=%s", band), 1)
}

const (
	STATUS_RECV_DATA         = "at+recv=0,0,0"
	STATUS_TX_COMFIRMED      = "at+recv=1,0,0"
	STATUS_TX_UNCOMFIRMED    = "at+recv=2,0,0"
	STATUS_JOINED_SUCCESS    = "at+recv=3,0,0"
	STATUS_JOINED_FAILED     = "at+recv=4,0,0"
	STATUS_TX_TIMEOUT        = "at+recv=5,0,0"
	STATUS_RX2_TIMEOUT       = "at+recv=6,0,0"
	STATUS_DOWNLINK_REPEATED = "at+recv=7,0,0"
	STATUS_WAKE_UP           = "at+recv=8,0,0"
	STATUS_P2PTX_COMPLETE    = "at+recv=9,0,0"
	STATUS_UNKNOWN           = "at+recv=100,0,0"
)

// JoinOTAA join the configured network in OTAA mode.
// The parent function needs to validate the response.
// Valid responses:  STATUS_JOINED_SUCCESS, STATUS_JOINED_FAILED, STATUS_TX_TIMEOUT
func (l *Lora) JoinOTAA() (string, error) {
	resp, err := l.txr("join=otaa", 1)
	if err != nil {
		return "", err
	}
	if resp != OK {
		return "", fmt.New(resp) // Convert the resp to an error so that the caller handle it properly.
	}
	// The module doesn't accept any other command before it returns a response
	// so need to wait for it.
	return l.tr(1)
}

// JoinABP join the configured network in ABP mode.
func (l *Lora) JoinABP() (string, error) {
	return l.txr("join=abp", 1)
}

// Signal check the radio rssi, snr, update by latest received radio packet.
func (l *Lora) Signal() (string, error) {
	return l.txr("signal", 1)
}

// GetDataRate get next send data rate.
func (l *Lora) GetDataRate() (string, error) {
	return l.txr("dr", 1)
}

// SetDataRate set next send data rate.
func (l *Lora) SetDataRate(datarate string) (string, error) {
	return l.txr(fmt.Sprintf("dr=%s", datarate), 1)
}

// GetLinkCnt get LoRaWAN uplink and downlink counter.
func (l *Lora) GetLinkCnt() (string, error) {
	return l.txr("link_cnt", 1)
}

// SetLinkCnt set LoRaWAN uplink and downlink counter.
func (l *Lora) SetLinkCnt(uplinkCnt, downlinkCnt float32) (string, error) {
	return l.txr(fmt.Sprintf("link_cnt=%f,%f", uplinkCnt, downlinkCnt), 1)
}

// GetABPInfo
func (l *Lora) GetABPInfo() (string, error) {
	return l.txr("abp_info", 1)
}

// Send send data to LoRaWAN network.
// The parent function should check for:
// STATUS_TX_UNCOMFIRMED when sending unconfirmed packets,
// STATUS_TX_COMFIRMED when sending confirmed packets.
func (l *Lora) Send(data string) (string, error) {
	resp, err := l.txr(fmt.Sprintf("send=%s", data), 1)
	if err != nil {
		return "", err
	}
	if resp != OK {
		return "", fmt.New(resp) // Convert the resp to an error so that the caller handle it properly.
	}
	// The module doesn't accept any other command before it returns a response
	// so need to wait for it.
	return l.tr(1)
}

// Recv receive event and data from LoRaWAN or LoRaP2P network.
func (l *Lora) Recv(data string) (string, error) {
	return l.txr(fmt.Sprintf("recv=%s", data), 1)
}

// GetRfConfig get RF parameters.
func (l *Lora) GetRfConfig() (string, error) {
	return l.txr("rf_config", 1)
}

// SetRfConfig Set RF parameters
func (l *Lora) SetRfConfig(parameters string) (string, error) {
	return l.txr(fmt.Sprintf("rf_config=%s", parameters), 1)
}

// Txc send LoraP2P message
func (l *Lora) Txc(parameters string) (string, error) {
	return l.txr(fmt.Sprintf("txc=%s", parameters), 1)
}

// Rxc set module in LoraP2P receive mode.
func (l *Lora) Rxc(enable int) (string, error) {
	return l.txr(fmt.Sprintf("rxc=%d", enable), 1)
}

// TxStop stops LoraP2P TX.
func (l *Lora) TxStop() (string, error) {
	return l.txr("tx_stop", 1)
}

// RxStop LoraP2P RX.
func (l *Lora) RxStop() (string, error) {
	return l.txr("rx_stop", 1)
}

//
// Radio commands
//

// GetRadioStatus get radio statistics.
func (l *Lora) GetRadioStatus() (string, error) {
	return l.txr("status", 1)
}

// ClearRadioStatus clear radio statistics.
func (l *Lora) ClearRadioStatus() (string, error) {
	return l.txr("status=0", 1)
}

func createCmd(cmd string) []byte {
	command := fmt.Sprintf("at+%s\r\n", cmd)
	return []byte(command)
}

//
// Peripheral commands
//

// GetUART get UART configurations.
func (l *Lora) GetUART() (string, error) {
	return l.txr("uart", 1)
}

// SetUART set UART configurations.
func (l *Lora) SetUART(configuration string) (string, error) {
	return l.txr(fmt.Sprintf("uart=%s", configuration), 1)
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
