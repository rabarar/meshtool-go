package meshtool

import (
	"github.com/rabarar/meshtool-go/github.com/meshtastic/go/meshtastic"
)

type Node struct {
	LongName      string
	ShortName     string
	ID            uint32
	HardwareModel meshtastic.HardwareModel
}

// EncryptPacket - Not actually in use yet 😅
func (n *Node) EncryptPacket(pkt *meshtastic.MeshPacket, channelName string, key []byte) *meshtastic.MeshPacket {
	payload := pkt.GetPayloadVariant()
	_ = payload
	switch p := payload.(type) {
	case *meshtastic.MeshPacket_Decoded:
		_ = p
		encrypted := meshtastic.MeshPacket_Encrypted{
			Encrypted: nil,
		}
		_ = encrypted
	}
	return nil
}
