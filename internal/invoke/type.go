package invoke

type RespHeaders struct {
	RespCode                      string
	Protocol                      string
	Date                          string
	ContentType                   string
	ContentLength                 string
	Connection                    string
	Server                        string
	AccessControlAllowOrigin      string
	AccessControlAllowCredentials bool
}

func newHeaders(code, proc, date, cType, cLength, conn, server, ACAllowOrigin string, ACAllowCred string) *RespHeaders {
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

type Response int

func NewResponse(i int) *Response {
	r := Response(i)
	return &r
}

func (r Response) String() string {
	switch int(r) {
	case 200:
		return "OK"
	}
	return "Response code unrecognized"
}
