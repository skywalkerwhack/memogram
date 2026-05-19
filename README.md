# memogram

`memogram` is a Telegram bot that saves messages into your [Memos](https://www.usememos.com/) instance.

It is designed for a simple workflow:

- send text to the bot and it becomes a memo
- send photos, documents, voice messages, or videos and they are attached to that memo
- forward messages from Telegram and the original sender is noted in the memo
- search your saved memos from Telegram
- change memo visibility, pin, or delete a memo from inline buttons

## What It Does

After you link your Telegram account to a Memos access token, you can use the bot as a personal capture inbox.

Supported Telegram actions:

- plain text messages
- captions on media
- documents
- photos
- voice messages
- videos
- forwarded messages

Supported bot commands:

- `/start <access_token>`: link your Telegram account to Memos
- `/help`: show help
- `/search <words>`: search your memos
- `/account`: show whether your Telegram account is linked to Memos
- `/me`: alias of `/account`
- `/unlink`: remove the saved Memos token for your Telegram account
- `/ping`: show admin-only backend diagnostics

## How It Works

`memogram` connects to:

- Telegram Bot API
- your Memos server API

Each Telegram user links their own Memos access token with `/start <access_token>`. The bot stores that token in a local data file so later messages can be saved to the correct Memos account.

## Requirements

- Go `1.25+` if you want to run from source
- a Telegram bot token from BotFather
- a reachable Memos instance
- a Memos access token for each Telegram user who will use the bot

## Configuration

The app reads environment variables directly, and also loads a local `.env` file if present.

You can start from `.env.example`:

```env
SERVER_ADDR=dns:localhost:5230
BOT_TOKEN=your_telegram_bot_token
BOT_PROXY_ADDR=
DATA=data.txt
ALLOWED_USERNAMES=
ADMIN_USERNAMES=
```

### Environment Variables

- `SERVER_ADDR`: Memos server address. If it does not start with `http://` or `https://`, `http://` is added automatically. Values like `dns:localhost:5230` are accepted.
- `BOT_TOKEN`: Telegram bot token.
- `BOT_PROXY_ADDR`: optional Telegram Bot API server URL or proxy endpoint.
- `DATA`: path to the local token storage file. Default: `data.txt`.
- `ALLOWED_USERNAMES`: optional comma-separated allowlist of Telegram usernames. If empty, any Telegram username can use the bot.
- `ADMIN_USERNAMES`: optional comma-separated list of Telegram usernames allowed to use `/ping`. If empty, diagnostic `/ping` is unavailable.

## Quick Start

1. Copy the example config:

```bash
cp .env.example .env
```

2. Edit `.env` and set:

- `SERVER_ADDR`
- `BOT_TOKEN`
- optionally `ALLOWED_USERNAMES`
- optionally `ADMIN_USERNAMES` if you want admin-only `/ping` diagnostics

3. Run the bot:

```bash
go run ./cmd/memogram
```

4. In Telegram, open your bot and link your account:

```text
/start <your_memos_access_token>
```

5. Send a message to the bot. It should reply with a saved memo link and action buttons.

## Docker

Build the image:

```bash
docker build -t memogram .
```

Run it:

```bash
docker run --rm \
  --name memogram \
  -e SERVER_ADDR=http://host.docker.internal:5230 \
  -e BOT_TOKEN=your_telegram_bot_token \
  -e ALLOWED_USERNAMES=yourtelegramusername \
  -e ADMIN_USERNAMES=yourtelegramusername \
  -v "$(pwd)/data.txt:/app/data.txt" \
  memogram
```

If you use a `.env` file, you can also mount or pass it in your usual Docker workflow.

## Cross-Platform Build

There is a helper script for release-style builds:

```bash
./scripts/build.sh
```

By default it builds:

- `linux/amd64`
- `freebsd/amd64`

Artifacts are written to `build/`.

## Notes

- The token store is a plain local file managed by the bot. Protect it like a secret.
- If `ALLOWED_USERNAMES` is set, users without a Telegram username cannot use the bot.
- If `ADMIN_USERNAMES` is empty, `/ping` diagnostics are disabled for everyone.
- Media albums are grouped so multiple items from the same Telegram media group attach to a single memo.
- The bot replies with a direct memo link based on the Memos instance URL when available.

## Development

Run tests:

```bash
go test ./...
```

## License

GPL-3.0-or-later. See [COPYING](COPYING).
