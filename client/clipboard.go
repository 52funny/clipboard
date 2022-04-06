package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
)

var mutex sync.Mutex
var change bool = false

// listen clipboard for change
func ListenClipboard(ctx context.Context, wait *sync.WaitGroup) {
	defer wait.Done()
	logrus.Infoln("start listen clipboard")
	buf := bytes.Buffer{}
	b := base64.NewEncoder(base64.StdEncoding, &buf)
	for {
		select {
		case <-ctx.Done():
			logrus.Infoln("ListenClipboard function got signal to interrupt")
			return
		case bs := <-clipboard.Watch(ctx, clipboard.FmtText):
			mutex.Lock()
			if !change {
				b.Write(bs)
				readCh <- buf.Bytes()
				buf.Reset()
			} else {
				change = false
			}
			mutex.Unlock()
			// readCh <- bs

		case bsi := <-clipboard.Watch(ctx, clipboard.FmtImage):
			mutex.Lock()
			if !change {
				b.Write(bsi)
				readCh <- buf.Bytes()
				buf.Reset()
			} else {
				change = false
			}
			mutex.Unlock()
			// readCh <- bsi
		}
	}
}

func SetClipboard(ctx context.Context, wait *sync.WaitGroup) {
	defer wait.Done()
	for {
		select {
		case <-ctx.Done():
			logrus.Info("SetClipboard function got signal to interrupt")
			return
		case bs := <-writeCh:
			buf := make([]byte, base64.StdEncoding.EncodedLen(len(bs)))
			base64.StdEncoding.Decode(buf, bs)
			t := http.DetectContentType(buf)
			fmt.Println(t)
			if strings.Contains(t, "image") {
				// set image to clipboard
				logrus.Info("set image")
				clipboard.Write(clipboard.FmtImage, buf)
			} else {
				// set text to clipboard
				logrus.Info("set text")
				clipboard.Write(clipboard.FmtText, buf)
			}
			mutex.Lock()
			change = true
			mutex.Unlock()
		}
	}
}
