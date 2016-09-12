package backend

import (
	"strings"
)

func MatchResource(resMap map[string]interface{}, accessPath []string) (interface{}, []string) {

	//match longest file path
	for i := len(accessPath); i > 0; i-- {
		joined := strings.Join(accessPath[0:i], SLASH)

		res := resMap[joined]
		if res != nil {
			return res, accessPath[i:]
		}
	}

	return nil, nil
}
