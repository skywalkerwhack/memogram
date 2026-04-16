# Memogram

Memogram connects a Telegram bot to your Memos instance. Send text, photos, or
documents to the bot and Memogram saves them as memos with optional attachments.

## Features

- Save Telegram text messages as Memos notes.
- Upload Telegram photos and documents as memo resources.
- Link each Telegram user to Memos with their own access token.
- Search your saved memos from Telegram.
- Restrict bot access to selected Telegram usernames.
- Check bot, backend, and account status with `/status`.

## Requirements

- A running [Memos](https://www.usememos.com/) instance.
- A Telegram bot token from [BotFather](https://t.me/BotFather).
- Go 1.25 or Docker, depending on how you want to run Memogram.

## Installation

Download a prebuilt archive from the
[Releases](https://github.com/usememos/memogram/releases) page, or build the
service from source:

```sh
go build -o build/memogram ./cmd/memogram
```

## Configuration

Memogram reads configuration from environment variables. If a `.env` file exists
in the working directory, it is loaded automatically.

Create a `.env` file:

```env
SERVER_ADDR=dns:localhost:5230
BOT_TOKEN=your_telegram_bot_token
BOT_PROXY_ADDR=
DATA=data.txt
ALLOWED_USERNAMES=
```

| Variable | Required | Description |
| --- | --- | --- |
| `SERVER_ADDR` | Yes | Memos server address. Values such as `localhost:5230`, `dns:localhost:5230`, `http://localhost:5230`, and `https://memos.example.com` are accepted. |
| `BOT_TOKEN` | Yes | Telegram bot token. |
| `BOT_PROXY_ADDR` | No | Custom Telegram API server URL, useful when routing bot traffic through a proxy. |
| `DATA` | No | Local file used to store Telegram user access tokens. Defaults to `data.txt`. |
| `ALLOWED_USERNAMES` | No | Comma-separated Telegram usernames allowed to use the bot, without `@`. Leave empty to allow any user. |

### Access Control

Set `ALLOWED_USERNAMES` to limit who can use the bot:

```env
ALLOWED_USERNAMES=alex,john,emily
```

Matching is case-insensitive and trims whitespace. Usernames must not include
the `@` prefix. When this option is set, Telegram accounts without a username
cannot use the bot.

## Running

### Binary

Place `.env` next to the binary, then start the service:

```sh
./memogram
```

### Docker

Build and run the image:

```sh
docker build -t memogram .

docker run -d --name memogram \
  -e SERVER_ADDR=dns:localhost:5230 \
  -e BOT_TOKEN=your_telegram_bot_token \
  -e DATA=/app/data.txt \
  memogram
```

Add optional variables such as `BOT_PROXY_ADDR` and `ALLOWED_USERNAMES` as
needed.

### Docker Compose

Use Compose when running Memogram next to Memos or when you want to keep
configuration in an env file:

```yaml
services:
  memogram:
    build: .
    container_name: memogram
    env_file: .env
    restart: unless-stopped
```

Start it with:

```sh
docker compose up -d
```

## Bot Usage

First, link your Telegram account to Memos:

```text
/start <memos_access_token>
```

After linking, you can use:

| Action | Result |
| --- | --- |
| Send a text message | Creates a memo with the message content. |
| Send a photo or document | Creates a memo and uploads the file as an attachment. |
| `/search <words>` | Searches your memos and returns up to 10 matches. |
| `/status` | Shows server, data file, backend latency, access control, and account link status. |
| `/ping` | Replies with `Pong!`. |

After a memo is created, the bot shows inline actions for changing visibility or
toggling the memo's pinned state.

## Development

Run tests:

```sh
go test ./...
```

Build the default local binary:

```sh
go build -o build/memogram ./cmd/memogram
```

Build the FreeBSD amd64 release binary:

```sh
./scripts/build.sh
```

The FreeBSD build is written to `build/memogram-freebsd-amd64`.
