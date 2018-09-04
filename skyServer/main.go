package main

import (
	"fmt"
	"github.com/andy-zhangtao/skyorg/pool"
	"github.com/andy-zhangtao/skyorg/tools"
	"github.com/urfave/cli"
	"log"
	"net"
	"os"
	"strconv"
)

var host string
var port string

var pools *pool.ConnPool

const (
	MAXCONN = 10
)

type ProxyConn struct {
	Id   int
	Conn *net.Conn
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host",
			Value:       "0.0.0.0",
			Usage:       "The addr listen to",
			Destination: &host,
		},
		cli.StringFlag{
			Name:        "port",
			Value:       "33333",
			Usage:       "The port listen to",
			Destination: &port,
		},
	}

	app.Action = func(c *cli.Context) error {

		return start(host, port)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func start(host, port string) error {

	var err error
	var l net.Listener

	connChan := make(chan int)
	clientOffLine := make(chan int)

	go createChanForClient(host, port)
	go createControlConn(host, port, connChan)

	laddr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		fmt.Println("Failed to resolve local address: %s", err)
		os.Exit(1)
	}

	l, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer l.Close()

	log.Println("Listening on " + host + ":" + port)

	go func() {
		current := MAXCONN
		for {
			select {
			case <-clientOffLine:
				current --
				fmt.Println(current)
				if current < MAXCONN/2 {
					connChan <- (MAXCONN - current)
					current = MAXCONN
				}
			}
		}
	}()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			os.Exit(1)
		}

		//logs an incoming message
		id, c := pools.NextConn()
		if c == nil {
			fmt.Println("No Idle Conn")
			conn.Close()
			continue
		}
		go proxyToClient(ProxyConn{
			Id:   id,
			Conn: c,
		}, conn, clientOffLine)
	}

	return nil
}

func proxyToClient(pc ProxyConn, src net.Conn, clientOffLine chan int) {
	fmt.Printf("Received message %s -> %s Get Idle Conn [%v] \n", src.RemoteAddr(), src.LocalAddr(), pc.Id)

	errMsg := make(chan bool)
	go tools.HandleRequest(src, *pc.Conn, errMsg)
	go tools.HandleRequest(*pc.Conn, src, errMsg)

	<-errMsg

	pools.Destory(pc.Id)
	fmt.Printf("Client %d Offline Cap %d \n", pc.Id, pools.Cap())
	clientOffLine <- 1
}

func createChanForClient(host, port string) (*net.Conn, error) {
	cp, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	cp += 1

	laddr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(cp))
	if err != nil {
		fmt.Println("Failed to resolve local address: %s", err)
		os.Exit(1)
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return nil, err
	}

	pools = pool.InitPool(MAXCONN)

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			return nil, err
		}

		pools.AddConn(conn)
		fmt.Printf("New Client Online %v  Cap %d \n", conn.LocalAddr(), pools.Cap())
	}

	return nil, nil
}

func createControlConn(host, port string, controlChan chan int) {
	cp, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	cp += 2

	laddr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(cp))
	if err != nil {
		fmt.Println("Failed to resolve local address: %s", err)
		os.Exit(1)
	}

	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println("Error listening:", err)
		panic(err)
	}

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Control Conn %v \n", conn.LocalAddr())
		for {
			select {
			case i := <-controlChan:
				_, err := conn.Write([]byte(fmt.Sprintf("%d", i)))
				if err != nil {
					goto HERE
				}
			}
		}
	HERE:
		fmt.Println("Wait New Control Conn")
	}

}
