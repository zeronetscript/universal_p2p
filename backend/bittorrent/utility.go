package bittorrent

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/zeronetscript/universal_p2p/backend"
	"strings"
)

func HexString(hashOrSpec interface{}) string {

	hash, ok := hashOrSpec.(metainfo.Hash)

	if ok {
		return hash.HexString()
	}

	spec, ook := hashOrSpec.(*torrent.TorrentSpec)

	if ook {
		return spec.InfoHash.HexString()
	}

	return ""
}

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

func MatchResource(resMap map[string]*Resource, accessPath []string) (*Resource, []string) {

	//match longest file path
	for i := len(accessPath); i > 0; i-- {
		joined := strings.Join(accessPath[0:i], backend.SLASH)

		res := resMap[joined]
		if res != nil {
			return res, accessPath[i:]
		}
	}

	return nil, nil
}
