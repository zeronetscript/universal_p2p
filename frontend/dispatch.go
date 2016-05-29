package frontend

import (
	"errors"
	"fmt"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"net/http"
	"strings"
)

const slash = "/"

var dispatchLog = loggo.GetLogger("Dispatch")

func parseHttpRequest(URL string) (commonRequest backend.CommonRequest, pathArray []string, err error) {
	dispatchLog.Tracef("accessing %s", URL)

	trimmed := strings.TrimRight(strings.TrimLeft(URL, slash), slash)

	allPathArray := strings.Split(trimmed, slash)

	if len(pathArray) < 3 {
		errStr := fmt.Sprintf("url access path is %s, less than needed (at least 3)", trimmed)
		dispatchLog.Errorf(errStr)
		return backend.CommonRequest{}, nil, errors.New(errStr)
	}

	commonRequest = backend.CommonRequest{
		RootProtocol: allPathArray[0],
		SubVersion:   allPathArray[1],
		RootCommand:  allPathArray[2],
	}

	pathArray = allPathArray[3:]

	dispatchLog.Debugf("RootProtocol:%s,SubVersion:%s,RootCommand:%s",
		commonRequest.RootProtocol, commonRequest.SubVersion, commonRequest.RootCommand)
	err = nil
	return
}

func Dispatch(w http.ResponseWriter, request *http.Request) {

	if request.Method != "GET" && request.Method != "POST" {
		dispatchLog.Warningf("unsupported method %s", request.Method)
		http.Error(w, "unsupported method except POST or GET", 404)
		return
	}

	commonRequest, pathArray, err := parseHttpRequest(request.URL.Path)

	if err != nil {
		frontLog.Errorf(err.Error())
		http.Error(w, err.Error(), 404)
		return
	}

	frontend, exist := AllFrontEnd[commonRequest.RootProtocol]

	if !exist {
		dispatchLog.Warningf("protocol %s not supported", commonRequest.RootProtocol)
		http.Error(w, fmt.Sprintf("not support protocol", commonRequest.RootProtocol), 404)
		return
	}

	var parsedRequest interface{}

	if request.Method == "GET" {
		parsedRequest = &backend.AccessRequest{
			CommonRequest: commonRequest,
			SubPath:       pathArray[3:],
		}
	} else {
		dispatchLog.Criticalf("POST data upload read not implemented")
		panic("not implemented")
		parsedRequest = &backend.UploadDataRequest{
			CommonRequest: commonRequest,
		}
	}
	//predefined command

	_, exist = backend.AllBackend[commonRequest.RootProtocol]

	if !exist {
		errStr := fmt.Sprintf("protocol %s not supported", commonRequest.RootProtocol)
		dispatchLog.Warningf(errStr)
		http.Error(w, errStr, http.StatusServiceUnavailable)
		return
	}

	frontend.HandleRequest(w, parsedRequest)
}
