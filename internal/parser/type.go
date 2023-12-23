package parser

import (
	"context"
)

type Url interface {
	String() string
	Bytes() []byte
	Protocol() *Proctocol
	Host() string
	Port() string
	Path() string
	Err() error
	Resolve(context.Context)
}
