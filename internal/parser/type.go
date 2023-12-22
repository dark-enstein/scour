package parser

type URLS interface {
	String() string
	Bytes() []byte
	Protocol() Proctocol
	Host() string
	Port() string
	Path() string
	Err() error
}
