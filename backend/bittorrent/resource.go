package bittorrent

import (
	"encoding/json"
	"github.com/anacrolix/torrent"
	"github.com/zeronetscript/universal_p2p/backend"
	"os"
	"strings"
	"time"
)

type Resource struct {
	lastAccess time.Time
	//for root
	Torrent *torrent.Torrent
	SubFile *torrent.File
	//key is path
	SubResources *map[string]*Resource
	rootRes      *Resource
	path         *string
}

func (this *Resource) UnmarshalJSON(data []byte) error {

	type Aux struct {
		SubResources *map[string]*Aux
		Path         *string
		LastAccess   time.Time
	}

	aux := &Aux{}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	this.lastAccess = aux.LastAccess

	if aux.SubResources != nil && len(*aux.SubResources) != 0 {

		this.Torrent = invalidTorrent

		tmp := make(map[string]*Resource)
		this.SubResources = &tmp
		for k, v := range *aux.SubResources {
			(*this.SubResources)[k] = &Resource{
				path:       v.Path,
				lastAccess: v.LastAccess,
			}
		}
	}

	return nil
}

func (this *Resource) MarshalJSON() ([]byte, error) {

	if !this.IsRoot() {
		//for none root
		return json.Marshal(&struct {
			Path       string
			LastAccess time.Time
		}{
			Path: this.SubFile.DisplayPath(),
		})
	}

	//for root
	return json.Marshal(&struct {
		SubResources *map[string]*Resource
		LastAccess   time.Time
	}{
		SubResources: this.SubResources,
		LastAccess:   this.lastAccess,
	})

}

var NEVER_ACCESSED = time.Unix(0, 0)

func CreateFromTorrent(t *torrent.Torrent, historyRoot *Resource) *Resource {
	//TODO loads lastAccess from serialized

	log.Debugf("create resource wrapper for torrent %s", t.Name())

	root := &Resource{
		Torrent: t,
	}

	if historyRoot != nil {
		log.Tracef("found history %s", t.InfoHash().HexString())
		root.lastAccess = historyRoot.lastAccess
	} else {
		root.lastAccess = time.Now()
	}

	tmp := make(map[string]*Resource)
	root.SubResources = &tmp

	for i, v := range t.Info().UpvertedFiles() {

		log.Debugf("create sub resource for torrent %s,%s", t.Name(), v.Path)

		sub := &Resource{
			//makes it old
			SubFile: &t.Files()[i],
			rootRes: root,
			Torrent: invalidTorrent,
		}

		joined := strings.Join(v.Path, backend.SLASH)

		(*root.SubResources)[joined] = sub

		if historyRoot != nil {
			subHistory := (*historyRoot.SubResources)[joined]
			if subHistory != nil {
				log.Tracef("found sub history %s", joined)
				sub.lastAccess = subHistory.lastAccess
				continue
			}
		}

		sub.lastAccess = NEVER_ACCESSED
	}

	return root
}

func (this *Resource) Protocol() string {
	return PROTOCOL
}

func (this *Resource) Size() int64 {
	if this.IsRoot() {
		return this.Torrent.Length()
	}

	return this.SubFile.Length()
}

func (this *Resource) DiskUsage() int64 {
	//TODO
	return this.Size()
}

func (this *Resource) DownloadedSize() int64 {
	//TODO
	return this.Size()
}

func (this *Resource) LastAccess() time.Time {
	return this.lastAccess
}

var invalidTorrent = &torrent.Torrent{}

func (this *Resource) IsRoot() bool {
	return this.Torrent != invalidTorrent
}

func (this *Resource) RootURL() string {
	if this.IsRoot() {
		return this.Torrent.InfoHash().HexString()
	}

	return this.rootRes.Torrent.InfoHash().HexString()
}

var emptyPathArray []string

//only sub resource have path
func (this *Resource) Path() []string {
	if this.IsRoot() {
		return emptyPathArray
	}

	return this.SubFile.FileInfo().Path
}

func (this *Resource) UpdateLastAccess() {
	this.lastAccess = time.Now()
	if this.rootRes != nil {
		this.rootRes.UpdateLastAccess()
	}
}

//even error happened ,you can still use result map(empty,not nil)
func loadLastAccessFile(ret *map[string]*Resource, filePath string) error {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		//remove unaccessible file
		_ = os.Remove(filePath)
		return err
	}

	dec := json.NewDecoder(f)

	err = dec.Decode(ret)

	return err
}

func saveLastAccessFile(res *map[string]*Resource, path string) error {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}
	f.Truncate(0)

	enc := json.NewEncoder(f)

	err = enc.Encode(res)
	if err != nil {
		_ = os.Remove(path)
	}
	return err
}
