package server

import (
	"bytes"
	"erri120/gotracker/protocol"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type testTorrent struct {
	leechers int32
	seeders  int32
	peers    []protocol.PeerAddr
}

func (t *testTorrent) GetLeechers() (int32, error) {
	return t.leechers, nil
}

func (t *testTorrent) GetSeeders() (int32, error) {
	return t.seeders, nil
}

func (t *testTorrent) AddPeer(addr *net.UDPAddr) error {
	t.peers = append(t.peers, protocol.PeerAddr{
		IP:   addr.IP,
		Port: uint16(addr.Port),
	})
	return nil
}

func (t *testTorrent) GetPeers(maxPeers int32) ([]protocol.PeerAddr, error) {
	return t.peers, nil
}

func TestServer(t *testing.T) {
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), os.Stdout, zap.DebugLevel))
	defer logger.Sync()

	torrents := map[protocol.InfoHash]Torrent{
		{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}: &testTorrent{
			leechers: 1,
			seeders:  2,
			peers: []protocol.PeerAddr{
				{
					IP:   net.IPv4(1, 2, 3, 4),
					Port: 12345,
				},
			},
		},
	}

	server := &Server{
		Logger: logger,
		GetTorrent: func(infoHash protocol.InfoHash) (Torrent, error) {
			torrent, ok := torrents[infoHash]
			if !ok {
				return nil, fmt.Errorf("torrent not found")
			}

			return torrent, nil
		},
	}

	serverAddr := &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 12345,
	}

	go func() {
		err := server.Listen(serverAddr)
		if err != nil {
			t.Fatalf("Failed to listen: %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 100)
	defer server.Close()

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// send connect request
	connectRequest := protocol.RequestHeader{
		Action:        protocol.ActionConnect,
		ConnectionId:  protocol.ConnectRequestMagic,
		TransactionId: 0x01,
	}

	connectRequestBytes, err := protocol.Marshal(8+4+4, connectRequest)
	if err != nil {
		t.Fatalf("Failed to marshal connect request: %v", err)
	}

	_, err = conn.Write(connectRequestBytes)
	if err != nil {
		t.Fatalf("Failed to write connect request: %v", err)
	}

	// read connect response
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		t.Fatalf("Failed to read connect response: %v", err)
	}

	reader := bytes.NewReader(data[:n])

	var connectResponse protocol.ConnectResponse
	err = protocol.Unmarshal(reader, &connectResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal connect response: %v", err)
	}

	// action should be the same
	if connectResponse.Action != connectRequest.Action {
		t.Fatalf("Unexpected connect response action: %v", connectResponse.Action)
	}

	// transaction id should be the same
	if connectResponse.TransactionId != connectRequest.TransactionId {
		t.Fatalf("Unexpected connect response transaction id: %v", connectResponse.TransactionId)
	}

	// on connect we should get a new connection id
	if connectResponse.ConnectionId == connectRequest.ConnectionId {
		t.Fatalf("Unexpected connect response connection id: %v", connectResponse.ConnectionId)
	}

	// send announce request
	announceRequestHeader := protocol.RequestHeader{
		Action:        protocol.ActionAnnounce,
		ConnectionId:  connectResponse.ConnectionId,
		TransactionId: 0x02,
	}

	announceRequest := protocol.IPv4AnnounceRequest{
		InfoHash:   protocol.InfoHash{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a},
		PeerId:     protocol.PeerId{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		Port:       0,
		Uploaded:   0,
		Downloaded: 0,
		Left:       0,
		Event:      protocol.AnnounceEventStarted,
		IpAddress:  0,
		Key:        0,
		NumWanted:  10,
	}

	announceUrl := "/announce"
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(0x2))
	buf.WriteByte(byte(len(announceUrl)))
	buf.WriteString(announceUrl)
	urlData := buf.Bytes()

	announceRequestBytes, err := protocol.Marshal(8+4+4+protocol.SizeOfIPv4AnnounceRequest+uint32(len(urlData)), announceRequestHeader, announceRequest, urlData)
	if err != nil {
		t.Fatalf("Failed to marshal announce request: %v", err)
	}

	_, err = conn.Write(announceRequestBytes)
	if err != nil {
		t.Fatalf("Failed to write announce request: %v", err)
	}

	// read announce response
	data = make([]byte, 1024)
	n, err = conn.Read(data)
	if err != nil {
		t.Fatalf("Failed to read announce response: %v", err)
	}

	reader = bytes.NewReader(data[:n])

	var responseHeader protocol.ResponseHeader
	err = protocol.Unmarshal(reader, &responseHeader)
	if err != nil {
		t.Fatalf("Failed to unmarshal announce response header: %v", err)
	}

	var announceResponse protocol.IPv4AnnounceResponse
	err = protocol.Unmarshal(reader, &announceResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal announce response: %v", err)
	}

	// leechers should match
	if announceResponse.Leechers != torrents[announceRequest.InfoHash].(*testTorrent).leechers {
		t.Fatalf("Unexpected leechers: %v", announceResponse.Leechers)
	}

	// seeders should match
	if announceResponse.Seeders != torrents[announceRequest.InfoHash].(*testTorrent).seeders {
		t.Fatalf("Unexpected seeders: %v", announceResponse.Seeders)
	}
}
