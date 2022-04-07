package client

import (
	"bytes"
	constrant "clipboard/constarnt"
	"context"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var readCh = make(chan []byte, 256)
var writeCh = make(chan []byte)

// listen websocket to prepare for client
func ReadAndWriteWebSocket(ctx context.Context, wait *sync.WaitGroup, host string) {
	defer wait.Done()
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	dial := websocket.DefaultDialer
	// dial.ReadBufferSize = 1024
	// dial.WriteBufferSize = 1024
	c, _, err := dial.DialContext(ctx, u.String(), nil)
	if err != nil {
		logrus.Fatal("dial:", err)
	}
	defer c.Close()
	// start goroutine to listen message from websocket server
	// read message from websocket server
	go func() {
		c.SetReadLimit(1 << 22)
		c.SetReadDeadline(time.Now().Add(constrant.PongWait))
		c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(constrant.PongWait)); return nil })
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logrus.Printf("error: %v", err)
					}
					break
				}
				message = bytes.TrimSpace(bytes.Replace(message, constrant.Newline, constrant.Space, -1))
				writeCh <- message
			}
		}
	}()
	// write message to websocket
	go func() {
		ticker := time.NewTicker(constrant.PingPeriod)
		for {
			select {
			case message, ok := <-readCh:
				c.SetWriteDeadline(time.Now().Add(constrant.WriteWait))
				if !ok {
					// The hub closed the channel.
					c.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := c.NextWriter(websocket.BinaryMessage)
				if err != nil {
					return
				}
				w.Write(message)
				// Add queued chat messages to the current websocket message.
				n := len(readCh)
				for i := 0; i < n; i++ {
					w.Write(constrant.Newline)
					w.Write(<-readCh)
				}

				if err := w.Close(); err != nil {
					return
				}
				logrus.Infoln("send done, bytes:", len(message))
			case <-ticker.C:
				c.SetWriteDeadline(time.Now().Add(constrant.WriteWait))
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				return
			}
		}
	}()
	<-ctx.Done()
	// wait done
}
