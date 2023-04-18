package docekr

import (
	"buildRun/pkg"
	"buildRun/utils"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
)

type Docker interface {
	PushImage(cli *client.Client, name string) error
	PullImage(cli *client.Client, name string) error
	LoadImage(cli *client.Client, name string) error
	ImportImage(cli *client.Client, name, newName string) error
	BuildImage(cli *client.Client, name string) error
	NewClient() (*client.Client, error)
}
type stiDocker struct {
	UserName string
	Password string
}

func (d stiDocker) NewClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}
func (d *stiDocker) PushImage(cli *client.Client, name string) error {
	//harbor认证
	authConfig := types.AuthConfig{
		Username: d.UserName,
		Password: d.Password,
	}
	authStr, err := encodeAuthToBase64(authConfig)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	//读取镜像文件
	zap.S().Info("start push image：", name)
	var pushReader io.ReadCloser
	pushReader, err = cli.ImagePush(context.Background(), name, types.ImagePushOptions{
		All:           false,
		RegistryAuth:  authStr,
		PrivilegeFunc: nil,
	})
	if err != nil {
		zap.S().Error(err)
		return err
	}
	defer pushReader.Close()
	//输出推送进度
	_ = logImage(pushReader)
	zap.S().Info("Image push success!")
	return nil
}
func (d *stiDocker) LoadImage(cli *client.Client, name string) error {
	//打开镜像文件
	imageFile, err := os.Open(name)
	if err != nil {
		zap.S().Error(err)
	}
	defer imageFile.Close()
	ctx := context.Background()
	zap.S().Info("Start load image")
	load, err := cli.ImageLoad(ctx, imageFile, true)
	if err != nil {
		zap.S().Error(err)
	}
	defer load.Body.Close()
	//load镜像
	str := logImage(load.Body)
	//获取load后的镜像名称
	start := strings.Index(str, ": ") + 2
	end := strings.Index(str[start:], "\\n")
	imageName := str[start : start+end]
	zap.S().Info("Image load success!")
	//导入镜像
	err = d.PushImage(cli, imageName)
	if err != nil {
		return err
	}
	return nil
}
func (d *stiDocker) ImportImage(cli *client.Client, name, imageName string) error {
	//读取镜像文件
	imageFile, err := os.Open(name)
	defer imageFile.Close()
	if err != nil {
		return err
	}
	options := types.ImageImportOptions{}
	source := types.ImageImportSource{
		Source:     imageFile,
		SourceName: "-",
	}
	//import镜像文件
	ctx := context.Background()
	zap.S().Info("Start import image!")
	imageImport, err := cli.ImageImport(ctx, source, imageName, options)
	defer imageImport.Close()
	if err != nil {
		return err
	}
	_ = logImage(imageImport)
	zap.S().Info("Image Import success!")
	//推送镜像
	err = d.PushImage(cli, imageName)

	if err != nil {
		return err
	}
	return nil
}
func (d *stiDocker) BuildImage(cli *client.Client, codeName, name string) error {
	var tags = []string{name}
	fileOptions := types.ImageBuildOptions{
		Tags:           tags,
		Dockerfile:     "docker/Dockerfile",
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
	}
	//拷贝文件到Dockerfile目录下
	err := utils.Copy(codeName, pkg.DOCKERFILEPATH+codeName)
	if err != nil {
		return err
	}
	var destTar = "docker.tar"
	//把文件打成tar包
	err = utils.Tar(pkg.DOCKERFILEPATH, destTar, false)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	//执行构建
	zap.S().Info("Start build image:", name)
	ctx := context.Background()
	dockerBuildContext, err := os.Open(destTar)
	if err != nil {
		return err
	}
	defer dockerBuildContext.Close()
	buildResponse, err := cli.ImageBuild(ctx, dockerBuildContext, fileOptions)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	_ = logImage(buildResponse.Body)
	zap.S().Info("Start build image:", name, "success!")
	err = d.PushImage(cli, name)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	return nil
}

// base64加密
func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	authJSON, err := json.Marshal(authConfig)
	return base64.URLEncoding.EncodeToString(authJSON), err
}

func NewDocker(user, password string) *stiDocker {
	return &stiDocker{
		UserName: user,
		Password: password,
	}
}

// 输出推送加载\上传进度
func logImage(reader io.Reader) string {
	buf1 := new(bytes.Buffer)
	_, err := buf1.ReadFrom(reader)
	if err != nil {
		zap.S().Error(err)
	}
	s1 := buf1.String()
	zap.S().Info(s1)
	return s1
}
