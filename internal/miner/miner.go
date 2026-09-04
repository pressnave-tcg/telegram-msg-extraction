package miner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"

	"github.com/pressnave-tcg/telegram-msg-extraction/internal/model"
	"github.com/pressnave-tcg/telegram-msg-extraction/internal/telegramx"
)

var errStop = errors.New("stop history iteration")

type Options struct {
	GroupName string
	Limit     int
	Since     *time.Time
	Output    string
}

type Result struct {
	Groups   int
	Messages int
	Output   string
}

func Mine(ctx context.Context, api *tg.Client, opts Options) (Result, error) {
	groups, err := telegramx.ListGroups(ctx, api)
	if err != nil {
		return Result{}, err
	}

	selected := selectGroups(groups, opts.GroupName)
	if len(selected) == 0 {
		if opts.GroupName == "" {
			return Result{}, fmt.Errorf("no Telegram groups found")
		}
		return Result{}, fmt.Errorf("group %q not found", opts.GroupName)
	}

	if opts.Output == "" {
		opts.Output = "data/messages.jsonl"
	}

	if dir := filepath.Dir(opts.Output); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return Result{}, fmt.Errorf("create output directory: %w", err)
		}
	}

	f, err := os.OpenFile(opts.Output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return Result{}, fmt.Errorf("open output: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	total := 0

	for _, group := range selected {
		count, err := mineGroup(ctx, api, group, opts, enc)
		if err != nil {
			return Result{}, fmt.Errorf("mine %q: %w", group.Group.Title, err)
		}
		total += count
	}

	return Result{
		Groups:   len(selected),
		Messages: total,
		Output:   opts.Output,
	}, nil
}

func selectGroups(groups []telegramx.GroupDialog, name string) []telegramx.GroupDialog {
	if strings.TrimSpace(name) == "" {
		return groups
	}

	name = strings.TrimSpace(name)
	var selected []telegramx.GroupDialog
	for _, group := range groups {
		if strings.EqualFold(group.Group.Title, name) {
			selected = append(selected, group)
		}
	}
	return selected
}

func mineGroup(
	ctx context.Context,
	api *tg.Client,
	group telegramx.GroupDialog,
	opts Options,
	enc *json.Encoder,
) (int, error) {
	count := 0

	err := query.Messages(api).GetHistory(group.Dialog.Peer).ForEach(ctx, func(_ context.Context, elem messages.Elem) error {
		msg, ok := elem.Msg.(*tg.Message)
		if !ok {
			return nil
		}

		msgTime := time.Unix(int64(msg.Date), 0).UTC()
		if opts.Since != nil && msgTime.Before(*opts.Since) {
			return errStop
		}

		if opts.Limit > 0 && count >= opts.Limit {
			return errStop
		}

		record := model.Message{
			Group:     group.Group,
			MessageID: msg.ID,
			Date:      msgTime,
			Text:      msg.Message,
			Outgoing:  msg.Out,
			Sender:    senderFromMessage(msg, elem),
		}

		if file, ok := elem.File(); ok {
			record.Attachment = &model.Attachment{
				Name:     file.Name,
				MIMEType: file.MIMEType,
			}
		}

		if err := enc.Encode(record); err != nil {
			return fmt.Errorf("encode message %d: %w", msg.ID, err)
		}

		count++
		return nil
	})

	if err != nil && !errors.Is(err, errStop) {
		return count, err
	}
	return count, nil
}

func senderFromMessage(msg *tg.Message, elem messages.Elem) model.Sender {
	switch p := msg.FromID.(type) {
	case *tg.PeerUser:
		if user, ok := elem.Entities.User(p.UserID); ok {
			name := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
			return model.Sender{
				ID:       p.UserID,
				Name:     name,
				Username: user.Username,
				Kind:     "user",
			}
		}
		return model.Sender{ID: p.UserID, Kind: "user"}

	case *tg.PeerChat:
		if chat, ok := elem.Entities.Chat(p.ChatID); ok {
			return model.Sender{
				ID:   p.ChatID,
				Name: chat.Title,
				Kind: "chat",
			}
		}
		return model.Sender{ID: p.ChatID, Kind: "chat"}

	case *tg.PeerChannel:
		if channel, ok := elem.Entities.Channel(p.ChannelID); ok {
			return model.Sender{
				ID:       p.ChannelID,
				Name:     channel.Title,
				Username: channel.Username,
				Kind:     "channel",
			}
		}
		return model.Sender{ID: p.ChannelID, Kind: "channel"}

	default:
		return model.Sender{}
	}
}
