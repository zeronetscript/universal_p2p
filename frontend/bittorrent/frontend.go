package bittorrent

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"github.com/zeronetscript/universal_p2p/backend/bittorrent"
	"github.com/zeronetscript/universal_p2p/frontend"
	"net/http"
	"time"
)

var log = loggo.GetLogger("BittorrentFrontend")

type Frontend struct {
	backend *bittorrent.Backend
}

var bittorrentFrontend Frontend = Frontend{
	backend: bittorrent.BittorrentBackend,
}

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

	var f *torrent.File

	if len(access.SubPath) == 1 {
		//ask for largest file in torrent
		f = getLargest(rootRes)
	} else {
		have := false

		//TODO support archive unpack
		this.backend.IterateSubResources(rootRes, func(res backend.P2PResource) bool {
			cast := res.(*bittorrent.Resource)
			if pathEqual(access.SubPath[1:], cast.SubFile.FileInfo().Path) {
				have = true
				f = cast.SubFile
				return true
			} else {
				return false
			}
		})
		if !have {
			errStr := fmt.Sprintf("no such file %s", access.SubPath)
			log.Errorf(errStr)
			http.Error(w, errStr, 404)
			return
		}
	}

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

func (this *Frontend) HandleRequest(w http.ResponseWriter, r *http.Request, request interface{}) {

	access := request.(*backend.AccessRequest)

	if len(access.SubPath) < 1 {
		log.Errorf("access url didn't have enough parameters")
		http.Error(w, "access url didin't have enough parameters", 404)
		return
	}

	if access.RootCommand == backend.STREAM {
		this.Stream(w, r, access)
		return
	} else {
		http.Error(w, "unsupport", http.StatusInternalServerError)
	}

}

func init() {
	frontend.RegisterFrontend(&bittorrentFrontend)
}
