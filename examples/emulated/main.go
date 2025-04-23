package main

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/rabarar/meshtool-go/github.com/meshtastic/go/meshtastic"
	"github.com/rabarar/meshtool-go/public/emulated"
	"github.com/rabarar/meshtool-go/public/meshtool"
	"github.com/rabarar/meshtool-go/public/mqtt"
	"github.com/rabarar/meshtool-go/public/radio"
	"github.com/rabarar/meshtool-go/public/transport"
	"golang.org/x/sync/errgroup"
)

func main() {
	// TODO: Flesh this example out and make it configurable
	ctx := context.Background()
	log.SetLevel(log.DebugLevel)

	nodeID, err := meshtool.RandomNodeID()
	if err != nil {
		panic(err)
	}
	r, err := emulated.NewRadio(emulated.Config{
		LongName:   "EXAMPLE",
		ShortName:  "EMPL",
		NodeID:     nodeID,
		MQTTClient: &mqtt.DefaultClient,
		Channels: &meshtastic.ChannelSet{
			Settings: []*meshtastic.ChannelSettings{
				{
					Name: "LongFast",
					Psk:  radio.DefaultKey,
				},
			},
		},
		BroadcastNodeInfoInterval: 5 * time.Minute,

		BroadcastPositionInterval: 5 * time.Minute,
		// Hardcoded to the position of Buckingham Palace.
		PositionLatitudeI:  515014760,
		PositionLongitudeI: -1406340,
		PositionAltitude:   2,

		TCPListenAddr: "127.0.0.1:4403",
	})
	if err != nil {
		panic(err)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := r.Run(egCtx); err != nil {
			return fmt.Errorf("running radio: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		conn, err := transport.NewClientStreamConn(r.Conn(egCtx))
		if err != nil {
			return fmt.Errorf("creating connection: %w", err)
		}

		msg := &meshtastic.ToRadio{
			PayloadVariant: &meshtastic.ToRadio_Packet{
				Packet: &meshtastic.MeshPacket{
					From: nodeID.Uint32(),
					// This is hard coded to Noah's node ID
					To: 2437877602,
					PayloadVariant: &meshtastic.MeshPacket_Decoded{
						Decoded: &meshtastic.Data{
							Portnum: meshtastic.PortNum_TEXT_MESSAGE_APP,
							Payload: []byte("from main!!"),
						},
					},
				},
			},
		}
		if err := conn.Write(msg); err != nil {
			return fmt.Errorf("writing to radio: %w", err)
		}

		for {
			select {
			case <-egCtx.Done():
				return nil
			default:
			}
			msg := &meshtastic.FromRadio{}
			if err := conn.Read(msg); err != nil {
				return fmt.Errorf("reading from radio: %w", err)
			}
			log.Info("FromRadio!!", "packet", msg)
		}
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}
