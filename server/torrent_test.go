package server

import (
	"erri120/gotracker/protocol"
	"net"
	"testing"
)

func TestTorrentAddExistingPeer(t *testing.T) {
	torrent := &Torrent{
		Leechers: 0,
		Seeders:  0,
		Peers: []protocol.PeerAddr{
			protocol.PeerAddr{
				Ip:   net.IPv4(127, 0, 0, 1),
				Port: 6881,
			},
		},
	}

	torrent.AddPeer(&net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 6881,
	})

	if len(torrent.Peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(torrent.Peers))
	}
}

func TestTorrentAddNewPeer(t *testing.T) {
	torrent := &Torrent{
		Leechers: 0,
		Seeders:  0,
		Peers: []protocol.PeerAddr{
			protocol.PeerAddr{
				Ip:   net.IPv4(127, 0, 0, 1),
				Port: 6881,
			},
		},
	}

	torrent.AddPeer(&net.UDPAddr{
		IP:   net.IPv4(192, 168, 178, 1),
		Port: 6881,
	})

	if len(torrent.Peers) != 2 {
		t.Errorf("expected 2 peers, got %d", len(torrent.Peers))
	}
}
