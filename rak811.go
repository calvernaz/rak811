package rak811

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/uart"
	"periph.io/x/conn/v3/uart/uartreg"
	"periph.io/x/host/v3"
	_ "periph.io/x/host/v3/serial"
)

const (
	CodeArgErr        = -1
	CodeArgNotFind    = -2
	CodeJoinAbpErr    = -3
	CodeJoinOtaaErr   = -4
	CodeNotJoin       = -5
	CodeMacBusyErr    = -6
	CodeTxErr         = -7
	CodeInterErr      = -8
	CodeWrCfgErr      = -11
	CodeRdCfgErr      = -12
	CodeTxLenLimitErr = -13
	CodeUnknownErr    = -20

	StatusRecvData         = 0
	StatusTxConfirmed      = 1
	StatusTxUnconfirmed    = 2
	StatusJoinedSuccess    = 3
	StatusJoinedFailed     = 4
	StatusTxTimeout        = 5
	StatusRx2Timeout       = 6
	StatusDownlinkRepeated = 7
	StatusWakeUp           = 8
	StatusP2pComplete      = 9
	StatusUnknown          = 100

	// event response prefix
	eventRespPrefix = "at+recv="

	// JoinSuccess successful join.
	JoinSuccess = "at+recv=3,0,0"
	// JoinFail incorrect join config parameters.
	JoinFail = "at+recv=4,0,0"
	// JoinTimeout no response from a gateway.
	JoinTimeout = "at+recv=6,0,0"

	CR_LF = "\r\n"

	OK    = "OK"
	ERROR = "ERROR"
)

// LoraError type that describes the LoRaWAN errors
type LoraError struct {
	code int
	desc string
}

// Code returns the error code
func (e *LoraError) Code() int {
	return e.code
}

// Error returns the error description
func (e *LoraError) Error() string {
	return e.desc
}

// WhichError translates an error string to a LoraError
func WhichError(error string) *LoraError {
	if !strings.HasPrefix(error, ERROR) {
		return nil
	}

	// TODO: turn error strings into constants
	errCode := strings.TrimPrefix(error, ERROR)
	switch errCode {
	case "-1":
		return whichError(CodeArgErr, "invalid argument")
	case "-2":
		return whichError(CodeArgNotFind, "argument is not available")
	case "-3":
		return whichError(CodeJoinAbpErr, "can't join network using ABP")
	case "-4":
		return whichError(CodeJoinOtaaErr, "can't join network using OTAA")
	case "-5":
		return whichError(CodeNotJoin, "can't send packet, failed to join network")
	case "-6":
		return whichError(CodeMacBusyErr, "can't send packet, busy channel")
	case "-7":
		return whichError(CodeTxErr, "can't send packet, transmission error")
	case "-8":
		return whichError(CodeInterErr, "")
	case "-11":
		return whichError(CodeWrCfgErr, "configuration write error")
	case "-12":
		return whichError(CodeRdCfgErr, "configuration read error")
	case "-13":
		return whichError(CodeTxLenLimitErr, "transmission length limit error")
	case "-20":
		return whichError(CodeUnknownErr, "unknown error")
	}
	return nil
}

type EventResponse struct {
	code int
	desc string
}

func (e *EventResponse) Code() int {
	return e.code
}

func (e *EventResponse) Description() string {
	return e.desc
}

func WhichEventResponse(resp string) *EventResponse {
	if !strings.HasPrefix(resp, "at+recv=") {
		return nil
	}

	evt := strings.TrimPrefix(resp, eventRespPrefix)
	status := strings.Split(evt, ",")[0]
	switch status {
	case "0":
		return whichEventResponse(StatusRecvData, "received data from server or P2P")
	case "1":
		return whichEventResponse(StatusTxConfirmed, "transmission succeeded and received ACK from server")
	case "2":
		return whichEventResponse(StatusTxUnconfirmed, "transmission succeeded")
	case "3":
		return whichEventResponse(StatusJoinedSuccess, "join network procedure was successful")
	case "4":
		return whichEventResponse(StatusJoinedFailed, "join network procedure failed")
	case "5":
		return whichEventResponse(StatusTxTimeout, "transmission timeout")
	case "6":
		return whichEventResponse(StatusRx2Timeout, "join network procedure timeout, no response from the gateway")
	case "7":
		return whichEventResponse(StatusDownlinkRepeated, "downlink repeated")
	case "8":
		return whichEventResponse(StatusWakeUp, "module is awake")
	case "9":
		return whichEventResponse(StatusP2pComplete, "lora P2P continues transmission has completed")
	case "100":
		return whichEventResponse(StatusUnknown, "unknown status")
	}
	return nil
}

type StopBits int8
type Parity byte

const (
	Stop1     StopBits = 1
	Stop1Half StopBits = 15
	Stop2     StopBits = 2
)

const (
	ParityNone  Parity = 'N'
	ParityOdd   Parity = 'O'
	ParityEven  Parity = 'E'
	ParityMark  Parity = 'M' // parity bit is always 1
	ParitySpace Parity = 'S' // parity bit is always 0
)

type extraConfig struct {
	debug bool
}

type Config struct {
	Name     string
	Baud     int64
	Parity   Parity
	StopBits StopBits
	Size     int
}

type config func(*Config)

type Lora struct {
	conn   conn.Conn
	config *extraConfig
	port   uart.PortCloser
}

func New(conf *Config) (*Lora, error) {
	defaultConfig := &Config{
		Name:     "/dev/ttyAMA0",
		Baud:     115200,
		Parity:   ParityNone,
		StopBits: Stop1,
		Size:     8,
	}

	newConfig(conf)(defaultConfig)

	// Load all the drivers:
	_, err := host.Init()
	if err != nil {
		return nil, err
	}

	p, err := uartreg.Open(defaultConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("open: %q", err)
	}
	c, err := p.Connect(
		physic.Frequency(defaultConfig.Baud)*physic.Hertz,
		uart.Stop(defaultConfig.StopBits),
		uart.Parity(defaultConfig.Parity),
		uart.NoFlow,
		defaultConfig.Size,
	)
	if err != nil {
		return nil, err
	}

	return &Lora{
		port:   p,
		conn:   c,
		config: &extraConfig{
			debug: false,
		},
	}, nil
}

func (l *Lora) tx(s string, fn func([]byte) (string, error)) (string, error) {
	cmd := createCmd(s)
	_, err := l.conn.(io.Writer).Write(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to write command %q with: %v", cmd, err)
	}
	debug(l, "tx: write:", string(cmd))

	buf := bytes.Buffer{}
	r, err := buf.ReadFrom(l.conn.(io.Reader))
	if err != nil {
		return "", fmt.Errorf("failed to read response from %q: %v", cmd, err)
	}
	debug(l, "tx: %d read: %v", r, buf.Bytes())

	return fn(buf.Bytes())
}

func debug(l *Lora, format string, a ...interface{}) {
	if l.config.debug {
		if a != nil {
			fmt.Println(strings.TrimSuffix(fmt.Sprintf(format, a), CR_LF))
		} else {
			fmt.Printf(strings.TrimSuffix(format, CR_LF))
		}
	}
}

//
// System Commands
//

// Version get module version
func (l *Lora) Version() (string, error) {
	return l.tx("version", readline)
}

// Sleep enter sleep mode
func (l *Lora) Sleep() (string, error) {
	return l.tx("sleep", readline)
}

// Debug set debug mode on or off
func (l *Lora) Debug(mode bool) {
	l.config.debug = mode
}

// Reset module or LoRaWAN stack
// 0: reset and restart module
// 1: reset LoRaWAN stack and the module will reload
// LoRa configuration from EEPROM
func (l *Lora) Reset(mode int) (string, error) {
	return l.tx(fmt.Sprintf("reset=%d", mode), readline)
}

// HardReset the module by reseting the hat pins.
func (l *Lora) HardReset() (string, error) {
	pin := gpioreg.ByName("GPIO17")

	if err := pin.Out(gpio.Low); err != nil {
		return "", err
	}
	time.Sleep(10 * time.Millisecond)

	if err := pin.Out(gpio.High); err != nil {
		return "", err
	}
	time.Sleep(2000 * time.Millisecond)

	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(l.conn.(io.Reader))
	if err != nil {
		return "", fmt.Errorf("failed reading response: %v", err)
	}
	return buf.String(), nil
}

// Reload set LoRaWAN and LoraP2P configurations to default
func (l *Lora) Reload() (string, error) {
	return l.tx("reload", readline)
}

// GetMode get module mode
func (l *Lora) GetMode() (string, error) {
	return l.tx("mode", readline)
}

// SetMode set module to work for LoRaWAN or LoraP2P mode, defaults to 0
func (l *Lora) SetMode(mode int) (string, error) {
	return l.tx(fmt.Sprintf("mode=%d", mode), readline)
}

// GetRecvEx get RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) GetRecvEx() (string, error) {
	return l.tx("recv_ex", readline)
}

// SetRecvEx set RSSI & SNR report on receive flag (Enabled/Disabled).
func (l *Lora) SetRecvEx(mode int) (string, error) {
	return l.tx(fmt.Sprintf("recv_ex=%d", mode), readline)
}

// Close the serial conn.
func (l *Lora) Close() {
	if err := l.port.Close(); err != nil {
		fmt.Printf("failed closing conn: %v", err)
	}
}

//
// LoRaWAN commands
//

// SetConfig set LoRaWAN configurations
func (l *Lora) SetConfig(config string) (string, error) {
	return l.tx(fmt.Sprintf("set_config=%v", config), readline)
}

// GetConfig LoRaWAN configuration
func (l *Lora) GetConfig(key string) (string, error) {
	return l.tx(fmt.Sprintf("get_config=%s", key), readline)
}

// GetBand LoRaWAN band region
func (l *Lora) GetBand() (string, error) {
	return l.tx("band", readline)
}

// SetBand LoRaWAN band region
func (l *Lora) SetBand(band string) (string, error) {
	return l.tx(fmt.Sprintf("band=%s", band), readline)
}

// JoinOTAA join the configured network in OTAA mode.
// The module doesn't accept any other command before it returns a response.
// Response: JoinSuccess, JoinFail, JoinTimeout
func (l *Lora) JoinOTAA() (string, error) {
	return l.tx("join=otaa", func(b []byte) (string, error) {
		scanner := bufio.NewScanner(bytes.NewReader(b))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := scanner.Text()
			switch line {
			case OK:
				continue
			case JoinSuccess:
				return JoinSuccess, nil
			case JoinFail:
				return JoinFail, nil
			case JoinTimeout:
				return JoinTimeout, nil
			default:
				return line, fmt.Errorf("invalid join response line: %v", line)
			}
		}

		err := scanner.Err()
		if err != nil {
			return "", err
		}
		return "", nil
	})
}

// JoinABP join the configured network in ABP mode
func (l *Lora) JoinABP() (string, error) {
	return l.tx("join=abp", readline)
}

// Signal check the radio rssi, snr, update by latest received radio packet
func (l *Lora) Signal() (string, error) {
	return l.tx("signal", readline)
}

// GetDataRate get next send data rate
func (l *Lora) GetDataRate() (string, error) {
	return l.tx("dr", readline)
}

// SetDataRate set next send data rate
func (l *Lora) SetDataRate(datarate string) (string, error) {
	return l.tx(fmt.Sprintf("dr=%s", datarate), readline)
}

// GetLinkCnt get LoRaWAN uplink and downlink counter
func (l *Lora) GetLinkCnt() (string, error) {
	return l.tx("link_cnt", readline)
}

// SetLinkCnt set LoRaWAN uplink and downlink counter
func (l *Lora) SetLinkCnt(uplinkCnt, downlinkCnt float32) (string, error) {
	return l.tx(fmt.Sprintf("link_cnt=%f,%f", uplinkCnt, downlinkCnt), readline)
}

// GetABPInfo
func (l *Lora) GetABPInfo() (string, error) {
	return l.tx("abp_info", readline)
}

// Send sends data to LoRaWAN network, returns the event response
func (l *Lora) Send(data string) (string, error) {
	return l.tx(fmt.Sprintf("send=%s", data), func(b []byte) (string, error) {
		scanner := bufio.NewScanner(bytes.NewReader(b))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case line == OK:
				continue
			case line != "":
				return line, nil
			}
		}

		err := scanner.Err()
		if err != nil {
			return "", err
		}
		return "", nil
	})
}

// Recv receive event and data from LoRaWAN or LoRaP2P network
func (l *Lora) Recv(data string) (string, error) {
	return l.tx(fmt.Sprintf("recv=%s", data), readline)
}

// GetRfConfig get RF parameters
func (l *Lora) GetRfConfig() (string, error) {
	return l.tx("rf_config", readline)
}

// SetRfConfig Set RF parameters
func (l *Lora) SetRfConfig(parameters string) (string, error) {
	return l.tx(fmt.Sprintf("rf_config=%s", parameters), readline)
}

// Txc send LoraP2P message
func (l *Lora) Txc(parameters string) (string, error) {
	return l.tx(fmt.Sprintf("txc=%s", parameters), readline)
}

// Rxc set module in LoraP2P receive mode
func (l *Lora) Rxc(enable int) (string, error) {
	return l.tx(fmt.Sprintf("rxc=%d", enable), readline)
}

// TxStop stops LoraP2P TX
func (l *Lora) TxStop() (string, error) {
	return l.tx("tx_stop", readline)
}

// RxStop LoraP2P RX
func (l *Lora) RxStop() (string, error) {
	return l.tx("rx_stop", readline)
}

//
// Radio commands
//

// GetStatus get radio statistics
func (l *Lora) GetRadioStatus() (string, error) {
	return l.tx("status", readline)
}

// SetStatus clear radio statistics
func (l *Lora) ClearRadioStatus() (string, error) {
	return l.tx("status=0", readline)
}

func createCmd(cmd string) []byte {
	return []byte(fmt.Sprintf("at+%v\r\n", cmd))
}

//
// Peripheral commands
//

// GetUART get UART configurations
func (l *Lora) GetUART() (string, error) {
	return l.tx("uart", readline)
}

// SetUART set UART configurations
func (l *Lora) SetUART(configuration string) (string, error) {
	return l.tx(fmt.Sprintf("uart=%s", configuration), readline)
}

func readline(b []byte) (string, error) {
	fmt.Printf("readline %v\n", string(b))
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if isOk(line) {
			return line, nil
		}

		if err := isError(line); err != nil {
			return "", err
		}
		continue
	}

	err := scanner.Err()
	if err != nil {
		return "", err
	}
	return "", nil
}

func newConfig(config *Config) config {
	return func(defaultConfig *Config) {
		if config.Baud > 0 {
			defaultConfig.Baud = config.Baud
		}
		if config.Name != "" {
			defaultConfig.Name = config.Name
		}
	}
}

func isOk(msg string) bool {
	if strings.HasPrefix(msg, OK) {
		return true
	}
	return false
}

func isError(msg string) error {
	if strings.HasPrefix(msg, ERROR) {
		return errors.New(msg)
	}
	return nil
}

func whichError(code int, desc string) *LoraError {
	return &LoraError{
		code: code,
		desc: desc,
	}
}

func whichEventResponse(statusCode int, desc string) *EventResponse {
	return &EventResponse{
		statusCode,
		desc,
	}
}
