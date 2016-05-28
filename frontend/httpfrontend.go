package frontend

import (
	"fmt"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"io"
	"path"
)

var frontLog = loggo.GetLogger("frontend")

type HttpFrontend interface {
	Protocolize
	SubVersion() string
	HandleRequest(io.Writer, *backend.CommonRequest)
}

var AllFrontEnd map[string]HttpFrontend

func RegisterFrontend(frontend HttpFrontend) bool {

	frontLog.Debugf("try to register frontend %s", frontend.Protocol())

	_, exist := AllFrontEnd[frontend.Protocol()]

	if exist {
		frontLog.Errorf("frontend %s already exist", frontend.Protocol())

		return false
	}
	AllFrontEnd[frontend.Protocol()] = frontend
	frontLog.Infof("frontend %s registered", frontend.Protocol())
	return true
}
