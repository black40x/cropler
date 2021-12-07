package config

import (
	"github.com/jessevdk/go-flags"
	"os"
)

var Options struct {
	Host        string `short:"a" long:"addr" description:"Server host" required:"true" default:"127.0.0.1:8000"`
	StoragePath string `short:"s" long:"storage" description:"Storage path" default:"./storage"`
	TempPath    string `short:"t" long:"temp" description:"Temp path" required:"true" default:"./temp"`
	Minio       bool   `short:"m" long:"minio" description:"Use minio as storage"`
	MinioAddr   string `short:"r" long:"minio-addr" description:"Minio host address"`
	MinioBucket string `short:"b" long:"minio-bucket" description:"Minio bucket name"`
	MinioKey    string `short:"k" long:"minio-key" description:"Minio auth key"`
	MinioSecret string `short:"e" long:"minio-secret" description:"Minio auth secret"`
	MinioSsl    bool   `long:"minio-ssl" description:"Minio auth secret"`
	KeepAlive   bool   `long:"keep-alive" description:"HTTP Keep alive"`
	ReadTimeout int    `long:"read-timeout" description:"HTTP Read timeout" default:"10"`
	IdleTimeout int    `long:"idle-timeout" description:"HTTP Idle timeout" default:"10"`
	CacheTime   int    `long:"cache-time" description:"Cache time in hours" default:"24"`
	Help        bool   `short:"h" long:"help" description:"Show this menu."`
}

var parser = flags.NewParser(&Options, flags.Default)

func WriteHelp() {
	parser.WriteHelp(os.Stdout)
}

func InitConfig() ([]string, error) {
	return parser.ParseArgs(os.Args)
}
