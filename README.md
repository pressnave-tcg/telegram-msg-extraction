# Telegram Message Extraction

A Go CLI for reading and exporting messages from Telegram groups that **your own Telegram user account** can already access.

It uses Telegram MTProto through [`github.com/gotd/td`](https://github.com/gotd/td), not the Telegram Bot API.

## Features

- Authenticate as your Telegram user account.
- Persist the Telegram authorization session locally.
- List joined basic groups and supergroups.
- Exclude direct messages and broadcast channels.
- Export group message history as JSONL.
- Filter by exact group title.
- Filter messages by date.
- Limit the number of messages exported per group.
- Export sender information and attachment metadata when available.

## Requirements

- Go 1.25+
- A Telegram account
- Telegram application credentials (`api_id` and `api_hash`)

Create Telegram application credentials at:

```text
https://my.telegram.org/apps
```

## Clone

```bash
git clone https://github.com/pressnave-tcg/telegram-msg-extraction.git
cd telegram-msg-extraction
```

## Configure

Copy the environment template:

```bash
cp .env.example .env
```

Edit `.env`:

```dotenv
APP_ID=12345678
APP_HASH=your_api_hash
TELEGRAM_PHONE=+971500000000
TELEGRAM_PASSWORD=
SESSION_FILE=.data/session.json
```

`TELEGRAM_PASSWORD` is only required when Telegram two-step verification is enabled.

Load the variables into your shell:

```bash
set -a
source .env
set +a
```

## Install dependencies

```bash
go mod tidy
```

## Build

```bash
make build
```

The binary will be created at:

```text
bin/tgminer
```

## First login

Run:

```bash
./bin/tgminer groups
```

On the first run Telegram sends a login code to your account. Enter that code in the terminal.

The reusable authorization session is stored at `.data/session.json` by default.

> Treat the session file like a credential. Do not commit, upload, or share it.

## List groups

```bash
./bin/tgminer groups
```

Example output:

```text
  1  supergroup    1234567890  Engineering
  2  basic_group   987654321   Family
```

## Mine one group

```bash
./bin/tgminer mine \
  --group "Engineering" \
  --limit 1000 \
  --output data/engineering.jsonl
```

## Mine messages since a date

```bash
./bin/tgminer mine \
  --group "Engineering" \
  --since 2026-09-01 \
  --limit 0 \
  --output data/engineering-september.jsonl
```

`--since` accepts either a date:

```text
2026-09-01
```

or RFC3339:

```text
2026-09-01T08:30:00Z
```

## Mine all joined groups

```bash
./bin/tgminer mine \
  --limit 500 \
  --output data/all-groups.jsonl
```

The message limit is applied **per group**. Set `--limit 0` for no message-count limit.

## Output format

Each line of the output file is an independent JSON object:

```json
{
  "group": {
    "id": 1234567890,
    "title": "Engineering",
    "kind": "supergroup"
  },
  "message_id": 4812,
  "date": "2026-09-04T05:42:10Z",
  "text": "Production latency is high",
  "outgoing": false,
  "sender": {
    "id": 12345,
    "name": "Example User",
    "username": "example",
    "kind": "user"
  }
}
```

For supported Telegram file messages, attachment metadata is added:

```json
{
  "attachment": {
    "name": "report.pdf",
    "mime_type": "application/pdf"
  }
}
```

The CLI records metadata only; it does not download attached files.

## Project structure

```text
cmd/tgminer/           CLI entry point
internal/config/       Environment configuration
internal/telegramx/    Telegram authentication and group discovery
internal/miner/        Message history extraction and JSONL export
internal/model/        Export models
```

## Commands

```bash
make build
make fmt
make test
make groups
make mine
```

Or run directly:

```bash
go run ./cmd/tgminer groups
```

```bash
go run ./cmd/tgminer mine --group "Engineering" --limit 100
```

## Security and privacy

Only extract messages from conversations that your own account is legitimately authorized to access.

Keep these out of Git and other shared systems:

- `.env`
- `APP_HASH`
- `TELEGRAM_PASSWORD`
- `.data/session.json`
- exported Telegram messages

The included `.gitignore` excludes these local files by default.

## Possible next improvements

- SQLite/PostgreSQL persistence
- Incremental checkpoints
- Live message streaming
- Keyword and regex filters
- n8n webhook forwarding
- Incident detection
- Embeddings/RAG over group history
- AI summaries and topic extraction
