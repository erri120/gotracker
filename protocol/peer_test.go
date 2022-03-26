package protocol

import (
	"bytes"
	"net"
	"testing"
)

func TestIPv4PeerMarshalBinary(t *testing.T) {
	var tests = []struct {
		peers    IPv4Peers
		expected []byte
	}{
		{
			peers: IPv4Peers{
				{
					Ip:   net.ParseIP("127.0.0.1"),
					Port: uint16(8080),
				},
				{
					Ip:   net.ParseIP("192.168.178.1"),
					Port: uint16(8080),
				},
			},
			expected: []byte{
				0x7f, 0x00, 0x00, 0x01, 0x1F, 0x90,
				0xC0, 0xA8, 0xB2, 0x01, 0x1F, 0x90,
			},
		},
	}

	for _, test := range tests {
		actual, err := test.peers.MarshalBinary()
		if err != nil {
			t.Fatalf("IPv4Peers.MarshalBinary() returned error: %v", err)
		}

		if !bytes.Equal(actual, test.expected) {
			t.Fatalf("IPv4Peers.MarshalBinary() returned %v, expected %v", actual, test.expected)
		}
	}
}
