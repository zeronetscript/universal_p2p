package main

import (
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/frontend"
)

var appLog = loggo.GetLogger("app")

func main() {

	// err is non-nil if and only if the name isn't found.
	loggo.ConfigureLoggers("<root>=TRACE")

	appLog.Infof("start app")
	frontend.StartHttpServer()
	appLog.Infof("app stop")
}
