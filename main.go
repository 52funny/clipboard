package main

import (
	"clipboard/client"
	"clipboard/server"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
)

var hub = server.NewHub()

func init() {
	// logrus.SetReportCaller(true)
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

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "server":
			wait.Add(1)
			go server.ListenWebSocket(ctx, &wait, hub)
		case "client":
			wait.Add(3)
			go client.ListenClipboard(ctx, &wait)
			go client.ReadAndWriteWebSocket(ctx, &wait)
			go client.SetClipboard(ctx, &wait)
		}
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
