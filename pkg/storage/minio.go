package storage

import (
	"bufio"
	"bytes"
	"context"
	"cropler/pkg/config"
	"cropler/pkg/logger"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
)

type Minio struct {
	client *minio.Client
}

func (r *Minio) Init() {
	logger.Log(fmt.Sprintf("Using Minio storage at %s\n", config.Options.MinioAddr))

	client, err := minio.New(config.Options.MinioAddr, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Options.MinioKey, config.Options.MinioSecret, ""),
		Secure: config.Options.MinioSsl,
	})

	if err != nil {
		log.Fatalln(err)
	}

	r.client = client

	logger.Log(fmt.Sprintf("Minio status: %t\n\n", r.client.IsOnline()))
}

func (r Minio) StoreOriginal() bool {
	return true
}

func (r *Minio) GetFile(fileName string) ([]byte, error) {
	obj, err := r.client.GetObject(context.Background(), config.Options.MinioBucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		logger.LogError("ERROR!")
		return nil, err
	}

	b := bytes.Buffer{}
	foo := bufio.NewWriter(&b)
	if _, err := io.Copy(foo, obj); err != nil {
		fmt.Println(err)
	}

	return b.Bytes(), nil
}
