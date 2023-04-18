package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinioOption struct {
	Endpoint        string `json:"endpoint,omitempty" `
	DisableSSL      bool   `json:"disableSSL,omitempty"`
	ForcePathStyle  string `json:"forcePathStyle,omitempty" `
	AccessKeyID     string `json:"accessKeyID,omitempty" `
	SecretAccessKey string `json:"secretAccessKey,omitempty" `
	SessionToken    string `json:"sessionToken,omitempty" `
	Bucket          string `json:"bucket,omitempty" `
	CodeName        string `json:"codeName,omitempty"`
	CodePath        string `json:"codePath,omitempty"`
}

func (o *MinioOption) MinioClient() (*minio.Client, error) {
	fmt.Println(o)
	cli, err := minio.New(o.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(o.AccessKeyID, o.SecretAccessKey, ""),
		Secure: o.DisableSSL,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil

}

func (o *MinioOption) Pull(cli *minio.Client) error {
	fmt.Println(&cli)
	zap.S().Info("Start download code!")
	ctx := context.Background()
	err := cli.FGetObject(ctx, o.Bucket, o.CodePath+o.CodeName, o.ForcePathStyle+o.CodeName, minio.GetObjectOptions{})
	if err != nil {
		zap.S().Error(err)
		return nil
	}
	zap.S().Info("Download the file successful!")
	return nil
}
