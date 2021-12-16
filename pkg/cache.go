package pkg

import (
	"cropler/pkg/config"
	"cropler/pkg/logger"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func byteFormat(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func SaveCacheImage(img *vips.ImageRef, fileName, ext string) {
	if _, err := os.Stat(fileName); err == nil {
		return
	}

	switch ext {
	case ".jpg", ".jpeg":
		ep := vips.NewJpegExportParams()
		ep.Quality = 90
		bytes, _, _ := img.ExportJpeg(ep)
		err := ioutil.WriteFile(fileName, bytes, 0644)
		if err != nil {
			logger.LogError(err.Error())
		}
		break
	case ".png":
		ep := vips.NewPngExportParams()
		bytes, _, _ := img.ExportPng(ep)
		err := ioutil.WriteFile(fileName, bytes, 0644)
		if err != nil {
			logger.LogError(err.Error())
		}
		break
	}
}

func cacheClear() {
	fs, err := ioutil.ReadDir(config.Options.TempPath)
	if err != nil {
		logger.LogError(err.Error())
		return
	}

	cleared := 0
	var size int64 = 0

	for _, info := range fs {
		if filepath.Ext(info.Name()) == ".cache" {
			if time.Since(info.ModTime()).Hours() > float64(config.Options.CacheTime) {
				os.Remove(fmt.Sprintf("%s/%s", config.Options.TempPath, info.Name()))
				size += info.Size()
				cleared++
			}
		}
	}

	logger.LogNotice(fmt.Sprintf("[%s] Cache clear %d files: %s\n", logger.CurrTime(), cleared, byteFormat(size)))
}

func cacheWorker() {
	for i := 0; ; i++ {
		time.Sleep(time.Minute)
		cacheClear()
		time.Sleep(time.Hour - time.Minute)
	}
}

func InitCacheWorker() {
    err := os.MkdirAll(config.Options.TempPath, 0644)
    if err != nil {
        logger.LogError("Error create temp path directory")
    }

	if config.Options.CacheTime > 0 {
		go cacheWorker()
	}
}
