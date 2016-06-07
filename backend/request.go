package backend

import (
	"io"
)

const (
	//ROOT protocol level command ,every backend should support
	STREAM  = "stream"
	STOP    = "stop"
	RECYCLE = "recycle"
	STATUS  = "status"
)

type CommonRequest struct {
	//top level protocol , backend register this top level rootProtocol
	//will be : bittorrent/ipfs/triblr/btsync/syncthing
	RootProtocol string
	//this field allow us to upgrade our protocol or p2p protocol easily
	SubVersion  string
	RootCommand string
}

//when access stream, or control command is simple
//no need to upload binary data
type AccessRequest struct {
	CommonRequest
	RootURL string
	SubPath []string
}

//add a torrent to backend
type UploadDataRequest struct {
	CommonRequest
	UploadReader io.Reader
}
