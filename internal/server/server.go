package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/tg-manager/internal/api"
	"github.com/tg-manager/internal/conf"
	"github.com/tg-manager/internal/forwarder"
	"github.com/tg-manager/internal/storage"
	"github.com/tg-manager/internal/telegram"
)

type Server struct {
	storage   *storage.Storage
	tgSvc     *telegram.Service
	engine    *forwarder.Engine
	apiServer *api.ApiServer
	conf      conf.Config
}

func NewServer(st *storage.Storage, conf conf.Config) *Server {
	// 1. Create forwarder engine (needs DB, api getter will be set after tg starts)
	engine := forwarder.NewEngine(st.GetDB())

	// 2. Create Telegram service (passes engine as update handler)
	tgSvc := telegram.NewService(
		conf.TelegramConfiguration.AppID,
		conf.TelegramConfiguration.AppHash,
		conf.TelegramConfiguration.SessionPath,
		engine,
	)

	// 3. Wire the API getter into the engine
	engine.SetAPIGetter(tgSvc.API)

	// 4. Create API server
	apiServer := api.NewApiServer(st, tgSvc, engine, conf)

	return &Server{
		storage:   st,
		tgSvc:     tgSvc,
		engine:    engine,
		apiServer: apiServer,
		conf:      conf,
	}
}

func (s *Server) Run(ctx context.Context) error {
	// Start Telegram client in background
	go func() {
		if err := s.tgSvc.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Telegram service stopped with error")
		}
	}()

	// Load forwarding rules
	if err := s.engine.ReloadRules(); err != nil {
		log.Error().Err(err).Msg("Failed to load forwarding rules")
	}

	// Start HTTP server (blocking)
	return s.apiServer.Run()
}
