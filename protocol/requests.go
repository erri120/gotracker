package protocol

// Magic Connection Id used when the Client wants to connect to the Tracker.
const ConnectRequestMagic ConnectionId = 0x41727101980

const SizeOfRequestHeader uint32 = 8 + 4 + 4

type RequestHeader struct {
	ConnectionId  ConnectionId  // Connection Id for the Client.
	Action        Action        // Action of the request.
	TransactionId TransactionId // Transaction Id of the request.
}

type AnnounceEvent int32

const (
	AnnounceEventNone      AnnounceEvent = 0
	AnnounceEventCompleted AnnounceEvent = 1 // The local peer just completed the torrent.
	AnnounceEventStarted   AnnounceEvent = 2 // The local peer has just resumed this torrent.
	AnnounceEventStopped   AnnounceEvent = 3 // The local peer is leaving the swarm.
)

const SizeOfIPv4AnnounceRequest uint32 = 20 + 20 + 8 + 8 + 8 + 4 + 4 + 4 + 4 + 2

type IPv4AnnounceRequest struct {
	InfoHash   InfoHash      //
	PeerId     PeerId        //
	Downloaded int64         // Number of bytes downloaded.
	Left       int64         // Number of bytes left.
	Uploaded   int64         // Number of bytes uploaded.
	Event      AnnounceEvent //
	IpAddress  uint32        //
	Key        int32         //
	NumWanted  int32         // Number of Peers the Client wants. -1 for default.
	Port       uint16        //
}

type BEP41OptionType uint8

const (
	BEP41OptionTypeEndOfOptions BEP41OptionType = 0
	BEP41OptionTypeNOP          BEP41OptionType = 1
	BEP41OptionTypeURLData      BEP41OptionType = 2
)
