package docekr

import (
	"buildRun/pkg"
	"buildRun/utils"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	dockermessage "github.com/docker/docker/pkg/jsonmessage"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	// RetriableErrors is a set of strings that indicate that an retriable error occurred.
	RetriableErrors = []string{
		"ping attempt failed with error",
		"is already in progress",
		"connection reset by peer",
		"transport closed before response was received",
		"connection refused",
	}
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
	authConfig := types.AuthConfig{
		Username: os.Getenv("HARBORK"),
		Password: os.Getenv("HARBORV"),
	}

	authConfigEncoded, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}
	pullOptions := types.ImagePullOptions{
		RegistryAuth: authConfigEncoded,
	}
	ctx := context.Background()
	zap.S().Info("获取基础镜像")

	out, err := cli.ImagePull(ctx, name, pullOptions)
	if err != nil {
		return err
	}
	defer out.Close()
	str := logImage(out)
	if strings.Contains(str, "error") {
		return fmt.Errorf("基础镜像获取失败！", os.Getenv("OLDIMAGE"))
	}

	return nil
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
	retriableError := false
	for retries := 0; retries <= pkg.DefaultPushRetryCount; retries++ {
		err = utils.TimeoutAfter(pkg.DefaultDockerTimeout, fmt.Sprintf("pushing image %q", name), func(timer *time.Timer) error {
			resp, pushErr := cli.ImagePush(context.Background(), name, types.ImagePushOptions{
				All:           false,
				RegistryAuth:  authStr,
				PrivilegeFunc: nil,
			})
			if pushErr != nil {
				return pushErr
			}
			defer resp.Close()

			decoder := json.NewDecoder(resp)
			for {
				if !timer.Stop() {
					return &utils.TimeoutError{}
				}
				timer.Reset(pkg.DefaultDockerTimeout)

				var msg dockermessage.JSONMessage
				pushErr = decoder.Decode(&msg)
				if pushErr == io.EOF {
					return nil
				}
				if pushErr != nil {
					return pushErr
				}

				if msg.Error != nil {
					return msg.Error
				}

				if msg.Progress != nil {
					if msg.Progress.Current != 0 {
						zap.S().Info("pushing image:", name, msg.Progress.String())
					}
				}
				//body, err := io.ReadAll(resp)
				//if err != nil {
				//	return err
				//}
				//if strings.Contains(string(body), "error") {
				//	return fmt.Errorf("build image error！请检查构建镜像参数是否正确！")
				//}
			}
		})
		if err == nil {
			break
		}
		zap.S().Error("pushing image error : %v", err)
		errMsg := fmt.Sprintf("%s", err)
		for _, errorString := range RetriableErrors {
			if strings.Contains(errMsg, errorString) {
				retriableError = true
				break
			}
		}

		if !retriableError {
			return errors.New("Push image failed")
		}

		zap.S().Info("retrying in %s ...", pkg.DefaultPullRetryDelay)
		time.Sleep(pkg.DefaultPullRetryDelay)

	}
	if err != nil {
		zap.S().Error("Push image failed!")
		return err
	}
	zap.S().Info("push success ! ", name)
	//读取镜像文件
	//zap.S().Info("start push image：", name)
	//var pushReader io.ReadCloser
	//pushReader, err = cli.ImagePush(context.Background(), name, types.ImagePushOptions{
	//	All:           false,
	//	RegistryAuth:  authStr,
	//	PrivilegeFunc: nil,
	//})
	//defer pushReader.Close()
	////输出推送进度
	//_ = logImage(pushReader)
	//if err != nil {
	//	os.Exit(pkg.PUSHIMAGEERROR)
	//	zap.S().Error(err)
	//	return err
	//}

	return nil
}
func (d *stiDocker) LoadImage(cli *client.Client, code, oldName, name string) error {
	//打开镜像文件
	imageFile, err := os.Open(code)
	if err != nil {
		zap.S().Error(err)
	}
	defer imageFile.Close()
	ctx := context.Background()
	zap.S().Info("Start load image")
	load, err := cli.ImageLoad(ctx, imageFile, true)

	defer load.Body.Close()
	//load镜像
	str := logImage(load.Body)
	//_ = logImage(load.Body)
	if err != nil {
		os.Exit(pkg.LOADIMAGEERROR)
		zap.S().Error(err)
	}
	//获取load后的镜像名称
	start := strings.Index(str, ": ") + 2
	end := strings.Index(str[start:], "\\n")
	imageName := str[start : start+end]
	zap.S().Info("Image load success!")
	if name == "" {
		name = imageName
	}
	err = cli.ImageTag(ctx, oldName, name)
	if err != nil {
		return err
	}
	//导入镜像
	err = d.PushImage(cli, name)
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
	_ = logImage(imageImport)
	if err != nil {
		os.Exit(pkg.IMPORTIMAGEERROR)
		return err
	}
	//推送镜像
	err = d.PushImage(cli, imageName)

	if err != nil {
		return err
	}
	return nil
}
func (d *stiDocker) BuildImage(cli *client.Client, name string) error {
	var tags = []string{name}
	fileOptions := types.ImageBuildOptions{
		Tags:           tags,
		Dockerfile:     "docker/Dockerfile",
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     false,
	}
	//拷贝Dockerfile
	err := utils.Copy(pkg.DOCKERFILE, pkg.DOCKERFILEPATH+"Dockerfile")
	if err != nil {
		return err
	}
	//拷贝要放到容器内的文件
	nfsPath := os.Getenv("NFSPATH")
	if nfsPath != "" {
		fmt.Println("开始拷贝文件！")
		pathList := strings.Split(nfsPath, ",")
		for _, path := range pathList {
			pathList := strings.Split(path, "/")
			dir := pathList[len(pathList)-1]
			cmd := exec.Command("cp", "-r", path, pkg.DOCKERFILEPATH+dir)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("执行命令时出错:", err)
				return err
			}
			fmt.Println(output)
		}
	}

	var destTar = "docker.tar"
	//把文件打成tar包
	err = utils.Tar(pkg.DOCKERFILEPATH, destTar, false)
	if err != nil {
		zap.S().Error(err)
		return err
	}

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

	//_ = logImage(buildResponse.Body)
	str := logImage(buildResponse.Body)
	if strings.Contains(str, "error") {
		return fmt.Errorf("build image error！请检查构建镜像参数是否正确！")
	}

	zap.S().Info("build image:", name, "success!")
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

//// // 判断需要拷贝的文件类型
//func fileType(path string) error {
//	fileInfo, err := os.Stat(path)
//	if err != nil {
//		return err
//	}
//	//判断是否为目录
//	if fileInfo.IsDir() {
//		pathList := strings.Split(path, "/")
//		dir := pathList[len(pathList)-1]
//		err := os.MkdirAll(pkg.DOCKERFILEPATH+dir, 0755)
//		if err != nil {
//			return err
//		}
//	}
//	err = utils.Copy(path, pkg.DOCKERFILEPATH+path)
//	if err != nil {
//		return err
//	}
//	return nil
//}
