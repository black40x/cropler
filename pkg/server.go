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

func handleCropRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	imagePath := vars["image"]
	imagePath = strings.ReplaceAll(imagePath, "../", "")
	imagePath = strings.ReplaceAll(imagePath, "./", "")

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
	    logger.Log(fmt.Sprintf("[%s] File open error: \"%s\"\n", logger.CurrTime(), err.Error()))

		response(w, http.StatusNotFound, map[string]interface{}{
			"error": fmt.Sprintf("File %s not found", imagePath),
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

func InitServer(port string) {
    addr := ":" + port

	concurrency := runtime.NumCPU() * 2
	listener, _ := net.Listen("tcp", addr)
	listener = netutil.LimitListener(listener, concurrency*10)

	router := mux.NewRouter()
	router.HandleFunc("/{width}/{height}/{image:.*}", handleCropRequest).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(handleNotFound)

	srv := &http.Server{
		Addr:        addr,
		Handler:     router,
		ReadTimeout: time.Duration(config.Options.ReadTimeout) * time.Second,
		IdleTimeout: time.Duration(config.Options.IdleTimeout) * time.Second,
	}
	srv.SetKeepAlivesEnabled(config.Options.KeepAlive)

	logger.LogInfo(fmt.Sprintf("🚀 Server started at %s \n", addr))
	log.Fatal(srv.Serve(listener))
}
