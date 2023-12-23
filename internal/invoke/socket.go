package invoke

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/google/uuid"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	CONSOLE_CLEAR = iota
	CONSOLE_ERR
	CONSOLE_EXIT
)

var (
	UNIXSUMMARY = `
Session summary
Session ID: %s
All Socket Communications:
%s
Errors:
%s
`
	DEFAULT_LIMIT = 5
	RCV_PAGESIZE  = 1024
	CONN_TIMEOUT  = time.Duration(2)
)

var (
	ERR_PATHNOTSOCKET = errors.New("file path not a socket")
)

// UnixSock establishes a Unix socket connection and initiates a console session.
// It returns a summary of the session including any errors encountered.
func UnixSock(ctx context.Context, path string, url *parser.HTTP, it bool) ([]byte, error) {
	var mux sync.Mutex
	mux.Lock()
	if !IsSocket(path) {
		return []byte{}, ERR_PATHNOTSOCKET
	}
	mux.Unlock()
	conn, err := net.Dial("unix", path)
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	if err != nil {
		log.Printf("Error connecting to unix socket %s: %s\n", path, err.Error())
		return nil, err
	}

	id, _, communication, err := NewConsole(ctx, conn, url, it).Enter()
	flat := flatten(communication, []byte("\n"))
	fmt.Printf(UNIXSUMMARY+"\n", id, flat)
	return flat, err
}

// flatten concatenates a slice of byte slices into a single byte slice,
// separated by the provided delimiter.
func flatten(b [][]byte, delim []byte) (bflat []byte) {
	for i := 0; i < len(b); i++ {
		bflat = append(bflat, b[i]...)
		bflat = append(bflat, delim...)
	}
	return
}

// IsSocket checks if the provided path is a Unix socket.
func IsSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("Error opening file: %s\n", err.Error())
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

// Console represents a console session over a network connection.
type Console struct {
	ctx     context.Context
	url     *parser.HTTP
	it      bool
	conn    net.Conn
	recurse recurse
	sync.Mutex
}

// recurse is a utility struct for managing iterative retry logic.
type recurse struct {
	iter  int
	limit int
}

// Reset resets the iteration counter of recurse to zero.
func (r *recurse) Reset() {
	r.iter = 0
}

// Iter increments the iteration counter of recurse by one.
func (r *recurse) Iter() {
	r.iter++
}

// SetLim sets the iteration limit for the recurse struct.
func (r *recurse) SetLim(n int) {
	r.limit = n
}

// Can check if the current iteration count is within the set limit.
func (r *recurse) Can() bool {
	return r.iter <= r.limit
}

// Write sends a message over the console's network connection.
func (c *Console) Write(msg []byte) (int, error) {
	return c.conn.Write(msg)
}

// Read receives a message from the console's network connection.
func (c *Console) Read(msg []byte) (int, error) {
	return c.conn.Read(msg)
}

// NewConsole creates a new Console instance with the given context, network connection,
// HTTP, and interactive mode flag.
func NewConsole(ctx context.Context, conn net.Conn, url *parser.HTTP, it bool) *Console {
	return &Console{ctx: ctx, url: url, conn: conn, recurse: recurse{limit: DEFAULT_LIMIT}, it: it}
}

// Enter starts the console session and handles communication based on the interactive mode.
// It returns the session ID, exit code, communication log, and any error encountered.
func (c *Console) Enter() (sessID string, code int, communication [][]byte, err error) {
	sessUUID := uuid.New().String()
	switch c.it {
	case true:
		var scn = bufio.NewScanner(os.Stdin)
		fmt.Printf("Starting console session: %s\n", sessUUID)
	cursor:
		for scn.Scan() {
			lineReq := scn.Bytes()
			if len(lineReq) == 0 {
				break cursor
			}
			if string(lineReq) == ":close" {
				return sessUUID, 0, communication, err
			}
			// Requests are delimited by "\"
			c.recurse.Reset()
			c.Lock()
			ctx, cancel := context.WithTimeout(c.ctx, CONN_TIMEOUT*time.Second)
			defer cancel()
			_, errSend := c.socSend(ctx, lineReq)
			err = fmt.Errorf("%s: %w", err, errSend)
			if err != nil {
				break cursor
			}
			communication = append(communication, lineReq)

			respBuf, errRcv := c.socRcv(ctx)
			err = fmt.Errorf("%s: %w", err, errRcv)
			c.Unlock()
			fmt.Printf("< %s\n", string(respBuf))
			communication = append(communication, respBuf)
		}
	case false:
		c.Lock()
		ctx, cancel := context.WithTimeout(c.ctx, CONN_TIMEOUT*time.Second)
		defer cancel()
		_, err := c.socSend(ctx, c.url.Bytes())
		communication = append(communication, c.url.Bytes())
		if err != nil {
			return sessUUID, 1, communication, err
		}
		respBuf, err2 := c.socRcv(ctx)
		err = fmt.Errorf("%s: %w", err, err2)
		c.Unlock()
		communication = append(communication, respBuf)
	}
	return sessUUID, 0, communication, err

}

// socRcv handles receiving data from the network connection.
// It returns the received data and any error encountered.
func (c *Console) socRcv(ctx context.Context) ([]byte, error) {
	fmt.Printf("< rcvin:\n")
	var buf bytes.Buffer
	tmp := make([]byte, RCV_PAGESIZE)
	for {
		n, err := c.conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error receiving response:", err.Error())
				buf.Write(tmp[:n])
				return buf.Bytes(), err
			}
			break
		}
		buf.Write(tmp[:n])
	}
	return buf.Bytes(), nil
}

// socSend handles sending data over the network connection.
// It returns the number of bytes sent and any error encountered.
func (c *Console) socSend(ctx context.Context, lineReq []byte) (int, error) {
	fmt.Printf("> %s:", string(lineReq))
	_, err := c.Write(lineReq)
	if err != nil {
		fmt.Printf("&> Error sending message: %s\n", err.Error())
		if c.it && retrySend(string(lineReq)) {
			if c.recurse.Can() {
				c.recurse.Iter()
				_, err = c.socSend(ctx, lineReq)
				if err != nil {
					return -1, fmt.Errorf("err: %w", err)
				}
				return 0, err
			}
		}
		return -1, err
	}
	return 0, nil
}

// retrySend prompts the user to decide whether to resend a request.
// It returns true if the user opts to retry, false otherwise.
func retrySend(req string) bool {
	var s string
	var help = `
Respond yes with "y|Y|YES|Yes|yes"
Respond no with "n|N|NO|No|no"
`
	scn := bufio.NewScanner(os.Stdin)
	i := 0
try:
	for {
		fmt.Printf("? Do you want to resend %v\n %s\n?", req, help)
		if !scn.Scan() {
			// If Scan returns false, we need to check for an error.
			if err := scn.Err(); err != nil {
				log.Printf("Error reading input: %v\n", err)
				return false
			}
			// If there's no error, it means we've reached EOF or similar.
			return false
		}

		s = scn.Text() // or scn.Bytes() if you want []byte
		switch s {
		case "y", "Y", "YES", "Yes", "yes":
			return true
		case "n", "N", "NO", "No", "no":
			return false
		default:
			fmt.Printf("Response %s unrecognized. %s\n", s, help)
			if i < 5 {
				i++
				break try
			}
			log.Println("Invalid response limit reached. Disabling retry")
		}
	}
	return false
}
