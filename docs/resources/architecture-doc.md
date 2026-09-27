# Arquitectura

El proyecto sigue una **arquitectura hexagonal (puertos y adaptadores)** aplicada sobre **DDD**, organizada en **módulos autocontenidos**: cada módulo tiene sus propias capas `domain`, `application` e `infrastructure`, y nadie afuera toca sus internos.

![Arquitectura hexagonal + DDD + modular](../mediaresources/artquitectura_hexagonal_ddd_modular.png)

---

## Estructura general

```
cmd/server/main.go            → instancia módulos + globales, monta rutas
config/                       → conexión a BD, migraciones, modelos GORM, envs
modules/
├── auth/                     → un módulo (ejemplo de todos)
│   ├── auth_module.go        → BOOTSTRAP del módulo (inyección de dependencias)
│   ├── domain/               → lo más interno: no importa de nadie
│   │   ├── entities/
│   │   ├── valueobjects/
│   │   ├── enums/
│   │   └── services/         → (opcional) lógica de dominio
│   ├── applicastion/         → la intención / qué se quiere hacer
│   │   ├── dtos/
│   │   │   ├── requests/
│   │   │   └── responses/
│   │   ├── usecases/
│   │   ├── pipelines/        → (opcional, casos raros)
│   │   └── ports/            → contratos que cumplen los adapters
│   └── infrastructure/       → el cómo: implementaciones concretas
│       ├── adapters/
│       ├── controllers/
│       ├── middlewares/
│       ├── raws/
│       ├── mappers/
│       └── sse/              → (opcional) server-sent events
└── shared/                   → cosas compartidas por >1 módulo
```

**Regla de dependencia (la flecha de la hexágono):** las dependencias siempre apuntan hacia adentro:

```
infrastructure ──► application ──► domain
```

- `domain` no importa de ningún otro layer (es puro Go).
- `application` conoce a `domain` y define **puertos** (interfaces), nunca concretas.
- `infrastructure` implementa los puertos y adapta el mundo exterior (BD, librerías, HTTP).

---

## Domain (el núcleo)

Representa las reglas del negocio. No sabe que existen HTTP, BD ni librerías.

| Pieza | Qué es | Ejemplo |
|---|---|---|
| **Value Objects** | objetos **inmutables y válidos**: no existen en un estado inválido; se crean con una fábrica que valida | `NewUsernameVO("je")` → error; `NewUsernameVO("jean")` → VO con `Value()` |
| **Entities** | objetos con identidad con los que trabaja la app; pueden cambiar con el tiempo | `User`, `Session`, `UserAuth` |
| **Enums** | opciones **estáticas que no cambian** y tienen significado propio | `authenums.AccountValid = "VALID"`, `RoleAdmin = "ADM"` |
| **Domain Services** *(opcional)* | lógica que no pertenece a una entidad ni a un VO, encapsulada | cálculos, reglas compuestas |

```go
// valueobject: inmutable y validado en la construcción
usernameVO, err := authvalueobjects.NewUsernameVO(dto.Username)
if err != nil { return nil, err }          // ya sabés que es válido
user, err := repo.FindUserByUserName(ctx, usernameVO.Value())
```

---

## Application (la intención)

Define **qué** se quiere hacer, sin saber **cómo**.

### DTOs (`dtos/requests` y `dtos/responses`)

Transportan datos. Solo eso:

- **requests**: lo que entra (body/query de HTTP) → validados con el `DTOValidator`.
- **responses**: lo que sale al cliente.

No tienen lógica y no se confunden con entidades: si el cliente recibe menos campos que la entidad, el DTO recorta.

### Ports (`ports/`)

**Contrato (interfaz) que los adapters deben cumplir.** Es lo que permite invertir la dependencia: el use-case pide un `AuthRepositoryPort` y no le importa si detrás hay GORM, pgx o un mock.

```go
type AuthRepositoryPort interface {
    FindUserByUserName(ctx context.Context, username string) (*authentities.UserAuth, error)
    // ...
}
```

Los puertos también existen para adaptadores de librerías: `EncryptorPort`, `JwtPort`.

### Use Cases (`usecases/`)

Definen **qué se quiere hacer** y **orquestan** adapters y/u otros use-cases para lograrlo. Son el único lugar donde se decide el orden de las cosas:

```go
func (l *LoginAccountUC) Execute(ctx, dto) (*authentities.Session, error) {
    usernameVO := NewUsernameVO(dto.Username)   // validar dominio
    user := repo.FindUserByUserName(...)        // adapter (BD)
    match := encrypter.Compare(...)             // adapter (librería argon2)
    tokens := jwtService.GenerateToken(...)     // adapter (librería JWT)
    repo.UpdateSessionTokens(...)               // adapter (BD)
    return entities.CreateSession(...)          // armar entidad
}
```

El nombre del método es **`Execute`**.

### Pipelines (`pipelines/`) — casos raros

Parecidas a un use-case, pero **orquestan use-cases**: representan una **tubería de transformación** donde cada caso de uso ejecuta su transformación y devuelve data que pasa al siguiente.

| | Use case | Pipeline |
|---|---|---|
| Orquesta | adapters / servicios | **use cases** |
| Método | `Execute` | **`Run`** |
| Cuándo | siempre | solo cuando la transformación es un **proceso largo** |
| Ejecución | inline en la request | se puede mandar a **segundo plano** |

```go
p := pipelines.NewMiPipeline(uc1, uc2, uc3)
go p.Run(ctx)   // proceso largo en background
```

---

## Infrastructure (el cómo)

### Adapters (`adapters/`)

**Implementan los puertos.** Son la traducción concreta del contrato:

- `AuthRepository` implementa `AuthRepositoryPort` (GORM contra la BD).
- **Adaptadores de librerías** encapsulan una librería para no depender de ella en el resto de la app: `JwtAdapterService` envuelve `golang-jwt`, `EncryptorAdapter` envuelve `argon2`. Si mañana cambiás de librería, solo tocás el adapter; el use-case ni se entera.

### Middlewares (`middlewares/`)

Se ejecutan **antes** de llegar a los controllers (auth, rate limit, logging…). Pueden cortar la request (401) o dejar datos en el `context` para que el controller los use.

### Controllers (`controllers/`)

**Punto de entrada HTTP del módulo.** Reciben la request, validan el DTO con el `DTOValidator`, llaman al use-case correspondiente y traducen la respuesta/error a HTTP (con los mappers compartidos de response). Acá también viven las rutas del módulo (`AuthMapRoutes`).

### Raws (`raws/`)

Las **respuestas crudas que bota la BD** (o un servicio externo): structs planos con todos los campos del JOIN, sin interpretación de negocio.

### Mappers (`mappers/`)

Traducen: el mapper de infra lleva del **raw → entidad de dominio** (o raw → DTO, según el caso), **eligiendo solo la data que le importa a la aplicación** (por ejemplo descarta `created_by`). El repositorio **devuelve siempre entidades de dominio**, nunca raws:

```
BD → raw (authraws.UserFoundRaw) → infra mapper → entidad (authentities.UserAuth) → use-case
```

### SSE (`sse/`) — opcional

Capa para **Server-Sent Events**: managers/conectores que emiten eventos en tiempo real al cliente (notificaciones, progreso…). Vive en infra porque es transporte.

---

## Bootstrap de cada módulo (archivo raíz)

Cada módulo tiene en su raíz un archivo `xxx_module.go` (**`auth_module.go`**) con su `XxxBootstrap(...)`. Ese archivo es el único que sabe de las implementaciones concretas: **hace la inyección de dependencias invirtiéndolas** — construye los adapters, se los pasa a los use-cases y arma el controller/router del módulo.

```go
func AuthBootstrap(db *gorm.DB, v *shareddtovalidators.DTOValidator) chi.Router {
    authRepository := authadapters.NewAuthRepository(db)      // adapter
    jwtService     := authadapters.NewJwtAdapterService()     // adapter de librería
    encrypter      := authadapters.NewEncryptorAdapter()      // adapter de librería

    loginUC := authusecases.NewLoginAccountUC(authRepository, encrypter, jwtService)

    controller := authcontrollers.NewAuthController(v, middleware, loginUC, ...)

    r := chi.NewRouter()
    r.Mount("/", authcontrollers.AuthMapRoutes(controller))
    return r
}
```

Fuera de este archivo, **nadie** importa `infrastructure/adapters` — si algo necesita una implementación concreta, es el bootstrap quien la inyecta.

## `main` (composition root)

`cmd/server/main.go` instancia **todos los módulos** llamando a sus bootstraps, más las **instancias globales compartidas** (validator, middlewares globales, etc.) y las rutas raíz (`/health`):

```go
dtoValidator := shareddtovalidators.NewDTOValidator()   // global compartida

r.Get("/health", ...)                                   // ruta raíz

r.Mount("/auth", auth.AuthBootstrap(db, dtoValidator))  // módulo 1
// r.Mount("/futuro", futuro.FuturoBootstrap(db, dtoValidator))  // módulo 2
```

No hay un "bootstrap global": el `main` llama al bootstrap de cada módulo.

## `shared`

`modules/shared/` pone todo lo **compartido por más de un módulo**: value objects genéricos (`UUIDVO`), errores de dominio/infra, mappers de respuesta HTTP, el DTO validator, códigos de error. Un módulo **nunca** importa de otro módulo directamente — si dos módulos necesitan lo mismo, eso sube a `shared`.

---

## Flujo de una petición

```
main
 └─► bootstrap del módulo (DI: invierte dependencias)
      └─► controller (entrada HTTP)
           └─► DTO validator (valida el request)
                └─► use case (orquesta adapters/servicios)
                     └─► adapter o servicio
                          └─► consulta a la BD / ejecuta lógica
                               (raw → mapper de infra → entidad de dominio)
```

Y de vuelta: la entidad se convierte en **response DTO** en el controller y se responde por HTTP.

---

## Convenciones rápidas

- **Rutas** se declaran en el controller del módulo (`AuthMapRoutes`), montadas por el bootstrap.
- **Nombres**: `XxxAdapter` = implementación de puerto; `XxxUC` = use case; `XxxVO` = value object; `XxxRaw` = crudo de BD; `XxxBootstrap` = DI del módulo.
- **Errores**: de dominio (`DomainError`) o de infra (`AppError`); el response mapper compartido los traduce a JSON + status.
- **Métodos**: `Execute` en use cases, `Run` en pipelines, `New*` como fábrica validada de VOs/objetos.
