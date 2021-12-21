package config

import (
	"cropler/pkg/logger"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"os"
)

var Options struct {
	Host        string `short:"a" long:"addr" description:"Server host" required:"true" env:"APP_PORT" default:"5000"`
	RouteRoot   string `long:"route-root" description:"Route root path" env:"APP_ROUTE_ROOT" default:""`
	StoragePath string `short:"s" long:"storage" description:"Storage path" env:"APP_STORAGE_PATH" default:"./storage"`
	TempPath    string `short:"t" long:"temp" description:"Temp path" required:"true" env:"APP_TEMP_PATH" default:"./temp"`
	Minio       bool   `short:"m" long:"minio" description:"Use minio as storage" env:"USE_MINIO"`
	MinioAddr   string `short:"r" long:"minio-addr" description:"Minio host address" env:"MINIO_ADDR"`
	MinioBucket string `short:"b" long:"minio-bucket" description:"Minio bucket name" env:"MINIO_BUCKET"`
	MinioKey    string `short:"k" long:"minio-key" description:"Minio auth key" env:"MINIO_KEY"`
	MinioSecret string `short:"e" long:"minio-secret" description:"Minio auth secret" env:"MINIO_SECRET"`
	MinioSsl    bool   `long:"minio-ssl" description:"Minio auth secret" env:"MINIO_SSL"`
	KeepAlive   bool   `long:"keep-alive" description:"HTTP Keep alive" env:"HTTP_KEEP_ALIVE"`
	ReadTimeout int    `long:"read-timeout" description:"HTTP Read timeout" env:"HTTP_READ_TIMEOUT" default:"10"`
	IdleTimeout int    `long:"idle-timeout" description:"HTTP Idle timeout" env:"HTTP_IDLE_TIMEOUT" default:"10"`
	CacheTime   int    `long:"cache-time" description:"Cache time in hours" env:"CACHE_TIME" default:"24"`
	Help        bool   `short:"h" long:"help" description:"Show this menu."`
}

var parser = flags.NewParser(&Options, flags.Default)

func WriteHelp() {
	parser.WriteHelp(os.Stdout)
}

func InitConfig() ([]string, error) {
	err := godotenv.Load()
	if err != nil {
		logger.LogError("Error loading .env file")
	}

	return parser.ParseArgs(os.Args)
}
