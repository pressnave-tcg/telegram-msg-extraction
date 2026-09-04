package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gotd/td/tg"

	"github.com/pressnave-tcg/telegram-msg-extraction/internal/config"
	"github.com/pressnave-tcg/telegram-msg-extraction/internal/miner"
	"github.com/pressnave-tcg/telegram-msg-extraction/internal/telegramx"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	runner, err := telegramx.New(cfg)
	if err != nil {
		return err
	}

	switch args[0] {
	case "groups":
		return runner.Run(ctx, func(ctx context.Context, api *tg.Client) error {
			return listGroups(ctx, api)
		})

	case "mine":
		fs := flag.NewFlagSet("mine", flag.ContinueOnError)
		groupName := fs.String("group", "", "exact Telegram group title; empty mines all groups")
		limit := fs.Int("limit", 1000, "maximum messages per group; 0 means no limit")
		sinceRaw := fs.String("since", "", "only messages at/after YYYY-MM-DD or RFC3339 time")
		output := fs.String("output", "data/messages.jsonl", "JSONL output path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *limit < 0 {
			return fmt.Errorf("--limit cannot be negative")
		}

		since, err := parseSince(*sinceRaw)
		if err != nil {
			return err
		}

		return runner.Run(ctx, func(ctx context.Context, api *tg.Client) error {
			result, err := miner.Mine(ctx, api, miner.Options{
				GroupName: *groupName,
				Limit:     *limit,
				Since:     since,
				Output:    *output,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Exported %d messages from %d group(s) to %s\n",
				result.Messages, result.Groups, result.Output)
			return nil
		})

	case "help", "-h", "--help":
		printUsage()
		return nil

	default:
		return fmt.Errorf("unknown command %q; use groups or mine", args[0])
	}
}

func listGroups(ctx context.Context, api *tg.Client) error {
	groups, err := telegramx.ListGroups(ctx, api)
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}

	for i, group := range groups {
		fmt.Printf("%3d  %-12s  %d  %s\n",
			i+1,
			group.Group.Kind,
			group.Group.ID,
			group.Group.Title,
		)
	}
	return nil
}

func parseSince(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, raw)
		if err == nil {
			t = t.UTC()
			return &t, nil
		}
	}

	return nil, fmt.Errorf("--since must be YYYY-MM-DD or RFC3339")
}

func printUsage() {
	fmt.Println(`tgminer - mine messages from Telegram groups your own account can access

Usage:
  tgminer groups
  tgminer mine [options]

Mine options:
  --group "Group title"   Mine one exact group title. Omit to mine all groups.
  --limit 1000            Maximum messages per group. 0 means no limit.
  --since 2026-09-01      Only messages at/after this date or RFC3339 timestamp.
  --output path.jsonl     Output JSONL file.

Required environment:
  APP_ID
  APP_HASH
  TELEGRAM_PHONE

Optional:
  TELEGRAM_PASSWORD       Required only when Telegram 2FA is enabled.
  SESSION_FILE            Defaults to .data/session.json.`)
}
