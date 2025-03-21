package core

type NonSecretConf struct {
	DevMode                   bool
	DisableStdoutLog          bool
	EnableFileLog             bool
	LogDir                    string
	LogLevel                  string
	LogRotationMaxDays        int
	UseDummyDevice            bool
	DeviceSettingsPath        string
	QueueMaxSize              int
	QueueRefillThreshold      int
	GRPCTranspilerHost        string
	GRPCTranspilerPort        string
	TranspilerPluginPath      string
	ServiceDBEndpoint         string
	ServiceDBRegion           string
	DisableStartDevicePolling bool
}

type Info struct {
	Conf *NonSecretConf
}

var CurrentInfo *Info

func SetInfo(c *Conf) {
	conf := &NonSecretConf{
		DevMode:                   c.DevMode,
		DisableStdoutLog:          c.DisableStdoutLog,
		EnableFileLog:             c.EnableFileLog,
		LogDir:                    c.LogDir,
		LogLevel:                  c.LogLevel,
		LogRotationMaxDays:        c.LogRotationMaxDays,
		UseDummyDevice:            c.UseDummyDevice,
		DeviceSettingsPath:        c.DeviceSettingPath,
		QueueMaxSize:              c.QueueMaxSize,
		QueueRefillThreshold:      c.QueueRefillThreshold,
		GRPCTranspilerHost:        c.GRPCTranspilerHost,
		GRPCTranspilerPort:        c.GRPCTranspilerPort,
		TranspilerPluginPath:      c.TranspilerPluginPath,
		ServiceDBEndpoint:         c.ServiceDBEndpoint,
		DisableStartDevicePolling: c.DisableStartDevicePolling,
	}

	CurrentInfo = &Info{
		Conf: conf,
	}
}
