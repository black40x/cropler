package pkg

import (
	"cropler/pkg/config"
	"cropler/pkg/storage"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"os"
	"path/filepath"
	"strings"
)

func ResizeImage(fileName string, width, height, cx, cy, cw, ch, cmw, cmh int) (outputFile string, err error) {
	// Check bad work for resize
	if !storage.StoreOriginal() && (width == 0 && height == 0 && cw == 0 && ch == 0) {
		outputFile = fmt.Sprintf("%s/%s", config.Options.StoragePath, fileName)
		return outputFile, nil
	}

	// Check cache!
	ext := strings.ToLower(filepath.Ext(strings.ReplaceAll(fileName, ".cache", "")))
	cacheName := []byte(fmt.Sprintf("%dx%dc%d_%dx%d_%dcw%dx%d_%s", width, height, cx, cy, cw, ch, cmw, cmh, fileName))
	outputFile = fmt.Sprintf("%s/%x.cache", config.Options.TempPath, md5.Sum(cacheName))
	if _, err := os.Stat(outputFile); err == nil {
		return outputFile, nil
	}

	// Work with storage file
	bytes, err := storage.GetFile(fileName)
	if err != nil {
		return "", errors.New(err.Error())
	}

	file, err := vips.NewImageFromBuffer(bytes)
	if err != nil {
		return "", errors.New("image not found")
	}
	defer file.Close()

	if file.Width() < width {
		width = 0
	}

	if file.Height() < height {
		height = 0
	}

	widthScale := float64(width) / float64(file.Width())
	heightScale := float64(height) / float64(file.Height())

	if width == 0 {
		widthScale = (float64(file.Width()) * heightScale) / float64(file.Width())
	}

	// Resize
	if width != 0 || height != 0 {
		err := file.ResizeWithVScale(widthScale, heightScale, vips.KernelLanczos2)
		if err != nil {
			return "", errors.New("image resize invalid scale factor")
		}
	}

	if cw != 0 && ch != 0 {
		err := file.ExtractArea(cx, cy, cw, ch)
		if err != nil {
			return "", errors.New("image crop invalid rect")
		}

		if (cmw != 0 || cmh != 0) && (cmw <= cw || cmh <= ch) {
			if file.Width() < cmw {
				cmw = 0
			}

			if file.Height() < cmh {
				cmh = 0
			}

			cmwScale := float64(cmw) / float64(cw)
			cmhScale := float64(cmh) / float64(ch)

			if cmw == 0 && cmh == 0 {
				cmwScale = 1
			} else if cmw == 0 {
				cmwScale = (float64(file.Width()) * cmhScale) / float64(file.Width())
			}

			err := file.ResizeWithVScale(cmwScale, cmhScale, vips.KernelLanczos2)
			if err != nil {
				return "", errors.New("image crop resize invalid scale factor")
			}
		}
	}

	SaveCacheImage(file, outputFile, ext)

	return outputFile, nil
}

func VipsLog(messageDomain string, messageLevel vips.LogLevel, message string) {}

func ShutdownVips() {
	defer vips.Shutdown()
}

func InitVips() {
	config := vips.Config{
		ReportLeaks:  false,
		CollectStats: false,
		CacheTrace:   false,
	}
	vips.LoggingSettings(VipsLog, 0)
	vips.Startup(&config)
}
