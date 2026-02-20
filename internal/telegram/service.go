package telegram

import (
	"context"
	"runtime"
	"sync"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
)

// UpdateHandler is the interface the forwarder engine implements.
type UpdateHandler interface {
	Handle(ctx context.Context, updates tg.UpdatesClass) error
}

type Service struct {
	appID          int
	appHash        string
	sessionStorage session.Storage
	handler        UpdateHandler

	client *telegram.Client
	api    *tg.Client
	ready  chan struct{}
	cancel context.CancelFunc

	mu           sync.Mutex
	authPhone    string
	authCodeHash string
}

func NewService(appID int, appHash string, sessionStorage session.Storage, handler UpdateHandler) *Service {
	return &Service{
		appID:          appID,
		appHash:        appHash,
		sessionStorage: sessionStorage,
		handler:        handler,
		ready:          make(chan struct{}),
	}
}

// Start creates the Telegram client and runs it. Blocks until ctx is cancelled.
func (s *Service) Start(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)

	dispatcher := tg.NewUpdateDispatcher()
	dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		if s.handler != nil {
			return s.handler.Handle(ctx, &tg.Updates{
				Updates: []tg.UpdateClass{update},
			})
		}
		return nil
	})

	s.client = telegram.NewClient(s.appID, s.appHash, telegram.Options{
		SessionStorage: s.sessionStorage,
		UpdateHandler:  dispatcher,
		Device: telegram.DeviceConfig{
			DeviceModel:    "tg-manager",
			SystemVersion:  runtime.GOOS + "/" + runtime.GOARCH,
			AppVersion:     "0.1.0",
			SystemLangCode: "en",
			LangCode:       "en",
		},
	})

	return s.client.Run(ctx, func(ctx context.Context) error {
		s.api = s.client.API()
		close(s.ready)
		log.Info().Msg("Telegram client ready")
		<-ctx.Done()
		return ctx.Err()
	})
}

// Stop cancels the Telegram client context.
func (s *Service) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// API returns the underlying tg.Client, waiting until the client is ready.
func (s *Service) API() *tg.Client {
	<-s.ready
	return s.api
}

// Client returns the telegram.Client for auth operations.
func (s *Service) Client() *telegram.Client {
	<-s.ready
	return s.client
}

// Ready returns a channel that is closed when the client is ready.
func (s *Service) Ready() <-chan struct{} {
	return s.ready
}

// SetAuthState stores phone and code hash for multi-step auth.
func (s *Service) SetAuthState(phone, codeHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authPhone = phone
	s.authCodeHash = codeHash
}

// GetAuthState retrieves stored phone and code hash.
func (s *Service) GetAuthState() (phone, codeHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.authPhone, s.authCodeHash
}
