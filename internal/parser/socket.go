package parser

import (
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/utils"
	"log"
	"regexp"
	"strings"
)

var (
	// No agreed concensus on socket url format, but going with this. UnixTransportOnlyEx and UnixPlusHTTPSocketEx defines template format used by scour for parsin Unix socket url, and params.
	UnixTransportOnlyEx  = "unix:/path/to/socket.sock"                        // https://github.com/whatwg/url/issues/577
	UnixHTTPPlusSocketEx = "http://localhost:[/path/to/socket.sock]/resource" // https://github.com/whatwg/url/issues/577
)

const (
	PROC_TRANSPORTONLY = iota
	PROC_HTTPPLUSTRANSPORT
)

type Socket struct {
	// rawString holds the socket url string as parsed from user cli options
	rawString string
	// host defined the referenced unix host. By default it is localhost.
	host string
	// pathToSocket defines the path to the UNIX socket
	pathToSocket string
	// resource defines the api resource to be accessed
	resource string
	// t defines the class of unix urls received. It is either PROC_TRANSPORTONLY or PROC_HTTPPLUSTRANSPORT
	t int
	// err defines the socket error
	err error
}

// Creates a Socket resource
func NewSocket(s string) *Socket {
	return &Socket{rawString: s}
}

// Resolve resolves the UNIX socket url string into the Socket struct. It returns an error on encountering any issue.
func (s *Socket) Resolve(ctx context.Context) error {
	if v, err := Check(s.rawString); !v || err != nil {
		log.Printf("error: %s\n", err.Error())
		s.err = fmt.Errorf("error: %sn", err.Error())
		return s.err
	}
	splice := strings.Split(s.rawString, ":")
	if len(splice) == 0 {
		log.Printf("error: unrecognized url format: %s\n", splice[1])
		s.err = fmt.Errorf("error unrecognized url format: %sn", splice[1])
		return s.err
	}
	if splice[1] == "unix" {
		log.Printf("recognized protocol: %s\n", splice[1])
		s.t = PROC_TRANSPORTONLY
	} else if splice[1] == "http" {
		log.Printf("recognized protocol: %s\n", splice[1])
		s.t = PROC_HTTPPLUSTRANSPORT
	} else {
		log.Printf("error: unrecognized protocol: %s\n", splice[1])
		s.err = fmt.Errorf("error: unrecognized protocol: %sn", splice[1])
		return s.err
	}
	splice = splice[1:]
	switch s.t {
	case PROC_TRANSPORTONLY:
		s.pathToSocket = splice[0]
	case PROC_HTTPPLUSTRANSPORT:
		s.host = splice[0]
		s.pathToSocket = splice[1]
		resourcePath := strings.Split(splice[1], "sock")
		if !(len(resourcePath) == 2) {
			log.Printf("error: unrecognized url format: %s\n", splice[1])
			s.err = fmt.Errorf("error: unrecognized url format: %sn", splice[1])
			return s.err
		}
		s.pathToSocket, s.resource = resourcePath[0]+"sock", resourcePath[1]
		if !utils.IsSocket(s.pathToSocket) {
			log.Printf("error: invalid socket file: %s\n", s.pathToSocket)
			s.err = fmt.Errorf("error: invalid socket path: %sn", s.pathToSocket)
			return s.err
		}
	}
	log.Printf("Resolved URL: %#v\n", s)
	return nil
}

// IsValid does the socket url validation, attempting to resolve it, and returning false and the error on ony error
func (s *Socket) IsValid(ctx context.Context) (bool, error) {
	if err := s.Resolve(ctx); err == nil {
		return false, err
	}
	return true, nil
}

func Check(sUrl string) (bool, error) {
	if strings.Contains(sUrl, " ") {
		return false, fmt.Errorf("url contains string: %s\n", sUrl)
	}
	re := regexp.MustCompile("^(unix:/[^ ]+|http://localhost:/?[^ ]+/*)$")
	if !re.MatchString(sUrl) {
		return false, fmt.Errorf("hostname: %s invalid\n", sUrl)
	}
	return true, nil
}

func (s *Socket) Error() string {
	return fmt.Sprintf("error with socket. msg: [%s] host: [%s] path: %s resource: [%s] rawString: [%s]\n", s.err, s.host, s.pathToSocket, s.resource, s.rawString)
}
func (s *Socket) Err(e error) string {
	s.err = e
	return s.Error()
}
