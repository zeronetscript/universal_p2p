package bittorrent

import(

	"github.com/anacrolix/torrent"
)
	

const torrentBlockListURL = "http://john.bitsurge.net/public/biglist.p2p.gz"


type BittorrentBackend struct{
	client *torrent.Client

	infoHashMapTorrent map[string][*torrent.Torrent]
}

var BitTorrentBackend *backend = newBackend()

func newBackend() *BitTorrentBackend{

	BitTorrentBackend t{
		client := torrent.NewClient(

		)
	}
}

func init(){


}
