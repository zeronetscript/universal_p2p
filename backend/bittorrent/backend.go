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

func Resources(this *Backend) []*backend.P2PResource {
	return nil
}

func SeedExternal(this *Backend, path string) {
}

func Stream(this *Backend, w io.Writer, res *backend.P2PResource) {

}

func (this *Backend) AddTorrentInfoHash(infoHash string) error {
	{
		this.rwLock.RLock()
		defer this.rwLock.RUnlock()
		resource, exist := this.resources[infoHash]
		if exist {

			log.Debugf("%s already exist", infoHash)
			// we didin't change struct ,no need to lock write
			resource.lastAccess = time.Now()
			return nil
		}
	}

	var hash metainfo.Hash

	log.Infof("try to add torrent for %s", infoHash)

	err := hash.FromHexString(infoHash)
	if err != nil {
		errStr := fmt.Sprintf("access path %s is not a valid info hash", infoHash)
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	var t torrent.Torrent

	{
		this.rwLock.Lock()
		defer this.rwLock.Unlock()
		var new bool
		t, new := this.client.AddTorrentInfoHash(hash)
		if !new {
			log.Debugf("other goroutine is also adding %s", infoHash)
			return nil
		}

		info := t.Info()
		if info != nil {
			//already got
			log.Debugf("info already got for %s", infoHash)
			this.resources[infoHash] = CreateFromInfo(info)
			return nil
		}
		//we add this first, wait get info complete
	}

	var lock sync.Mutex
	cond := sync.NewCond(&lock)

	got := false

	go func() {
		<-t.GotInfo()

		lock.Lock()
		got = true
		lock.Unlock()
		log.Debugf("torrent downloaded...")
		cond.Signal()
	}()

	lock.Lock()
	if !got {
		log.Debugf("waiting torrent downloaded...")
		cond.Wait()
		log.Debugf("waiting complete ...")
	}
	lock.Unlock()

	this.rwLock.Lock()
	defer this.rwLock.Unlock()
	this.resources[infoHash] = CreateFromInfo(t.Info())

	return nil
}

func (this *Backend) Command(w io.Writer, r *backend.CommonRequest) {}

func (this *Backend) IterateRootResources(iterFunc backend.ResourceIterFunc) {
	this.rwLock.RLock()
	for _, v := range this.resources {
		iterFunc(v)
	}
	defer this.rwLock.RUnlock()
}

func (this *Backend) Recycle(r *backend.P2PResource) {
	panic("not implemented")
}
