package log

import (
	"github.com/juju/loggo"
)

func init() {
	// err is non-nil if and only if the name isn't found.
	loggo.ConfigureLoggers("<root>=TRACE")
}
