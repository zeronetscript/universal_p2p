package frontend

import (
	"fmt"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"net/http"
	"strings"
)

const slash = "/"

var dispatchLog = loggo.GetLogger("Dispatch")

func Dispatch(w http.ResponseWriter, request *http.Request) {

	dispatchLog.Tracef("accessing %s", request.URL)

	if request.Method != "GET" && request.Method != "POST" {
		dispatchLog.Warningf("unsupported method %s", request.Method)
		http.Error(w, "unsupported method except POST or GET", 404)
		return
	}

	trimmed := strings.TrimRight(strings.TrimLeft(request.URL.Path, slash), slash)

	pathArray := strings.Split(trimmed, slash)

	if len(pathArray) < 3 {
		dispatchLog.Warningf("url access path is %s, less than needed (at least 3)", trimmed)
		http.Error(w, "can not access ROOT ", 404)
		return
	}

	rootProtocol := pathArray[0]

	frontend, exist := AllFrontEnd[rootProtocol]

	if !exist {
		dispatchLog.Warningf("protocol %s not supported", rootProtocol)
		http.Error(w, fmt.Sprintf("not support protocol", rootProtocol), 404)
		return
	}

	subVersion := pathArray[1]
	rootCommand := pathArray[2]
	dispatchLog.Debugf("RootProtocol:%s,SubVersion:%s,RootCommand:%s",
		rootProtocol, subVersion, rootCommand)

	var parsedRequest interface{}

	if request.Method == "GET" {
		parsedRequest = &backend.AccessRequest{
			SubPath: pathArray[3:],
		}
	} else {
		dispatchLog.Criticalf("POST data upload read not implemented")
		panic("not implemented")
		parsedRequest = &backend.UploadDataRequest{}
	}

	commonRequest := parsedRequest.(*backend.CommonRequest)
	commonRequest.RootProtocol = rootProtocol

	//predefined command

	p2pBackend, exist := backend.AllBackend[rootProtocol]

	if !exist {
		dispatchLog.Warningf("protocol %s not supported", rootProtocol)
		http.Error(w, fmt.Sprintf("not support protocol", rootProtocol), http.StatusServiceUnavailable)
		return
	}

	frontend.HandleRequest(w, commonRequest)
}
