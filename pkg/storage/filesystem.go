package storage

import (
	"cropler/pkg/config"
	"cropler/pkg/logger"
	"fmt"
	"os"
)

type FileSystem struct{}

func (r *FileSystem) Init() {
	logger.LogNotice("Using file system storage. \n")
}

func (r FileSystem) StoreOriginal() bool {
	return false
}

func (r *FileSystem) GetFile(fileName string) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("%s/%s", config.Options.StoragePath, fileName))
}
