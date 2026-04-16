# Production Deployment — edemos.patrio.dev

Staging environment. Self-contained Docker Compose setup with no conflicts with other projects on the server.

## Ports used (all bound to 127.0.0.1)

- 3407 — MariaDB
- 8180 — Go server (API + Astro website)
- 4280 — Angular client (nginx)

## 1. Clone and configure

```bash
cd ~/projects
git clone <repo-url> edemos
cd edemos
cp .env.example .env
```

Edit `.env` with production values:

```env
MARIADB_ROOT_PASSWORD=<generate>
MARIADB_PASSWORD=<generate>
JWT_SECRET=<generate-min-32-chars>
SMTP_HOST=mail.hory.app
SMTP_PORT=465
SMTP_USER=info@atrain.app
SMTP_PASSWORD=<from-dev-env>
SMTP_FROM=info@atrain.app
BASE_URL=https://edemos.patrio.dev
SERVE_WEBSITE=true
```

Generate secrets with: `openssl rand -base64 24`

## 2. Build and start

```bash
docker compose up -d --build
```

Watch logs until all services are ready:

```bash
docker compose logs -f
```

MariaDB auto-applies `db/schema.sql` on first start. No manual migration needed.

## 3. Verify containers

```bash
docker compose ps
curl -s http://127.0.0.1:8180/api/v1/survey/public
curl -s http://127.0.0.1:4280/ | head -c 200
```

## 4. Add Caddy block

Add to `/etc/caddy/Caddyfile` (or wherever the global config lives):

```caddy
edemos.patrio.dev {
    handle /api/* {
        reverse_proxy localhost:8180
    }

    @website {
        path / /about /contact /cs /cs/* /_astro/*
    }
    handle @website {
        reverse_proxy localhost:8180
    }

    handle {
        reverse_proxy localhost:4280
    }
}
```

```bash
sudo systemctl reload caddy
```

## 5. Verify end-to-end

```bash
curl -s https://edemos.patrio.dev/              # Astro homepage
curl -s https://edemos.patrio.dev/login          # Angular SPA (HTML)
curl -s https://edemos.patrio.dev/api/v1/survey/public  # API (JSON)
```

## Updating

```bash
cd ~/projects/edemos
git pull
docker compose up -d --build
```

## Notes

- Angular prod build uses `apiUrl: ""` (relative paths) — Caddy proxies `/api/*` to the Go server on the same domain. No build-time API URL needed.
- DB data persists in Docker volume `edemos_mariadb-data`.
- SMTP credentials reused from atrain (staging only).
- No port/container/network conflicts with mapriot or other projects.
