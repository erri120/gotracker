package protocol

import (
	"bytes"
	"encoding/binary"
	"net"
)

type PeerAddr struct {
	Ip   net.IP
	Port uint16
}

type IPv4Peers []PeerAddr

type IPv6Peers []PeerAddr

func (peers IPv4Peers) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(peers)*(net.IPv4len+2)))

	for _, peer := range peers {
		ip := peer.Ip.To4()

		// skip this peer if the IP is not IPv4
		if ip == nil {
			continue
		}

		if _, err := buf.Write(ip); err != nil {
			return nil, err
		}

		if err := binary.Write(buf, binary.BigEndian, &peer.Port); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
