package main

import (
	"buildRun/pkg"
	"buildRun/pkg/docekr"
	"fmt"
	"go.uber.org/zap"
)

func main() {
	pkg.NewConfig()
	dc := docekr.NewDocker("admin", "Dyg@12345")
	cli, err := dc.NewClient()
	if err != nil {
		zap.S().Error(err)
	}
	//err = dc.PushImage(cli, "192.168.2.106:1180/456/alpine:3.6")
	//if err != nil {
	//	zap.S().Error(err)
	//}
	fmt.Println("开始load镜像")
	err = dc.LoadImage(cli, "run.rar")
	if err != nil {
		zap.S().Error(err)
	}
	fmt.Println("开始import镜像")
	err = dc.ImportImage(cli, "run.img", "192.168.2.106:1180/456/run:3.6")
	if err != nil {
		zap.S().Error(err)
	}

}
