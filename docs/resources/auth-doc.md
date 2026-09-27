# Auth — cómo funciona actualmente

Módulo `modules/auth`. Aplicación de **un solo usuario en app de escritorio (Wails)**:
sin tenants, sin permissions. El esquema en BD es `auth.users` ↔ `auth.sessions` ↔ `auth.roles`.

> Referencia de endpoints y Docker: [endpoints-doc.md](./endpoints-doc.md).
> Arquitectura general: [architecture-doc.md](./architecture-doc.md).

---

## Las 3 credenciales

| Credencial | Vida | Dónde sale | Dónde se envía |
|---|---|---|---|
| **Access token** (JWT HS256) | **5 min** | body → `sessionData.accessToken` | `Authorization: Bearer <token>` |
| **Refresh token** (JWT HS256) | **72 h** | header `X-Refresh-Token` | header `X-Refresh-Token` |
| **Refresh code** (payload.payload, no es JWT) | se rota con cada uso | header `X-Refresh-Code` | header `X-Refresh-Code` |

- El access token viaja **siempre en el body**; el refresh token y el refresh code
  **siempre en headers** (nunca en el body). El servidor los devuelve igual en los
  headers en login/register.
- Claims del JWT: `{"user_data": {"ID": "<uuid>", "UserName": "..."}, "exp", "iat"}`.
- El refresh code es `payloadDelAccessToken.payloadDelRefreshToken` (base64url de
  ambos payloads, sin firma): es el **fallback server-side** cuando el refresh token vence.

---

## Qué debe hacer el cliente (Wails)

1. **Guardar** los 3 valores que devuelven `register`/`login`:
   `accessToken` (body) + `X-Refresh-Token` + `X-Refresh-Code` (headers).
2. **Al hacer una petición protegida**: `Authorization: Bearer <accessToken>`.
3. **Al recibir `401`** (o antes de que venzan los 5 min): llamar
   `GET /auth/refresh-token` con los 2 headers guardados → **reemplazar los 3
   valores** con los nuevos (body + headers de la respuesta) → reintentar la
   petición original.
4. **Nunca reutilizar** un refresh token/code ya usado: quedan rotados en BD y
   el siguiente intento con los viejos dará `401`.

> Los headers de refresh son legibles desde JS porque el server los expone vía CORS
> (`ExposedHeaders: X-Refresh-Token, X-Refresh-Code`). Si las peticiones salen desde
> **Go** (bindings de Wails), no aplica CORS y se leen directo de la respuesta.

---

## Flujo completo

### 1. Primer acceso — `POST /auth/register`

```http
POST /auth/register
Content-Type: application/json

{"userName": "jean", "password": "secret123"}
```

```
controller → DTOValidator (userName 3-50, password min 8)
  → RegisterAccountUC:
      1. valida VOs (UsernameVO / PasswordVO)
      2. CheckUsernameExists      → si existe: 409
      3. argon2id hash            → Encode()
      4. CreateSession            → INSERT auth.sessions
      5. CreateUser (role 'ADM')   → INSERT auth.users
      6. genera access (5m) + refresh (72h)
      7. GenerateRefreshCode
      8. UpdateSessionTokens      → UPDATE auth.sessions (rota ambos)
```

**`201 Created`**

```json
{
  "userData":    { "id": "...", "createdAt": "...", "userName": "jean", "role": "ADM" },
  "sessionData": { "accessToken": "eyJ...", "createdAt": "...", "isValid": true }
}
```
```
X-Refresh-Token: eyJ...
X-Refresh-Code:  eyJ...
```

### 2. Login — `POST /auth/login`

Mismo body/respuesta que register. Use case: `FindUserByUserName` → `Compare`
(argon2id, tiempo constante) → genera y rota los 3 valores igual que register.

Errores: `404` usuario inexistente · `400` password incorrecta
(`{"field":"password","message":"password invalid"}`) · `422` campos inválidos.

### 3. Recarga — `GET /auth/refresh-token`

**Envío:**

```http
GET /auth/refresh-token
X-Refresh-Token: <guardado>
X-Refresh-Code:  <guardado>
```

**Qué pasa por dentro** (middleware `RefreshToken` → controller → `CheckRefreshCodeUC`):

```
¿Los 2 headers presentes?
 ├─ NO  → 401 "Missing or invalid refresh token"
 └─ SÍ
     VerifyToken(refresh token)
      ├─ VÁLIDO (no venció) ──► confía en la firma JWT,
      │                         busca usuario por ID (nivel 1)
      └─ VENCIDO ─────────────► busca usuario por el X-Refresh-Code
                                en la BD (nivel 2, fuente de verdad)

     En ambos casos:
       genera access (5m) + refresh (72h) + code NUEVOS
       → UpdateSessionTokens (rota refresh_token Y refresh_code en BD)
       → 200 con la misma forma de login (access nuevo en body,
         refresh token y code nuevos en headers)
```

Respuesta: **`200 OK`**, idéntica a login, con headers nuevos.

Errores: `401 {"title":"Refresh code","message":"session not valid"}`
(refresh token vencido + code incorrecto o desconocido) · `401` faltan headers.

### 4. Rutas protegidas (hoy)

El middleware **`AccessToken`** está implementado y disponible
(`Authorization: Bearer ...` → verifica JWT → deja los claims en el context),
pero **hoy ninguna ruta lo usa montado**: `login` y `register` son públicas y
`refresh-token` usa `RefreshToken`. Las rutas futuras del módulo u otros módulos
se montan con `r.Use(am.AccessToken)`.

---

## Reglas de rotación (las importantes)

| Evento | access | refresh | code | persiste en BD |
|---|---|---|---|---|
| `register` | nuevo | nuevo | nuevo | sí |
| `login` | nuevo | nuevo | nuevo | sí |
| `refresh-token` (vencido o no) | nuevo | nuevo | nuevo | sí |

- **Siempre** se rotan los **3** juntos y **siempre** se persisten los 2 últimos.
- El cliente **debe** reemplazar sus valores guardados con cada respuesta;
  si guarda los viejos, el siguiente refresh da `401`.
- Un token viejo usado "re-engancha" la sesión solo mientras su firma no venza
  (nivel 1); una vez vencido, el único camino es el refresh code vigente en BD.

---

## Seguridad actual

- **Passwords**: argon2id (`m=64MiB, t=3, p=2`, salt 16B, key 32B), formato PHC,
  comparación en tiempo constante. Nunca se guarda ni se devuelve la password.
- **JWT**: HS256 con `JWT_SECRET` (obligatorio en `ENVIROMENT=production`,
  la app no arranca con el default).
- **Rotación obligatoria** de refresh token + code en cada uso (una fuga del
  refresh token viejo no sirve tras el primer refresh).
- **Single-user**: no hay roles/permisos en juego más allá del `role: "ADM"` fijo
  y no existen tenants.

---

## Errores — catálogo

| Status | Title | Cuándo |
|---|---|---|
| `400` | Validation Error | password incorrecta (login), campos de dominio inválidos |
| `401` | Unauthorized | faltan headers de refresh, token ausente/vencido en middleware |
| `401` | Refresh code | refresh token vencido + refresh code inválido → `session not valid` |
| `404` | User Not Found | usuario inexistente (login) |
| `409` | Conflict | `userName` ya existe (register) |
| `422` | Validation Failed | DTO inválido (validator de `go-playground`) |
| `500` | Database Error | falla de BD (no expone detalles al cliente) |

Formato de respuesta de error:

```json
{ "title": "...", "message": "...", "status": 400 }
```
(los de dominio agregan `"field": "password"`).

---

## Secuencia tipo (cliente Wails)

```
┌─ Cliente                          ─ Backend ─┐
│ POST /auth/login {user, pass}                │
│ ◄── 200 + accessToken(body) + RT/RC(headers) │   guardo los 3
│                                              │
│ GET /ruta-protegida                          │
│ Authorization: Bearer accessToken            │
│ ◄── 401 (access venció a los 5 min)          │
│                                              │
│ GET /auth/refresh-token                      │
│ X-Refresh-Token + X-Refresh-Code             │
│ ◄── 200 + 3 valores NUEVOS                   │   reemplazo los 3
│                                              │
│ GET /ruta-protegida  (retry)                 │
│ Authorization: Bearer accessTokenNuevo        │
│ ◄── 200                                     │
└──────────────────────────────────────────────┘
```
