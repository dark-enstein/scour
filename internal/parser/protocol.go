package parser

import (
	"fmt"
	"github.com/dark-enstein/scour/internal/config"
	"strings"
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
		return &Proctocol{s, config.HTTPVer}
	case "https":
		return &Proctocol{s, config.HTTPSVer}
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
