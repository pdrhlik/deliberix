# Deliberix

Deliberative survey platform (wiki survey) where participants rate statements, mark them as important, and submit new ones for moderation.

Inspired by Pol.is and All Our Ideas.

## Tech Stack

- **Frontend**: Ionic / Angular / Capacitor
- **Backend**: Go (chi router, mibk/dali ORM)
- **Database**: MariaDB
- **Infra**: Docker Compose

## Prerequisites

- Docker & Docker Compose
- Node.js 24+ and npm (for local Angular CLI)
- Go 1.26+ (for local tooling)

## Setup

```bash
cp .env.example .env
# Edit .env with your values (defaults work for local dev)
```

## Development

### Start all services

```bash
docker-compose -f docker-compose-dev.yml up
```

This starts:

| Service | URL                   | Description                   |
| ------- | --------------------- | ----------------------------- |
| Client  | http://localhost:4200 | Angular dev server (ng serve) |
| Server  | http://localhost:8080 | Go API with hot reload (air)  |
| MariaDB | localhost:3307        | Database                      |
| Adminer | http://localhost:8081 | Database admin UI             |

### Hot reload

- **Client**: changes to `client/src/` are picked up automatically
- **Server**: changes to `server/*.go` trigger automatic rebuild via air

### Rebuild after dependency changes

```bash
# After changes to client/package.json
docker-compose -f docker-compose-dev.yml up --build client

# After changes to server/go.mod
docker-compose -f docker-compose-dev.yml up --build server
```

### Rebuild everything from scratch

```bash
docker-compose -f docker-compose-dev.yml up --build
```

### Reset database

```bash
docker-compose -f docker-compose-dev.yml down -v
docker-compose -f docker-compose-dev.yml up
```

The `-v` flag removes the MariaDB volume. Schema is re-applied from `db/schema.sql` on next start.

### Adminer login

- System: **MySQL**
- Server: **mariadb**
- Username: **deliberix**
- Password: _(MARIADB_PASSWORD from .env)_
- Database: **deliberix**

## Production

```bash
docker-compose --env-file .env.production up -d
```

Services bind to `127.0.0.1` only. Set up a reverse proxy (Caddy, nginx, Traefik) to handle TLS and routing.

## API

All endpoints are prefixed with `/api/v1/`. Health check:

```bash
curl http://localhost:8080/api/v1/health
```
