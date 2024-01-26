package parser

import (
	"context"
	"errors"
	"github.com/dark-enstein/scour/internal/config"
	"log"
	"regexp"
	"strings"
)

// Key and default values used in the package.
var (
	KeyV               = "VERBOSE"                            // Key for checking verbose logging in context.
	DefaultHTTPPort    = "80"                                 // Default port for HTTP.
	DefaultHTTPSPort   = "443"                                // Default port for HTTPS.
	ErrIPV6            = "IPv6 Address detected: Unsupported" // Error message for IPv6.
	ErrIPv6tError      = errors.New(ErrIPV6)                  // Error for unsupported IPv6 addresses.
	ErrHostnameInvalid = errors.New("Hostname invalid")       // Error for invalid hostname.
)

// HTTP struct represents the components of a URL.
type HTTP struct {
	rawString string // The original URL string.
	protocol  string // The protocol (http/https) used.
	host      string // The hostname or IP address.
	port      string // The port number.
	path      string // The path component of the URL.
	err       error  // Any error encountered during URL parsing.
}

// String method returns the raw URL string.
func (u *HTTP) String() string {
	return u.rawString
}

// Bytes method returns the raw URL as a byte slice.
func (u *HTTP) Bytes() []byte {
	return []byte(u.rawString)
}

// Protocol method returns a new Protocol struct representing the URL's protocol.
func (u *HTTP) Protocol() *Proctocol {
	return NewProtocol(u.protocol)
}

// Host method returns the URL's host component.
func (u *HTTP) Host() string {
	return u.host
}

// Port method returns the URL's port component.
func (u *HTTP) Port() string {
	return u.port
}

// Path method returns the URL's path component.
func (u *HTTP) Path() string {
	return u.path
}

// Err method returns any error encountered during URL parsing.
func (u *HTTP) Err() error {
	return u.err
}

// NewUrl creates a new HTTP struct from a raw URL string.
// It resolves the URL components and handles default values for protocol and port.
func NewUrl(ctx context.Context, url string) (*HTTP, error) {
	u := &HTTP{}
	split := strings.Split(url, ":")
	if len(split) < 2 {
		re := regexp.MustCompile("^([a-zA-Z0-9]+(\\.[a-zA-Z0-9]+)+.*)$")
		if !re.MatchString(url) {
			return nil, ErrHostnameInvalid
		}
		u.rawString = "https://" + url
	} else {
		u.rawString = url
	}
	u.Resolve(ctx)
	return u, u.err
}

// Resolve resolves a raw URL string into its components and updates the HTTP struct.
func (u *HTTP) Resolve(ctx context.Context) {
	v, proc, h, po, pa, err := resolve(ctx, u.rawString)
	u.err = err
	if !v && err != nil && ParseLogLevelFromCtx(ctx, KeyV) == true {
		log.Printf("Error occured: %s\n", err.Error())
		return
	}
	if ParseLogLevelFromCtx(ctx, KeyV) == true {
		log.Printf("Error: %v\n", u.err)
	}
	u.path, u.port, u.host, u.protocol = pa, po, h, proc
}

// expectation contract: http://eu.httpbin.org:80/get || http://eu.httpbin.org/get
// resolve does the low-level resolution of a URL string into its component parts.
func resolve(ctx context.Context, u string) (valid bool, protocol, host, port, path string, err error) {
	urlColSplit := strings.Split(u, ":")
	switch {
	case len(urlColSplit) == 2:
		valid = true
		protocol = urlColSplit[0]
		switch protocol {
		case config.HTTP:
			port = DefaultHTTPPort
		case config.HTTPS:
			port = DefaultHTTPSPort
		}
		hostPath := strings.TrimLeft(urlColSplit[1], "/")
		hostPathArr := strings.Split(hostPath, "/")
		if len(hostPathArr) < 2 {
			host, path, err = hostPathArr[0], "/", nil
		} else {
			host, path, err = hostPathArr[0], strings.Join(hostPathArr[1:], "/"), nil
		}
	case len(urlColSplit) == 3:
		valid = true
		protocol = urlColSplit[0]
		host = strings.TrimLeft(urlColSplit[1], "/")
		portPath := strings.Split(urlColSplit[2], "/")
		port, path, err = portPath[0], strings.Join(portPath[1:], "/"), nil
	case len(urlColSplit) > 3:
		err = ErrIPv6tError
		return
	}
	return
}

// ParseLogLevelFromCtx extracts the verbose logging setting from a context.
func ParseLogLevelFromCtx(ctx context.Context, k string) bool {
	return ctx.Value(k).(bool)
}
