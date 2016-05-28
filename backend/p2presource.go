package backend

import (
	"time"
)

type P2PResource interface {
	//same as Frontend
	Protocol() string
	//if IsRoot ,return total size of all (bittorrent alike)
	//or if total size unknown currently , return -1
	Size() int64
	//return currently used disk size (0 if not downloaded any)
	DiskUsage() uint64
	//how much downloaded .p2p may allocate whole file but only used a little
	DownloadedSize() uint64
	//for collect history status, for recycle
	LastAccess() time.Time
	//for bittorrent , this is the "torrent" which contains all sub resources
	IsRoot() bool
	// unique URL to identify .for bittorrent, this is the info hash
	RootURL() string

	//only sub resource have path
	Path() string
}
