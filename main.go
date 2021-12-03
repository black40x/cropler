package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	"golang.org/x/net/netutil"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Color

const (
	LogInfoColor    = "\033[1;34m%s\033[0m"
	LogNoticeColor  = "\033[1;36m%s\033[0m"
	LogWarningColor = "\033[1;33m%s\033[0m"
	LogErrorColor   = "\033[1;31m%s\033[0m"
	LogDebugColor   = "\033[0;36m%s\033[0m"
	LogNone         = "%s"
)

func Log(color string, str string) {
	fmt.Printf(color, str)
}

// Crop image

// Helpers

func response(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(toJson(data))
}

func toJson(r interface{}) []byte {
	res, _ := json.Marshal(r)
	return res
}

// Crop

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
			Log(LogErrorColor, err.Error())
		}
		break
	case ".png":
		ep := vips.NewPngExportParams()
		bytes, _, _ := img.ExportPng(ep)
		err := ioutil.WriteFile(fileName, bytes, 0644)
		if err != nil {
			Log(LogErrorColor, err.Error())
		}
		break
	}
}

func ResizeImage(fileName string, width, height, cx, cy, cw, ch, cmw, cmh int) (outputFile string, err error) {
	// Check bad work for resize
	if width == 0 && height == 0 && cw == 0 && ch == 0 {
		outputFile = fmt.Sprintf("%s/%s", Options.StoragePath, fileName)
		return outputFile, nil
	}

	// Check cache!
	ext := strings.ToLower(filepath.Ext(strings.ReplaceAll(fileName, ".cache", "")))
	cacheName := []byte(fmt.Sprintf("%dx%dc%d_%dx%d_%dcw%dx%d_%s", width, height, cx, cy, cw, ch, cmw, cmh, fileName))
	outputFile = fmt.Sprintf("%s/%x.cache", Options.TempPath, md5.Sum(cacheName))
	if _, err := os.Stat(outputFile); err == nil {
		return outputFile, nil
	}

	// Work with storage file
	file, err := vips.NewImageFromFile(fmt.Sprintf("%s/%s", Options.StoragePath, fileName))
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

// Server

func HandleCropRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	imagePath := vars["image"]
	imagePath = strings.ReplaceAll(imagePath, "../", "")
	imagePath = strings.ReplaceAll(imagePath, "./", "")

	if Options.Debug {
		Log(LogDebugColor, fmt.Sprintf("-> Resize image request \"%s\"\n", imagePath))
	}

	cx, _ := strconv.Atoi(r.URL.Query().Get("cx"))
	cy, _ := strconv.Atoi(r.URL.Query().Get("cy"))
	cw, _ := strconv.Atoi(r.URL.Query().Get("cw"))
	ch, _ := strconv.Atoi(r.URL.Query().Get("ch"))
	cmw, _ := strconv.Atoi(r.URL.Query().Get("cmw"))
	cmh, _ := strconv.Atoi(r.URL.Query().Get("cmh"))
	width, _ := strconv.Atoi(vars["width"])
	height, _ := strconv.Atoi(vars["height"])
	timeStart := time.Now()

	outputFile, err := ResizeImage(imagePath, width, height, cx, cy, cw, ch, cmw, cmh)

	if err != nil {
		response(w, http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		file, err := os.Open(outputFile)
		defer file.Close()
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			io.Copy(w, file)
		} else {
			response(w, http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}

		if Options.Debug {
			Log(LogDebugColor, fmt.Sprintf("-> Resize image \"%s\" at %s\n", imagePath, time.Since(timeStart)))
		}
	}
}

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	response(w, http.StatusNotFound, map[string]interface{}{
		"error": "Method not found.",
	})
}

func InitServer(host string, port int) {
	concurrency := runtime.NumCPU() * 2
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, _ := net.Listen("tcp", addr)
	listener = netutil.LimitListener(listener, concurrency*10)

	router := mux.NewRouter()
	router.HandleFunc("/{width}/{height}/{image:.*}", HandleCropRequest).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(HandleNotFound)

	srv := &http.Server{
		Addr:        addr,
		Handler:     router,
		ReadTimeout: time.Duration(Options.ReadTimeout) * time.Second,
		IdleTimeout: time.Duration(Options.IdleTimeout) * time.Second,
	}
	srv.SetKeepAlivesEnabled(Options.KeepAlive)

	Log(LogInfoColor, fmt.Sprintf("🚀 Server started at %s \n", addr))
	log.Fatal(srv.Serve(listener))
}

// Main

var Options struct {
	Host        string `short:"h" long:"host" description:"Server host" required:"true" default:"127.0.0.1"`
	Port        int    `short:"p" long:"port" description:"Server port" required:"true" default:"8080"`
	StoragePath string `short:"s" long:"storage" description:"Storage path" required:"true" default:"./storage"`
	TempPath    string `short:"t" long:"temp" description:"Temp path" required:"true" default:"./temp"`
	Debug       bool   `short:"d" long:"debug" description:"Show debug information"`
	KeepAlive   bool   `short:"k" long:"keep-alive" description:"HTTP Keep alive"`
	ReadTimeout int    `short:"r" long:"read-timeout" description:"HTTP Read timeout" default:"10"`
	IdleTimeout int    `short:"i" long:"idle-timeout" description:"HTTP Idle timeout" default:"10"`
	CacheTime   int    `short:"c" long:"cache-time" description:"Cache time in hours" default:"24"`
}

func VipsLog(messageDomain string, messageLevel vips.LogLevel, message string) {
}

// Cache clear

func ByteFormat(size int64) string {
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

func CacheClear() {
	fs, err := ioutil.ReadDir(Options.TempPath)
	if err != nil {
		Log(LogErrorColor, err.Error())
		return
	}

	cleared := 0
	var size int64 = 0

	for _, info := range fs {
		if filepath.Ext(info.Name()) == ".cache" {
			if time.Since(info.ModTime()).Seconds() > float64(Options.CacheTime) {
				os.Remove(fmt.Sprintf("%s/%s", Options.TempPath, info.Name()))
				size += info.Size()
				cleared++
			}
		}
	}

	if Options.Debug {
		Log(LogNoticeColor, fmt.Sprintf("---> Cache clear %d files, %s size.\n", cleared, ByteFormat(size)))
	}
}

func CacheWorker() {
	for i := 0; ; i++ {
		time.Sleep(time.Minute)
		CacheClear()
		time.Sleep((time.Hour * time.Duration(Options.CacheTime)) - time.Minute)
	}
}

// Main server

func main() {
	config := vips.Config{
		ReportLeaks:  false,
		CollectStats: false,
		CacheTrace:   false,
	}
	vips.LoggingSettings(VipsLog, 0)
	vips.Startup(&config)
	defer vips.Shutdown()

	Log(LogInfoColor, "🌄 Welcome to Cropler image resize server\n")
	Log(LogNone, "Commands:\n")
	Log(LogNone, "	-host       	Server host. Default 'localhost'.\n")
	Log(LogNone, "	-port       	Server port. Default '8080'.\n")
	Log(LogNone, "	-storage    	Image storage path. Default './storage'.\n")
	Log(LogNone, "	-temp       	Image temp path. Default './temp'.\n")
	Log(LogNone, "	-keep-alive 	HTTP Keep alive. Default 'false'.\n")
	Log(LogNone, "	-read-timeout	HTTP Read timeout. Default '10'.\n")
	Log(LogNone, "	-idle-timeout	HTTP Idle timeout. Default '10'.\n")
	Log(LogNone, "	-cache-time	Cache time in hours, 0 to disable. Default '24'.\n")
	Log(LogNone, "	-debug      	Show debug information. Default 'false'.\n\n")

	_, err := flags.ParseArgs(&Options, os.Args)

	if err != nil {
		panic(err)
	} else {
		if Options.CacheTime > 0 {
			go CacheWorker()
		}
		InitServer(Options.Host, Options.Port)
	}
}
