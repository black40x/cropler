package storage

import (
	"cropler/pkg/config"
)

type Adapter interface {
	Init()
	GetFile(fileName string) ([]byte, error)
	StoreOriginal() bool
}

var fileAdapter Adapter

func StoreOriginal() bool {
	return fileAdapter.StoreOriginal()
}

func GetFile(fileName string) ([]byte, error) {
	return fileAdapter.GetFile(fileName)
}

func InitAdapter() {
	if config.Options.Minio {
		fileAdapter = new(Minio)
	} else {
		fileAdapter = new(FileSystem)
	}

	fileAdapter.Init()
}
