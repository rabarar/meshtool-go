package radio

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/rabarar/meshtastic"
	"google.golang.org/protobuf/proto"
)

// not sure what i was getting at with this
type comMode uint8

const (
	ComModeProtobuf comMode = iota + 1
	ComModeSerialDebug
)

// DefaultKey encryption key, commonly referenced as AQ==
// as base64: 1PG7OiApB1nwvP+rz05pAQ==
var DefaultKey = []byte{0xd4, 0xf1, 0xbb, 0x3a, 0x20, 0x29, 0x07, 0x59, 0xf0, 0xbc, 0xff, 0xab, 0xcf, 0x4e, 0x69, 0x01}

// ParseKey converts the most common representation of a channel encryption key (URL encoded base64) to a byte slice
func ParseKey(key string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(key)
}

// GenerateByteSlices creates a bunch of weak keys for use when interfacing on MQTT.
// This creates 128, 192, and 256 bit AES keys with only a single byte specified
func GenerateByteSlices() [][]byte {
	// There are 256 possible values for a single byte
	// We create 1536 slices: 512 with 16 bytes, 512 with 24 bytes, and 512 with 32 bytes
	allSlices := make([][]byte, 256*3)

	for i := 0; i < 256; i++ {
		// Create a slice of 16 bytes for the first 256 slices
		slice16 := make([]byte, 16)
		// Set the last byte to the current iteration value.
		slice16[15] = byte(i)
		// Assign the slice to our slice of slices.
		allSlices[i] = slice16

		// Create a slice of 24 bytes (192 bits) for the next 256 slices
		slice24 := make([]byte, 24)
		// Set the last byte to the current iteration value.
		slice24[23] = byte(i)
		// Assign the slice to our slice of slices, offset by 256.
		allSlices[i+256] = slice24

		// Create a slice of 32 bytes for the last 256 slices
		slice32 := make([]byte, 32)
		// Set the last byte to the current iteration value.
		slice32[31] = byte(i)
		// Assign the slice to our slice of slices, offset by 512.
		allSlices[i+512] = slice32
	}

	return allSlices
}

// xorHash computes a simple XOR hash of the provided byte slice.
func xorHash(p []byte) uint8 {
	var code uint8
	for _, b := range p {
		code ^= b
	}
	return code
}

// ChannelHash returns the hash for a given channel by XORing the channel name and PSK.
func ChannelHash(channelName string, channelKey []byte) (uint32, error) {
	if len(channelKey) == 0 {
		return 0, fmt.Errorf("channel key cannot be empty")
	}

	h := xorHash([]byte(channelName))
	h ^= xorHash(channelKey)

	return uint32(h), nil
}

// TryDecode attempts to decrypt a packet with the specified key, or return the already decrypted data if present.
func TryDecode(packet *meshtastic.MeshPacket, key []byte) (*meshtastic.Data, error) {

	switch packet.GetPayloadVariant().(type) {
	case *meshtastic.MeshPacket_Decoded:
		//fmt.Println("decoded")
		return packet.GetDecoded(), nil
	case *meshtastic.MeshPacket_Encrypted:
		decrypted, err := XOR(packet.GetEncrypted(), key, packet.Id, packet.From)
		if err != nil {
			log.Warnf("Failed decrypting packet: %s", err)
			return nil, ErrDecrypt
		}
		log.Warnf("PLAINTEXT: [%s]", hex.EncodeToString(decrypted))

		useOriginal := true
		if useOriginal {
			var meshPacket meshtastic.Data
			err = proto.Unmarshal(decrypted, &meshPacket)
			if err != nil {
				log.Warnf("Failed to unmarshal Meshtastic Data packet: %s", err)
				return nil, ErrDecrypt
			}
			return &meshPacket, nil
		} else {

			var dataPacket meshtastic.Data
			err = proto.Unmarshal(decrypted, &dataPacket)
			if err != nil {
				log.Warnf("Failed to unmarshal Meshtastic Data packet: %s", err)
				return nil, ErrDecrypt
			}

			switch dataPacket.Portnum {
			case meshtastic.PortNum_TEXT_MESSAGE_APP:
				txt := dataPacket.Payload
				fmt.Println("Got Text:", string(txt))
			case meshtastic.PortNum_TELEMETRY_APP:
				var telemetry meshtastic.Telemetry
				proto.Unmarshal(dataPacket.Payload, &telemetry)
				fmt.Printf("Got Telemetry:")
			default:
				fmt.Println("Unknown portnum:", dataPacket.Portnum)
			}

			return &dataPacket, nil
		}
	default:
		return nil, ErrUnkownPayloadType
	}
}
