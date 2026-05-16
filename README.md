# Memogram

Memogram is a Telegram bot for saving Telegram messages to a
[Memos](https://www.usememos.com/) instance.

Send the bot a message, photo, voice message, video, or document and it will
create a memo in Memos. After saving, the bot replies with a link to the memo
and buttons for changing visibility or pinning it.

This project continues the work from
[usememos/telegram-integration](https://github.com/usememos/telegram-integration)
as a standalone bot.

## Features

- Save plain text messages as memos.
- Save photos, voice messages, videos, and documents as memo attachments.
- Use captions as memo content for uploaded files.
- Preserve supported Telegram formatting.
- Include forwarded-message source information when Telegram provides it.
- Link each Telegram account to its own Memos access token.
- Search saved memos with `/search`.
- Check bot and account status with `/status`.
- Update memo visibility and pin state from Telegram inline buttons.

## Requirements

You need:

- A running Memos instance.
- A Telegram bot token from [BotFather](https://t.me/BotFather).
- A Memos access token for each Telegram user who will use the bot.
- Either Docker or Go 1.25+.

## Quick Start

### Run with Docker

```sh
docker run -d --name memogram \
  -e SERVER_ADDR=http://host.docker.internal:5230 \
  -e BOT_TOKEN=your_telegram_bot_token \
  -e DATA=/app/data/data.txt \
  -v memogram-data:/app/data \
  conch0601/memogram
```

Replace:

- `SERVER_ADDR` with the address of your Memos instance.
- `BOT_TOKEN` with the token from BotFather.

Check logs:

```sh
docker logs -f memogram
```

Docker notes:

- `http://host.docker.internal:5230` usually works on Docker Desktop when
  Memos runs on the same machine.
- On Linux, `host.docker.internal` may not exist by default. Use a reachable
  host/IP instead, or put Memogram and Memos on the same Docker network.
- The `memogram-data` volume stores linked Telegram access tokens.

### Run with Go

Create a `.env` file in the project directory:

```env
SERVER_ADDR=http://localhost:5230
BOT_TOKEN=your_telegram_bot_token
BOT_PROXY_ADDR=
DATA=data.txt
ALLOWED_USERNAMES=
```

Then build and run:

```sh
go build -o build/memogram ./cmd/memogram
./build/memogram
```

## Configuration

Memogram reads configuration from environment variables. If a `.env` file is
present in the working directory, it is loaded automatically.

| Variable | Required | Description |
| --- | --- | --- |
| `SERVER_ADDR` | Yes | Address of the Memos server. Examples: `localhost:5230`, `dns:localhost:5230`, `http://localhost:5230`, `https://memos.example.com`. If no scheme is provided, Memogram uses `http://`. |
| `BOT_TOKEN` | Yes | Telegram bot token from BotFather. |
| `BOT_PROXY_ADDR` | No | Custom Telegram Bot API server URL. |
| `DATA` | No | Path to the file used to store linked Telegram user access tokens. Defaults to `data.txt`. The file is created automatically if it does not exist. |
| `ALLOWED_USERNAMES` | No | Comma-separated Telegram usernames allowed to use the bot, without `@`. Leave empty to allow any Telegram user. |

### Restrict access to specific Telegram users

```env
ALLOWED_USERNAMES=alex,john,emily
```

Rules:

- Matching is case-insensitive.
- Surrounding whitespace is ignored.
- Do not include `@`.
- If this setting is used, Telegram accounts without a username cannot use the
  bot.

## Using the Bot

### 1. Link a Telegram account

Send this message to the bot:

```text
/start <memos_access_token>
```

The bot verifies the token and stores it for that Telegram account.

### 2. Save content

After linking your account, send any of the following:

- A text message to create a text memo.
- A photo, voice message, video, or document to create a memo with an
  attachment.
- A captioned upload to use the caption as the memo content.
- A forwarded message to include source information when Telegram exposes it.

After saving, the bot replies with a memo link and buttons for:

- `Public`
- `Private`
- `Pin`

### 3. Use commands

| Command | Description |
| --- | --- |
| `/start <memos_access_token>` | Link your Telegram account to Memos. |
| `/search <words>` | Search your memos and return up to 10 matches. |
| `/status` | Show backend, storage, access-control, and account-link status. |
| `/ping` | Reply with `Pong!`. |

## Docker with an env file

If you already have a `.env` file:

```sh
docker run -d --name memogram \
  --env-file .env \
  -e DATA=/app/data/data.txt \
  -v memogram-data:/app/data \
  conch0601/memogram
```

## Development

Run tests:

```sh
go test ./...
```

Build a local binary:

```sh
go build -o build/memogram ./cmd/memogram
```

Build release binaries:

```sh
./scripts/build.sh
```

By default, `scripts/build.sh` writes:

- `build/memogram-linux-amd64`
- `build/memogram-freebsd-amd64`

To build different targets:

```sh
TARGETS="linux/amd64 linux/arm64 freebsd/amd64" ./scripts/build.sh
```

## Releases

Prebuilt archives are available from
[GitHub Releases](https://github.com/usememos/memogram/releases).

The release workflow can also publish Docker images for tags such as
`v1.2.3` or `v1.2.3-rc1`.

Required repository secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

Stable release tags also update `latest`.

## License

Memogram is distributed under the license in [COPYING](COPYING).
