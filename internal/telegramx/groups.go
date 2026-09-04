package telegramx

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"

	"github.com/pressnave-tcg/telegram-msg-extraction/internal/model"
)

type GroupDialog struct {
	Group  model.Group
	Dialog dialogs.Elem
}

func ListGroups(ctx context.Context, api *tg.Client) ([]GroupDialog, error) {
	var groups []GroupDialog

	err := query.GetDialogs(api).ForEach(ctx, func(_ context.Context, dlg dialogs.Elem) error {
		if dlg.Deleted() {
			return nil
		}

		group, ok := groupFromDialog(dlg)
		if !ok {
			return nil
		}

		groups = append(groups, GroupDialog{
			Group:  group,
			Dialog: dlg,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list Telegram dialogs: %w", err)
	}

	return groups, nil
}

func groupFromDialog(dlg dialogs.Elem) (model.Group, bool) {
	switch p := dlg.Peer.(type) {
	case *tg.InputPeerChat:
		chat, ok := dlg.Entities.Chat(p.ChatID)
		if !ok {
			return model.Group{}, false
		}

		return model.Group{
			ID:    p.ChatID,
			Title: chat.Title,
			Kind:  "basic_group",
		}, true

	case *tg.InputPeerChannel:
		channel, ok := dlg.Entities.Channel(p.ChannelID)
		if !ok || channel.Broadcast {
			return model.Group{}, false
		}

		return model.Group{
			ID:    p.ChannelID,
			Title: channel.Title,
			Kind:  "supergroup",
		}, true

	default:
		return model.Group{}, false
	}
}
