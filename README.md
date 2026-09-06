## Stack

- [Gin](https://github.com/gin-gonic/gin) — HTTP router
- [GORM](https://gorm.io/) + PostgreSQL — persistence
- [golang-jwt](https://github.com/golang-jwt/jwt) — auth tokens

## Getting started

### Prerequisites

- Go 1.25+
- Docker (for the local PostgreSQL instance)

## Development

```bash
make test   # run tests
make up     # start PostgreSQL
make down   # stop PostgreSQL
make run    # run the server
```

Tables are auto-migrated on startup.

### Configuration

Configuration is read from environment variables (a `.env` file is loaded automatically if present):

| Variable      | Default                 | Description                                                        |
| ------------- | ----------------------- | ------------------------------------------------------------------ |
| `SERVER_PORT` | `8080`                  | Port the HTTP server listens on                                    |
| `BASE_URL`    | `http://localhost:8080` | Base URL used to build short links                                 |
| `DB_HOST`     | `localhost`             | PostgreSQL host                                                    |
| `DB_PORT`     | `5432`                  | PostgreSQL port                                                    |
| `DB_USER`     | `postgres`              | PostgreSQL user                                                    |
| `DB_PASSWORD` | `postgres`              | PostgreSQL password                                                |
| `DB_NAME`     | `url_shortener`         | PostgreSQL database name                                           |
| `DB_SSLMODE`  | `disable`               | PostgreSQL SSL mode                                                |
| `JWT_SECRET`  | —                       | Secret used to sign JWTs. **Required** — the server exits if unset |

## API

### `POST /api/register`

```bash
curl --request POST \
  --url http://localhost:8080/api/register \
  --header 'content-type: application/json' \
  --data '{
  "email": "test@hotmail.com",
  "password": "12345678"
}'
```

### `POST /api/login`

```bash
curl --request POST \
  --url http://localhost:8080/api/login \
  --header 'content-type: application/json' \
  --data '{
  "email": "test@hotmail.com",
  "password": "12345678"
}'
```

### `POST /api/shorten`

```bash
curl --request POST \
  --url http://localhost:8080/api/shorten \
  --header 'content-type: application/json' \
  --data '{
  "original_url": "https://facebook.com"
}'
```

### `GET /api/urls`

```bash
curl --request GET \
  --url http://localhost:8080/api/urls \
  --header 'authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1IiwiZXhwIjoxNzg4NzY1NjI0LCJpYXQiOjE3ODg2NzkyMjR9.RRjIqQORtUOtxHVpU2MSZfZnaxjCHUuaBAgETnf0ujs'
```

### `DELETE /api/urls/:id`

```bash
curl --request DELETE \
  --url http://localhost:8080/api/urls/1 \
  --header 'authorization: Bearer <token>'
```

### `GET /:code`

```bash
curl --request GET \
  --url http://localhost:8080/1MLMZqp
```
