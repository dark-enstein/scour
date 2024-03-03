package invoke

// RespHeaders defines the structure for storing HTTP response headers.
type RespHeaders struct {
	RespCode                      string // HTTP response status code.
	Protocol                      string // Protocol used in the response (e.g., HTTP/1.1).
	Date                          string // Date of the response.
	ContentType                   string // MIME type of the response content.
	ContentLength                 string // Length of the response content in bytes.
	Connection                    string // Connection status.
	Server                        string // Server information.
	AccessControlAllowOrigin      string // Allowed origins for cross-origin requests.
	AccessControlAllowCredentials bool   // Indicates if credentials are allowed in cross-origin requests.
}

// NewHeaders creates a new instance of RespHeaders with provided header values.
// ACAllowCred is interpreted as a boolean based on its string value.
func NewHeaders(code, proc, date, cType, cLength, conn, server, ACAllowOrigin, ACAllowCred string) *RespHeaders {
	var ac bool
	switch ACAllowCred {
	case "true":
		ac = true
	case "false":
		ac = false
	}
	return &RespHeaders{
		RespCode:                      code,
		Protocol:                      proc,
		Date:                          date,
		ContentType:                   cType,
		ContentLength:                 cLength,
		Connection:                    conn,
		Server:                        server,
		AccessControlAllowOrigin:      ACAllowOrigin,
		AccessControlAllowCredentials: ac,
	}
}

// Response is a custom type for representing HTTP response status codes.
type Response int

// NewResponse creates a new instance of Response from an integer status code.
func NewResponse(i int) *Response {
	r := Response(i)
	return &r
}

// String provides a string representation of the Response.
// Currently, it only recognizes the 200 OK status code.
func (r Response) String() string {
	switch int(r) {
	case 200:
		return "OK"
	}
	return "Response code unrecognized"
}
