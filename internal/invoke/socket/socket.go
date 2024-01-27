package socket

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unicode"
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
	SOCKET_GET    = "get"
)

var (
	ERR_PATHNOTSOCKET = errors.New("file resource not a socket") // Error when the provided resource is not a socket.
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
	if err != nil {
		log.Printf("Error connecting to unix socket %s: %s\n", url.Path(), err.Error())
		return nil, err
	}
	defer func(conn net.Conn) {
		err = conn.Close()
		if err != nil {
			log.Println("error encountered closing socket connection:", err)
		}
	}(conn)

	_, _, communication, err := NewConsole(ctx, conn, url, it).Enter()
	flat := flatten(communication, []byte("\n"))
	//fmt.Printf(UNIXSUMMARY, id, flat, err.Error())
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
			c.resource = lineReq
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
		//resource, resource := c.url.Path(), c.url.Resource()
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
	color.Green("<< receiving:\n")
	stream, err := io.ReadAll(c.conn)
	if err != nil {
		log.Println("error encountered from socket connection:", err.Error())
		return nil, err
	}
	fmt.Printf("Received stream: %s\n", stream)
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

type Creator struct {
	socket   string
	resource string
	jsonLoc  string
	sync.Mutex
}

func NewCreator() *Creator {
	return &Creator{}
}

func (c *Creator) StartSocket(ctx context.Context, socketUpchan chan struct{}, errChan chan error) {
	// delete if socket file already exists
	_, err := os.Stat(c.socket)
	//fmt.Println("Socket stat:", err, c.socket)
	if !errors.Is(err, os.ErrNotExist) {
		fmt.Println("Socket exist. Removing...")
		err := os.RemoveAll(c.socket)
		if err != nil {
			log.Println("failed to delete used socket")
		}
	}

	//<-time.After(2 * time.Second)

	// Create a Unix domain socket and listen for incoming connections.
	socket, err := net.Listen("unix", c.socket)
	if err != nil {
		log.Println("failed to create socket:", err)
		errChan <- err
		return
	}
	defer socket.Close()

	// return a response via channel once socket is created
	socketUpchan <- struct{}{}

	// Cleanup the socketfile on syscall.SIGTERM in a separate go routine
	ch := make(chan os.Signal, 1)
	wg := sync.WaitGroup{}
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ch
		fmt.Println("detected interrupt signal. cleaning socket")
		err := os.RemoveAll(c.socket)
		if err != nil {
			log.Println("failed to delete used socket")
		}
		return
	}()

	var expectingConn = true

	// serverLoop
	for expectingConn {
		// Accept an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Println("error accepting connection:", err.Error())
			errChan <- err
			return
		}

		// Handle the connection in a separate goroutine.
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Read all client request body
			log.Println("handling api request")
			var cli = make(chan []byte)
			var cliBytes = []byte{}
			var alive bool
			go func(stream chan []byte) {
				clientStream, err := io.ReadAll(conn)
				if err != nil {
					log.Printf("error receiving client bytes stream: %s\n", err.Error())
				}
				stream <- clientStream
				return
			}(cli)
			select {
			case cliBytes = <-cli:
				fmt.Println("cli bytes:", cliBytes)
				if err != nil {
					errChan <- err
				}
				alive = true
			case <-time.After(5 * time.Second):
				log.Println("socket timeout. socket stayed up for too long without receiving client connection")
				alive = false
			}

			if alive {
				api, err := handleSocketApi(cliBytes)
				if err != nil {
					log.Println("api response not equals nil")
					errorBytes := []byte(fmt.Sprintf("ERROR: %s\n", cliBytes))
					c.Lock()
					_, err = conn.Write(errorBytes)
					c.Unlock()
					if err != nil {
						errChan <- err
					}
					return
				}
				log.Println("handling api request successful")

				// Echo the api response back to the client
				c.Lock()
				_, err = conn.Write([]byte(api))
				c.Unlock()
				if err != nil {
					errChan <- err
				}
			} else {
				log.Println("client dead. closing connection")
				c.Lock()
				expectingConn = false
				c.Unlock()
				return
			}
			conn.Close()
			return
		}()
	}
	//ch <- syscall.SIGTERM
	wg.Wait()
}

// handleSocketApi handles the client connection and according toe the resource requested TODO: implement this in the http.Handler/HandlerFunc fashion
func handleSocketApi(b []byte) (string, error) {
	//b = []byte("/" + SOCKET_GET)
	if b[0] != []byte("/")[0] {
		return "", fmt.Errorf("url string invalid\n")
	}

	// remove the protocol if appended
	splice := bytes.Split(b, []byte("/"))
	len := len(splice)

	// check for the parent resource of the resource
	switch string(splice[1]) {
	case SOCKET_GET:
		//fmt.Println("in socket get")
		var childPath []byte
		// if sub resource is present
		if len > 2 {
			childPath = bytes.Join(splice[2:], []byte("/"))
		}
		b, err := getHandler(childPath)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}

type UUIDs struct {
	List []string `json:"uuid_list"`
}

func (u UUIDs) ByteSlice() (b [][]byte) {
	for i := 0; i < len(u.List); i++ {
		if i < len(u.List)-1 {
			b = append(b, []byte(u.List[i]+"\n"))
		} else {
			b = append(b, []byte(u.List[i]))
		}
	}
	return
}

func getHandler(path []byte) ([]byte, error) {
	// check is resource is absent, or it is present and empty
	if len(path) == 0 || path == nil {
		g, err := getGenAll()
		return tflatten(g), err
	}
	if !unicode.IsNumber(rune(path[0])) {
		return nil, fmt.Errorf("get subroute passed in isn't a number")
	}
	fmt.Println("index str:", string(path))
	i, _ := strconv.Atoi(string(path))
	return findIndexGen(i)
}

func findIndexGen(i int) ([]byte, error) {
	fmt.Println("index received:", i)
	if i < 0 || i > 15 {
		return nil, fmt.Errorf("index less than 0 or greater than 15\n")
	}
	b, err := getGenAll()
	if err != nil {
		return nil, err
	}

	return b[i], nil
}

func getGenAll() ([][]byte, error) {
	b, err := os.ReadFile(filepath.Join("./sock", "uuid.json"))
	if err != nil {
		return nil, err
	}
	var uds UUIDs
	err = json.Unmarshal(b, &uds)
	return uds.ByteSlice(), nil
}

func tflatten(bb [][]byte) (bflat []byte) {
	for i := 0; i < len(bb); i++ {
		bflat = append(bflat, bb[i]...)
	}
	return
}

// CreateSocketSubProc creates a socket for testing
func CreateSocketSubProc(name string) error {
	// delete socket if already exist
	_, err := os.Stat(name)
	if errors.Is(err, os.ErrExist) {
		log.Println("exists")
		err := os.RemoveAll(name)
		if err != nil {
			log.Println("Failed to remove socket")
			return errors.New("failed to remove socket")
		}
	}

	var cmdChan = make(chan string)
	defer close(cmdChan)
	const exitErrPrefix = "INTERNAL"

	wg := sync.WaitGroup{}
	// create socket TODO: make this function into a fullyfledged feature, if the approach will be to use nc, then as part of the init checks of scour, nc sould be confirmed to exist on the system it is run first
	wg.Add(1)
	go func(ch chan string) {
		args := fmt.Sprintf("echo -e this is the sample response | nc -lk -U %s", name)
		cmd := exec.Command("bash", "-c", args)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println("INTERNAL: could not create stdout pipe:", err.Error())
			ch <- fmt.Errorf("INTERNAL: could not create stdout pipe: %s\n", err.Error()).Error()
		}

		if err := cmd.Start(); err != nil {
			log.Println("INTERNAL: running command errored with:", err.Error())
			ch <- fmt.Errorf("INTERNAL: could not create socket due to: %s\n", err.Error()).Error()
		}

		output, err := io.ReadAll(stdout)
		if err != nil {
			log.Println("INTERNAL: could not read stdout pipe stream:", err.Error())
			ch <- fmt.Errorf("INTERNAL: could not read stdout pipe stream: %s\n", err.Error()).Error()
		}

		if err := cmd.Wait(); err != nil {
			log.Println("INTERNAL: could not create socket:", err.Error())
			ch <- fmt.Errorf("INTERNAL: could not create socket: %s\n", err.Error()).Error()
		}

		fmt.Printf("Created socket: %s\n", output)
		wg.Done()
	}(cmdChan)

	for output := range cmdChan {
		fmt.Println("Socket output:", output)
	}

	wg.Wait()
	return nil
}
