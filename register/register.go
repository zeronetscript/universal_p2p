package register

import (
	btbe "github.com/zeronetscript/universal_p2p/backend/bittorrent"
	btfe "github.com/zeronetscript/universal_p2p/frontend/bittorrent"

	"github.com/juju/loggo"
)

var log = loggo.GetLogger("backend")

func init() {

	btbe.BittorrentBackend = btbe.NewBittorrentBackend()
	btfe.BittorrentFrontend = btfe.NewBittorrentFrontend(btbe.BittorrentBackend)

}
