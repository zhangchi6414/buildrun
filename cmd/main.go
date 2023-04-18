package main

import (
	"buildRun/pkg"
	"buildRun/pkg/docekr"
	"buildRun/pkg/minio"
	"go.uber.org/zap"
	"os"
)

const (
	RUNERROR   = 500
	MINIOERROR = 501
)

func main() {
	pkg.NewConfig()
	//dc := docekr.NewDocker(os.Getenv("HARBORK"), os.Getenv("HARBORV"))
	dc := docekr.NewDocker("admin", "Dyg@12345")
	cli, err := dc.NewClient()
	if err != nil {
		zap.S().Error(err)
	}
	//newImage := os.Getenv("NewImage")
	newImage := "192.168.2.106:1180/456/build:v123"
	if os.Getenv("JobType") == "git" {
		//TODO 执行git拉取代码并构建镜像
	}

	//minioOption := &minio.MinioOption{
	//	Endpoint:        os.Getenv("MinioUrl"),
	//	DisableSSL:      false,
	//	ForcePathStyle:  "./",
	//	AccessKeyID:     os.Getenv("MinioId"),
	//	SecretAccessKey: os.Getenv("MinioKey"),
	//	Bucket:          os.Getenv("MinioBucket"),
	//	CodeName:        os.Getenv("Code"),
	//	CodePath:        os.Getenv("CodePath"),
	//}
	minioOption := &minio.MinioOption{
		Endpoint:        "192.168.2.108:30900",
		DisableSSL:      false,
		ForcePathStyle:  "./",
		AccessKeyID:     "admin",
		SecretAccessKey: "abcdefg123456",
		Bucket:          "dyg-fzzn",
		CodeName:        "address.png",
		CodePath:        "/fz-1/ay/",
	}

	minioCli, err := minioOption.MinioClient()
	if err != nil {
		zap.S().Error(err)
		os.Exit(MINIOERROR)
	}
	err = minioOption.Pull(minioCli)
	if err != nil {
		zap.S().Error(err)
		os.Exit(MINIOERROR)
	}
	switch os.Getenv("JobType") {
	case "save":
		zap.S().Info("The type save!")
		err := dc.LoadImage(cli, newImage)
		if err != nil {
			zap.S().Error(err)
			os.Exit(RUNERROR)
		}
	case "export":
		zap.S().Info("The type export!")
		err = dc.ImportImage(cli, minioOption.CodeName, newImage)
		if err != nil {
			zap.S().Error(err)
			os.Exit(RUNERROR)
		}
	case "minio":
		zap.S().Info("The type build!")
		err = dc.BuildImage(cli, minioOption.CodeName, newImage)
		if err != nil {
			zap.S().Error(err)
			os.Exit(RUNERROR)
		}
	}
	zap.S().Info("The type build!")
	err = dc.BuildImage(cli, minioOption.CodeName, newImage)
	if err != nil {
		zap.S().Error(err)
		os.Exit(RUNERROR)
	}
}
