package bittorrent

import (
	"bytes"
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

func (this *Frontend) Stream(w http.ResponseWriter,
	r *http.Request, access *backend.AccessRequest) {

	hashOrSpec, er := bittorrent.ParseHashOrSpec(access.SubPath[0])

	if er != nil {
		errStr := fmt.Sprintf("%s is not a info hash or magnet link,%s", access.SubPath[0], er)
		log.Errorf(errStr)
		http.Error(w, errStr, 404)
		return
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

func infoRootRes(res *bittorrent.Resource, full bool) (ret map[string]interface{}) {

	ret = make(map[string]interface{})
	ret["hash"] = res.Torrent.InfoHash().HexString()
	ret["name"] = *res.OriginalName
	ret["bytes_completed"] = res.Torrent.BytesCompleted()
	ret["length"] = res.Torrent.Length()

	if !full {
		return
	}

	file_list := make([]string, len(*res.SubResources))

	i := 0
	for k, _ := range *res.SubResources {
		file_list[i] = k
		i++
	}

	ret["file_list"] = file_list

	return
}

func (this *Frontend) Status(w http.ResponseWriter, subPath []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var jsonMap map[string]interface{} = make(map[string]interface{})

	this.backend.RwLock.RLock()
	defer this.backend.RwLock.RUnlock()

	if check1Arg(false, w, subPath) {
		//single torrent
		hashOrSpec, err := bittorrent.ParseHashOrSpec(subPath[0])
		if err != nil {
			frontend.HttpAndLogError(fmt.Sprintf("first arg is not infoHash or magnet link:%s", err), &log, w)
			return
		}

		res := this.backend.Resources[bittorrent.HexString(hashOrSpec)]
		jsonMap = infoRootRes(res, true)
	} else {
		//global status
		jsonMap["dht"] = this.backend.Client.DHT().Stats()
		infoArray := make([]interface{}, len(this.backend.Resources))
		i := 0
		for _, v := range this.backend.Resources {
			infoArray[i] = infoRootRes(v, false)
			i += 1
		}

		jsonMap["torrents"] = infoArray
	}

	enc := json.NewEncoder(w)

	enc.Encode(jsonMap)
}

func check1Arg(writeResponse bool, w http.ResponseWriter, subPath []string) bool {

	if len(subPath) < 1 {
		if writeResponse {
			log.Errorf("access url didn't have enough parameters")
			http.Error(w, "access url didin't have enough parameters", 404)
		}
		return false
	}

	return true
}

func (this *Frontend) addTorrent(w http.ResponseWriter, u *backend.UploadDataRequest) {

	metaInfo, er := metainfo.Load(u.UploadReader)
	if er != nil {
		frontend.HttpAndLogError(fmt.Sprintf("this is not a torrent file :%s", er), &log, w)
		return
	}

	res, err := this.backend.AddTorrent(metaInfo)
	if err != nil {
		frontend.HttpAndLogError(fmt.Sprintf("error adding torrent:%s", err), &log, w)
		return
	}
	log.Debugf("adding torrent complete")

	jsonMap := infoRootRes(res, true)

	enc := json.NewEncoder(w)

	enc.Encode(jsonMap)

}

func (this *Frontend) getTorrent(w http.ResponseWriter, req *http.Request, a *backend.AccessRequest) {

	hashOrSpec, er := bittorrent.ParseHashOrSpec(a.SubPath[1])

	if er != nil {
		errStr := fmt.Sprintf("%s is not  a hash info or magnet link:%s", a.SubPath[1], er)
		log.Errorf(errStr)
		http.Error(w, errStr, 404)
		return
	}

	res, err := this.backend.AddTorrentHashOrSpec(hashOrSpec)

	if err != nil {

		errStr := fmt.Sprintf("error adding torrent :%s", err)
		log.Errorf(errStr)
		http.Error(w, errStr, 404)
		return
	}

	bin, err := res.Torrent.Info().MarshalBencode()
	if err != nil {
		errStr := fmt.Sprintf("error getting torrent content:%s", err)
		log.Errorf(errStr)
		http.Error(w, errStr, 404)
		return
	}

	rd := bytes.NewReader(bin)

	http.ServeContent(w, req, *res.OriginalName+".torrent", time.Now(), rd)
}

func (this *Frontend) HandleRequest(w http.ResponseWriter, r *http.Request, request interface{}) {

	access, isAccess := request.(*backend.AccessRequest)

	if isAccess {
		switch access.RootCommand {
		case backend.STREAM:
			if !check1Arg(true, w, access.SubPath) {
				return
			}

			this.Stream(w, r, access)
			return
		case backend.STATUS:
			this.Status(w, access.SubPath)
			return

		case bittorrent.GET_TORRENT:

			if !check1Arg(true, w, access.SubPath) {
				return
			}
			this.getTorrent(w, r, access)

		default:
			http.Error(w, "unsupport", http.StatusInternalServerError)
			return
		}
	}

	upload := request.(*backend.UploadDataRequest)
	this.addTorrent(w, upload)

}

func NewBittorrentFrontend(be *bittorrent.Backend) *Frontend {

	ret := &Frontend{
		backend: be,
	}

	frontend.RegisterFrontend(ret)

	return ret
}
