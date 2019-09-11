# rak811
RAK811 Golang Library for use with LoRa pHAT &amp; MicroBIT Node

[![CircleCI](https://circleci.com/gh/calvernaz/rak811.svg?style=svg)](https://circleci.com/gh/calvernaz/rak811)

# Example

```
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

	fmt.Println(resp)
}
```

To run the example, use `sudo`:

	sudo go run main.go

You can find a more complete example and usage in resources.

# Resources

[Wiki](https://github.com/calvernaz/rak811/wiki/Development)
[Usage Example](https://github.com/calvernaz/lorawan-rpi-temp)