package server

import (
	"erri120/gotracker/protocol"
	"net"
)

type Torrent struct {
	Leechers int32
	Seeders  int32
	Peers    []protocol.PeerAddr
}

func (torrent *Torrent) AddPeer(addr *net.UDPAddr) {
	var exists bool
	for _, peer := range torrent.Peers {
		// TODO: allow same IP but different ports?
		if peer.Ip.Equal(addr.IP) {
			exists = true
			break
		}
	}

	if exists {
		return
	}

	torrent.Peers = append(torrent.Peers, protocol.PeerAddr{
		Ip:   addr.IP,
		Port: uint16(addr.Port),
	})
}

func (torrent Torrent) GetPeers(maxPeers int32) []protocol.PeerAddr {
	if maxPeers == -1 {
		maxPeers = defaultPeerCount
	}

	if len(torrent.Peers) < int(maxPeers) {
		return torrent.Peers
	}

	if maxPeers > maxPeerCount {
		return torrent.Peers[:maxPeerCount]
	}

	return torrent.Peers[:maxPeers]
}
