package config

import (
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/exp/slices"
	"net/http"
)

const (
	MODE_HTTP = iota + 1
	MODE_SOCKET
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

// Flags struct holds the flag values passed in via the commandline to Scour.
type Flags struct {
	// Verbose to turn on/off the verbose output mode
	Verbose bool
	// Method denotes the http method the current http request is using
	Method string
	// Data denotes the payload to be sent to the server parsed via command line
	Data string
	// Headers denotes the header information to be sent to the server
	Headers string
	// UnixSocket flag sets scour into unixsocket mode
	UnixSocket bool
	// InteractiveMode opens scour console where requests can be sent and received interactively
	InteractiveMode bool
}

// NewFlags is a consuructor function for Flags
func NewFlags() *Flags {
	return &Flags{}
}

// ValidateAll implements validation for Flags values
func (f *Flags) ValidateAll() error {
	if !slices.Contains(AllSupportedConn, f.Method) {
		return fmt.Errorf("connection type \"%s\" passed is not supported. please pass in a supported type: GET, DELETE, PUT, POST. Use --unix-socket or --abstract-unix-socket flags for socket connection", f.Method)
	}
	if f.UnixSocket {
		color.Green("Socket mode enabled")
		f.Method = MethodSocket
	}
	return nil
}

// Resolve resolves the mode of the current request
func (f *Flags) Resolve() int {
	if f.UnixSocket {
		return MODE_SOCKET
	}
	return MODE_HTTP
}
