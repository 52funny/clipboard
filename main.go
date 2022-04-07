package main

import (
	"clipboard/client"
	"clipboard/server"
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
)

var hub = server.NewHub()
var isServer *bool
var isClient *bool
var port string
var host string

func init() {
	isServer = flag.Bool("s", false, "for server")
	isClient = flag.Bool("c", false, "for client")
	flag.StringVar(&port, "p", "3355", "server port")
	flag.StringVar(&host, "h", "localhost:3355", "host server address")
	flag.Parse()
}
func main() {
	err := clipboard.Init()
	if err != nil {
		logrus.Errorln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wait sync.WaitGroup

	// handle the signal
	go HandleSignal(cancel)

	if *isServer {
		// is server
		wait.Add(1)
		go server.ListenWebSocket(ctx, &wait, hub, port)
	} else if *isClient {
		// is client
		wait.Add(3)
		go client.ListenClipboard(ctx, &wait)
		go client.ReadAndWriteWebSocket(ctx, &wait, host)
		go client.SetClipboard(ctx, &wait)
	}
	wait.Wait()
}

// Handle the signal
func HandleSignal(cancel context.CancelFunc) {
	// handle the signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT)
	signal := <-interrupt
	logrus.Infoln("got signal:", signal)
	cancel()
}
