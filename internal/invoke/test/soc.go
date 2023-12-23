package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"unicode"
)

var (
	PRIMARY_GET = "get"
)

type input struct {
	socket string
	path   string
}

func main() {
	//_mainwithflags()
	// Create a Unix domain socket and listen for incoming connections.
	socket, err := net.Listen("unix", "/tmp/echo.sock")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove("/tmp/echo.sock")
		os.Exit(1)
	}()

	for {
		// Accept an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a separate goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			// Create a buffer for incoming data.
			buf := make([]byte, 512)
			log.Println("making buffer")

			// Read data from the client connection
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("reading buffer")

			log.Println("handling api request")
			api, err := handleApi(buf)
			if err != nil {
				log.Println("api response equals nil")
				ergh := []byte(fmt.Sprintf("ERROR:") + string(buf[:n]) + err.Error()) // super not optimized
				_, err = conn.Write(ergh)
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			log.Println("handling api request successful")

			// Echo the api response back to the client
			_, err = conn.Write([]byte(api))
			if err != nil {
				log.Fatal(err)
			}
		}(conn)
	}
}

func handleApi(b []byte) (string, error) {
	if b[0] != []byte("/")[0] {
		return "", fmt.Errorf("url string invalid\n")
	}
	fmt.Println([]byte(":"))
	first, rest, f := bytes.Cut(b[1:], []byte("/"))
	fmt.Println(string(first[:3]))
	firstStrp := first[:3]
	fmt.Println(PRIMARY_GET)
	fmt.Println(firstStrp)
	fmt.Println([]byte(PRIMARY_GET))
	switch string(firstStrp) {
	case PRIMARY_GET:
		var b []byte
		var err error
		if !f {
			b, err = getHandler([]byte(""))
		} else {
			b, err = getHandler(rest)
		}
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
	if len(path) == 0 {
		g, err := getGenAll()
		return flatten(g), err
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
	b, err := os.ReadFile("/Users/ayobamibamigboye/GolandProjects/scour/internal/invoke/test/uuid.json")
	if err != nil {
		return nil, err
	}
	var uds UUIDs
	err = json.Unmarshal(b, &uds)
	return uds.ByteSlice(), nil
}

func flatten(bb [][]byte) (bflat []byte) {
	for i := 0; i < len(bb); i++ {
		bflat = append(bflat, bb[i]...)
	}
	return
}

func _mainwithflags() {
	var in input
	// Parse flags
	flag.StringVar(&in.socket, "name", "/tmp/api.sock", "pass in preferred path to socket. default /tmp/api.sock")
	flag.StringVar(&in.path, "resource", "", "pass in api resource to call on UNIX socket. if this isn't passed, resource is expected via stdin.")
	flag.Parse()

	// Create a Unix domain socket and listen for incoming connections.
	socket, err := net.Listen("unix", in.socket)
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(in.socket)
		os.Exit(1)
	}()

	for {
		// Accept an incoming connection.
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a separate goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			// Create a buffer for incoming data.
			buf := make([]byte, 512)
			log.Println("making buffer")
			n, err := 0, errors.New("")

			// Read data from the client connection if --resource flag not passed
			if in.path == "" {
				buf = []byte(in.path)
			} else {
				n, err = conn.Read(buf)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("reading buffer")
			}

			log.Println("handling api request")
			api, err := handleApi(buf)
			if err != nil {
				log.Println("api response equals nil")
				ergh := []byte(fmt.Sprintf("ERROR:") + string(buf[:n]) + err.Error()) // super not optimized
				_, err = conn.Write(ergh)
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			log.Println("handling api request successful")

			// Echo the api response back to the client
			_, err = conn.Write([]byte(api))
			if err != nil {
				log.Fatal(err)
			}
		}(conn)
	}
}
