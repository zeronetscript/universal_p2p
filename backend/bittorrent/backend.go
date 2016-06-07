package bittorrent

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/dht/krpc"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	_ "github.com/zeronetscript/universal_p2p/log"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

const (
	ADD_TORRENT = "add_torrent"
	GET_TORRENT = "get_torrent"
)

const PROTOCOL = "bittorrent"

const torrentBlockListURL = "http://john.bitsurge.net/public/biglist.p2p.gz"

//30GB
const MAX_KEEP_FILE_SIZE int64 = 30 * 1024 * 1024 * 1024 * 1024

type Backend struct {
	Client *torrent.Client
	RwLock *sync.RWMutex

	Resources map[string]*Resource
}

func (this Backend) Protocol() string {
	return PROTOCOL
}

var BittorrentBackend *Backend

type dummy struct {
}

func (dummy) Protocol() string {
	return PROTOCOL
}

func NewBittorrentBackend() *Backend {

	cfg := &torrent.Config{
		DataDir: backend.GetDownloadRootPath(dummy{}),
		Debug:   false,
	}

	client, err := torrent.NewClient(cfg)

	if err != nil {
		log.Errorf("error creating bittorrent backend %s", err)
		return nil
	}

	ret := &Backend{
		Resources: make(map[string]*Resource),
		Client:    client,
		RwLock:    new(sync.RWMutex),
	}

	ret.loadSaved()

	backend.RegisterBackend(ret)

	return ret

}

var log = loggo.GetLogger("bittorrent")

func (this *Backend) saveAndRename(t *torrent.Torrent) *torrent.Torrent {

	metaDir := this.getSingleMetaPath(t.InfoHash().HexString())

	er := os.MkdirAll(metaDir, os.ModeDir|os.ModePerm)

	log.Debugf("creating torrent dir %s", metaDir)

	if er == nil {

		torrentPath := path.Join(metaDir, "torrent.torrent")

		log.Debugf("creating torrent %s", torrentPath)
		//save before rename
		f, err := os.Create(torrentPath)

		if err != nil {
			log.Errorf("create file %s failed: %s", torrentPath, err)
			//ignore torrent not save problem
			return t
		}

		log.Debugf("create file %s ", torrentPath)

		err = t.Metainfo().Write(f)
		if err != nil {
			log.Errorf("error saving %s, %s", torrentPath, err)
			//ignore not saving problem. we can still work
		}
	} else {
		log.Errorf("error creating torrent path %s", metaDir)
	}

	return this.renameAddTorrent(t)

}

func (this *Backend) renameAddTorrent(t *torrent.Torrent) *torrent.Torrent {
	renameTorrent(t)
	newT, _ := this.Client.AddTorrent(t.Metainfo())
	return newT
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
		this.RwLock.Lock()
		func() {
			log.Tracef("wunlock")
			defer this.RwLock.Unlock()
		}()
		resource, exist := this.Resources[hashHexString]
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
		this.RwLock.Lock()
		func() {
			log.Tracef("wunlock")
			defer this.RwLock.Unlock()
		}()

		var new bool
		if isHash {
			t, new = this.Client.AddTorrentInfoHash(*hash)
			t.AddTrackers(DefaultTrackers)
		} else {
			var err error
			t, new, err = this.Client.AddTorrentSpec(torrentSpec)
			if err != nil {
				return nil, err
			}
		}

		if !new && t.Info() != nil {
			log.Debugf("other goroutine is also adding %s, we not save torrent,only return resource from it", hashHexString)

			originalName := t.Name()
			t = this.saveAndRename(t)

			ret := CreateFromTorrent(t, originalName, this.Resources[hashHexString])
			return ret, nil
		}

		log.Tracef("this is new added torrent or waiting,%s", t)
	}

	log.Tracef("call GotInfo")
	<-t.GotInfo()
	log.Tracef("GotInfo completed")

	log.Tracef("wLock")
	this.RwLock.Lock()
	originalName := t.Name()
	t = this.saveAndRename(t)
	log.Debugf("torrent downloaded...")
	defer func() {
		log.Tracef("wunLock")
		this.RwLock.Unlock()
	}()
	ret := CreateFromTorrent(t, originalName, this.Resources[hashHexString])
	this.Resources[hashHexString] = ret

	return ret, nil
}

func (this *Backend) Command(w io.Writer, r *backend.CommonRequest) {
	panic("not implemented")
}

func (this *Backend) IterateRootResources(iterFunc backend.ResourceIterFunc) {
	this.RwLock.RLock()
	defer this.RwLock.RUnlock()
	for _, v := range this.Resources {
		if iterFunc(v) {
			v.UpdateLastAccess()
		}
	}
}

func (this *Backend) IterateSubResources(res backend.P2PResource, iterFunc backend.ResourceIterFunc) error {
	this.RwLock.RLock()
	defer this.RwLock.RUnlock()

	v, exist := this.Resources[res.RootURL()]

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

	for _, v := range *v.SubResources {
		if iterFunc(v) {
			v.UpdateLastAccess()
		}
	}
	return nil
}

func (this *Backend) Recycle(r *backend.P2PResource) {
	panic("not implemented")
}

func (this *Backend) loadSingleTorrent(dirPath string, wg *sync.WaitGroup) {
	defer wg.Done()
	torrentFile := path.Join(dirPath, "torrent.torrent")

	//will not clash ,since this is init action
	this.RwLock.Lock()
	defer this.RwLock.Unlock()

	t, err := this.Client.AddTorrentFromFile(torrentFile)
	if err != nil {
		log.Errorf("add torrent %s failed ,deleting :%s", torrentFile, dirPath)
		//deleting whole dir
		os.RemoveAll(dirPath)
		return
	}

	originalName := t.Name()
	renameTorrent(t)
	this.Resources[t.InfoHash().HexString()] =
		CreateFromTorrent(t, originalName, this.Resources[t.InfoHash().HexString()])

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
	return path.Join(this.getTorrentsPath(), infoHash)
}

func (this *Backend) getSingleDownloadPath(infoHash string) string {
	return path.Join(backend.GetDownloadRootPath(this), infoHash)
}

func (this *Backend) recycle() {
	//now we are in init stage, no need to lock
	//currently

	s := make(ByLastAccessTime, len(this.Resources))
	i := 0
	for _, v := range this.Resources {
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
			log.Debugf("sumrize %s size %d,total %d", *v.OriginalName, v.DiskUsage(), totalSize)
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
			log.Debugf("keeping %s", *v.OriginalName)
		}
	}

}

func (this *Backend) getTorrentsPath() string {
	return path.Join(backend.GetMetaRootPath(this), "torrents")
}

func (this *Backend) getInfosPath() string {
	return path.Join(backend.GetMetaRootPath(this), "infos")
}

func (this *Backend) getDHTNodesPath() string {
	return path.Join(this.getInfosPath(), "dht.nodes")
}

// load from previously save torrents, dht nodes, access history
func (this *Backend) loadSaved() {

	log.Debugf("loading dht nodes:%s", this.getDHTNodesPath())
	nodes, er := loadDHTNodes(this.getDHTNodesPath())
	if er != nil {
		log.Errorf("error loading dht nodes:%s, ignore saved dht nodes", er)
	} else {
		for _, n := range nodes {
			this.Client.DHT().AddNode(n)
		}
	}

	log.Debugf("loading last access history:%s", this.getLastAccessPath())
	er = loadLastAccessFile(&this.Resources, this.getLastAccessPath())
	if er != nil {
		log.Errorf("ignore last access history,error :%s", er)
	}

	torrentsPath := this.getTorrentsPath()
	log.Infof("loading saved torrents from %s", torrentsPath)
	fileInfos, err := ioutil.ReadDir(torrentsPath)

	if err != nil {
		log.Errorf("read dir %s failed, not loading old torrent", torrentsPath)
		return
	}

	var hash metainfo.Hash

	var wg sync.WaitGroup

	for _, d := range fileInfos {

		fullPath := path.Join(torrentsPath, d.Name())
		log.Tracef("testing %s", fullPath)
		err := hash.FromHexString(d.Name())

		if err == nil {
			log.Debugf("found info hash dir %s", d.Name())
			wg.Add(1)
			go this.loadSingleTorrent(fullPath, &wg)
		} else {
			log.Errorf("%s is not a infohash ,will delete it", fullPath)
			//TODO danger
			//os.RemoveAll(fullPath)
		}
	}

	wg.Wait()

	//delete every access history which do not have corresponding torrent
	for k, v := range this.Resources {
		if v.Torrent == nil {
			log.Tracef("history %s not have corresponding torrent ", *v.OriginalName)
			delete(this.Resources, k)
		}
	}

	this.recycle()

}

func saveDHTNodes(filePath string, nodes krpc.CompactIPv4NodeInfo) error {

	nodesFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	binary, er := nodes.MarshalBencode()
	if er != nil {
		return er
	}

	_, errr := nodesFile.Write(binary)
	if errr != nil {
		return errr
	}

	return nil
}

func loadDHTNodes(filePath string) (krpc.CompactIPv4NodeInfo, error) {

	f, er := os.Open(filePath)
	var ret krpc.CompactIPv4NodeInfo

	if er != nil {
		return nil, er
	}

	bytes, err := ioutil.ReadAll(f)

	if err != nil {

		return nil, err
	}

	errr := ret.UnmarshalBencode(bytes)
	if errr != nil {
		return nil, errr
	}

	return ret, nil
}

func (this *Backend) getLastAccessPath() string {
	return path.Join(this.getInfosPath(), "lastAccess.json")
}

func (this *Backend) AddTorrent(mi *metainfo.MetaInfo) error {
	this.RwLock.Lock()
	defer this.RwLock.Unlock()

	t, err := this.Client.AddTorrent(mi)
	if err != nil {
		return err
	}

	this.renameAddTorrent(t)

	return nil

}

func (this *Backend) Shutdown() {
	os.MkdirAll(this.getInfosPath(), os.ModePerm)
	log.Debugf("saving dht nodes:%s", this.getDHTNodesPath())
	err := saveDHTNodes(this.getDHTNodesPath(), this.Client.DHT().Nodes())
	if err != nil {
		log.Errorf("error saving dht nodes:%s", err)
	}

	log.Debugf("saving access history file :%s", this.getLastAccessPath())

	err = saveLastAccessFile(&this.Resources, this.getLastAccessPath())

	if err != nil {
		log.Errorf("error saving last access file %s", err)
	}

}
