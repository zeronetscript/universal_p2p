package bittorrent

import (
	"encoding/json"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"github.com/zeronetscript/universal_p2p/backend/bittorrent"
	"github.com/zeronetscript/universal_p2p/frontend"
	"net/http"
	"strings"
	"time"
)

var log = loggo.GetLogger("BittorrentFrontend")

type Frontend struct {
	backend *bittorrent.Backend
}

var BittorrentFrontend *Frontend

func (this *Frontend) Protocol() string {
	return bittorrent.PROTOCOL
}

func (this *Frontend) SubVersion() string {
	return "v0"
}

func getLargest(rootRes *bittorrent.Resource) *torrent.File {

	log.Debugf("getting larget file from %s", *rootRes.OriginalName)

	var target torrent.File
	var maxSize int64
	for _, file := range rootRes.Torrent.Files() {
		log.Tracef("testing %s size %d", file.DisplayPath(), file.Length())
		if maxSize < file.Length() {
			log.Tracef("choose largest file as %s", file.DisplayPath())
			maxSize = file.Length()
			target = file
		}
	}

	log.Debugf("final choose %s as largest", target.DisplayPath())

	return &target
}

func pathEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (this *Frontend) Stream(w http.ResponseWriter,
	r *http.Request, access *backend.AccessRequest) {

	var hashOrSpec interface{}
	if len(access.SubPath[0]) == 40 {
		log.Debugf("try to parse %s as info hash", access.SubPath[0])
		var hash metainfo.Hash
		var err error
		err = hash.FromHexString(access.SubPath[0])
		if err != nil {
			errStr := fmt.Sprintf("%s is not a info hash", access.SubPath[0])
			log.Errorf(errStr)
			http.Error(w, errStr, 404)
			return
		}
		log.Debugf("seems %s is a magnet infohash", access.SubPath[0])
		hashOrSpec = &hash
	}

	if hashOrSpec == nil {
		//not 20 hex string, can be a encodeURIComponent magnet link
		log.Debugf("trying to parse %s as encodeURIComponent magnet link", access.SubPath[0])

		var err error
		hashOrSpec, err = torrent.TorrentSpecFromMagnetURI(access.SubPath[0])
		if err != nil {
			errStr := fmt.Sprintf("%s is not a encodeURIComponent encoded magnet link", access.SubPath[0])
			http.Error(w, errStr, 404)
			return
		}
	}

	rootRes, err := this.backend.AddTorrentHashOrSpec(hashOrSpec)

	if err != nil {
		log.Errorf(err.Error())
		http.Error(w, err.Error(), 404)
		return
	}

	var subRes *bittorrent.Resource

	if len(access.SubPath) == 1 {
		//ask for largest file in torrent
		this.backend.IterateSubResources(rootRes, func(res backend.P2PResource) bool {
			cast := res.(*bittorrent.Resource)
			if subRes == nil {
				subRes = cast
				return false
			}

			if cast.Size() > subRes.Size() {
				subRes = cast
			}

			return false
		})
	} else {
		//TODO support archive unpack
		subRes = (*rootRes.SubResources)[strings.Join(access.SubPath[1:], backend.SLASH)]

		if subRes == nil {
			errStr := fmt.Sprintf("no such file %s", access.SubPath)
			log.Errorf(errStr)
			http.Error(w, errStr, 404)
			return
		}

	}

	subRes.UpdateLastAccess()
	f := subRes.SubFile
	log.Tracef("streaming %s", f.DisplayPath())

	f.Download()

	reader, err := NewFileReader(f)

	defer func() {
		if err := reader.Close(); err != nil {
			log.Errorf("Error closing file reader: %s\n", err)
		}
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+rootRes.Torrent.Info().Name+"\"")
	http.ServeContent(w, r, f.DisplayPath(), time.Now(), reader)

	log.Tracef("client disconnected")

}

func infoRootRes(res *bittorrent.Resource) (ret map[string]interface{}) {

	ret = make(map[string]interface{})
	ret["hash"] = res.Torrent.InfoHash().HexString()
	ret["name"] = *res.OriginalName
	ret["bytes_completed"] = res.Torrent.BytesCompleted()
	ret["length"] = res.Torrent.Length()
	return
}

func (this *Frontend) Info(w http.ResponseWriter) {
	var jsonMap map[string]interface{} = make(map[string]interface{})

	jsonMap["dht"] = this.backend.Client.DHT().Stats()

	this.backend.RwLock.RLock()
	defer this.backend.RwLock.RUnlock()

	infoArray := make([]interface{}, len(this.backend.Resources))
	i := 0
	for _, v := range this.backend.Resources {
		infoArray[i] = infoRootRes(v)
		i += 1
	}

	jsonMap["torrents"] = infoArray

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	enc.Encode(jsonMap)
}

func (this *Frontend) HandleRequest(w http.ResponseWriter, r *http.Request, request interface{}) {

	access := request.(*backend.AccessRequest)

	if access.RootCommand == backend.STREAM {
		if len(access.SubPath) < 1 {
			log.Errorf("access url didn't have enough parameters")
			http.Error(w, "access url didin't have enough parameters", 404)
			return
		}
		this.Stream(w, r, access)
		return
	} else if access.RootCommand == backend.STATUS {
		this.Info(w)
		return
	} else {
		http.Error(w, "unsupport", http.StatusInternalServerError)
		return
	}

}

func NewBittorrentFrontend(be *bittorrent.Backend) *Frontend {

	ret := &Frontend{
		backend: be,
	}

	frontend.RegisterFrontend(ret)

	return ret
}
