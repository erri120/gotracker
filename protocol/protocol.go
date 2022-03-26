package protocol

// Id generated by the Client, unique to each request.
type TransactionId int32

// Id generated by the Tracker, Clients need to supply this Id with each request.
type ConnectionId int64

// Action of the request, also has to be included in the response.
type Action int32

const (
	ActionConnect  Action = 0 // Used for Connection requests/responses.
	ActionAnnounce Action = 1 // Used for Announce requests/responses.
	ActionScrape   Action = 2 // Used for Scrape requests/responses.
	ActionError    Action = 3 // Used for error responses.
)

// 20 byte SHA1 hash to identify a torrent.
type InfoHash [20]byte

// 20 byte Id, generated by the Client for each new download.
type PeerId [20]byte
