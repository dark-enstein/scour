package config

import (
	"fmt"
	"golang.org/x/exp/slices"
	"net/http"
)

var (
	HTTPVer          = "1.1"
	HTTPSVer         = "1.1"
	HTTP             = "http"
	HTTPS            = "https"
	MethodSocket     = "SOCKET"
	MethodAbsSocket  = "ABSSOCKET"
	AllSupportedConn = []string{MethodSocket, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete}
	Help             = `
    Usage:
	scour [flags] <url>

	Flags:	
	--verbose or -v: Enable verbose mode.
	-X: Specify the request method (GET, POST, etc.).
	-d: Pass request data.
	-H: Custom request headers.
	--unix-socket or -aus: Use an Unix domain socket.
	--abstract-unix-socket or -aus: Use an abstract Unix domain socket.

	Example:
    scour -v -X GET https://example.com
`
)

type Flags struct {
	Verbose         bool
	Method          string
	Data            string
	Headers         string
	UnixSocket      bool
	InteractiveMode bool
}

func NewFlags() *Flags {
	return &Flags{}
}

func (f *Flags) ValidateAll() error {
	if !slices.Contains(AllSupportedConn, f.Method) {
		return fmt.Errorf("connection type %s passed is not supported. please pass in a supported type: GET, DELETE, PUT, POST. Use --unix-socket or --abstract-unix-socket flags for socket connection.\n")
	}
	return nil
}
