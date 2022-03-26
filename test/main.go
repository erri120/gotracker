package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"erri120/gotracker"
	"erri120/gotracker/protocol"

	"github.com/anacrolix/torrent/tracker"
	"github.com/anacrolix/torrent/tracker/http"
	"github.com/anacrolix/torrent/tracker/udp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	go startServer()

	time.Sleep(time.Second * 2)

	for i := 0; i < 1; i++ {
		go startClient(i)
	}

	time.Sleep(time.Second * 2)
	return nil
}

func startServer() error {
	host := flag.String("host", "0.0.0.0", "host to listen on")
	port := flag.Int("port", 9000, "port to listen on")

	flag.Parse()

	ip := net.ParseIP(*host)

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()), os.Stdout, zap.DebugLevel)
	logger := zap.New(core).With(zap.String("name", "server"))

	defer logger.Sync()

	server := gotracker.Server{
		Logger:           logger,
		Conn:             nil,
		ConnectedClients: make(map[protocol.ConnectionId]gotracker.ConnectedClient),
		Torrents:         make(map[protocol.InfoHash]gotracker.Torrent),
	}

	server.Torrents[protocol.InfoHash{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1}] = gotracker.Torrent{
		Leechers: 0,
		Seeders:  0,
		Peers:    make([]protocol.PeerAddr, 0),
	}

	defer server.Close()
	return server.Listen(&net.UDPAddr{IP: ip, Port: *port})
}

func startClient(id int) {
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()), os.Stdout, zap.DebugLevel)
	logger := zap.New(core).With(zap.String("name", fmt.Sprint("client", id)))

	client, err := tracker.NewClient("udp://localhost:9000", tracker.NewClientOpts{})
	defer client.Close()

	if err != nil {
		logger.Error("error creating client", zap.Error(err))
		return
	}

	req := udp.AnnounceRequest{
		InfoHash:   [20]byte{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1},
		PeerId:     [20]byte{0xa3, 0x56, 0x41, 0x43, 0x74, 0x23, 0xe6, 0x26, 0xd9, 0x38, 0x25, 0x4a, 0x6b, 0x80, 0x49, 0x10, 0xa6, 0x67, 0xa, 0xc1},
		Downloaded: 0,
		Left:       0,
		Uploaded:   0,
		Event:      udp.AnnounceEvent(protocol.AnnounceEventStarted),
		IPAddress:  0,
		Key:        0,
		NumWant:    -1,
		Port:       0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), tracker.DefaultTrackerAnnounceTimeout)
	defer cancel()

	res, err := client.Announce(ctx, req, http.AnnounceOpt{})
	if err != nil {
		logger.Error("error announcing", zap.Error(err))
		return
	}

	logger.Info("announce response", zap.Int32("interval", res.Interval))
}
