package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"github.com/tg-manager/internal/telegram"
)

type DialogInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	UnreadCount int    `json:"unread_count"`
	LastMessage string `json:"last_message"`
	AccessHash  int64  `json:"access_hash"`
}

// dialogs.list
type DialogsListMethod struct {
	tgSvc *telegram.Service
}

type dialogsListParams struct {
	Limit int `json:"limit"`
}

func (m *DialogsListMethod) Name() string { return "dialogs.list" }
func (m *DialogsListMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p dialogsListParams
	if len(params) > 0 {
		_ = json.Unmarshal(params, &p)
	}
	if p.Limit <= 0 {
		p.Limit = 50
	}

	api := m.tgSvc.API()
	return fetchDialogs(ctx, api, p.Limit, "")
}

// channels.list
type ChannelsListMethod struct {
	tgSvc *telegram.Service
}

func (m *ChannelsListMethod) Name() string { return "channels.list" }
func (m *ChannelsListMethod) Execute(ctx context.Context, _ json.RawMessage) (interface{}, error) {
	api := m.tgSvc.API()
	return fetchDialogs(ctx, api, 100, "channel")
}

func fetchDialogs(ctx context.Context, api *tg.Client, limit int, filterType string) ([]DialogInfo, error) {
	resp, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get dialogs: %w", err)
	}

	modified, ok := resp.AsModified()
	if !ok {
		return []DialogInfo{}, nil
	}

	users := make(map[int64]*tg.User)
	for _, u := range modified.GetUsers() {
		if user, ok := u.(*tg.User); ok {
			users[user.ID] = user
		}
	}

	chats := make(map[int64]*tg.Chat)
	channels := make(map[int64]*tg.Channel)
	for _, c := range modified.GetChats() {
		switch chat := c.(type) {
		case *tg.Chat:
			chats[chat.ID] = chat
		case *tg.Channel:
			channels[chat.ID] = chat
		}
	}

	topMessages := make(map[int]*tg.Message)
	for _, m := range modified.GetMessages() {
		if msg, ok := m.(*tg.Message); ok {
			topMessages[msg.ID] = msg
		}
	}

	var result []DialogInfo
	for _, d := range modified.GetDialogs() {
		dlg, ok := d.(*tg.Dialog)
		if !ok {
			continue
		}

		info := resolveDialogInfo(dlg, users, chats, channels)
		if filterType != "" && info.Type != filterType {
			continue
		}

		// Add last message preview
		if msg, ok := topMessages[dlg.TopMessage]; ok {
			ts := time.Unix(int64(msg.Date), 0).Format("2006-01-02 15:04")
			text := msg.Message
			if len(text) > 80 {
				text = text[:80] + "..."
			}
			if text == "" {
				text = "[media/service message]"
			}
			info.LastMessage = fmt.Sprintf("%s: %s", ts, text)
		}

		result = append(result, info)
	}

	return result, nil
}

func resolveDialogInfo(dlg *tg.Dialog, users map[int64]*tg.User, chats map[int64]*tg.Chat, channels map[int64]*tg.Channel) DialogInfo {
	var info DialogInfo
	info.UnreadCount = dlg.UnreadCount

	switch p := dlg.Peer.(type) {
	case *tg.PeerUser:
		info.ID = p.UserID
		info.Type = "user"
		if u, ok := users[p.UserID]; ok {
			firstName, _ := u.GetFirstName()
			lastName, _ := u.GetLastName()
			info.Name = strings.TrimSpace(firstName + " " + lastName)
			info.AccessHash, _ = u.GetAccessHash()
		} else {
			info.Name = fmt.Sprintf("User#%d", p.UserID)
		}
	case *tg.PeerChat:
		info.ID = p.ChatID
		info.Type = "group"
		if c, ok := chats[p.ChatID]; ok {
			info.Name = c.Title
		} else {
			info.Name = fmt.Sprintf("Chat#%d", p.ChatID)
		}
	case *tg.PeerChannel:
		info.ID = p.ChannelID
		info.Type = "channel"
		if ch, ok := channels[p.ChannelID]; ok {
			info.Name = ch.Title
			info.AccessHash, _ = ch.GetAccessHash()
		} else {
			info.Name = fmt.Sprintf("Channel#%d", p.ChannelID)
		}
	}

	return info
}
