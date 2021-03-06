// interative web user interface
package web

import (
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"net/http"
	"ngrok/client/ui"
	"ngrok/client/views/web/static"
	"ngrok/log"
	"ngrok/proto"
	"ngrok/util"
	"strings"
)

type WebView struct {
	wsMessages *util.Broadcast
}

func NewWebView(ctl *ui.Controller, state ui.State, port int) *WebView {
	v := &WebView{
		wsMessages: util.NewBroadcast(),
	}

	switch p := state.GetProtocol().(type) {
	case *proto.Http:
		NewWebHttpView(v, ctl, p)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/http/in", 302)
	})

	http.HandleFunc("/_ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)

		if err != nil {
			http.Error(w, "Failed websocket upgrade", 400)
			log.Warn("Failed websocket upgrade: %v", err)
			return
		}

		msgs := v.wsMessages.Reg()
		defer v.wsMessages.UnReg(msgs)
		for m := range msgs {
			err := conn.WriteMessage(websocket.OpText, m.([]byte))
			if err != nil {
				// connection is closed
				break
			}
		}
	})

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		name := parts[len(parts)-1]
		fn, ok := static.AssetMap[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(fn())
	})

	log.Info("Serving web interface on localhost:%d", port)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return v
}
