package pkg

import (
	"cropler/pkg/config"
	"cropler/pkg/logger"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/net/netutil"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func response(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(toJson(data))
}

func toJson(r interface{}) []byte {
	res, _ := json.Marshal(r)
	return res
}

func handlePingRequest(w http.ResponseWriter, r *http.Request) {
	response(w, http.StatusOK, map[string]interface{}{
		"result": true,
	})
}

func handleCropRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	imagePath := vars["image"]
	imagePath = strings.ReplaceAll(imagePath, "../", "")
	imagePath = strings.ReplaceAll(imagePath, "./", "")

	inUv := false
	if uv, _ := strconv.Atoi(r.URL.Query().Get("uv")); uv == 1 {
		inUv = true
	}

	cx, _ := strconv.ParseFloat(r.URL.Query().Get("cx"), 64)
	cy, _ := strconv.ParseFloat(r.URL.Query().Get("cy"), 64)
	cw, _ := strconv.ParseFloat(r.URL.Query().Get("cw"), 64)
	ch, _ := strconv.ParseFloat(r.URL.Query().Get("ch"), 64)
	cmw, _ := strconv.Atoi(r.URL.Query().Get("cmw"))
	cmh, _ := strconv.Atoi(r.URL.Query().Get("cmh"))
	width, _ := strconv.Atoi(vars["width"])
	height, _ := strconv.Atoi(vars["height"])
	timeStart := time.Now()

	points := ParsePoints(r.URL.Query().Get("points"))

	outputFile, err := ResizeImage(imagePath, width, height, cx, cy, cw, ch, cmw, cmh, points, inUv)

	if err != nil {
		logger.Log(fmt.Sprintf("[%s] File open error: \"%s\"\n", logger.CurrTime(), err.Error()))

		response(w, http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		file, err := os.Open(outputFile)
		defer file.Close()
		if err == nil {
			logger.Log(fmt.Sprintf("[%s] Resize \"%s\" (%s)\n", logger.CurrTime(), imagePath, time.Since(timeStart)))

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/octet-stream")
			io.Copy(w, file)
		} else {
			logger.Log(fmt.Sprintf("[%s] Resize error \"%s\"\n", logger.CurrTime(), imagePath))

			response(w, http.StatusBadRequest, map[string]interface{}{
				"error": fmt.Sprintf("Resize file %s error", imagePath),
			})
		}
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	response(w, http.StatusNotFound, map[string]interface{}{
		"error": "Method not found.",
	})
}

func InitServer(addr string, port string) {
	appAddr := addr + ":" + port

	concurrency := runtime.NumCPU() * 2
	listener, _ := net.Listen("tcp", appAddr)
	listener = netutil.LimitListener(listener, concurrency*10)

	router := mux.NewRouter()
	router.HandleFunc(
		fmt.Sprintf("%s/{width}/{height}/{image:.*}", config.Options.RouteRoot), handleCropRequest,
	).Methods("GET")

	router.HandleFunc(
		fmt.Sprintf("%s/ping", config.Options.RouteRoot), handlePingRequest,
	).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(handleNotFound)

	srv := &http.Server{
		Addr:        appAddr,
		Handler:     router,
		ReadTimeout: time.Duration(config.Options.ReadTimeout) * time.Second,
		IdleTimeout: time.Duration(config.Options.IdleTimeout) * time.Second,
	}
	srv.SetKeepAlivesEnabled(config.Options.KeepAlive)

	logger.LogInfo(fmt.Sprintf("🚀 Server started at %s \n", appAddr))
	log.Fatal(srv.Serve(listener))
}
