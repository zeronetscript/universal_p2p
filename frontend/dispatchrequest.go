package frontend

interface DispatchedRequest {

	rootProtocol string
	//this field allow us to upgrade our protocol or p2p protocol easily
	subVersion     string
	controlCommand string

	controlArgs interface
}

type CommonRequest struct {
	//top level protocol , backend register this top level rootProtocol
	//will be : bittorrent/ipfs/triblr/btsync/syncthing
	rootProtocol string
	//this field allow us to upgrade our protocol or p2p protocol easily
	subVersion     string
	controlCommand string
}

type URLRequest struct {
	CommonRequest
	controlArgs []string
}

type PostData struct{
	subPath []string
	rawData []byte
}

type PostRequest struct{
	CommonRequest
	controlArgs PostData
}
