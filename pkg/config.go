package pkg

import (
	"buildRun/pkg/logger"
	"time"
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
	FROMIMAGEERROR = 104
	IMPORTIMAGEERROR
	PUSHIMAGEERROR
	DOCKERERROR
)

const (
	DefaultDockerTimeout  = 4 * time.Minute
	DefaultPullRetryCount = 6
	DefaultPushRetryCount = 2
	DefaultPullRetryDelay = 10 * time.Second
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
