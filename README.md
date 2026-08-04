# AniWays

An anime tracking platform for discovering anime, managing a personal library, and
keeping lists in sync across AniList and MyAnimeList. Go backend, SvelteKit frontend,
Docker-based development.

> **Archived — no longer maintained.**
> Development has moved to a rewrite. This repository is kept for reference. There is
> no hosted instance; `aniways.xyz` is no longer online. The code is not deployable as
> it stands — the streaming proxy component has been removed (see below).

> **Disclaimer**
> Educational project. It integrates with third-party sources and provides no content
> of its own. Use responsibly and comply with the terms of any service you connect to.

---

## Features

- **Anime discovery and metadata** — data from AniList and MyAnimeList
- **Personal library management** — track anime with custom statuses (watching, completed, planned)
- **OAuth authentication** — login via AniList and MAL
- **Multi-source sync** — keep lists consistent across platforms
- **Episode sources** — resolves episode links from third-party providers
- **REST API** — OpenAPI-documented backend
- **Web app** — SvelteKit and TailwindCSS
- **Background jobs** — workers handle scraping, syncing, and metadata updates

---

## Architecture

Three services:

| Service          | Path         | Description                                             |
| ---------------- | ------------ | ------------------------------------------------------- |
| **API Server**   | `cmd/api`    | REST API handling authentication, library, and metadata |
| **Worker**       | `cmd/worker` | Background processor for scraping and sync jobs         |
| **Web Frontend** | `web/`       | SvelteKit UI                                            |

### Removed component

A dedicated HLS proxy (`cmd/proxy`) previously sat alongside these and has been
removed from this repository, along with its Docker and Makefile wiring. The API still
returns proxy-style URLs and the web player still expects them, so playback does not
work in this tree. Since the project is archived, those references were left in place
rather than unpicked.

### Tech stack

**Backend** — Go 1.24+ with Chi, PostgreSQL + SQLC, Redis, GraphQL (AniList), JWT + OAuth2
**Frontend** — SvelteKit, TypeScript, TailwindCSS, Shadcn-Svelte, Vite
**Infra** — Docker, `golang-migrate`, cron-driven background tasks

---

## Worker CLI

```bash
./worker                              # daemon mode

./worker scrape recently-updated      # fetch 40 most recent anime
./worker scrape all-recently-updated  # fetch all recent anime
./worker scrape full-seed             # complete database seed (A–Z)

./worker library retry-failed         # retry failed syncs
./worker auth refresh-tokens          # refresh OAuth tokens
```

---

## Local development

```bash
make setup      # install dependencies, generate code, start containers
```

```bash
make dev-api                # API server
make dev-worker             # worker
cd web && bun run dev       # frontend
```

API docs at `http://localhost:8080/swagger/`, frontend at `http://localhost:3000`.

---

Built by [Coeeter](https://github.com/coeeter). Superseded by a rewrite built on Bun,
TanStack Start, Effect, and Drizzle.
