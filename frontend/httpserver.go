package frontend

import (
	"github.com/juju/loggo"
	_ "github.com/zeronetscript/universal_p2p/frontend/bittorrent"
	"net/http"
	"strconv"
)

var httpLog = loggo.GetLogger("httpserver")

// Exit statuses.
const (
	_ = iota
	exitNoTorrentProvided
	exitErrorInClient
)

func StartHttpServer() {

	host := "127.0.0.1"
	port := 7788

	httpLog.Infof("start listening: 7788")

	http.HandleFunc("/*", Dispatch)

	err := http.ListenAndServe(host+":"+strconv.Itoa(port), nil)
	if err != nil {
		httpLog.Errorf("error listening http server %s", err)
	}
}
