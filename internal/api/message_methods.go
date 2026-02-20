package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"github.com/tg-manager/internal/telegram"
)

type MessageInfo struct {
	ID         int    `json:"id"`
	Date       string `json:"date"`
	Text       string `json:"text"`
	SenderName string `json:"sender_name"`
	IsOutgoing bool   `json:"is_outgoing"`
}

type MessagesHistoryMethod struct {
	tgSvc *telegram.Service
}

type messagesHistoryParams struct {
	PeerID     int64  `json:"peer_id"`
	PeerType   string `json:"peer_type"`
	AccessHash int64  `json:"access_hash,string"`
	Limit      int    `json:"limit"`
}

func (m *MessagesHistoryMethod) Name() string { return "messages.history" }

func (m *MessagesHistoryMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p messagesHistoryParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}
	if p.PeerID == 0 {
		return nil, fmt.Errorf("peer_id is required")
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}

	var peer tg.InputPeerClass
	switch p.PeerType {
	case "user":
		peer = &tg.InputPeerUser{UserID: p.PeerID, AccessHash: p.AccessHash}
	case "group":
		peer = &tg.InputPeerChat{ChatID: p.PeerID}
	case "channel":
		peer = &tg.InputPeerChannel{ChannelID: p.PeerID, AccessHash: p.AccessHash}
	default:
		return nil, fmt.Errorf("unsupported peer_type: %s", p.PeerType)
	}

	api := m.tgSvc.API()
	resp, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  peer,
		Limit: p.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}

	modified, ok := resp.AsModified()
	if !ok {
		return []MessageInfo{}, nil
	}

	users := make(map[int64]string)
	for _, u := range modified.GetUsers() {
		if user, ok := u.(*tg.User); ok {
			first, _ := user.GetFirstName()
			last, _ := user.GetLastName()
			name := first
			if last != "" {
				name += " " + last
			}
			users[user.ID] = name
		}
	}

	var result []MessageInfo
	for _, m := range modified.GetMessages() {
		msg, ok := m.(*tg.Message)
		if !ok {
			continue
		}

		text := msg.Message
		if text == "" {
			if msg.Media != nil {
				text = "[media]"
			} else {
				text = "[service message]"
			}
		}

		senderName := ""
		if msg.FromID != nil {
			if pu, ok := msg.FromID.(*tg.PeerUser); ok {
				if name, found := users[pu.UserID]; found {
					senderName = name
				}
			}
		}

		result = append(result, MessageInfo{
			ID:         msg.ID,
			Date:       time.Unix(int64(msg.Date), 0).Format("2006-01-02 15:04"),
			Text:       text,
			SenderName: senderName,
			IsOutgoing: msg.Out,
		})
	}

	return result, nil
}
