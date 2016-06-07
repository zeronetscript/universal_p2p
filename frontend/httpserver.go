package frontend

import (
	"github.com/juju/loggo"
	"github.com/tylerb/graceful"
	"github.com/zeronetscript/universal_p2p/backend"
	"net/http"
	"strconv"
	"time"
)

var httpLog = loggo.GetLogger("httpserver")

// Exit statuses.
const (
	_ = iota
	exitNoTorrentProvided
	exitErrorInClient
)

func StartHttpServer() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", Dispatch)

	host := "127.0.0.1"
	port := 7788

	listenAddr := host + ":" + strconv.Itoa(port)

	srv := &graceful.Server{
		Timeout: 1 * time.Second,

		Server: &http.Server{
			Addr:    listenAddr,
			Handler: mux,
		},
	}

	err := srv.ListenAndServe()

	if err != nil {
		httpLog.Errorf("error listening http server %s", listenAddr)
	}

	backend.ShutdownAll()
}
