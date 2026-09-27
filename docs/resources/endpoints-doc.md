# Endpoints — Auth

Base URL: `http://localhost:3000`

Todos los bodies son JSON. Los tokens de refresco **nunca van en el body**: van en headers.

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| GET | `/health` | pública | Liveness/readiness: `200` si la BD responde, `503` si no |
| POST | `/auth/register` | pública | Crea el usuario, su sesión y devuelve los tokens |
| POST | `/auth/login` | pública | Valida credenciales y devuelve los tokens |
| GET | `/auth/refresh-token` | headers de refresco | Rota los tokens (renueva access + refresh + refresh code) |

---

## 1. `POST /auth/register`

**Request**

```json
{
  "userName": "jean",
  "password": "secret123"
}
```

| Campo | Reglas |
|---|---|
| `userName` | requerido, 3–50 caracteres |
| `password` | requerido, mínimo 8 caracteres |

**Response: `201 Created`**

```json
{
  "userData": {
    "id": "98a61206-7539-4bb6-aac0-62bd2db1f31e",
    "createdAt": "2026-09-26T23:35:46.575309Z",
    "userName": "jean",
    "role": "ADM"
  },
  "sessionData": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "createdAt": "2026-09-26T23:35:46.570852Z",
    "isValid": true
  }
}
```

**Headers de respuesta (guardar en el cliente):**

```
X-Refresh-Token: eyJhbGciOiJIUzI1NiIs...
X-Refresh-Code:  eyJleHAiOjE3OTA0NjYwNDYs...
```

**Errores**

| Código | Cuándo |
|---|---|
| `409 Conflict` | el `userName` ya existe |
| `422 Validation Failed` | campos inválidos (username < 3, password < 8, etc.) |

---

## 2. `POST /auth/login`

**Request**

```json
{
  "userName": "jean",
  "password": "secret123"
}
```

**Response: `200 OK`** — misma forma que register, con los datos de la sesión.

**Headers de respuesta:**

```
X-Refresh-Token: <refresh token nuevo>
X-Refresh-Code:  <refresh code nuevo>
```

> **Rotación confirmada:** cada login genera access token, refresh token y
> refresh code **nuevos** y los persiste en `auth.sessions`
> (`UPDATE auth.sessions SET refresh_token=..., refresh_code=...`).

**Errores**

| Código | Cuándo |
|---|---|
| `404 Not Found` | usuario inexistente |
| `400 Validation Error` | password incorrecta o campos inválidos |

---

## 3. `GET /auth/refresh-token`

**Headers requeridos**

```
X-Refresh-Token: <refresh token>
X-Refresh-Code:  <refresh code>
```

**Response: `200 OK`** — misma forma que login (access token nuevo en el body,
refresh token y refresh code nuevos en los headers).

### Flujo de recarga de sesión (confirmado con pruebas)

```
                  ┌─ Headers X-Refresh-Token / X-Refresh-Code presentes? ─┐
                  │  NO  → 401 "Missing or invalid refresh token"        │
                  │  SÍ                                                 │
                  ▼                                                     │
        VerifyToken(refresh token)                                      │
         ├─ VÁLIDO (no vencido) ──► confía en el JWT,                   │
         │                          busca el usuario por ID              │
         │                          (refresh normal)                     │
         │                                                              │
         └─ VENCIDO ─────────────► usa el X-Refresh-Code de los headers │
                                  y busca el usuario por                 │
                                  refresh_code en la BD                  │
                                  (fallback server-side)                 │
                                                                               │
        En ambos casos: genera access (5m) + refresh (72h) + code nuevos    │
        y los persiste en BD (UpdateSessionTokens)                         │
        └─► 200 OK con los tokens nuevos en body + headers                │
```

Comportamientos verificados con pruebas reales:

1. **Refresh token NO vencido** → renueva "normal": responde `200` y rota
   ambos tokens. No necesita validar el refresh code contra la BD
   (confía en la firma/expiración del JWT).
2. **Refresh token VENCIDO** → responde `200` usando el `X-Refresh-Code`
   de los headers como fuente de verdad (`SELECT ... WHERE refresh_code = $1`).
3. **Refresh token vencido + refresh code incorrecto** → `401`
   `{"title":"Refresh code","message":"session not valid"}`.
4. **Faltan los headers** → `401 Unauthorized`.

> **Rotación confirmada:** cada recarga (vencido o no) actualiza
> `refresh_token` **y** `refresh_code` en la BD. El cliente debe reemplazar
> ambos valores guardados con los nuevos headers de la respuesta; si no lo
> hace, la siguiente recarga fallará con `401`.

---

## Tokens

| Token | Vida | Dónde se entrega |
|---|---|---|
| Access token | 5 minutos | `sessionData.accessToken` (body) |
| Refresh token | 72 horas | header `X-Refresh-Token` |
| Refresh code | n/a (rotado con cada uso) | header `X-Refresh-Code` |

El access token se envía en rutas protegidas como
`Authorization: Bearer <accessToken>` (middleware `AccessToken`).

---

## Ejemplos con curl

```bash
# registrar
curl -i -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"userName":"jean","password":"secret123"}'

# login (guarda los headers X-Refresh-*)
curl -i -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"userName":"jean","password":"secret123"}'

# recargar la sesión (rotar tokens)
curl -i http://localhost:3000/auth/refresh-token \
  -H "X-Refresh-Token: $REFRESH_TOKEN" \
  -H "X-Refresh-Code: $REFRESH_CODE"
```

---

## Correr con Docker (compose)

```bash
# 1. configurar envs (obligatorio: JWT_SECRET)
cp .env.example .env
#    completar JWT_SECRET (p. ej.: openssl rand -hex 32) y la password de la BD

# 2. levantar BD + backend
docker compose up -d

# 3. verificar
docker compose ps          # app y db deben estar (healthy)
curl http://localhost:3000/health   # {"status":"ok"}
```

`docker-compose.yml`:

| Servicio | Imagen | Notas |
|---|---|---|
| `db` | `postgres:16-alpine` | volumen `pgdata`, healthcheck `pg_isready`, crea el schema `dw_lubrisur` en el primer arranque (`docker/db/init/`), publicado **solo en 127.0.0.1** |
| `app` | build del `Dockerfile` (`scrapper-ai:prod`) | espera a que `db` esté `healthy`, `restart: unless-stopped`, `stop_grace_period: 15s` |

**Envs que recibe el backend** (desde `.env` vía compose):

| Env | Requerido | Default | Descripción |
|---|---|---|---|
| `JWT_SECRET` | **sí** | — | sin él (o con el default) en `ENVIROMENT=production` la app **no arranca** |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | no | `postgres` / `postgres` / `scrapper` | credenciales (las mismas para el servicio `db`) |
| `DB_SSLMODE` | no | `disable` | |
| `ENVIROMENT` | no | `production` | `production` activa el guard de `JWT_SECRET` |
| `APP_PORT` | no | `3000` | puerto publicado del backend |
| `DB_PUBLISHED_PORT` | no | `5432` | puerto local de la BD (solo loopback) |
| `CORS_ORIGINS` | no | localhost:34115/5173 | orígenes separados por coma |

Nota: dentro de la red del compose la app se conecta con `DB_HOST=db` (nombre del servicio), no con `localhost`.

### Verificado en producción

- Imagen `scratch` de **~9 MB**, binario estático (UPX), corre como no-root `65532:65532`, sin shell.
- `.env` **no** entra a la imagen (`.dockerignore`) → los secretos viajan solo por compose.
- `HEALTHCHECK` interno (`/app/server healthcheck` → `GET /health`, sin shell/curl):
  responde `200` con BD, `503` sin ella → `docker compose ps` marca unhealthy.
- Falta `JWT_SECRET` en producción → exit 1 con mensaje claro (antes de conectar).
- BD caída al arrancar → exit 1 (no levanta a medias).
- `SIGTERM` (`docker stop` / `compose down`) → **graceful shutdown** (drena requests 10s).
- Recuperación: si la BD vuelve, `/health` recupera `200` solo.

```bash
docker compose down          # parar (el volumen pgdata persiste)
docker compose down -v       # parar y borrar la BD
```
