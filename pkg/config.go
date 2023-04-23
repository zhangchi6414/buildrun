package pkg

import (
	"buildRun/pkg/logger"
)

const (
	DOCKERFILEPATH = "/root/docker/" //env DOCKERFILE
	DOCKERFILE     = "/config/Dockerfile"
)

const (
	RUNERROR       = 100
	MINIOERROR     = 101
	GITERROR       = 102
	LOADIMAGEERROR = 103
	IMPORTIMAGEERROR
	PUSHIMAGEERROR
	DOCKERERROR
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
