package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSizeOfRequestHeader(t *testing.T) {
	requestHeader := RequestHeader{
		ConnectionId:  0x0,
		Action:        0x0,
		TransactionId: 0x0,
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, requestHeader)
	if err != nil {
		t.Error(err)
	}

	actualSize := uint32(buf.Len())
	expectedSize := SizeOfRequestHeader

	if actualSize != expectedSize {
		t.Errorf("SizeOfRequestHeader: expected %d, actual %d", expectedSize, actualSize)
	}
}

func TestSizeOfIPv4AnnounceRequest(t *testing.T) {
	ipv4AnnounceRequest := IPv4AnnounceRequest{
		InfoHash:   [20]byte{},
		PeerId:     [20]byte{},
		Downloaded: 0,
		Left:       0,
		Uploaded:   0,
		Event:      0,
		IpAddress:  0,
		Key:        0,
		NumWanted:  0,
		Port:       0,
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, ipv4AnnounceRequest)
	if err != nil {
		t.Error(err)
	}

	actualSize := uint32(buf.Len())
	expectedSize := SizeOfIPv4AnnounceRequest

	if actualSize != expectedSize {
		t.Errorf("SizeOfIPv4AnnounceRequest: expected %d, actual %d", expectedSize, actualSize)
	}
}
