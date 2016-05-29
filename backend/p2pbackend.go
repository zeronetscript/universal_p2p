package backend

import (
	"github.com/juju/loggo"
	"io"
	"path"
)

var backendLog = loggo.GetLogger("backend")

type Protocolize interface {
	Protocol() string
}

type P2PBackend interface {
	Protocolize
	IterateRootResources(iterFunc ResourceIterFunc)
	IterateSubResources(P2PResource, ResourceIterFunc) error
	Command(io.Writer, *CommonRequest)
	Recycle(*P2PResource)
}

var AllBackend map[string]P2PBackend = make(map[string]P2PBackend)

func RegisterBackend(backend P2PBackend) bool {

	log.Debugf("try to register backend %s", backend.Protocol())

	_, exist := AllBackend[backend.Protocol()]
	if exist {
		log.Errorf("backend %s already exist", backend.Protocol())

		return false
	}

	AllBackend[backend.Protocol()] = backend
	backendLog.Infof("backend %s registered", backend.Protocol())

	return true
}

//callback to iterator resources, return true means accessed(will update lastAccess)
//should get needed resource and return,not do block operator inside
type ResourceIterFunc func(P2PResource) bool

func GetProtocolRootPath(i Protocolize) string {

	return path.Join(GlobalConfig.RunningDir, i.Protocol())
}
