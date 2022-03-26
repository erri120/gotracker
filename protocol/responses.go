package protocol

const SizeOfResponseHeader uint32 = 4 + 4

type ResponseHeader struct {
	Action        Action
	TransactionId TransactionId
}

const SizeOfConnectResponse uint32 = 4 + 4 + 8

type ConnectResponse struct {
	Action        Action
	TransactionId TransactionId
	ConnectionId  ConnectionId
}

const SizeOfIPv4AnnounceResponse uint32 = 4 + 4 + 4

type IPv4AnnounceResponse struct {
	Interval int32
	Leechers int32
	Seeders  int32
}
