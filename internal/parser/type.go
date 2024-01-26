package parser

import (
	"context"
)

// Url is an interface that defines methods for working with URLs.
type Url interface {
	String() string          // String returns the URL as a string.
	Bytes() []byte           // Bytes returns the URL as a byte slice.
	Protocol() *Proctocol    // Protocol returns a Proctocol instance representing the URL's protocol.
	Host() string            // Host returns the host component of the URL.
	Port() string            // Port returns the port component of the URL.
	Path() string            // Path returns the path component of the URL.
	Err() error              // Err returns any error encountered during URL parsing or resolution.
	Resolve(context.Context) // Resolve parses the URL into its components within a given context.
}
