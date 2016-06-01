package main

import (
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/frontend"
	_ "github.com/zeronetscript/universal_p2p/log"
	_ "github.com/zeronetscript/universal_p2p/register"
	"io/ioutil"
	"log"
)

var appLog = loggo.GetLogger("app")

func main() {

	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	appLog.Infof("start app")
	frontend.StartHttpServer()
	appLog.Infof("app stop")
}
