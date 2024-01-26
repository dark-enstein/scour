package socketparser

import (
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser/httparser"
	"github.com/dark-enstein/scour/internal/utils"
	"log"
	"regexp"
	"strings"
)

const (
	REGEX_RESOURCE   = "^(https?:)(\\/[^\\/\\s]+)+$"
	SOCKET_ARG_DELIM = " "
)

// Socket struct represents the components of a socket connection.
type Socket struct {
	rawString string // The original socket string. expected format: /var/run/docker.sock http:/images/json TODO
	path      string // The path component of the socket connection.
	resource  string // The specific resource being accessed on the socket.
	err       error  // Any error encountered during socket parsing.
}

// NewSocket creates a new Socket instance from a raw string.
func NewSocket(ctx context.Context, s string) *Socket {
	soc := &Socket{rawString: s}
	soc.Resolve(ctx)
	return soc
}

// String method returns the raw Socket string.
func (s *Socket) String() string {
	return s.rawString
}

// Bytes method returns the raw Socket as a byte slice.
func (s *Socket) Bytes() []byte {
	return []byte(s.rawString)
}

// TODO: refactor Protocol, Host, Port, and the URL interface
// Protocol method returns a new Protocol struct representing the URL's protocol.
func (s *Socket) Protocol() *httparser.Proctocol {
	return httparser.NewProtocol("UNIMPLEMENTED")
}

// Host method returns the Socket's host component.
func (s *Socket) Host() string {
	return "UNIMPLEMENTED"
}

// Port method returns the Socket's port component.
func (s *Socket) Port() string {
	return "UNIMPLEMENTED"
}

// Path method returns the Socket's file path component.
func (s *Socket) Path() string {
	return s.path
}

// Resource method returns the Socket's resource component.
func (s *Socket) Resource() string {
	return s.resource
}

// Err method returns any error encountered during Socket parsing.
func (s *Socket) Err() error {
	return s.err
}

// Resolve parses the raw socket string into its component parts.
// It handles different formats and logs errors if the format is incorrect.
func (s *Socket) Resolve(ctx context.Context) {
	splice := strings.Split(s.rawString, SOCKET_ARG_DELIM)
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
		if httparser.ParseLogLevelFromCtx(ctx, httparser.KeyV) == true {
			log.Printf("Error occurred: url %s malformed. Expecting only two colons in URL\n", s.rawString)
			// It's important to note that the function returns without setting the error field in the Socket struct.
			return
		}
	}
}

// IsValid checks that the contents of the referenced Socket is valid.
func (s *Socket) IsValid() error {
	// check that socket path is valid
	validSocket, err := utils.IsSocket(s.path)
	if !validSocket || err != nil {
		return fmt.Errorf("path %s isn't a valid socket or doesn't exist\n", s.path)
	}

	// check if resource is a valid resource format
	validResource := isValidResource(s.resource)
	if !validResource {
		return fmt.Errorf("resource %s isn't a socket resource or doesn't exist\n", s.resource)
	}
	return nil
}

// isValidResource checks if the resource path passed in is valid: Regex: ^(https?:)(\/[^\/\s]+)+$
func isValidResource(resource string) bool {
	re, err := regexp.Compile(REGEX_RESOURCE)
	if err != nil {
		log.Println("INTERNAL_ERR: regexp not valid") // TODO: Use INTERNAL_ERR for internal error, and add issue creation workflow for those errors
	}

	return re.MatchString(resource)
}
