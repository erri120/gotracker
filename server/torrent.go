package server

import (
	"net"

	"erri120/gotracker/protocol"
)

type Torrent interface {
	GetLeechers() (int32, error)
	GetSeeders() (int32, error)
	AddPeer(addr *net.UDPAddr) error
	GetPeers(maxPeers int32) ([]protocol.PeerAddr, error)
}
