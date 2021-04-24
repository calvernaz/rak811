package rak811

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
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
		{"send=0,2,000000000000007F0000000000000000", "at+send=0,2,000000000000007F0000000000000000\r\n"}, /*APP conn:2, battery level 50%, unconfirmed message*/
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

func Test_WhichError(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{"ERROR-1", CodeArgErr},
		{"ERROR-2", CodeArgNotFind},
		{"ERROR-3", CodeJoinAbpErr},
		{"ERROR-4", CodeJoinOtaaErr},
		{"ERROR-5", CodeNotJoin},
		{"ERROR-6", CodeMacBusyErr},
		{"ERROR-7", CodeTxErr},
		{"ERROR-8", CodeInterErr},
		{"ERROR-11", CodeWrCfgErr},
		{"ERROR-12", CodeRdCfgErr},
		{"ERROR-13", CodeTxLenLimitErr},
		{"ERROR-20", CodeUnknownErr},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			err := WhichError(tt.in)
			fmt.Println(err)
			code := err.Code()
			if code != tt.out {
				t.Errorf("want %d, got %d", tt.out, code)
			}
		})
	}
}

func TestWhichEventResponse(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{"at+recv=0,0,0", StatusRecvData},
		{"at+recv=1,0,0", StatusTxConfirmed},
		{"at+recv=2,0,0", StatusTxUnconfirmed},
		{"at+recv=3,0,0", StatusJoinedSuccess},
		{"at+recv=4,0,0", StatusJoinedFailed},
		{"at+recv=5,0,0", StatusTxTimeout},
		{"at+recv=6,0,0", StatusRx2Timeout},
		{"at+recv=7,0,0", StatusDownlinkRepeated},
		{"at+recv=8,0,0", StatusWakeUp},
		{"at+recv=9,0,0", StatusP2pComplete},
		{"at+recv=100,0,0", StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			code := WhichEventResponse(tt.in).Code()
			if code != tt.out {
				t.Errorf("want %d, got %d", tt.out, code)
			}
		})
	}
}

func TestLora_Version(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OK2.0.3.0\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate lora")
	}

	t.Run("get software version", func(t *testing.T) {
		actual, err := lora.Version()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(actual), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", actual, fsp.At())
		}
	})
}

func TestLora_Sleep(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("module enter sleep", func(t *testing.T) {
		res, err := lora.Sleep()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Reset(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("reset module", func(t *testing.T) {
		res, err := lora.Reset(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Reload(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("reload the default parameters", func(t *testing.T) {
		res, err := lora.Reset(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_SetMode(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("set module work on", func(t *testing.T) {
		res, err := lora.SetMode(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_SetRecvEx(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("set enable or disable rssi and snr messages", func(t *testing.T) {
		res, err := lora.SetRecvEx(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_SetConfig(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("set enable or disable rssi and snr messages", func(t *testing.T) {
		res, err := lora.SetConfig("EU868:20")
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetConfig(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("get lora configuration", func(t *testing.T) {
		res, err := lora.GetConfig("dev_addr")
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetBand(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OKEU868\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("get lora region", func(t *testing.T) {
		res, err := lora.GetBand()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_JoinOTAA(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK+"\r\n"), []byte(JoinSuccess+"\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("over the air activation", func(t *testing.T) {
		res, err := lora.JoinOTAA()
		if err != nil {
			t.Errorf("error %v", err)
		}

		if bytes.Compare([]byte(JoinSuccess), fsp.At()) != 0 {
			t.Fatalf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_JoinOTAA_Failed(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK+"\r\n"), []byte(JoinFail+"\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("failed over the air activation", func(t *testing.T) {
		res, err := lora.JoinOTAA()
		if err != nil {
			t.Errorf("error %v", err)
		}

		if bytes.Compare([]byte(JoinFail), fsp.At()) != 0 {
			t.Fatalf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_JoinOTAA_Cant_Join(t *testing.T) {
	fsp := newFakeSerialConn([]byte("ERROR-4\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("failed over the air activation", func(t *testing.T) {
		res, err := lora.JoinOTAA()
		if err != nil {
			lerr := WhichError(res)
			if lerr.code != -4 || strings.Compare(lerr.desc, "can't join network using OTAA") != 0 {
				t.Fatalf("got %q, want %q", lerr.code, CodeJoinOtaaErr)
			}
		}
	})
}

func TestLora_Signal(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OK10 11\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

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
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("change next data rate", func(t *testing.T) {
		res, err := lora.SetDataRate("EU868")
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetLinkCnt(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OK1,2\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("get lora link info", func(t *testing.T) {
		res, err := lora.SetLinkCnt(1.0, 2.0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetABPInfo(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OK1,2,64,32\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("abp info query", func(t *testing.T) {
		res, err := lora.GetABPInfo()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Send(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK+"\r\n"), []byte("at+recv=2,0,0\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

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
	fsp := newFakeSerialConn([]byte("OK1,2\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("receive the module data", func(t *testing.T) {
		res, err := lora.Recv("STATUS_TX_CONFIRMED,10\r\n")
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetRfConfig(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OK868100000,12,0,1,8,20\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("get lorap2p configuration", func(t *testing.T) {
		res, err := lora.GetRfConfig()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_Txc(t *testing.T) {
	fsp := newFakeSerialConn([]byte(OK))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

	t.Run("set lorap2p tx continues", func(t *testing.T) {
		res, err := lora.Txc("1,10,64")
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.At()) != 0 {
			t.Errorf("got %q, want %q", res, fsp.At())
		}
	})
}

func TestLora_GetRadioStatus(t *testing.T) {
	fsp := newFakeSerialConn([]byte("OK1,2,3,4,5,6,7\r\n"))
	lora, err := newLora(fsp)
	if err != nil {
		t.Error("failed to instantiate Lora")
	}

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

func TestLora_Debug(t *testing.T) {
	cfg := &extraConfig{debug: false}
	lora := &Lora{config: cfg}

	lora.Debug(true)

	if lora.config.debug != true {
		t.Errorf("got %v, want true", lora.config.debug)
	}
}

func newFakeSerialConn(data ...[]byte) *FakeSerialConn {
	return &FakeSerialConn{
		responses: data,
	}
}

type FakeSerialConn struct {
	responses [][]byte // Each element is returns as a separate response.
	current   []byte
}

func (f *FakeSerialConn) Read(p []byte) (n int, err error) {
	if len(f.responses) == 0 {
		return 0, io.EOF
	}

	f.current = f.responses[0]
	copy(p, f.current)
	f.responses = f.responses[1:]
	return len(f.current), nil
}

func (f *FakeSerialConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (*FakeSerialConn) Close() error {
	return nil
}

func (*FakeSerialConn) Flush() {
	return
}

func (f FakeSerialConn) At() []byte {
	return []byte(strings.TrimSuffix(string(f.current), "\r\n"))
}
