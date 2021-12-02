package main

import (
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func SaveCacheImage(img image.Image, fileName string) {
	out, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	ext := strings.ToLower(filepath.Ext(fileName))

	switch ext {
	case ".jpg", ".jpeg":
		opt := jpeg.Options{Quality: 90}
		jpeg.Encode(out, img, &opt)
		break
	case ".png":
		png.Encode(out, img)
		break
	}
}

func ResizeImage(fileName string, width, height, cx, cy, cw, ch, cmw, cmh int) (outputFile string, err error) {
	// Check cache!
	outputFile = fmt.Sprintf("%s/%dx%dc%d_%dx%d_%dcw%dx%d:%s", Options.TempPath, width, height, cx, cy, cw, ch, cmw, cmh, fileName)
	if _, err := os.Stat(outputFile); err == nil {
		return outputFile, nil
	}

	// Work with storage file
	file, err := os.Open(fmt.Sprintf("%s/%s", Options.StoragePath, fileName))
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	file.Seek(0, 0)
	imgConfig, _, err := image.DecodeConfig(file)

	if imgConfig.Width < width {
		width = 0
	}

	if imgConfig.Width < height {
		height = 0
	}

	// Resize
	imgResized := img

	if width != 0 || height != 0 {
		imgResized = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	if cw != 0 && ch != 0 {
		imgResized = imaging.Crop(imgResized, image.Rect(cx, cy, cx+cw, cy+ch))

		if (cmw != 0 || cmh != 0) && (cmw <= cw || cmh <= ch) {
			imgResized = imaging.Resize(imgResized, cmw, cmh, imaging.Lanczos)
		}
	}

	SaveCacheImage(imgResized, outputFile)

	// Free vars
	imgResized = nil
	img = nil
	file = nil

	return outputFile, nil
}

// Server

func HandleCropRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cx, _ := strconv.Atoi(r.URL.Query().Get("cx"))
	cy, _ := strconv.Atoi(r.URL.Query().Get("cy"))
	cw, _ := strconv.Atoi(r.URL.Query().Get("cw"))
	ch, _ := strconv.Atoi(r.URL.Query().Get("ch"))
	cmw, _ := strconv.Atoi(r.URL.Query().Get("cmw"))
	cmh, _ := strconv.Atoi(r.URL.Query().Get("cmh"))
	width, _ := strconv.Atoi(vars["width"])
	height, _ := strconv.Atoi(vars["height"])
	timeStart := time.Now()
	outputFile, err := ResizeImage(vars["image"], width, height, cx, cy, cw, ch, cmw, cmh)

	if err != nil {
		response(w, http.StatusNotFound, map[string]interface{}{
			"error": "Image not found",
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
				"error": "Image process error",
			})
		}

		if Options.Debug {
			Log(LogDebugColor, fmt.Sprintf("-> Resize image \"%s\" at %s\n", vars["image"], time.Since(timeStart)))
		}
	}
}

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	response(w, http.StatusNotFound, map[string]interface{}{
		"error": "Method not found.",
	})
}

func InitServer(host string, port int) {
	router := mux.NewRouter()
	router.HandleFunc("/{width}/{height}/{image}", HandleCropRequest).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(HandleNotFound)

	http.Handle("/", router)

	Log(LogInfoColor, fmt.Sprintf("🚀 Server started at %s:%d \n", host, port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}

// Main

var Options struct {
	Host        string `short:"h" long:"host" description:"Server host" required:"true" default:"127.0.0.1"`
	Port        int    `short:"p" long:"port" description:"Server port" required:"true" default:"8080"`
	StoragePath string `short:"s" long:"storage" description:"Storage path" required:"true" default:"./storage"`
	TempPath    string `short:"t" long:"temp" description:"Temp path" required:"true" default:"./temp"`
	Debug       bool   `short:"d" long:"debug" description:"Show debug information"`
}

func main() {
	Log(LogNoticeColor, "🌄 Welcome to Cropler image resize server\n")
	Log(LogNone, "Commands:\n")
	Log(LogNone, "	-host       Server host\n")
	Log(LogNone, "	-port       Server port\n")
	Log(LogNone, "	-storage    Image storage path\n")
	Log(LogNone, "	-temp       Image temp path\n")
	Log(LogNone, "	-debug      Show debug information\n\n")

	_, err := flags.ParseArgs(&Options, os.Args)

	if err != nil {
		panic(err)
	} else {
		InitServer(Options.Host, Options.Port)
	}
}
