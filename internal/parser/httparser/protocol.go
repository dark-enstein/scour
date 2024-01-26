package httparser

import (
	"fmt"
	"github.com/dark-enstein/scour/internal/config"
	"strings"
)

// Protocol is an interface defining methods for handling protocol strings.
type Protocol interface {
	String() string    // String returns the protocol as a string.
	MustUpper() string // MustUpper returns the protocol in uppercase.
	MustLower() string // MustLower returns the protocol in lowercase.
}

// Proctocol is a struct implementing the Protocol interface.
// It holds information about a specific protocol and its version.
type Proctocol struct {
	t   string // The protocol type (e.g., "http" or "https").
	ver string // The protocol version.
}

// NewProtocol creates a new Proctocol instance based on the provided string.
// It sets the protocol version based on the configuration in the config package.
func NewProtocol(s string) *Proctocol {
	switch s {
	case "http":
		return &Proctocol{s, config.HTTPVer}
	case "https":
		return &Proctocol{s, config.HTTPSVer}
	}
	return nil // Returns nil if the protocol is not recognized.
}

// String formats the protocol as a string.
func (p *Proctocol) String() string {
	return fmt.Sprintf("%s", p.t)
}

// Stringln formats the protocol as a string followed by a line feed.
func (p *Proctocol) Stringln() string {
	return fmt.Sprintf("%s\n", p.t)
}

// MustUpper converts the protocol string to uppercase.
func (p *Proctocol) MustUpper() string {
	return strings.ToUpper(p.String())
}

// MustLower converts the protocol string to lowercase.
func (p *Proctocol) MustLower() string {
	return strings.ToLower(p.String())
}
