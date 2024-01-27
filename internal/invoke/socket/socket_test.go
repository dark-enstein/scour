package socket

import (
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type SocketTestSuite struct {
	suite.Suite
}

var (
	testSockets = map[string]bool{
		"file1.txt":      false,
		"socket1.sock":   true,
		"image1.png":     false,
		"socket2.sock":   true,
		"document1.docx": false,
		"archive.tar.gz": false,
		"script.sh":      false,
		"socket3.sock":   true,
		"data.csv":       false,
		"note.md":        false,
		"config.conf":    false,
		"backup.zip":     false,
	}
	Order = `
-------------------------------
Test Order: %d
-------------------------------
`
	TESTDIR   = "./sock"
	JsonStore = `
{
  "uuid_list": [
    "f42d83e4-632a-4f3c-a0c1-23e7af3b7d7a",
    "9e1c42c2-38f9-4a28-bd68-98e77f5e2b5e",
    "0d89258b-0b58-43a5-9a87-2e8f3d9e8ba0",
    "6f2cc106-4ff5-4ae6-afcc-f413c2b84952",
    "dfc6d43d-8497-4b09-aee6-1a7a31a0ae77",
    "be8b8c63-fc3f-487d-a8d4-5e3139c78d12",
    "90f59e32-bdf7-46b7-b028-93be1d178828",
    "79b4739b-c0b7-40e5-82f8-314b1bb793f5",
    "a9aa01f6-74dc-4b81-bf5b-6e7938df36bb",
    "791a47bc-c7e2-4865-b3d2-ff02a4a0c1a5",
    "b9962e2e-5a46-4cb0-8469-1453b1df01a7",
    "c63b39ad-2df5-4398-b60f-346fcf6f0b7b",
    "4e5f0d4c-ff19-4c3a-b8fe-42d73a44c6b5",
    "51f8e9cf-7b97-490f-81a2-8e234e9e1e0a",
    "218f94d7-c2a6-4ed4-b99d-80a14f86e59d"
  ]
}
`
	jsonStoreLoc = filepath.Join(TESTDIR, "uuid.json")
	socketLoc    = "/tmp/api.sock"
)

func (suite *SocketTestSuite) SetupTest() {
	err := os.MkdirAll(TESTDIR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = _writeJson(jsonStoreLoc, JsonStore)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *SocketTestSuite) TestIsSocket() {
	i := 1
	for k, v := range testSockets {
		fmt.Printf(Order, i)
		if v {
			l, err := net.Listen("unix", k)
			if err != nil {
				log.Fatal("listen error:", err)
			}
			fileInfo, _ := os.Stat(k)
			log.Printf("FilePath: %s\n", k)
			log.Printf("Filemode: %s\n", fileInfo.Mode().Type())
			what, err := utils.IsSocket(k)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), v, what)
			defer func(l net.Listener) {
				_ = l.Close()
			}(l)
		} else {
			tmpFile, err := os.Create(k)
			if err != nil {
				log.Fatalf("Error creating file: %s\n", err.Error())
			}
			filePath := tmpFile.Name()
			fileInfo, _ := os.Stat(filePath)
			log.Printf("FilePath: %s\n", filePath)
			log.Printf("Filemode: %s\n", fileInfo.Mode().Type())

			what, err := utils.IsSocket(filePath)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), v, what)
			// Close and remove the temp file, so we can use its name for the socket
			_ = tmpFile.Close()
			_ = os.Remove(filePath)
		}
		i++
	}
}

func (suite *SocketTestSuite) TestHandleSocketAPI() {
	storm := []byte("/get")
	str, err := handleSocketApi(storm)
	if err != nil {
		fmt.Println("received err from handleSocketAPI:", err.Error())
	}
	fmt.Println("response received from socket api:", str)
}

// TODO: it hangs on socket connection
//func (suite *SocketTestSuite) TestUnixSockIt() {
//	ctx := context.WithValue(context.Background(), httparser.KeyV, true)
//	var err = make(chan error, 1)
//	var commsUpChan = make(chan struct{})
//	var resource = "/get"
//	go func() {
//		suite.sockServer(ctx, commsUpChan, &input{
//			socket:   socketLoc,
//			resource: resource,
//			jsonLoc:  jsonStoreLoc,
//		}, err)
//	}()
//
//	<-commsUpChan
//
//	//errConcrete := <-err
//	//if errConcrete != nil {
//	//	fmt.Printf("detected error: %s\n", errConcrete.Error())
//	//	return
//	//}
//
//	fmt.Printf("detected no error\n")
//
//	socket := socketparser.NewSocket(ctx, fmt.Sprintf("%s %s", socketLoc, resource))
//	suite.Require().NoError(socket.Err())
//
//	byteStream, errConn := UnixSock(ctx, socket, false)
//	suite.Require().NoError(errConn)
//	log.Println("Test Response:", string(byteStream))
//}

func clean() {
	err := os.RemoveAll(TESTDIR)
	if err != nil {
		log.Printf("failed to delete test dir: %s\n", TESTDIR)
	}
}

func (suite *SocketTestSuite) TearDownTest() {
	clean()
}

func TestSocketSuite(t *testing.T) {
	suite.Run(t, new(SocketTestSuite))
}

var (
	PRIMARY_GET = "get"
)

type input struct {
	socket   string
	resource string
	jsonLoc  string
}

func (suite *SocketTestSuite) sockServer(ctx context.Context, socketUpChan chan struct{}, in *input, errChan chan error) {

	// Create a Unix domain socket and listen for incoming connections.
	creator := Creator{
		socket:   in.socket,
		resource: in.resource,
		jsonLoc:  jsonStoreLoc,
	}
	creator.StartSocket(ctx, socketUpChan, errChan)

	select {
	case err := <-errChan:
		log.Println("received error:", err)
		return
	case <-time.After(10 * time.Second):
		return
	}

}

//func handleApi(b []byte) (string, error) {
//	if b[0] != []byte("/")[0] {
//		return "", fmt.Errorf("url string invalid\n")
//	}
//	fmt.Println([]byte(":"))
//	first, rest, f := bytes.Cut(b[1:], []byte("/"))
//	fmt.Println(string(first[:3]))
//	firstStrip := first[:3]
//	fmt.Println(PRIMARY_GET)
//	fmt.Println(firstStrip)
//	fmt.Println([]byte(PRIMARY_GET))
//	switch string(firstStrip) {
//	case PRIMARY_GET:
//		var b []byte
//		var err error
//		if !f {
//			b, err = getHandler([]byte(""))
//		} else {
//			b, err = getHandler(rest)
//		}
//		if err != nil {
//			return "", err
//		}
//		return string(b), nil
//	}
//	return "", nil
//}

//	type UUIDs struct {
//		List []string `json:"uuid_list"`
//	}
//
//	func (u UUIDs) ByteSlice() (b [][]byte) {
//		for i := 0; i < len(u.List); i++ {
//			if i < len(u.List)-1 {
//				b = append(b, []byte(u.List[i]+"\n"))
//			} else {
//				b = append(b, []byte(u.List[i]))
//			}
//		}
//		return
//	}
//
//	func getHandler(resource []byte) ([]byte, error) {
//		if len(resource) == 0 {
//			g, err := getGenAll()
//			return tflatten(g), err
//		}
//		if !unicode.IsNumber(rune(resource[0])) {
//			return nil, fmt.Errorf("get subroute passed in isn't a number")
//		}
//		fmt.Println("index str:", string(resource))
//		i, _ := strconv.Atoi(string(resource))
//		return findIndexGen(i)
//	}
//
//	func findIndexGen(i int) ([]byte, error) {
//		fmt.Println("index received:", i)
//		if i < 0 || i > 15 {
//			return nil, fmt.Errorf("index less than 0 or greater than 15\n")
//		}
//		b, err := getGenAll()
//		if err != nil {
//			return nil, err
//		}
//
//		return b[i], nil
//	}
//
//	func getGenAll() ([][]byte, error) {
//		b, err := os.ReadFile(jsonStoreLoc)
//		if err != nil {
//			return nil, err
//		}
//		var uds UUIDs
//		err = json.Unmarshal(b, &uds)
//		return uds.ByteSlice(), nil
//	}
//
//	func tflatten(bb [][]byte) (bflat []byte) {
//		for i := 0; i < len(bb); i++ {
//			bflat = append(bflat, bb[i]...)
//		}
//		return
//	}
func _writeJson(path string, content string) error {
	err := os.WriteFile(path, []byte(content), 0755)
	if err != nil {
		return err
	}
	b, _ := os.ReadFile(path)
	fmt.Println(string(b))
	return nil
}
