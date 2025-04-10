package serial

import (
	"go.bug.st/serial"
)

const (
	DefaultPortSpeed = 115200 //921600
)

func Connect(port string) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: DefaultPortSpeed,
	}
	p, err := serial.Open(port, mode)
	if err != nil {
		return nil, err
	}
	return p, nil
}
