package socketparser

import (
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"testing"
)

type SocketParserTestSuite struct {
	suite.Suite
}

var (
	testSocketURLMap = map[string]bool{
		"unix:/path/to/socket.sock":                           true,
		"http://localhost/path/to/socket.sock/resource":       true,
		"unix:/var/run/mysocket.sock":                         true,
		"http://localhost/var/run/mysocket.sock/resource":     true,
		"unix:/tmp/socket.sock":                               true,
		"http://localhost/tmp/socket.sock/resource":           true,
		"unix:/home/user/custom.sock":                         true,
		"http://localhost/home/user/custom.sock/resource":     true,
		"unix:/var/socket/app.sock":                           true,
		"http://localhost/var/socket/app.sock/resource":       true,
		"unix://path/to/socket.sock":                          false,
		"http://localhost://path/to/socket.sock/resource":     false,
		"unix:/path/to/socket.sock/resource":                  false,
		"http://localhostpath/to/socket.sock/resource":        false,
		"unix:/path/to socket.sock":                           false,
		"http://localhost/path/to socket.sock/resource":       false,
		"unix:/path/to/socket.sock?query":                     false,
		"http://localhost/path/to/socket.sock?query/resource": false,
		"unix:/path/to/sock et.sock":                          false,
		"http://localhost/path/to/sock et.sock/resource":      false,
	}
	Order = `
-------------------------------
Test Order: %d
-------------------------------
`
	TESTDIR = "./sock"
)

func (suite *SocketParserTestSuite) SetupTest() {
	// do nothing
}

//func (suite *SocketParserTestSuite) TestIsValid() {
//	ctx := context.Background()
//	i := 1
//	for k, _ := range testSocketURLMap {
//		fmt.Printf(Order, i)
//		log.Printf("URL: %s\n", k)
//		sckt := NewSocket(ctx, k)
//		err := sckt.IsValid()
//		suite.Assert().NoErrorf(err, "error occured: %s", err.Error())
//		//suite.Assert().Equalf(v, b, "expected %v, but got %v", v, b)
//		i++
//	}
//}

//func (suite *SocketParserTestSuite) TestResolve() {
//	ctx := context.WithValue(context.Background(), parser.KeyV, true)
//	var err = make(chan error, 1)
//	go suite.sockServer(ctx, &input{
//		socket:  socketLoc,
//		path:    "/get",
//		jsonLoc: jsonStoreLoc,
//	}, err)
//
//	//errConcrete := <-err
//	//if errConcrete != nil {
//	//	fmt.Printf("detected error: %s\n", errConcrete.Error())
//	//	return
//	//}
//
//	fmt.Printf("detected no error: \n")
//
//	url, errCon := parser.NewUrl(ctx, "http://localhost/get")
//	suite.Require().NoError(errCon)
//
//	b, errCon := UnixSock(ctx, socketLoc, url, false)
//	suite.Require().NoError(errCon)
//	log.Println("Test Response:", string(b))
//}

func clean() {
	err := os.RemoveAll(TESTDIR)
	if err != nil {
		log.Printf("failed to delete test dir: %s\n", TESTDIR)
	}
}

func (suite *SocketParserTestSuite) TearDownTest() {
	clean()
}

func TestSocketParserSuite(t *testing.T) {
	suite.Run(t, new(SocketParserTestSuite))
}
