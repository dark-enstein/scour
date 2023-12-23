package parser

import (
	"context"
	"log"
	"strings"
)

type Socket struct {
	rawString string
	protocol  string
	host      string
	path      string
	resource  string
	err       error
}

func NewSocket(s string) *Socket {
	return &Socket{rawString: s}
}

func (s *Socket) Resolve(ctx context.Context) {
	splice := strings.Split(s.rawString, ":")
	switch {
	case len(splice) == 1:
		s.path = splice[0]
	case len(splice) == 2:
		s.path = splice[0]
		s.resource = splice[1]
	case len(splice) > 2:
		if ParseLogLevelFromCtx(ctx, KeyV) == true {
			log.Printf("Error occured: url %s malformed. Expecting only two colon in URL\n", s.rawString)
			return
		}
	}
}
