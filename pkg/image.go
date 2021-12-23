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

func fromUV(x float64, size int) int {
	return int(x * float64(size))
}

func ResizeImage(fileName string, width, height int, _cx, _cy, _cw, _ch float64, cmw, cmh int, points []DefinitionPoint, inUV bool) (outputFile string, err error) {
	// Check bad work for resize
	if !storage.StoreOriginal() && (width == 0 && height == 0 && _cw == 0 && _ch == 0) {
		outputFile = fmt.Sprintf("%s/%s", config.Options.StoragePath, fileName)
		return outputFile, nil
	}

	// Check cache!
	ext := strings.ToLower(filepath.Ext(strings.ReplaceAll(fileName, ".cache", "")))
	cacheName := []byte(fmt.Sprintf(
		"%t_%dx%dc%f_%fx%f_%fcw%dx%d_%s_%s", inUV, width, height, _cx, _cy, _cw, _ch, cmw, cmh, GetPointsCache(points), fileName,
	))
	outputFile = fmt.Sprintf("%s/%x.cache", config.Options.TempPath, md5.Sum(cacheName))
	if _, err := os.Stat(outputFile); err == nil {
		return outputFile, nil
	}

	// Work with storage file
	imBytes, err := storage.GetFile(fileName)
	if err != nil {
		return "", errors.New("image not found")
	}

	file, err := vips.NewImageFromBuffer(imBytes)
	if err != nil {
		return "", errors.New("image read error")
	}
	defer file.Close()

	// Convert coords
	var cx, cy, cw, ch int

	if inUV {
		var newWidth, newHeight int

		if width != 0 && height != 0 {
			newWidth = width
			newHeight = height
		} else if width != 0 && height == 0 {
			newWidth = width
			newHeight = int(float32(file.Height()) * ((float32(width) / (float32(file.Width()) * 0.01)) / 100))
		} else if height != 0 && width == 0 {
			newHeight = height
			newWidth = int(float32(file.Width()) * ((float32(height) / (float32(file.Height()) * 0.01)) / 100))
		} else {
			newHeight = file.Height()
			newWidth = file.Width()
		}

		cw = fromUV(_cw, newWidth)
		ch = fromUV(_ch, newHeight)
		cx = fromUV(_cx, newWidth) - (cw / 2)
		cy = fromUV(_cy, newHeight) - (ch / 2)
	} else {
		cx = int(_cx)
		cy = int(_cy)
		cw = int(_cw)
		ch = int(_ch)
	}

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

	cropped := false
	if cw != 0 && ch != 0 {
		cropped = true
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

	if !cropped && len(points) > 0 {
		DrawPoints(outputFile, points, inUV)
	}

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
