package bittorrent

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func ParseHashOrSpec(str string) (interface{}, error) {

	if len(str) == 40 {
		log.Tracef("try to parse %s as info hash", str)
		var hash metainfo.Hash
		err := hash.FromHexString(str)
		if err != nil {
			return nil, err
		}

		return hash, nil
	}

	//not 20 hex string, can be a encodeURIComponent magnet link
	log.Tracef("trying to parse %s as magnet link", str)

	spec, err := torrent.TorrentSpecFromMagnetURI(str)
	if err != nil {
		return nil, err
	}

	return spec, nil
}
