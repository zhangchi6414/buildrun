package docekr

import (
	"buildRun/pkg"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"io"
	"os"
	"regexp"
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
func (d *stiDocker) PullImage(cli *client.Client, name string) error {
	return nil
}
func (d *stiDocker) PushImage(cli *client.Client, name string) error {
	authConfig := types.AuthConfig{
		Username: d.UserName,
		Password: d.Password,
	}
	authStr, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}
	zap.S().Info("start push image：", name)
	var pushReader io.ReadCloser

	pushReader, err = cli.ImagePush(context.Background(), name, types.ImagePushOptions{
		All:           false,
		RegistryAuth:  authStr,
		PrivilegeFunc: nil,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer pushReader.Close()
	_ = logImage(pushReader)
	zap.S().Info("Image push success!")
	return nil
}
func (d *stiDocker) LoadImage(cli *client.Client, name string) error {
	path := pkg.CODEPATH + name
	imageFile, err := os.Open(path)
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
	str := logImage(load.Body)
	//re := regexp.MustCompile(`Loaded image: ([^\n]+)`)
	//math := re.FindStringSubmatch(s)
	//re = regexp.MustCompile(`\\n"}`)
	//imageName := re.ReplaceAllString(math[1],"")
	//err = errors.Errorf("not found image Name")
	//if len(imageName) < 1 {
	//	return err
	//}
	re := regexp.MustCompile(`Loaded image: (\S+)`) // 匹配 "Loaded image: " 后的非空白字符
	imageName := re.FindStringSubmatch(str) // 查找匹配项
	if len(imageName) > 1 {
		fmt.Println(imageName[1]) // 输出 "192.168.2.106:1180/456/alpine:3.6"
	}

	fmt.Println(imageName[1])
	zap.S().Info("Image load success!")
	err = d.PushImage(cli, imageName[1])
	if err != nil {
		return err
	}
	return nil
}
func (d *stiDocker) ImportImage(cli *client.Client, name, imageName string) error {
	path := pkg.CODEPATH + name
	imageFile, err := os.Open(path)
	defer imageFile.Close()
	if err != nil {
		return err
	}
	options := types.ImageImportOptions{}
	source := types.ImageImportSource{
		Source:     imageFile,
		SourceName: "-",
	}
	ctx := context.Background()
	zap.S().Info("Start import image!")
	imageImport, err := cli.ImageImport(ctx, source, imageName, options)
	defer imageImport.Close()
	if err != nil {
		return err
	}
	logImage(imageImport)
	zap.S().Info("Image Import success!")
	err = d.PushImage(cli, imageName)
	if err != nil {
		return err
	}
	return nil
}
func (d *stiDocker) BuildImage(cli *client.Client, name string) error {
	return nil
}
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

func logImage(reader io.Reader) string {
	buf1 := new(bytes.Buffer)
	buf1.ReadFrom(reader)
	s1 := buf1.String()
	zap.S().Info(s1)
	return s1
}
