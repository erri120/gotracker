package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSizeOfResponseHeader(t *testing.T) {
	responseHeader := ResponseHeader{
		Action:        0x0,
		TransactionId: 0x0,
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, responseHeader)
	if err != nil {
		t.Error(err)
	}

	actualSize := uint32(buf.Len())
	expectedSize := SizeOfResponseHeader

	if actualSize != expectedSize {
		t.Errorf("SizeOfResponseHeader: expected %d, actual %d", expectedSize, actualSize)
	}
}

func TestSizeOfConnectResponse(t *testing.T) {
	connectResponse := ConnectResponse{
		Action:        0x0,
		TransactionId: 0x0,
		ConnectionId:  0x0,
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, connectResponse)
	if err != nil {
		t.Error(err)
	}

	actualSize := uint32(buf.Len())
	expectedSize := SizeOfConnectResponse

	if actualSize != expectedSize {
		t.Errorf("SizeOfConnectResponse: expected %d, actual %d", expectedSize, actualSize)
	}
}

func TestSizeOfIPv4AnnounceResponse(t *testing.T) {
	ipv4AnnounceResponse := IPv4AnnounceResponse{
		Interval: 0,
		Leechers: 0,
		Seeders:  0,
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, ipv4AnnounceResponse)
	if err != nil {
		t.Error(err)
	}

	actualSize := uint32(buf.Len())
	expectedSize := SizeOfIPv4AnnounceResponse

	if actualSize != expectedSize {
		t.Errorf("SizeOfIPv4AnnounceResponse: expected %d, actual %d", expectedSize, actualSize)
	}
}
