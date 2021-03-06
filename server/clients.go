package server

import (
	"time"

	"erri120/gotracker/protocol"
)

type ConnectedClient struct {
	ConnectionId protocol.ConnectionId
	TimeIdIssued time.Time
}

func (connection ConnectedClient) IsValid() bool {
	now := time.Now()
	diff := now.Sub(connection.TimeIdIssued)
	return diff < connectionIdLifetime
}
