package invoke

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/dark-enstein/scour/internal/utils"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	CONSOLE_CLEAR = iota // Represents a command to clear the console.
	CONSOLE_ERR          // Represents a console error.
	CONSOLE_EXIT         // Represents a command to exit the console.
)

var (
	// Format string for Unix socket session summary.
	UNIXSUMMARY = `
Session summary
Session ID: %s
All Socket Communications:
%s
Errors:
%s
`
	DEFAULT_LIMIT = 5                // Default limit for iterative operations.
	RCV_PAGESIZE  = 1024             // Default page size for receiving data.
	CONN_TIMEOUT  = time.Duration(2) // Default connection timeout duration.
)

var (
	ERR_PATHNOTSOCKET = errors.New("file path not a socket") // Error when the provided path is not a socket.
)

// UnixSock establishes a Unix socket connection and initiates a console session.
// It returns a summary of the session including any errors encountered.
func UnixSock(ctx context.Context, url parser.Url, it bool) ([]byte, error) {
	var mux sync.Mutex
	mux.Lock()
	if isSoc, err := utils.IsSocket(url.Path()); !isSoc || err != nil {
		return []byte{}, ERR_PATHNOTSOCKET
	}
	mux.Unlock()
	conn, err := net.Dial("unix", url.Path())
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Println("error encountered closing socket connection:", err)
		}
	}(conn)
	if err != nil {
		log.Printf("Error connecting to unix socket %s: %s\n", url.Path(), err.Error())
		return nil, err
	}

	id, _, communication, err := NewConsole(ctx, conn, url, it).Enter()
	flat := flatten(communication, []byte("\n"))
	fmt.Printf(UNIXSUMMARY, id, flat, err.Error())
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

// Console represents a console session over a network connection.
type Console struct {
	ctx      context.Context
	url      parser.Url
	resource []byte
	it       bool
	conn     net.Conn
	recurse  recurse
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
func NewConsole(ctx context.Context, conn net.Conn, url parser.Url, it bool) *Console {
	return &Console{ctx: ctx, url: url, resource: []byte(url.Resource()), conn: conn, recurse: recurse{limit: DEFAULT_LIMIT}, it: it}
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
			_, errSend := c.socSend(ctx)
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
		//path, resource := c.url.Path(), c.url.Resource()
		_, err := c.socSend(ctx)
		communication = append(communication, c.url.Bytes())
		if err != nil {
			return sessUUID, 1, communication, err
		}
		respBuf, err := c.socRcv(ctx)
		err = fmt.Errorf("%s", err)
		c.Unlock()
		communication = append(communication, respBuf)
	}
	return sessUUID, 0, communication, err

}

// socRcv handles receiving data from the network connection.
// It returns the received data and any error encountered.
func (c *Console) socRcv(ctx context.Context) ([]byte, error) {
	color.Green("<< rcvin:\n")
	stream, err := io.ReadAll(c.conn)
	if err != nil {
		log.Println("error encountered from socker connection:", err.Error())
		return nil, err
	}
	return stream, nil
}

// socSend handles sending data over the network connection.
// It returns the number of bytes sent and any error encountered.
func (c *Console) socSend(ctx context.Context) (int, error) {
	color.Yellow(">> sending to %s: %s\n", c.url.Path(), c.resource)
	_, err := c.Write([]byte(c.resource))
	if err != nil {
		fmt.Printf("&> Error sending message: %s\n", err.Error())
		if c.it && retrySend(string(c.resource)) {
			if c.recurse.Can() {
				c.recurse.Iter()
				_, err = c.socSend(ctx)
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
