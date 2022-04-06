package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

// listen websocket server
func ListenWebSocket(ctx context.Context, wait *sync.WaitGroup, hub *Hub) {
	defer wait.Done()
	go hub.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	go func() {
		err := http.ListenAndServe(":800", nil)
		if err != nil {
			logrus.Println(err)
		}
	}()
	logrus.Infoln("start websocket server")
	<-ctx.Done()
}
