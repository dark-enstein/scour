package parser

import (
	"context"
	"errors"
	"fmt"
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
	HTTPVer            = "1.1"
	HTTPSVer           = "1.1"
	HTTP               = "http"
	HTTPS              = "https"
)

type Protocol interface {
	String() string
	MustUpper() string
	MustLower() string
}

type Proctocol struct {
	t   string
	ver string
}

func NewProtocol(s string) *Proctocol {
	switch s {
	case "http":
		return &Proctocol{s, HTTPSVer}
	case "https":
		return &Proctocol{s, HTTPSVer}
	}
	return nil
}

func (p *Proctocol) String() string {
	return fmt.Sprintf("%s", p.t)
}

func (p *Proctocol) Stringln() string {
	return fmt.Sprintf("%s\f", p.t)
}

func (p *Proctocol) MustUpper() string {
	return strings.ToUpper(p.String())
}

func (p *Proctocol) MustLower() string {
	return strings.ToLower(p.String())
}

type URL struct {
	rawString string
	protocol  string
	host      string
	port      string
	path      string
	err       error
}

func (u *URL) String() string {
	return u.rawString
}

func (u *URL) Bytes() []byte {
	return []byte(u.rawString)
}

func (u *URL) ProtocolStruct() Proctocol {
	return *NewProtocol(u.protocol)
}

func (u *URL) Protocol() string {
	return u.protocol
}

func (u *URL) UProtocol() string {
	p := NewProtocol(u.protocol)
	return p.MustUpper()
}

func (u *URL) LProtocol() string {
	p := NewProtocol(u.protocol)
	return p.MustLower()
}

func (u *URL) Host() string {
	return u.host
}

func (u *URL) Port() string {
	return u.port
}

func (u *URL) Path() string {
	return u.path
}

func (u *URL) Err() error {
	return u.err
}

func NewUrl(ctx context.Context, url string) (*URL, error) {
	u := &URL{}
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

// Resolve is the high level api that resolves a raw url string into URL
func (u *URL) Resolve(ctx context.Context) {
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
		case HTTP:
			port = DefaultHTTPPort
		case HTTPS:
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
