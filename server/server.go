package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"time"

	"erri120/gotracker/protocol"

	"go.uber.org/zap"
)

const (
	blockSize            = 1024
	announceInterval     = 900
	connectionIdLifetime = time.Minute * 2
)

type GetTorrentFunc func(infoHash protocol.InfoHash) (Torrent, error)

type IsBannedFunc func(remoteAddr *net.UDPAddr) bool

type Server struct {
	Logger     *zap.Logger
	GetTorrent GetTorrentFunc
	IsBanned   IsBannedFunc

	conn             *net.UDPConn
	connectedClients map[protocol.ConnectionId]ConnectedClient
	closed           bool
}

// Responds to a client with the given data.
func (server *Server) Respond(remoteAddr *net.UDPAddr, bufSize uint32, responseHeader protocol.ResponseHeader, parts ...interface{}) error {
	bytes, err := protocol.Marshal(bufSize, append([]interface{}{responseHeader}, parts...)...)
	if err != nil {
		return err
	}

	n, err := server.conn.WriteToUDP(bytes, remoteAddr)
	if err != nil {
		return err
	}

	if n != len(bytes) {
		return fmt.Errorf("Wrote %d instead of %d bytes", n, len(bytes))
	}

	return nil
}

// Responds to a client with an error message.
func (server *Server) RespondWithError(remoteAddr *net.UDPAddr, transationId protocol.TransactionId, message string) error {
	b := []byte(message)
	return server.Respond(remoteAddr, protocol.SizeOfResponseHeader+uint32(len(b)), protocol.ResponseHeader{
		Action:        protocol.ActionError,
		TransactionId: transationId,
	}, b)
}

// Closes the underlying connection if it is open.
func (server *Server) Close() error {
	if server.closed {
		return fmt.Errorf("Server is already closed!")
	}

	server.closed = true

	if server.conn == nil {
		return nil
	}

	err := server.conn.Close()
	server.conn = nil
	return err
}

// Registers a new connection and returns the connection id.
func (server *Server) RegisterNewConnection() protocol.ConnectionId {
	// connection ids should not be guessable by the client, can look into crypto/rand
	connectionId := protocol.ConnectionId(rand.Int63())

	server.connectedClients[connectionId] = ConnectedClient{
		ConnectionId: connectionId,
		TimeIdIssued: time.Now(),
	}

	return connectionId
}

// Unregisters a connection.
func (server *Server) UnregisterConnection(connectionId protocol.ConnectionId) {
	delete(server.connectedClients, connectionId)
}

// Checks if the given connection id is valid.
func (server *Server) IsConnected(connectionId protocol.ConnectionId) bool {
	_, ok := server.connectedClients[connectionId]
	return ok
}

// Checks if the connection of the given connection id is valid.
func (server *Server) IsValidConnection(connectionId protocol.ConnectionId) bool {
	connection, ok := server.connectedClients[connectionId]
	if !ok {
		return false
	}

	return connection.IsValid()
}

// Starts the cleanup routine which removes old connections.
func (server *Server) StartCleanup(sleepTime time.Duration) {
	for {
		for _, connection := range server.connectedClients {
			if connection.IsValid() {
				continue
			}

			server.UnregisterConnection(connection.ConnectionId)
		}

		// TODO: also cleanup peers of torrents, would need to add timeAdded to peers or something

		time.Sleep(sleepTime)
	}
}

// Starts the routine which handles incoming messages.
func (server *Server) Listen(addr *net.UDPAddr) (err error) {
	if server.closed {
		return fmt.Errorf("Server is closed!")
	}

	if server.conn != nil {
		return fmt.Errorf("Server is already listening!")
	}

	if server.Logger == nil {
		server.Logger = zap.NewNop()
	}

	if server.GetTorrent == nil {
		return fmt.Errorf("GetTorrent function is not set!")
	}

	if server.IsBanned == nil {
		server.IsBanned = func(remoteAddr *net.UDPAddr) bool {
			return false
		}
	}

	server.connectedClients = make(map[protocol.ConnectionId]ConnectedClient)

	// TODO: UDP over IPv4 vs IPv6, the network name has to be changed to "udp4" or "udp6"
	listener, err := net.ListenUDP("udp4", addr)
	if err != nil {
		server.Logger.Error("Unable to start listening for packages", zap.Error(err))
		return
	}

	server.conn = listener

	data := make([]byte, blockSize)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)

		if server.closed {
			server.Logger.Info("Server closed, stopping listening")
			return nil
		}

		if err != nil {
			server.Logger.Error("Unable to read package", zap.Error(err))
			continue
		}

		if server.IsBanned(remoteAddr) {
			server.Logger.Warn("Banned client tried to send data", zap.String("remote", remoteAddr.String()))
			continue
		}

		if n < int(protocol.SizeOfRequestHeader) {
			server.Logger.Error("Received package is too small", zap.Int("bytes", n))
			continue
		}

		reader := bytes.NewReader(data[:n])
		go server.handleRequest(reader, remoteAddr)
	}
}

func (server *Server) handleRequest(reader *bytes.Reader, remoteAddr *net.UDPAddr) {
	n := reader.Len()

	var requestHeader protocol.RequestHeader
	if err := protocol.Unmarshal(reader, &requestHeader); err != nil {
		server.Logger.Error("Unable to unmarshal request header", zap.Error(err))
		return
	}

	logger := server.Logger.With(
		zap.String("remote", remoteAddr.String()),
		zap.Int32("action", int32(requestHeader.Action)),
		zap.Int64("connectionId", int64(requestHeader.ConnectionId)),
	)

	logger.Debug("Handling request from client", zap.Int("bytes", n))

	switch requestHeader.Action {
	case protocol.ActionConnect:
		if n > int(protocol.SizeOfRequestHeader) {
			logger.Warn("Client sent more data than expected for connect request", zap.Int("diff", n-int(protocol.SizeOfRequestHeader)))
		}

		if requestHeader.ConnectionId != protocol.ConnectRequestMagic {
			logger.Error("Client used wrong connection id for connect request")
			return
		}

		// TODO: maybe check if the Client is already connected based on the remote address

		connectionId := server.RegisterNewConnection()

		err := server.Respond(remoteAddr, protocol.SizeOfConnectResponse, protocol.ResponseHeader{
			Action:        protocol.ActionConnect,
			TransactionId: requestHeader.TransactionId,
		}, connectionId)

		if err != nil {
			logger.Error("Unable to respond to client with connect response", zap.Error(err))
			return
		}
	case protocol.ActionAnnounce:
		if !server.IsConnected(requestHeader.ConnectionId) {
			logger.Warn("Client tried to announce with unregistered connection id")
			return
		}

		if !server.IsValidConnection(requestHeader.ConnectionId) {
			logger.Warn("Client tired to use an expired connection id")
			server.UnregisterConnection(requestHeader.ConnectionId)
			return
		}

		// TODO: IPv4/IPv6
		if n < int(protocol.SizeOfRequestHeader+protocol.SizeOfIPv4AnnounceRequest) {
			logger.Error("Client send not enough data for an announce request",
				zap.Int("diff", n-int(protocol.SizeOfRequestHeader+protocol.SizeOfIPv4AnnounceRequest)),
			)
			return
		}

		var announceRequest protocol.IPv4AnnounceRequest
		if err := protocol.Unmarshal(reader, &announceRequest); err != nil {
			logger.Error("Unable to unmarshal IPv4 announce request", zap.Error(err))
			return
		}

		if n > int(protocol.SizeOfRequestHeader+protocol.SizeOfIPv4AnnounceRequest) {
			// check for BEP 41 as an extension
			url, err := protocol.ExtractExtensionData(reader)
			if err != nil {
				logger.Error("Unable to extract extension data", zap.Error(err))
				return
			}

			// TODO: do something with the url
			logger.Debug("Extracted extension data", zap.String("url", url))
		}

		torrent, err := server.GetTorrent(announceRequest.InfoHash)
		if err != nil {
			logger.Error("Unable to get torrent", zap.Error(err), zap.Binary("infoHash", announceRequest.InfoHash[:]))
			return
		}

		peers, err := torrent.GetPeers(announceRequest.NumWanted)
		if err != nil {
			logger.Error("Unable to get peers", zap.Error(err), zap.Binary("infoHash", announceRequest.InfoHash[:]))
			return
		}

		leechers, err := torrent.GetLeechers()
		if err != nil {
			logger.Error("Unable to get leechers", zap.Error(err), zap.Binary("infoHash", announceRequest.InfoHash[:]))
			return
		}

		seeders, err := torrent.GetSeeders()
		if err != nil {
			logger.Error("Unable to get seeders", zap.Error(err), zap.Binary("infoHash", announceRequest.InfoHash[:]))
			return
		}

		// add current client as peer
		if err := torrent.AddPeer(remoteAddr); err != nil {
			logger.Error("Unable to add peer", zap.Error(err), zap.Binary("infoHash", announceRequest.InfoHash[:]))
			return
		}

		// TODO: IPv4/IPv6
		ipv4Peers := protocol.IPv4Peers(peers)

		b, err := ipv4Peers.MarshalBinary()
		if err != nil {
			logger.Error("Unable to marshal IPv4 peers", zap.Error(err))
			return
		}

		err = server.Respond(remoteAddr, protocol.SizeOfIPv4AnnounceResponse+uint32(len(b)), protocol.ResponseHeader{
			Action:        protocol.ActionAnnounce,
			TransactionId: requestHeader.TransactionId,
		}, protocol.IPv4AnnounceResponse{
			Interval: announceInterval,
			Leechers: leechers,
			Seeders:  seeders,
		}, b)

		if err != nil {
			logger.Error("Unable to respond to client with announce response", zap.Error(err))
			return
		}
	case protocol.ActionScrape:
		// TODO: implement scraping
		err := server.RespondWithError(remoteAddr, requestHeader.TransactionId, fmt.Sprintf("Unsupported"))
		if err != nil {
			logger.Error("Unable to respond to client with error", zap.Error(err))
			return
		}
	default:
		logger.Warn("Unknown action")

		err := server.RespondWithError(remoteAddr, requestHeader.TransactionId, fmt.Sprintf("Unknown Action: %d", requestHeader.Action))
		if err != nil {
			logger.Error("Unable to respond to client with error", zap.Error(err))
			return
		}
	}
}
