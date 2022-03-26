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
	maxPeerCount         = 50
	defaultPeerCount     = 10
)

type Server struct {
	Logger           *zap.Logger
	Conn             *net.UDPConn
	ConnectedClients map[protocol.ConnectionId]ConnectedClient
	Torrents         map[protocol.InfoHash]Torrent
}

func (server *Server) Respond(remoteAddr *net.UDPAddr, bufSize uint32, responseHeader protocol.ResponseHeader, parts ...interface{}) error {
	bytes, err := protocol.Marshal(bufSize, append([]interface{}{responseHeader}, parts...)...)
	if err != nil {
		return err
	}

	n, err := server.Conn.WriteToUDP(bytes, remoteAddr)
	if err != nil {
		return err
	}

	if n != len(bytes) {
		return fmt.Errorf("Wrote %d instead of %d bytes", n, len(bytes))
	}

	return nil
}

func (server *Server) RespondWithError(remoteAddr *net.UDPAddr, transationId protocol.TransactionId, message string) error {
	b := []byte(message)
	return server.Respond(remoteAddr, protocol.SizeOfResponseHeader+uint32(len(b)), protocol.ResponseHeader{
		Action:        protocol.ActionError,
		TransactionId: transationId,
	}, b)
}

func (server *Server) Close() error {
	if server.Conn == nil {
		return nil
	}

	err := server.Conn.Close()
	server.Conn = nil
	return err
}

func (server *Server) RegisterNewConnection() protocol.ConnectionId {
	// connection ids should not be guessable by the client, can look into crypto/rand
	connectionId := protocol.ConnectionId(rand.Int63())

	server.ConnectedClients[connectionId] = ConnectedClient{
		ConnectionId: connectionId,
		TimeIdIssued: time.Now(),
	}

	return connectionId
}

func (server *Server) UnregisterConnection(connectionId protocol.ConnectionId) {
	delete(server.ConnectedClients, connectionId)
}

func (server *Server) IsConnected(connectionId protocol.ConnectionId) bool {
	_, ok := server.ConnectedClients[connectionId]
	return ok
}

func (server *Server) IsValidConnection(connectionId protocol.ConnectionId) bool {
	connection, ok := server.ConnectedClients[connectionId]
	if !ok {
		return false
	}

	return connection.IsValid()
}

func (server *Server) StartCleanup(sleepTime time.Duration) {
	for {
		for _, connection := range server.ConnectedClients {
			if connection.IsValid() {
				continue
			}

			server.UnregisterConnection(connection.ConnectionId)
		}

		// TODO: also cleanup peers of torrents, would need to add timeAdded to peers or something

		time.Sleep(sleepTime)
	}
}

func (server *Server) Listen(addr *net.UDPAddr) (err error) {
	if server.Conn != nil {
		return fmt.Errorf("Server is already listening!")
	}

	// TODO: UDP over IPv4 vs IPv6, the network name has to be changed to "udp4" or "udp6"
	listener, err := net.ListenUDP("udp4", addr)
	if err != nil {
		server.Logger.Error("Unable to start listening for packages", zap.Error(err))
		return
	}

	server.Conn = listener
	// called in Server.Close()
	// defer listener.Close()

	data := make([]byte, blockSize)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)

		if n < int(protocol.SizeOfRequestHeader) {
			server.Logger.Error("Received package is too small", zap.Int("bytes", n))
			continue
		}

		if err != nil {
			server.Logger.Error("Unable to read next package", zap.Error(err))
			continue
		}

		// TODO: handle each request in a goroutine?
		// copying the slice might not be the best approach
		// we could do the switch statement here or in the HandleRequest function
		// performance vs memory usage, I guess
		// go server.HandleRequest(remoteAddr, data[:n])

		reader := bytes.NewReader(data[:n])

		var requestHeader protocol.RequestHeader
		if err = protocol.Unmarshal(reader, &requestHeader); err != nil {
			server.Logger.Error("Unable to unmarshal request header", zap.Error(err))
			continue
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
				continue
			}

			// TODO: maybe check if the Client is already connected based on the remote address

			connectionId := server.RegisterNewConnection()

			err = server.Respond(remoteAddr, protocol.SizeOfConnectResponse, protocol.ResponseHeader{
				Action:        protocol.ActionConnect,
				TransactionId: requestHeader.TransactionId,
			}, connectionId)

			if err != nil {
				logger.Error("Unable to respond to client with connect response", zap.Error(err))
				continue
			}
		case protocol.ActionAnnounce:
			if !server.IsConnected(requestHeader.ConnectionId) {
				logger.Warn("Client tried to announce with unregistered connection id")
				continue
			}

			if !server.IsValidConnection(requestHeader.ConnectionId) {
				logger.Warn("Client tired to use an expired connection id")
				server.UnregisterConnection(requestHeader.ConnectionId)
				continue
			}

			// TODO: IPv4/IPv6

			if n < int(protocol.SizeOfRequestHeader+protocol.SizeOfIPv4AnnounceRequest) {
				logger.Error("Client send not enough data for an announce request",
					zap.Int("diff", n-int(protocol.SizeOfRequestHeader+protocol.SizeOfIPv4AnnounceRequest)),
				)
				continue
			}

			var announceRequest protocol.IPv4AnnounceRequest
			if err = protocol.Unmarshal(reader, &announceRequest); err != nil {
				logger.Error("Unable to unmarshal IPv4 announce request", zap.Error(err))
				continue
			}

			if n > int(protocol.SizeOfRequestHeader+protocol.SizeOfIPv4AnnounceRequest) {
				// check for BEP 41 as an extension
				url, err := protocol.ExtractExtensionData(reader)
				if err != nil {
					logger.Error("Unable to extract extension data", zap.Error(err))
					continue
				}

				// TODO: do something with the url
				logger.Debug("Extracted extension data", zap.String("url", url))
			}

			torrent, ok := server.Torrents[announceRequest.InfoHash]
			if !ok {
				logger.Warn("Client tried to announce unknown torrent", zap.Binary("infoHash", announceRequest.InfoHash[:]))
				continue
			}

			// TODO: IPv4/IPv6
			peers := protocol.IPv4Peers(torrent.GetPeers(announceRequest.NumWanted))

			// add current client as peer
			torrent.AddPeer(remoteAddr)

			b, err := peers.MarshalBinary()
			if err != nil {
				logger.Error("Unable to marshal IPv4 peers", zap.Error(err))
				continue
			}

			err = server.Respond(remoteAddr, protocol.SizeOfIPv4AnnounceResponse+uint32(len(b)), protocol.ResponseHeader{
				Action:        protocol.ActionAnnounce,
				TransactionId: requestHeader.TransactionId,
			}, protocol.IPv4AnnounceResponse{
				Interval: announceInterval,
				Leechers: torrent.Leechers,
				Seeders:  torrent.Seeders,
			}, b)

			if err != nil {
				logger.Error("Unable to respond to client with announce response", zap.Error(err))
				continue
			}
		case protocol.ActionScrape:
			// TODO: implement scraping
			err = server.RespondWithError(addr, requestHeader.TransactionId, fmt.Sprintf("Unsupported"))
			if err != nil {
				logger.Error("Unable to respond to client with error", zap.Error(err))
				continue
			}
		default:
			logger.Warn("Unknown action")

			err = server.RespondWithError(addr, requestHeader.TransactionId, fmt.Sprintf("Unknown Action: %d", requestHeader.Action))
			if err != nil {
				logger.Error("Unable to respond to client with error", zap.Error(err))
				continue
			}
		}
	}
}
