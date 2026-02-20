package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

const (
	appID   = 35447762
	appHash = "5d035d8fb6b8d4935f8289b80c5601db"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: "session.json"},
		Device: telegram.DeviceConfig{
			DeviceModel:    "tg-manager",
			SystemVersion:  runtime.GOOS + "/" + runtime.GOARCH,
			AppVersion:     "0.1.0",
			SystemLangCode: "en",
			LangCode:       "en",
		},
	})

	return client.Run(ctx, func(ctx context.Context) error {
		// Authenticate if needed.
		if err := authenticate(ctx, client); err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		fmt.Println("\nAuthenticated successfully!")
		fmt.Println(strings.Repeat("=", 60))

		// List dialogs and recent messages.
		return listDialogs(ctx, client.API())
	})
}

func authenticate(ctx context.Context, client *telegram.Client) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("auth status: %w", err)
	}
	if status.Authorized {
		u := status.User
		firstName, _ := u.GetFirstName()
		lastName, _ := u.GetLastName()
		fmt.Printf("Logged in as: %s %s\n", firstName, lastName)
		return nil
	}

	// If session file exists but we're not authorized, the session was revoked/expired.
	if _, err := os.Stat("session.json"); err == nil {
		fmt.Println("Saved session is no longer valid (revoked or expired). Re-authenticating...")
	}

	fmt.Println("NOTE: Telegram will send a code to your app.")
	fmt.Println("      Do NOT tap 'Terminate all other sessions' on the popup!")

	flow := auth.NewFlow(terminalAuth{}, auth.SendCodeOptions{})
	return flow.Run(ctx, client.Auth())
}

// terminalAuth implements auth.UserAuthenticator by prompting in the terminal.
type terminalAuth struct{}

func (terminalAuth) Phone(_ context.Context) (string, error) {
	return strings.TrimSpace(prompt("Enter phone number (e.g. +1234567890): ")), nil
}

func (terminalAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return strings.TrimSpace(prompt("Enter verification code: ")), nil
}

func (terminalAuth) Password(_ context.Context) (string, error) {
	pw := strings.TrimSpace(prompt("Enter 2FA password: "))
	if pw == "" {
		return "", auth.ErrPasswordNotProvided
	}
	return pw, nil
}

func (terminalAuth) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}

func (terminalAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign up not supported")
}

func prompt(msg string) string {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func listDialogs(ctx context.Context, api *tg.Client) error {
	resp, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      20,
	})
	if err != nil {
		return fmt.Errorf("get dialogs: %w", err)
	}

	modified, ok := resp.AsModified()
	if !ok {
		fmt.Println("No dialogs found.")
		return nil
	}

	// Build lookup maps for users, chats, and channels.
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

	// Build a map of top messages for quick lookup.
	topMessages := make(map[int]*tg.Message)
	for _, m := range modified.GetMessages() {
		if msg, ok := m.(*tg.Message); ok {
			topMessages[msg.ID] = msg
		}
	}

	dialogs := modified.GetDialogs()
	fmt.Printf("\nFound %d dialogs:\n\n", len(dialogs))

	for i, d := range dialogs {
		dlg, ok := d.(*tg.Dialog)
		if !ok {
			continue
		}

		name, peerType, inputPeer := resolveDialog(dlg, users, chats, channels)
		fmt.Printf("[%d] %s (%s) — unread: %d\n", i+1, name, peerType, dlg.UnreadCount)

		// Show the top message preview.
		if msg, ok := topMessages[dlg.TopMessage]; ok {
			ts := time.Unix(int64(msg.Date), 0).Format("2006-01-02 15:04")
			senderName := resolveSender(msg, users)
			text := msg.Message
			if len(text) > 80 {
				text = text[:80] + "..."
			}
			if text == "" {
				text = "[media/service message]"
			}
			fmt.Printf("    Last: %s | %s: %s\n", ts, senderName, text)
		}

		// Fetch recent message history.
		if inputPeer != nil {
			if err := printHistory(ctx, api, inputPeer, name); err != nil {
				fmt.Printf("    (could not fetch history: %v)\n", err)
			}
		}
		fmt.Println()
	}

	return nil
}

func resolveDialog(
	dlg *tg.Dialog,
	users map[int64]*tg.User,
	chats map[int64]*tg.Chat,
	channels map[int64]*tg.Channel,
) (name, peerType string, inputPeer tg.InputPeerClass) {
	switch p := dlg.Peer.(type) {
	case *tg.PeerUser:
		if u, ok := users[p.UserID]; ok {
			firstName, _ := u.GetFirstName()
			lastName, _ := u.GetLastName()
			name = strings.TrimSpace(firstName + " " + lastName)
			accessHash, _ := u.GetAccessHash()
			inputPeer = &tg.InputPeerUser{UserID: p.UserID, AccessHash: accessHash}
		} else {
			name = fmt.Sprintf("User#%d", p.UserID)
		}
		peerType = "user"
	case *tg.PeerChat:
		if c, ok := chats[p.ChatID]; ok {
			name = c.Title
		} else {
			name = fmt.Sprintf("Chat#%d", p.ChatID)
		}
		peerType = "group"
		inputPeer = &tg.InputPeerChat{ChatID: p.ChatID}
	case *tg.PeerChannel:
		if ch, ok := channels[p.ChannelID]; ok {
			name = ch.Title
			accessHash, _ := ch.GetAccessHash()
			inputPeer = &tg.InputPeerChannel{ChannelID: p.ChannelID, AccessHash: accessHash}
		} else {
			name = fmt.Sprintf("Channel#%d", p.ChannelID)
		}
		peerType = "channel"
	default:
		name = "Unknown"
		peerType = "unknown"
	}
	return
}

func resolveSender(msg *tg.Message, users map[int64]*tg.User) string {
	fromID, ok := msg.GetFromID()
	if !ok {
		return "unknown"
	}
	if p, ok := fromID.(*tg.PeerUser); ok {
		if u, found := users[p.UserID]; found {
			firstName, _ := u.GetFirstName()
			return firstName
		}
	}
	return "unknown"
}

func printHistory(ctx context.Context, api *tg.Client, peer tg.InputPeerClass, chatName string) error {
	resp, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  peer,
		Limit: 5,
	})
	if err != nil {
		return err
	}

	modified, ok := resp.AsModified()
	if !ok {
		return nil
	}

	// Build user map from history response.
	histUsers := make(map[int64]*tg.User)
	for _, u := range modified.GetUsers() {
		if user, ok := u.(*tg.User); ok {
			histUsers[user.ID] = user
		}
	}

	msgs := modified.GetMessages()
	if len(msgs) == 0 {
		return nil
	}

	fmt.Printf("    Recent messages (%d):\n", len(msgs))
	for _, m := range msgs {
		msg, ok := m.(*tg.Message)
		if !ok {
			continue
		}
		ts := time.Unix(int64(msg.Date), 0).Format("15:04")
		sender := "unknown"
		if fromID, ok := msg.GetFromID(); ok {
			if p, ok := fromID.(*tg.PeerUser); ok {
				if u, found := histUsers[p.UserID]; found {
					firstName, _ := u.GetFirstName()
					sender = firstName
				}
			}
		}
		text := msg.Message
		if len(text) > 60 {
			text = text[:60] + "..."
		}
		if text == "" {
			text = "[media/service message]"
		}
		fmt.Printf("      [%s] %s: %s\n", ts, sender, text)
	}

	return nil
}
