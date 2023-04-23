package api

import (
	"buildRun/pkg"
	"buildRun/pkg/docekr"
	"buildRun/pkg/minio"
	"go.uber.org/zap"
	"os"
)

func Run() {

	//minioOption := &minio.MinioOption{
	//	Endpoint:        "192.168.2.108:30900",
	//	DisableSSL:      false,
	//	ForcePathStyle:  "./",
	//	AccessKeyID:     "admin",
	//	SecretAccessKey: "abcdefg123456",
	//	Bucket:          "dyg-fzzn",
	//	CodeName:        "address.png",
	//	CodePath:        "/fz-1/ay/",
	//}
	pkg.NewConfig()
	newImage := os.Getenv("NewImageName") + ":" + os.Getenv("NewTag")
	if os.Getenv("JobType") == "git" {

	}
	dc := docekr.NewDocker(os.Getenv("HARBORK"), os.Getenv("HARBORV"))
	cli, err := dc.NewClient()
	if err != nil {
		zap.S().Error(err)
		os.Exit(pkg.DOCKERERROR)
	}
	if os.Getenv("JobType") == "minio" {
		zap.S().Info("The type build!")
		err = dc.BuildImage(cli, newImage)
		if err != nil {
			zap.S().Error(err)
			os.Exit(pkg.RUNERROR)
		}
	}

	minioOption := minioFucn()
	switch os.Getenv("JobType") {
	case "save":
		zap.S().Info("The type save!")
		saveImage := os.Getenv("SaveImageName")
		err := dc.LoadImage(cli, minioOption.CodeName, saveImage, newImage)
		if err != nil {
			zap.S().Error(err)
			os.Exit(pkg.MINIOERROR)
		}
	case "export":
		zap.S().Info("The type export!")
		err = dc.ImportImage(cli, minioOption.CodeName, newImage)
		if err != nil {
			zap.S().Error(err)
			os.Exit(pkg.RUNERROR)
		}

	}
	//zap.S().Info("The type build!")
	//err = dc.BuildImage(cli, minioOption.CodeName, newImage)
	//if err != nil {
	//	zap.S().Error(err)
	//	os.Exit(pkg.RUNERROR)
	//}
}

func minioFucn() *minio.MinioOption {
	minioOption := &minio.MinioOption{
		Endpoint:        os.Getenv("MinioUrl"),
		DisableSSL:      false,
		ForcePathStyle:  "./",
		AccessKeyID:     os.Getenv("MINIOK"),
		SecretAccessKey: os.Getenv("MINIOV"),
		Bucket:          os.Getenv("MinioBucket"),
		CodeName:        os.Getenv("Code"),
		CodePath:        os.Getenv("CodePath"),
	}
	minioCli, err := minioOption.MinioClient()
	if err != nil {
		zap.S().Error(err)
		os.Exit(pkg.MINIOERROR)
	}
	err = minioOption.Pull(minioCli)
	if err != nil {
		zap.S().Error(err)
		os.Exit(pkg.MINIOERROR)
	}
	return minioOption
}
