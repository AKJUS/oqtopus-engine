package core

import (
	"fmt"

	"go.uber.org/zap"
)

var Version string

const NoVersion = "no_version_info"

func SetVersion(c *Conf, versionByBuildFlag string) {
	if versionByBuildFlag != "" {
		Version = versionByBuildFlag
	} else if c.Version != "" {
		Version = c.Version
	} else {
		Version = NoVersion
	}
	zap.L().Info(fmt.Sprintf("Version is %s", Version))
}
