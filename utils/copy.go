package utils

import (
	"go.uber.org/zap"
	"io"
	"os"
)

func Copy(src, dit string) error {
	fs, err := os.Open(src)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	defer fs.Close()
	fd, err := os.Create(dit)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	defer fd.Close()
	_, err = io.Copy(fd, fs)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	return nil
}
