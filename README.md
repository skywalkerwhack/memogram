<!--
memogram is a Telegram bot for saving messages into a Memos instance.
Copyright (C) 2026  skywalkerwhack

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
-->

# Memogram

Memogram is a Telegram bot for saving messages into a
[Memos](https://www.usememos.com/) instance. Send the bot text, photos, voice
messages, videos, or documents, and it creates a memo with any supported files
attached.

## What It Does

- Saves Telegram messages as Memos notes.
- Uploads photos, voice messages, videos, and documents as memo attachments.
- Preserves supported Telegram text formatting in memo content.
- Adds source information for forwarded messages when Telegram provides it.
- Lets each Telegram user connect with their own Memos access token.
- Searches saved memos from Telegram.
- Restricts bot usage to selected Telegram usernames when configured.
- Reports bot, storage, backend, and account status with `/status`.
- Provides inline actions to make a memo public or private, and to toggle pinning.

## Requirements

- A running Memos instance.
- A Telegram bot token from [BotFather](https://t.me/BotFather).
- Go 1.25 or Docker.

## Configuration

Memogram reads configuration from environment variables. When a `.env` file is
present in the working directory, it is loaded automatically.

Start from the example file:

```sh
cp .env.example .env
```

Then edit `.env`:

```env
SERVER_ADDR=dns:localhost:5230
BOT_TOKEN=your_telegram_bot_token
BOT_PROXY_ADDR=
DATA=data.txt
ALLOWED_USERNAMES=
```

| Variable | Required | Description |
| --- | --- | --- |
| `SERVER_ADDR` | Yes | Memos server address. `localhost:5230`, `dns:localhost:5230`, `http://localhost:5230`, and `https://memos.example.com` are supported. Addresses without `http://` or `https://` use `http://`. |
| `BOT_TOKEN` | Yes | Telegram bot token. |
| `BOT_PROXY_ADDR` | No | Custom Telegram Bot API server URL. |
| `DATA` | No | File used to store Telegram user access tokens. Defaults to `data.txt`. The file is created automatically if it does not exist. |
| `ALLOWED_USERNAMES` | No | Comma-separated Telegram usernames allowed to use the bot, without `@`. Leave empty to allow any Telegram user. |

### Access Control

Use `ALLOWED_USERNAMES` to limit who can use the bot:

```env
ALLOWED_USERNAMES=alex,john,emily
```

Matching is case-insensitive and surrounding whitespace is ignored. Do not
include the `@` prefix. When this setting is used, Telegram accounts without a
username cannot use the bot.

## Run

### Binary

Build Memogram:

```sh
go build -o build/memogram ./cmd/memogram
```

Run it from the directory containing `.env`:

```sh
./build/memogram
```

You can also download a prebuilt archive from the
[Releases](https://github.com/usememos/memogram/releases) page.

### Docker

Build the image:

```sh
docker build -t memogram .
```

Run the container:

```sh
docker run -d --name memogram \
  -e SERVER_ADDR=dns:localhost:5230 \
  -e BOT_TOKEN=your_telegram_bot_token \
  -e DATA=/app/data.txt \
  memogram
```

Add `BOT_PROXY_ADDR` or `ALLOWED_USERNAMES` if your deployment needs them. Mount
a volume for `/app/data.txt` if you want linked Telegram accounts to survive
container replacement.

### Docker Compose

```yaml
services:
  memogram:
    build: .
    container_name: memogram
    env_file: .env
    restart: unless-stopped
```

Start it:

```sh
docker compose up -d
```

### Automated Docker Hub publishing

The GitHub release workflow also publishes a Docker image to Docker Hub when a
tag such as `v1.2.3` or `v1.2.3-rc1` is released.

Set these GitHub repository secrets before using it:

- `DOCKERHUB_USERNAME`: your Docker Hub username.
- `DOCKERHUB_TOKEN`: a Docker Hub access token with permission to push images.

The workflow pushes `${DOCKERHUB_USERNAME}/memogram:<tag>` for every release
tag. For stable tags without a prerelease suffix, it also updates
`${DOCKERHUB_USERNAME}/memogram:latest`.

## Bot Usage

Connect your Telegram account to Memos:

```text
/start <memos_access_token>
```

After the account is connected:

| Telegram action | Result |
| --- | --- |
| Send text | Creates a memo with that text. |
| Send a photo, voice message, video, or document | Creates a memo and uploads the file as an attachment. |
| Send a captioned file | Uses the caption as memo content and attaches the file. |
| Forward a message | Adds forwarded-source information when available. |
| `/search <words>` | Searches your memos and returns up to 10 matches. |
| `/status` | Shows bot, backend, storage, access-control, and account-link status. |
| `/ping` | Replies with `Pong!`. |

When a memo is created, Memogram replies with a link to the memo and inline
buttons for changing visibility or pinning.

## Development

Run tests:

```sh
go test ./...
```

Build a local binary:

```sh
go build -o build/memogram ./cmd/memogram
```

Build the default cross-platform binaries:

```sh
./scripts/build.sh
```

By default, the helper writes `build/memogram-linux-amd64` and
`build/memogram-freebsd-amd64`. Set `TARGETS` to choose other Go
`GOOS/GOARCH` pairs:

```sh
TARGETS="linux/amd64 linux/arm64 freebsd/amd64" ./scripts/build.sh
```

## License

Memogram is distributed under the license in [COPYING](COPYING).
