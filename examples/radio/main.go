package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"meshtool-go/github.com/meshtastic/go/meshtastic"

	"github.com/charmbracelet/log"
	"github.com/rabarar/meshtool-go/public/transport"
	"github.com/rabarar/meshtool-go/public/transport/serial"
	"google.golang.org/protobuf/proto"
)

var port string

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.SetLevel(log.DebugLevel)

	if len(os.Args) > 1 {
		port = os.Args[1]
	} else {
		port = serial.GetPorts()[0]
	}
	serialConn, err := serial.Connect(port)
	if err != nil {
		panic(err)
	}
	streamConn, err := transport.NewClientStreamConn(serialConn)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := streamConn.Close(); err != nil {
			panic(err)
		}
	}()

	client := transport.NewClient(streamConn, false)
	client.Handle(new(meshtastic.MeshPacket), func(msg proto.Message) {
		pkt := msg.(*meshtastic.MeshPacket)
		data := pkt.GetDecoded()
		log.Info("Received message from radio", "msg", processMessage(data), "from", pkt.From, "portnum", data.Portnum.String())
	})
	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTimeout()
	if client.Connect(ctxTimeout) != nil {
		panic("Failed to connect to the radio")
	}

	log.Info("Waiting for interrupt signal")
	<-ctx.Done()
}

func processMessage(message *meshtastic.Data) string {
	if message.Portnum == meshtastic.PortNum_NODEINFO_APP {
		var user = meshtastic.User{}
		proto.Unmarshal(message.Payload, &user)
		return user.String()
	}
	if message.Portnum == meshtastic.PortNum_POSITION_APP {
		var pos = meshtastic.Position{}
		proto.Unmarshal(message.Payload, &pos)
		return pos.String()
	}
	if message.Portnum == meshtastic.PortNum_TELEMETRY_APP {
		var t = meshtastic.Telemetry{}
		proto.Unmarshal(message.Payload, &t)
		return t.String()
	}
	if message.Portnum == meshtastic.PortNum_NEIGHBORINFO_APP {
		var n = meshtastic.NeighborInfo{}
		proto.Unmarshal(message.Payload, &n)
		return n.String()
	}
	if message.Portnum == meshtastic.PortNum_STORE_FORWARD_APP {
		var s = meshtastic.StoreAndForward{}
		proto.Unmarshal(message.Payload, &s)
		return s.String()
	}

	return "unknown message type"
}
