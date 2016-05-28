package backend

import (
	"github.com/juju/loggo"
	"github.com/kardianos/osext"
)

var log = loggo.GetLogger("config")

var GlobalConfig = initConfig()

func initConfig() Config {

	log.Debugf("init config")

	var ret Config
	var err error
	ret.RunningDir, err = osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}

	log.Infof("running dir is %s", ret.RunningDir)

	return ret
}

type Config struct {
	RunningDir string
}
