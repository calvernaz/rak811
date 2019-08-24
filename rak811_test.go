package rak811

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestCreateCmd(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"mode=0", "at+mode=0\r\n"},                         /* SET LoraWAN work mode */
		{"join=otaa", "at+join=otaa\r\n"},                   /* Join OTAA type*/
		{"join=abp", "at+join=abp\r\n"},                     /* Join ABP type*/
		{"get_config=dev_eui", "at+get_config=dev_eui\r\n"}, /* GET Dev_EUI check */
		{"set_config=rx2:3,868500000", "at+set_config=rx2:3,868500000\r\n"},
		{"set_config=app_eui:39d7119f920f7952&app_key:a6b08140dae1d795ebfa5a6dee1f4dbd",
			"at+set_config=app_eui:39d7119f920f7952&app_key:a6b08140dae1d795ebfa5a6dee1f4dbd\r\n"}, /* SET LoraGateway app_eui and app_key , big endian*/
		{"recv=3,0,0", "at+recv=3,0,0\r\n"}, /* Join status success*/
		{"send=0,2,000000000000007F0000000000000000", "at+send=0,2,000000000000007F0000000000000000\r\n"}, /*APP port:2, battery level 50%, unconfirmed message*/
		{"recv=1,0,0", "at+recv=1,0,0\r\n"}, /*confirmed mean receive ack from gateway*/
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			s := createCmd(tt.in)
			if bytes.Compare(s, []byte(tt.out)) != 0 {
				t.Errorf("got %q, want %q", s, tt.out)
			}
		})
	}
}

func TestLora_Version(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK2.0.3.0"))
	lora := newTestLora(fsp)

	t.Run("get software version", func(t *testing.T) {
		actual, err := lora.Version()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(actual), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", actual, fsp.At())
		}
	})
}

func TestLora_Sleep(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK))
	lora := newTestLora(fsp)

	t.Run("module enter sleep", func(t *testing.T) {
		res, err := lora.Sleep()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Reset(t *testing.T) {
	// The module firmware is broken and returns 4 lines instead of 1 so
	// the reader needs to provide 4 lines.
	fsp := newFakeFakeSerialPort([]byte(""), []byte(""), []byte(""), []byte(OK))
	lora := newTestLora(fsp)

	t.Run("reset module", func(t *testing.T) {
		res, err := lora.Reset(0)
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Reload(t *testing.T) {
	// The module firmware is broken and returns 4 lines instead of 1 so
	// the reader needs to provide 4 lines.
	fsp := newFakeFakeSerialPort([]byte(""), []byte(""), []byte(""), []byte(OK))
	lora := newTestLora(fsp)

	t.Run("reload the default parameters", func(t *testing.T) {
		res, err := lora.Reset(0)
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_SetMode(t *testing.T) {
	// The module firmware is broken and returns 4 lines instead of 1 so
	// the reader needs to provide 4 lines.
	fsp := newFakeFakeSerialPort([]byte(""), []byte(""), []byte(""), []byte(OK))
	lora := newTestLora(fsp)

	t.Run("set module work on", func(t *testing.T) {
		res, err := lora.SetMode(0)
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_SetRecvEx(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK))
	lora := newTestLora(fsp)

	t.Run("set enable or disable rssi and snr messages", func(t *testing.T) {
		res, err := lora.SetRecvEx(0)
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_SetConfig(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK))
	lora := newTestLora(fsp)

	t.Run("set enable or disable rssi and snr messages", func(t *testing.T) {
		res, err := lora.SetConfig("EU868:20")
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetConfig(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK))
	lora := newTestLora(fsp)

	t.Run("get lora configuration", func(t *testing.T) {
		res, err := lora.GetConfig("dev_addr")
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetBand(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OKEU868"))
	lora := newTestLora(fsp)

	t.Run("get lora region", func(t *testing.T) {
		res, err := lora.GetBand()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_JoinOTAA(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK), []byte(STATUS_JOINED_SUCCESS))
	lora := newTestLora(fsp)

	t.Run("activation over the air", func(t *testing.T) {
		res, err := lora.JoinOTAA()
		if err != nil {
			t.Fatalf("error %v", err)
		}

		if bytes.Compare([]byte(STATUS_JOINED_SUCCESS), fsp.At()) != 0 {
			t.Fatalf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Signal(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK10 11"))
	lora := newTestLora(fsp)

	t.Run("signal from Lora gateway", func(t *testing.T) {
		res, err := lora.Signal()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetDataRate(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK))
	lora := newTestLora(fsp)

	t.Run("change next data rate", func(t *testing.T) {
		res, err := lora.SetDataRate("EU868")
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetLinkCnt(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK1,2"))
	lora := newTestLora(fsp)

	t.Run("get lora link info", func(t *testing.T) {
		res, err := lora.SetLinkCnt(1.0, 2.0)
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetABPInfo(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK1,2,64,32"))
	lora := newTestLora(fsp)

	t.Run("abp info query", func(t *testing.T) {
		res, err := lora.GetABPInfo()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Send(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK), []byte("at+recv=2,0,0"))
	lora := newTestLora(fsp)

	t.Run("send packet string", func(t *testing.T) {
		res, err := lora.Send("0,1,DEADBEFF")
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Recv(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK1,2"))
	lora := newTestLora(fsp)

	t.Run("receive the module data", func(t *testing.T) {
		res, err := lora.Recv("STATUS_TX_CONFIRMED,10")
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetRfConfig(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK868100000,12,0,1,8,20"))
	lora := newTestLora(fsp)

	t.Run("get lorap2p configuration", func(t *testing.T) {
		res, err := lora.GetRfConfig()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Txc(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte(OK))
	lora := newTestLora(fsp)

	t.Run("set lorap2p tx continues", func(t *testing.T) {
		res, err := lora.Txc("1,10,64")
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetRadioStatus(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK1,2,3,4,5,6,7"))
	lora := newTestLora(fsp)

	t.Run("get the radio statistics", func(t *testing.T) {
		res, err := lora.GetRadioStatus()
		if err != nil {
			t.Fatalf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Timeout(t *testing.T) {
	fsp := newFakeFakeSerialPort([]byte("OK"))
	lora := newTestLora(fsp)
	lora.timeout = 100 * time.Millisecond

	t.Run("timeout", func(t *testing.T) {
		_, err := lora.JoinOTAA()
		if err == nil {
			t.Fatalf("expected error , but got none")
		}
		if err != ErrTimeout {
			t.Fatalf("expcted ErrTimeout error , but got:%v", err)

		}
	})
}

func newFakeFakeSerialPort(data ...[]byte) *FakeSerialPort {
	return &FakeSerialPort{
		responses: data,
	}
}

type FakeSerialPort struct {
	responses [][]byte // Each element is returns as a separate response.
	current   []byte
}

func (f *FakeSerialPort) Read(p []byte) (n int, err error) {
	if len(f.responses) == 0 {
		return 0, io.EOF
	}

	f.current = append(f.responses[0], []byte("\r\n")...)
	copy(p, f.current)
	f.responses = f.responses[1:]
	return len(f.current), nil
}

func (f *FakeSerialPort) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (*FakeSerialPort) Close() error {
	return nil
}

func (f FakeSerialPort) At() []byte {
	return []byte(strings.TrimSuffix(string(f.current), "\r\n"))
}

func newTestLora(fsp io.ReadWriteCloser) *Lora {
	return &Lora{
		port:    fsp,
		timeout: defaultConfig.ReadTimeout,
	}

}
