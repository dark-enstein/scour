package parser

// IsIPv6Unsupported checks if the given error is specifically an IPv6 unsupported error.
// It returns true if the error is exactly ErrIPv6tError, indicating an issue with IPv6 support.
// Otherwise, it returns false, suggesting that the error is of a different type or nil.
func IsIPv6Unsupported(e error) bool {
	// Check if the provided error is the specific IPv6 unsupported error (ErrIPv6tError).
	if e != ErrIPv6tError {
		return false // Return false if the error is different.
	}
	return true // Return true if the error matches ErrIPv6tError.
}
