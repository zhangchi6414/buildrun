package pkg

import (
	"buildRun/pkg/logger"
)

const (
	CODEPATH = "/tmp/app/minio/"
)

func NewConfig() {
	Conf := logger.LogConfigs{
		LogLevel:    "debug",
		LogFormat:   "logfmt",
		LogPath:     "/var/log/build/",
		LogFileName: "build.log",
		LogStdout:   true,
	}
	err := logger.InitLogger(Conf)
	if err != nil {
		panic(err)
	}
}
