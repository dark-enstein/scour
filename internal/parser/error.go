package parser

func IsIPv6Unsupported(e error) bool {
	if e != ErrIPv6tError {
		return false
	}
	return true
}
