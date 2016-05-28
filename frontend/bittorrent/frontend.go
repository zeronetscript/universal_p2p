package bittorrent

import (
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"github.com/zeronetscript/universal_p2p/backend/bittorrent"
	"net/http"
)

var log = loggo.GetLogger("BittorrentFrontend")

type Frontend struct {
	backend *bittorrent.Backend
}

func Protocol(this *Frontend) string {
	return bittorrent.PROTOCOL
}

func SubVersion(this *Frontend) string {
	return "v0"
}

func Stream(this *Frontend, w http.ResponseWriter, access *AccessRequest) {

	this.backend.AddTorrentInfoHash(access.SubPath[0])
}

func HandleRequest(this *Frontend, w http.ResponseWriter, request *backend.CommonRequest) {

	access := request.(*AccessRequest)

	if len(access.SubPath) < 1 {
		log.Errorf("access url didn't have enough parameters")
		http.Error(w, "access url didin't have enough parameters", 404)
		return
	}

	if access.RootCommand == backend.STREAM {
		this.Stream(w, access)
		return
	}

}
