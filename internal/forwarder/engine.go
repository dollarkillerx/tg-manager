package forwarder

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"github.com/tg-manager/internal/storage"
	"gorm.io/gorm"
)

type Engine struct {
	ctx       context.Context // app-lifecycle context for cancellation
	db        *gorm.DB
	apiGetter func() *tg.Client

	mu          sync.RWMutex
	rules       []storage.ForwardRule
	compiled    map[uint]*regexp.Regexp
	lastForward map[uint]time.Time
}

func NewEngine(db *gorm.DB) *Engine {
	return &Engine{
		db:          db,
		compiled:    make(map[uint]*regexp.Regexp),
		lastForward: make(map[uint]time.Time),
	}
}

// SetContext stores the app-lifecycle context so background goroutines
// (like backfill) can be cancelled on shutdown.
func (e *Engine) SetContext(ctx context.Context) {
	e.ctx = ctx
}

// SetAPIGetter sets the function to retrieve the Telegram API client.
func (e *Engine) SetAPIGetter(getter func() *tg.Client) {
	e.apiGetter = getter
}

// ReloadRules loads all enabled rules from DB and compiles regex patterns.
func (e *Engine) ReloadRules() error {
	var rules []storage.ForwardRule
	if err := e.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return err
	}

	compiled := make(map[uint]*regexp.Regexp)
	for _, r := range rules {
		re, err := regexp.Compile(r.MatchPattern)
		if err != nil {
			log.Warn().Uint("rule_id", r.ID).Err(err).Msg("Failed to compile rule pattern, skipping")
			continue
		}
		compiled[r.ID] = re
	}

	e.mu.Lock()
	e.rules = rules
	e.compiled = compiled
	e.lastForward = make(map[uint]time.Time)
	e.mu.Unlock()

	log.Info().Int("count", len(rules)).Msg("Forwarding rules loaded")
	return nil
}

// Handle processes incoming Telegram updates and forwards matching messages.
func (e *Engine) Handle(ctx context.Context, updates tg.UpdatesClass) error {
	switch u := updates.(type) {
	case *tg.Updates:
		for _, update := range u.Updates {
			if newMsg, ok := update.(*tg.UpdateNewChannelMessage); ok {
				e.handleChannelMessage(ctx, newMsg)
			}
		}
	case *tg.UpdateShort:
		if newMsg, ok := u.Update.(*tg.UpdateNewChannelMessage); ok {
			e.handleChannelMessage(ctx, newMsg)
		}
	}
	return nil
}

func (e *Engine) handleChannelMessage(ctx context.Context, update *tg.UpdateNewChannelMessage) {
	msg, ok := update.Message.(*tg.Message)
	if !ok || msg.Message == "" {
		return
	}

	// Extract channel ID from PeerChannel
	peer, ok := msg.PeerID.(*tg.PeerChannel)
	if !ok {
		return
	}
	channelID := peer.ChannelID

	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, rule := range e.rules {
		if rule.SourceChannelID != channelID {
			continue
		}

		re, ok := e.compiled[rule.ID]
		if !ok {
			continue
		}

		if !re.MatchString(msg.Message) {
			continue
		}

		// Dedup: check if this message was already forwarded by this rule
		var count int64
		e.db.Model(&storage.ForwardLog{}).
			Where("rule_id = ? AND message_id = ?", rule.ID, msg.ID).
			Count(&count)
		if count > 0 {
			log.Debug().Uint("rule_id", rule.ID).Int("message_id", msg.ID).
				Msg("Message already forwarded, skipping (dedup)")
			continue
		}

		// Rate limit: at most 1 forward per rule per minute
		if last, ok := e.lastForward[rule.ID]; ok && time.Since(last) < time.Minute {
			log.Info().Uint("rule_id", rule.ID).Int("message_id", msg.ID).
				Msg("Rate limit hit, skipping forward")
			continue
		}

		log.Info().
			Int64("source", rule.SourceChannelID).
			Int64("target", rule.TargetChannelID).
			Uint("rule_id", rule.ID).
			Str("match", rule.MatchPattern).
			Msg("Forwarding message")

		go e.forwardMessage(ctx, msg, rule)
	}
}

// BackfillRule fetches the latest 50 messages from the rule's source channel,
// matches them against the rule's pattern, and forwards matches with a 1-per-minute
// rate limit. Runs entirely in the background.
func (e *Engine) BackfillRule(rule storage.ForwardRule) {
	logger := log.With().Uint("rule_id", rule.ID).Logger()

	if e.apiGetter == nil {
		logger.Error().Msg("Backfill: API getter not set")
		return
	}

	api := e.apiGetter()

	// 1. Fetch latest 50 messages from source channel
	sourcePeer := &tg.InputPeerChannel{
		ChannelID:  rule.SourceChannelID,
		AccessHash: rule.SourceHash,
	}

	history, err := api.MessagesGetHistory(e.ctx, &tg.MessagesGetHistoryRequest{
		Peer:  sourcePeer,
		Limit: 50,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Backfill: failed to fetch message history")
		return
	}

	// Extract messages from the response
	var msgs []tg.MessageClass
	switch h := history.(type) {
	case *tg.MessagesMessages:
		msgs = h.Messages
	case *tg.MessagesMessagesSlice:
		msgs = h.Messages
	case *tg.MessagesChannelMessages:
		msgs = h.Messages
	default:
		logger.Warn().Msg("Backfill: unexpected history response type")
		return
	}

	// 2. Compile regex
	re, err := regexp.Compile(rule.MatchPattern)
	if err != nil {
		logger.Error().Err(err).Msg("Backfill: failed to compile pattern")
		return
	}

	// 3. Collect matching messages (oldest-first for chronological forwarding)
	var matched []*tg.Message
	for i := len(msgs) - 1; i >= 0; i-- {
		msg, ok := msgs[i].(*tg.Message)
		if !ok || msg.Message == "" {
			continue
		}
		if !re.MatchString(msg.Message) {
			continue
		}

		// 4. Dedup check
		var count int64
		e.db.Model(&storage.ForwardLog{}).
			Where("rule_id = ? AND message_id = ?", rule.ID, msg.ID).
			Count(&count)
		if count > 0 {
			logger.Debug().Int("message_id", msg.ID).Msg("Backfill: already forwarded, skipping")
			continue
		}

		matched = append(matched, msg)
	}

	if len(matched) == 0 {
		logger.Info().Msg("Backfill: no new matching messages found")
		return
	}

	logger.Info().Int("count", len(matched)).Msg("Starting backfill")

	// 5. Channel-based rate-limited forwarding
	ch := make(chan *tg.Message, len(matched))
	for _, m := range matched {
		ch <- m
	}
	close(ch)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		first := true
		for msg := range ch {
			if !first {
				select {
				case <-ticker.C:
				case <-e.ctx.Done():
					logger.Info().Msg("Backfill: cancelled by shutdown")
					return
				}
			}
			first = false

			// Re-check dedup before forwarding (race with real-time forwarding)
			var count int64
			e.db.Model(&storage.ForwardLog{}).
				Where("rule_id = ? AND message_id = ?", rule.ID, msg.ID).
				Count(&count)
			if count > 0 {
				logger.Debug().Int("message_id", msg.ID).Msg("Backfill: already forwarded (race), skipping")
				continue
			}

			logger.Info().Int("message_id", msg.ID).Msg("Backfill: forwarding message")
			e.forwardMessage(e.ctx, msg, rule)
		}

		logger.Info().Msg("Backfill complete")
	}()
}

func (e *Engine) forwardMessage(ctx context.Context, msg *tg.Message, rule storage.ForwardRule) {
	if e.apiGetter == nil {
		log.Error().Msg("API getter not set, cannot forward")
		return
	}

	api := e.apiGetter()

	fromPeer := &tg.InputPeerChannel{
		ChannelID:  rule.SourceChannelID,
		AccessHash: rule.SourceHash,
	}
	toPeer := &tg.InputPeerChannel{
		ChannelID:  rule.TargetChannelID,
		AccessHash: rule.TargetHash,
	}

	_, err := api.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
		FromPeer: fromPeer,
		ToPeer:   toPeer,
		ID:       []int{msg.ID},
		RandomID: []int64{int64(msg.ID) * 1000},
	})
	if err != nil {
		log.Error().Err(err).
			Int64("source", rule.SourceChannelID).
			Int64("target", rule.TargetChannelID).
			Msg("Failed to forward message")
		return
	}

	// Record forward log for dedup
	e.db.Create(&storage.ForwardLog{
		RuleID:          rule.ID,
		MessageID:       msg.ID,
		SourceChannelID: rule.SourceChannelID,
		TargetChannelID: rule.TargetChannelID,
	})

	// Update rate limit timestamp
	e.mu.Lock()
	e.lastForward[rule.ID] = time.Now()
	e.mu.Unlock()
}
