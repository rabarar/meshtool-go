package transport

// Transport defines methods required for communicating with a radio via serial, ble, or tcp
// Probably need to reevaluate this to just use the ToRadio and FromRadio protobufs
type Transport interface {
	Connect() error
	SendPacket(data []byte) error
	RequestConfig() error

	//	Listen(ch chan)
	Close() error
}
