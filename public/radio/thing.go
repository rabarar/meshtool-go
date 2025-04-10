package radio

import "buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"

// Something is something created to track keys for packet decrypting
type Something struct {
	keys map[string][]byte
}

func NewThing() *Something {
	return &Something{keys: map[string][]byte{
		"LongFast":  DefaultKey,
		"LongSlow":  DefaultKey,
		"VLongSlow": DefaultKey,
	}}
}

// TryDecode decode a payload to a Data protobuf
func (s *Something) TryDecode(packet *meshtastic.MeshPacket, key []byte) (*meshtastic.Data, error) {
	return TryDecode(packet, key)
}
