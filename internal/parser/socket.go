package parser

import (
	"context"
	"log"
	"strings"
)

// Socket struct represents the components of a socket connection.
type Socket struct {
	rawString string // The original socket string.
	protocol  string // The protocol used (e.g., "ws" for WebSocket).
	host      string // The host address.
	path      string // The path component of the socket connection.
	resource  string // The specific resource being accessed on the socket.
	err       error  // Any error encountered during socket parsing.
}

// NewSocket creates a new Socket instance from a raw string.
func NewSocket(s string) *Socket {
	return &Socket{rawString: s}
}

// Resolve parses the raw socket string into its component parts.
// It handles different formats and logs errors if the format is incorrect.
func (s *Socket) Resolve(ctx context.Context) {
	splice := strings.Split(s.rawString, ":")
	switch {
	case len(splice) == 1:
		// If there's only one part, it's assumed to be the path.
		s.path = splice[0]
	case len(splice) == 2:
		// If there are two parts, the first is the path, and the second is the resource.
		s.path = splice[0]
		s.resource = splice[1]
	case len(splice) > 2:
		// More than two parts indicate a malformed URL.
		if ParseLogLevelFromCtx(ctx, KeyV) == true {
			log.Printf("Error occurred: url %s malformed. Expecting only two colons in URL\n", s.rawString)
			// It's important to note that the function returns without setting the error field in the Socket struct.
			return
		}
	}
}
