package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tg-manager/internal/conf"
	"github.com/tg-manager/internal/server"
	"github.com/tg-manager/internal/storage"
	"github.com/tg-manager/pkg/common/client"
	"github.com/tg-manager/pkg/common/config"
	logs "github.com/tg-manager/pkg/common/log"
)

var configFilename string
var configDirs string

func init() {
	const (
		defaultConfigFilename = "config"
		defaultConfigDirs     = "./,./configs/"
	)
	flag.StringVar(&configFilename, "c", defaultConfigFilename, "Name of the config file, without extension")
	flag.StringVar(&configDirs, "cPath", defaultConfigDirs, "Directories to search for config file, separated by ','")
}

func main() {
	flag.Parse()

	var appConfig conf.Config
	if err := config.InitConfiguration(configFilename, strings.Split(configDirs, ","), &appConfig); err != nil {
		panic(err)
	}
	if b, err := json.MarshalIndent(appConfig, "", "  "); err == nil {
		fmt.Println(string(b))
	}

	// Logger
	logs.InitLog(appConfig.LoggerConfiguration)

	// PostgreSQL
	db, err := client.PostgresClient(appConfig.PostgresConfiguration)
	if err != nil {
		log.Error().Msg("Failed to connect to PostgreSQL")
		panic(err)
	}

	// Storage + AutoMigrate
	st := storage.NewStorage(db)
	if err := st.AutoMigrate(); err != nil {
		log.Error().Err(err).Msg("Failed to auto-migrate")
		panic(err)
	}
	log.Info().Msg("Storage initialized")

	// Context with signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Server
	s := server.NewServer(st, appConfig)
	if err := s.Run(ctx); err != nil {
		log.Error().Msgf("Server stopped: %s", err)
	}
}
