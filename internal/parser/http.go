package parser

import (
	"context"
	"errors"
	"github.com/dark-enstein/scour/internal/config"
	"log"
	"regexp"
	"strings"
)

var (
	KeyV = "VERBOSE"
)

var (
	DefaultHTTPPort    = "80"
	DefaultHTTPSPort   = "443"
	ErrIPV6            = "IPv6 Address detected: Unsupported"
	ErrIPv6tError      = errors.New(ErrIPV6)
	ErrHostnameInvalid = errors.New("Hostname invalid")
)

type HTTP struct {
	rawString string
	protocol  string
	host      string
	port      string
	path      string
	err       error
}

func (u *HTTP) String() string {
	return u.rawString
}

func (u *HTTP) Bytes() []byte {
	return []byte(u.rawString)
}

func (u *HTTP) Protocol() *Proctocol {
	return NewProtocol(u.protocol)
}

func (u *HTTP) Host() string {
	return u.host
}

func (u *HTTP) Port() string {
	return u.port
}

func (u *HTTP) Path() string {
	return u.path
}

func (u *HTTP) Err() error {
	return u.err
}

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

// Resolve is the high level api that resolves a raw url string into HTTP
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
// resolve does the low-level resolution of url string into it's component part
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

func ParseLogLevelFromCtx(ctx context.Context, k string) bool {
	return ctx.Value(k).(bool)
}
