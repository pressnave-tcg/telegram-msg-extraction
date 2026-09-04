package telegramx

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"github.com/pressnave-tcg/telegram-msg-extraction/internal/config"
)

type Runner struct {
	client *telegram.Client
}

func New(cfg config.Config) (*Runner, error) {
	if dir := filepath.Dir(cfg.SessionFile); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create session directory: %w", err)
		}
	}

	client := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: cfg.SessionFile},
	})

	return &Runner{client: client}, nil
}

func (r *Runner) Run(ctx context.Context, fn func(context.Context, *tg.Client) error) error {
	codePrompt := auth.CodeAuthenticatorFunc(func(_ context.Context, _ *tg.AuthSentCode) (string, error) {
		fmt.Print("Telegram login code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read Telegram login code: %w", err)
		}
		return strings.TrimSpace(code), nil
	})

	flow := auth.NewFlow(
		auth.Env("TELEGRAM_", codePrompt),
		auth.SendCodeOptions{},
	)

	return r.client.Run(ctx, func(ctx context.Context) error {
		if err := r.client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("Telegram authentication failed: %w", err)
		}

		return fn(ctx, r.client.API())
	})
}
