package log

import (
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"go.uber.org/zap"
)

const VersionLogTaskName = "version_log"

type VersionLogTaskImpl struct {
	core.DefaultTaskImpl
}

func (v *VersionLogTaskImpl) Task() {
	zap.L().Debug("Edge version:" + core.Version)
}
