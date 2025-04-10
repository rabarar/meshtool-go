This example connects to a Meshtastic device connected to a serial port, decodes received messages, and logs them out

```shell
go run main.go  # search for first device connected to a serial port
go run main.go /dev/ttyUSB0  # explicitly provide a port to connect to
```
