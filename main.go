package main

import (
	"cropler/pkg"
	"cropler/pkg/config"
	"cropler/pkg/logger"
	"cropler/pkg/storage"
)

func main() {
	logger.LogInfo("Welcome to Cropler image resize server 🌄\n\n")

	if _, err := config.InitConfig(); err == nil {
		if config.Options.Help {
			config.WriteHelp()
		} else {
			if config.Options.CacheTime > 0 {
				pkg.InitCacheWorker()
			}
			storage.InitAdapter()
			pkg.InitVips()
			defer pkg.ShutdownVips()
			pkg.InitServer(config.Options.Host)
		}
	} else {
		logger.LogError("Error!\n")
		config.WriteHelp()
	}
}
