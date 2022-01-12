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

			err := pkg.InitFont()
			if err != nil {
				logger.LogError(err.Error())
			} else {
				pkg.InitServer(config.Options.Host, config.Options.Port)
			}
		}
	} else {
		logger.LogError("Error!\n")
		config.WriteHelp()
	}
}
