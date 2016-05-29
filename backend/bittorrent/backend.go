package bittorrent

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"io"
	"sync"
	"time"
)

const PROTOCOL = "bittorrent"

const torrentBlockListURL = "http://john.bitsurge.net/public/biglist.p2p.gz"

type Backend struct {
	client *torrent.Client
	rwLock *sync.RWMutex

	resources map[string]*Resource
}

func (this Backend) Protocol() string {
	return PROTOCOL
}

var BittorrentBackend Backend

var log = loggo.GetLogger("bittorrent")

func init() {

	cfg := &torrent.Config{
		DataDir: backend.GetProtocolRootPath(&BittorrentBackend),
	}

	var err error
	BittorrentBackend.client, err = torrent.NewClient(cfg)
	if err != nil {
		log.Errorf("error creating bittorrent backend %s", err)
		return
	}

	BittorrentBackend.rwLock = new(sync.RWMutex)
	backend.RegisterBackend(&BittorrentBackend)
}

func (this *Backend) AddTorrentInfoHash(infoHash string) (*Resource, error) {
	{
		this.rwLock.RLock()
		defer this.rwLock.RUnlock()
		resource, exist := this.resources[infoHash]
		if exist {

			log.Debugf("%s already exist", infoHash)
			// we didin't change struct ,no need to lock write
			resource.lastAccess = time.Now()
			return resource, nil
		}
	}

	var hash metainfo.Hash

	log.Infof("try to add torrent for %s", infoHash)

	err := hash.FromHexString(infoHash)
	if err != nil {
		errStr := fmt.Sprintf("access path %s is not a valid info hash", infoHash)
		log.Errorf(errStr)
		return nil, errors.New(errStr)
	}

	var t *torrent.Torrent

	{
		this.rwLock.Lock()
		defer this.rwLock.Unlock()
		var new bool
		t, new := this.client.AddTorrentInfoHash(hash)
		if !new {
			log.Debugf("other goroutine is also adding %s", infoHash)

			//TODO needs test ?
			if t.Info() == nil {
				panic("logic error")
			}

			ret := CreateFromTorrent(t)
			this.resources[infoHash] = ret
			return ret, nil
		}

		info := t.Info()
		if info != nil {
			//already got
			log.Debugf("info already got for %s", infoHash)
			ret := CreateFromTorrent(t)
			this.resources[infoHash] = ret
			return ret, nil
		}
		//we add this first, wait get info complete
	}

	<-t.GotInfo()

	log.Debugf("torrent downloaded...")

	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	ret := CreateFromTorrent(t)
	this.resources[infoHash] = ret

	return ret, nil
}

func (this *Backend) Command(w io.Writer, r *backend.CommonRequest) {
	panic("not implemented")
}

func (this *Backend) IterateRootResources(iterFunc backend.ResourceIterFunc) {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()
	for _, v := range this.resources {
		if iterFunc(v) {
			v.lastAccess = time.Now()
		}
	}
}

func (this *Backend) IterateSubResources(res backend.P2PResource, iterFunc backend.ResourceIterFunc) error {
	this.rwLock.RLock()
	defer this.rwLock.RUnlock()

	v, exist := this.resources[res.RootURL()]

	if res.Protocol() != this.Protocol() {
		errStr := fmt.Sprintf("res protocol %s not same as my protocol", res.Protocol(), this.Protocol())
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	if !exist {
		errStr := fmt.Sprintf("unknown resource %s", res.RootURL())
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	for _, r := range v.subResources {

		if iterFunc(r) {
			r.lastAccess = time.Now()
		}
	}
	return nil
}

func (this *Backend) Recycle(r *backend.P2PResource) {
	panic("not implemented")
}
