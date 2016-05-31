package bittorrent

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const PROTOCOL = "bittorrent"

const torrentBlockListURL = "http://john.bitsurge.net/public/biglist.p2p.gz"

//30GB
const MAX_KEEP_FILE_SIZE int64 = 30 * 1024 * 1024 * 1024 * 1024

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
		DataDir: backend.GetDownloadRootPath(&BittorrentBackend),
		//		Debug:   true,
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

func (this *Backend) preprocessTorrent(t *torrent.Torrent) {

	torrentPath := path.Join(this.getSingleMetaPath(t.InfoHash().HexString()),
		"torrent.torrent")

	//save before rename
	f, err := os.Create(torrentPath)

	if err != nil {
		log.Errorf("create file %s failed: %s", torrentPath, err)
		//ignore torrent not save problem
		return
	}

	log.Debugf("create file %s ", torrentPath)

	err = t.Metainfo().Write(f)
	if err != nil {
		log.Errorf("error saving %s, %s", torrentPath, err)
		//ignore not saving problem. we can still work
	}

	renameTorrent(t)
}

//rename torrent name to info hash, info must be got first
func renameTorrent(t *torrent.Torrent) {

	log.Tracef("renaming name from %s to %s", t.Info().Name, t.InfoHash().HexString())
	//naming it folder as info hash to avoid clash
	t.Info().Name = t.InfoHash().HexString()
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

			this.preprocessTorrent(t)

			ret := CreateFromTorrent(t)
			this.resources[hashHexString] = ret
			return ret, nil
		} else {
			log.Tracef("this is new added torrent,%s", t)
		}

		info := t.Info()
		if info != nil {
			this.preprocessTorrent(t)
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
	this.preprocessTorrent(t)
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

func (this *Backend) loadSingleTorrent(dirPath string) {
	torrentFile := path.Join(dirPath, "torrent.torrent")

	//will not clash ,since this is init action
	this.rwLock.Lock()
	defer this.rwLock.Unlock()

	t, err := this.client.AddTorrentFromFile(torrentFile)
	if err != nil {
		log.Errorf("add torrent %s failed ,deleting :%s", dirPath)
		//deleting whole dir
		os.RemoveAll(dirPath)
		return
	}

	renameTorrent(t)
	this.resources[t.InfoHash().HexString()] = CreateFromTorrent(t)

}

type ByLastAccessTime []*Resource

func (s ByLastAccessTime) Len() int {
	return len(s)
}

func (s ByLastAccessTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLastAccessTime) Less(i, j int) bool {
	return s[i].LastAccess().After(s[j].LastAccess())
}

func (this *Backend) getSingleMetaPath(infoHash string) string {
	return path.Join(backend.GetMetaRootPath(this), infoHash)
}

func (this *Backend) getSingleDownloadPath(infoHash string) string {
	return path.Join(backend.GetDownloadRootPath(this), infoHash)
}

func (this *Backend) recycle() {
	//now we are in init stage, no need to lock
	//currently

	s := make(ByLastAccessTime, len(this.resources))
	i := 0
	for _, v := range this.resources {
		s[i] = v
		i += 1
	}

	sort.Sort(s)

	var totalSize int64 = 0

	keep := true
	// nearest resource first
	for _, v := range s {

		if keep {
			totalSize += v.DiskUsage()
			log.Debugf("sumrize %s size %d,total %d", v.Torrent.Name(), v.DiskUsage(), totalSize)
			if totalSize > MAX_KEEP_FILE_SIZE {
				//delete all after this
				keep = false
				log.Debugf("%d over %d limit ", v.DiskUsage(), MAX_KEEP_FILE_SIZE)
			}
		}

		if !keep {
			//delete data only ,keep meta
			dataPath := this.getSingleDownloadPath(v.RootURL())

			log.Debugf("recycling %s", dataPath)

			os.RemoveAll(dataPath)
		} else {
			log.Debugf("keeping %s", v.Torrent.Name())
		}
	}

}

// load from previously save torrents
func (this *Backend) loadSaved() {
	metaPath := backend.GetMetaRootPath(this)

	fileInfos, err := ioutil.ReadDir(metaPath)

	if err != nil {
		log.Errorf("read dir %s failed, not loading old torrent", metaPath)
		return
	}

	var hash metainfo.Hash

	for _, d := range fileInfos {

		fullPath := path.Join(metaPath, d.Name())
		log.Tracef("testing %s", fullPath)
		err := hash.FromHexString(d.Name())

		if err == nil {
			log.Debugf("found info hash dir %s", d.Name())
			go this.loadSingleTorrent(fullPath)
		} else {
			log.Errorf("%s is not a infohash ,will delete it", fullPath)
			//TODO danger
			//os.RemoveAll(fullPath)
		}
	}

	this.recycle()

}
