package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"log"
	"net"
	"os"
	"strconv"
	"temp/sky/sky3/tools"
	"time"
)

var host string
var port string
var proxy string
var pool int

func main() {
	app := cli.NewApp()
	app.Author = "Andy Zhang"
	app.Version = "v0.1.0"
	app.Name = "SkyClient"
	app.Email = "ztao@gmail.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host",
			Value:       "localhost",
			Usage:       "The skyserver addr",
			Destination: &host,
		},
		cli.StringFlag{
			Name:        "port",
			Value:       "33334",
			Usage:       "The skyserver port",
			Destination: &port,
		},
		cli.StringFlag{
			Name:        "proxy",
			Usage:       "The real server endpoint",
			Destination: &proxy,
		},
		cli.IntFlag{
			Name:        "pool",
			Value:       10,
			Usage:       "The connect pool amount",
			Destination: &pool,
		},
	}

	app.Action = func(c *cli.Context) error {

		if proxy == "" {
			return errors.New(("Proxy Cannot Empty!"))
		}

		return start(host, port, proxy, pool)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func start(host, port, proxy string, pool int) error {

	initPools(host, port, proxy, pool)

	control(host, port, proxy)

	return nil
}

func control(host, port, proxy string) {
	p, _ := strconv.Atoi(port)
	p += 1
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, p))
	if err != nil {
		panic(err)
	}

	log.Printf("Control Conn [%s] Create \n", tcpAddr.String())
	serverConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err)
	}

	serverConn.SetKeepAlivePeriod(5 * time.Second)
	serverConn.SetKeepAlive(true)
	serverConn.SetDeadline(time.Time{})

	buf := make([]byte, 0xffff)
	for {
		n, err := serverConn.Read(buf)
		if err != nil {
			panic(err)
		}

		i, _ := strconv.Atoi(string(buf[:n]))
		initPools(host, port, proxy, i)
	}
}
func connectProxy(src net.Conn, proxy string) {

	buf := make([]byte, 0xffff)
	n, err := src.Read(buf)
	if err != nil {
		panic(err)
	}

	log.Println("New Proxy Request")
	proxyAddr, err := net.ResolveTCPAddr("tcp", proxy)
	if err != nil {
		panic(err)
	}

	proxyConn, err := net.DialTCP("tcp", nil, proxyAddr)
	if err != nil {
		panic(err)
	}

	proxyConn.SetKeepAlivePeriod(5 * time.Second)
	proxyConn.SetKeepAlive(true)
	proxyConn.SetDeadline(time.Time{})

	proxyConn.Write(buf[:n])
	errMsg := make(chan bool)

	go tools.HandleRequest(src, proxyConn, errMsg)
	go tools.HandleRequest(proxyConn, src, errMsg)

	<-errMsg

}

func initPools(host, port, proxy string, num int) {
	for i := 0; i < num; i ++ {
		tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
		if err != nil {
			panic(err)
		}

		serverConn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			panic(err)
		}

		serverConn.SetKeepAlivePeriod(5 * time.Second)
		serverConn.SetKeepAlive(true)
		serverConn.SetDeadline(time.Time{})

		go func(src net.Conn) {
			log.Printf("[%d] [%s] Message Connect Create \n", i, tcpAddr.String())
			connectProxy(src, proxy)
		}(serverConn)

	}
}
