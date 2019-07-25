# rak811
RAK811 Golang Library for use with LoRa pHAT &amp; MicroBIT Node

# Example

```
func main() {
	cfg := &serial.Config{
		Name:        "/dev/ttyS0",
	}

	lora, err := rak811.New(cfg)
	if err != nil {
		log.Fatal(err, "failed to create rak811 instance")
	}

	resp, err := lora.Version()
	if err != nil {
		log.Fatal("failed to get version")
	}

	fmt.Println(resp)
}
```