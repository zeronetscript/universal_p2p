package bittorrent

import (
	"github.com/anacrolix/torrent/metainfo"
	"time"
)

type Resource struct {
	lastAccess time.Time
	inUse      bool
	info       *metainfo.InfoEx
	isRoot     bool
	path       *string
}

func CreateFromInfo(info *metainfo.InfoEx) *Resource {

	return nil
}

func (this *Resource) Protocol() string {
	return PROTOCOL
}

func (this *Resource) Size() int64 {
	panic("not implemented")
	return 0
}

func (this *Resource) DiskUsage() uint64 {
	panic("not implemented")
	return 0
}
func (this *Resource) DownloadedSize() uint64 {
	panic("not implemented")
	return 0
}

func (this *Resource) LastAccess() time.Time {
	return this.lastAccess
}

func (this *Resource) IsRoot() bool {
	return this.isRoot
}

func (this *Resource) RootURL() string {
	return this.info.Hash().HexString()
}

//only sub resource have path
func (this *Resource) Path() string {
	return *this.path
}
