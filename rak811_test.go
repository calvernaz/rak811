package rak811

import (
	"bytes"
	"testing"
)

const OK = "OK\r\n"

func TestCreateCmd(t *testing.T) {
	tests := []struct {
		in string
		out string
	}{
		{"mode=0", "at+mode=0\r\n"}, /* SET LoraWAN work mode */
		{"join=otaa", "at+join=otaa\r\n"}, /* Join OTAA type*/
		{"join=abp", "at+join=abp\r\n"}, /* Join ABP type*/
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
	fsp := &FakeSerialPort{
		buf: []byte("OK11.22.33.44\r\n"),
	}

	lora := &Lora{
		port: fsp,
	}

	t.Run("get software version", func(t *testing.T) {
		res, err := lora.Version()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.buf) != 0 {
			t.Errorf("got %q, want %q", res, []byte(fsp.buf))
		}
	})
}

func TestLora_Sleep(t *testing.T) {
	fsp := &FakeSerialPort{
		buf: []byte(OK),
	}

	lora := &Lora{
		port: fsp,
	}

	t.Run("module enter sleep", func(t *testing.T) {
		res, err := lora.Sleep()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.buf) != 0 {
			t.Errorf("got %q, want %q", res, []byte(fsp.buf))
		}
	})
}

func TestLora_Reset(t *testing.T) {
	fsp := &FakeSerialPort{
		buf: []byte(OK),
	}

	lora := &Lora{
		port: fsp,
	}

	t.Run("reset module", func(t *testing.T) {
		res, err := lora.Reset(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.buf) != 0 {
			t.Errorf("got %q, want %q", res, []byte(fsp.buf))
		}
	})
}

func TestLora_Reload(t *testing.T) {
	fsp := &FakeSerialPort{
		buf: []byte(OK),
	}

	lora := &Lora{
		port: fsp,
	}

	t.Run("reload the default parameters", func(t *testing.T) {
		res, err := lora.Reset(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.buf) != 0 {
			t.Errorf("got %q, want %q", res, []byte(fsp.buf))
		}
	})
}

func TestLora_SetMode(t *testing.T) {
	fsp := &FakeSerialPort{
		buf: []byte(OK),
	}

	lora := &Lora{
		port: fsp,
	}

	t.Run("set module work on", func(t *testing.T) {
		res, err := lora.SetMode(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.buf) != 0 {
			t.Errorf("got %q, want %q", res, []byte(fsp.buf))
		}
	})
}

func TestLora_SetRecvEx(t *testing.T) {
	fsp := &FakeSerialPort{
		buf: []byte(OK),
	}

	lora := &Lora{
		port: fsp,
	}

	t.Run("set enable or disable rssi and snr messages", func(t *testing.T) {
		res, err := lora.SetRecvEx(0)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if bytes.Compare([]byte(res), fsp.buf) != 0 {
			t.Errorf("got %q, want %q", res, []byte(fsp.buf))
		}
	})
}

type FakeSerialPort struct {
	buf []byte
}

func (f FakeSerialPort) Read(p []byte) (n int, err error) {
	n = copy(p, f.buf)
	return n, nil
}

func (FakeSerialPort) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (FakeSerialPort) Close() error {
	return nil
}