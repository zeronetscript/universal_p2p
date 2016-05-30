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

var BittorrentBackend Backend = Backend{
	resources: make(map[string]*Resource),
}

var log = loggo.GetLogger("bittorrent")

func init() {

	cfg := &torrent.Config{
		DataDir: backend.GetProtocolRootPath(&BittorrentBackend),
		Debug:   true,
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

func (this *Backend) AddTorrentHashOrSpec(hashOrSpec interface{}) (*Resource, error) {

	hash, isHash := hashOrSpec.(*metainfo.Hash)
	var hashHexString string
	var torrentSpec *torrent.TorrentSpec
	if isHash {
		hashHexString = hash.HexString()
	} else {
		torrentSpec = hashOrSpec.(*torrent.TorrentSpec)
	}

	if isHash {
		log.Tracef("wlock")
		this.rwLock.Lock()
		func() {
			log.Tracef("wunlock")
			defer this.rwLock.Unlock()
		}()
		resource, exist := this.resources[hashHexString]
		if exist {
			log.Debugf("%s already exist", hashHexString)
			// we didin't change struct ,no need to lock write
			resource.lastAccess = time.Now()
			return resource, nil
		}
		log.Debugf("%s not exist", hashHexString)
	}

	var t *torrent.Torrent

	{
		log.Tracef("wlock")
		this.rwLock.Lock()
		func() {
			log.Tracef("wunlock")
			defer this.rwLock.Unlock()
		}()

		var new bool
		if isHash {
			t, new = this.client.AddTorrentInfoHash(*hash)
			t.AddTrackers(DefaultTrackers)
		} else {
			var err error
			t, new, err = this.client.AddTorrentSpec(torrentSpec)
			if err != nil {
				return nil, err
			}
		}

		if !new {
			log.Debugf("other goroutine is also adding %s", hashHexString)

			//TODO needs test ?
			if t.Info() == nil {
				panic("logic error")
			}

			ret := CreateFromTorrent(t)
			this.resources[hashHexString] = ret
			return ret, nil
		} else {
			log.Tracef("this is new added torrent,%s", t)
		}

		info := t.Info()
		if info != nil {
			//already got
			log.Debugf("info already got for %s", hashHexString)
			ret := CreateFromTorrent(t)
			this.resources[hashHexString] = ret
			return ret, nil
		}
		//we add this first, wait get info complete
	}

	log.Tracef("call GotInfo")
	<-t.GotInfo()
	log.Tracef("GotInfo completed")

	log.Debugf("torrent downloaded...")

	log.Tracef("wLock")
	this.rwLock.Lock()
	func() {
		log.Tracef("wunLock")
		this.rwLock.Unlock()
	}()
	ret := CreateFromTorrent(t)
	this.resources[hashHexString] = ret

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
