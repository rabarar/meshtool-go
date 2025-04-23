package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/rabarar/meshtastic"
	"github.com/rabarar/meshtool-go/public/mqtt"
	"github.com/rabarar/meshtool-go/public/radio"
	"google.golang.org/protobuf/proto"
)

func main() {
	var server, username, password, rootTopic, level string
	flag.StringVar(&server, "server", "tcp://mqtt.meshtastic.org:1883", "MQTT server")
	flag.StringVar(&username, "username", "meshdev", "MQTT username")
	flag.StringVar(&password, "password", "large4cats", "MQTT password")
	flag.StringVar(&rootTopic, "topic", "msh/EU_868", "MQTT topic root")
	flag.StringVar(&level, "level", "info", "Log level")
	flag.Parse()
	if lvl, err := log.ParseLevel(level); err == nil {
		log.SetLevel(lvl)
	} else {
		log.Fatal("failed to parse log level", "level", level, "err", err)
	}

	client := mqtt.NewClient(server, username, password, rootTopic)
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}
	// key, err := generateKey("1PG7OiApB1nwvP+rz05pAQ==")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	client.Handle("LongFast", channelHandler("LongFast", radio.DefaultKey))
	log.Info("Started")
	select {}
}

func channelHandler(channel string, key []byte) mqtt.HandlerFunc {

	return func(m mqtt.Message) {
		var env meshtastic.ServiceEnvelope
		err := proto.Unmarshal(m.Payload, &env)
		if err != nil {
			log.Fatal("failed unmarshalling to service envelope", "err", err, "payload", hex.EncodeToString(m.Payload))
			return
		}

		/* TODO - not HasPacket()
		if !env.HasPacket() {
			log.Error("no packet in envelope", "payload", hex.EncodeToString(m.Payload))
			return
		}
		*/
		messagePtr, err := radio.TryDecode(env.Packet, key)
		if err != nil {
			log.Error("failed to decode packet", "err", err, "payload", hex.EncodeToString(m.Payload))
			return
		}
		if out, err := processMessage(messagePtr); err != nil {
			if messagePtr.Portnum != 0 {
				log.Error("failed to process message", "err", err, "payload", hex.EncodeToString(m.Payload), "topic", m.Topic, "channel", channel, "portnum", messagePtr.Portnum.String())
			}
			return
		} else {
			log.Info(out, "topic", m.Topic, "channel", channel, "portnum", messagePtr.Portnum.String())
		}
	}
}

var ErrUnknownMessageType = errors.New("unknown message type")

func processMessage(message *meshtastic.Data) (string, error) {
	var err error
	if message.Portnum == meshtastic.PortNum_NODEINFO_APP {
		var user = meshtastic.User{}
		err = proto.Unmarshal(message.Payload, &user)
		return user.String(), err
	}
	if message.Portnum == meshtastic.PortNum_POSITION_APP {
		var pos = meshtastic.Position{}
		err = proto.Unmarshal(message.Payload, &pos)
		return pos.String(), err
	}
	if message.Portnum == meshtastic.PortNum_TELEMETRY_APP {
		var t = meshtastic.Telemetry{}
		err = proto.Unmarshal(message.Payload, &t)
		return t.String(), err
	}
	if message.Portnum == meshtastic.PortNum_NEIGHBORINFO_APP {
		var n = meshtastic.NeighborInfo{}
		err = proto.Unmarshal(message.Payload, &n)
		return n.String(), err
	}
	if message.Portnum == meshtastic.PortNum_STORE_FORWARD_APP {
		var s = meshtastic.StoreAndForward{}
		err = proto.Unmarshal(message.Payload, &s)
		return s.String(), err
	}

	return "", ErrUnknownMessageType
}

func generateKey(key string) ([]byte, error) {
	// Pad the key with '=' characters to ensure it's a valid base64 string
	padding := (4 - len(key)%4) % 4
	paddedKey := key + strings.Repeat("=", padding)

	// Replace '-' with '+' and '_' with '/'
	replacedKey := strings.ReplaceAll(paddedKey, "-", "+")
	replacedKey = strings.ReplaceAll(replacedKey, "_", "/")

	// Decode the base64-encoded key
	return base64.StdEncoding.DecodeString(replacedKey)
}
