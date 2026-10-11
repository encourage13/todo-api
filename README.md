# Todo API

REST API для ведения личного списка задач: регистрация, вход по JWT и CRUD задач, где каждый пользователь видит только свои задачи.
Написан на Go со стандартным `net/http` (роутинг через `http.ServeMux` с методами и параметрами пути), хранит данные в PostgreSQL через `pgx`, аутентифицирует по JWT (`golang-jwt`, HS256) и паролям в bcrypt.
Запускается в Docker Compose, документация API генерируется в Swagger (`swaggo`).

## Быстрый старт через Docker

Нужны Docker и Docker Compose.

```bash
git clone git@github.com:encourage13/todo-api.git
cd todo-api
cp .env.example .env
```

Откройте `.env` и заполните `JWT_SECRET` (минимум 32 байта). Сгенерировать значение:

```bash
openssl rand -base64 32
```

Без `JWT_SECRET` Compose не запустится: переменная помечена как обязательная (`${JWT_SECRET:?JWT_SECRET is not set}`). Для чего-то кроме локального запуска поменяйте и `DB_PASSWORD`.

Запуск:

```bash
docker compose up -d --build
```

Compose поднимает три сервиса по порядку:

1. `postgres`: PostgreSQL 17, дожидаемся готовности по healthcheck (`pg_isready`);
2. `migrate`: применяет миграции из `migrations/` и завершается;
3. `app`: API на `http://localhost:8080`, стартует только после успешных миграций.

Порты `8080` и `5432` открыты только на `127.0.0.1`.

### Пример запросов

Регистрация:

```bash
curl -i -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123"}'
```

Вход (в ответе `{"token":"..."}`):

```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123"}'
```

Сохраните токен в переменную. Если установлен `jq`:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123"}' | jq -r .token)
```

Без `jq` скопируйте значение `token` вручную: `TOKEN=<токен>`.

Создание задачи:

```bash
curl -i -X POST http://localhost:8080/todos \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy milk","description":"2 liters"}'
```

Список задач:

```bash
curl -s http://localhost:8080/todos -H "Authorization: Bearer $TOKEN"
```

### Остановка

```bash
docker compose down
```

Данные PostgreSQL лежат в именованном томе `todo_pgdata` и переживают `down`.

```bash
docker compose down -v
```

Флаг `-v` удаляет и том, то есть **все пользователи и задачи будут потеряны**. При следующем запуске миграции создадут пустую базу.


### Эндпоинты

| Метод    | Путь             | Токен | Успех                   | Ошибки             |
|----------|------------------|-------|-------------------------|--------------------|
| `POST`   | `/auth/register` | нет   | `201` `{id, email}`     | `400`, `409`, `500` |
| `POST`   | `/auth/login`    | нет   | `200` `{token}`         | `400`, `401`, `500` |
| `GET`    | `/me`            | да    | `200` `{id}`            | `401`, `500`        |
| `POST`   | `/todos`         | да    | `201` задача            | `400`, `401`, `500` |
| `GET`    | `/todos`         | да    | `200` массив задач      | `401`, `500`        |
| `GET`    | `/todos/{id}`    | да    | `200` задача            | `400`, `401`, `404`, `500` |
| `PATCH`  | `/todos/{id}`    | да    | `200` задача            | `400`, `401`, `404`, `500` |
| `DELETE` | `/todos/{id}`    | да    | `204` без тела          | `400`, `401`, `404`, `500` |
| `GET`    | `/swagger/`      | нет   | `200` Swagger UI        | |

Что означают коды:

- `400`: невалидный JSON, неизвестные поля в теле, тело больше 1 МБ, `id` не UUID или не прошла валидация;
- `401`: нет заголовка `Authorization: Bearer ...`, токен невалиден или истёк; на `/auth/login` так же отвечаем при неверном email или пароле;
- `404`: задача не найдена или принадлежит другому пользователю, а также несуществующий путь;
- `405`: путь есть, но метод не поддерживается;
- `409`: email уже зарегистрирован.

Правила валидации:

- email приводится к нижнему регистру и должен быть корректным адресом;
- пароль от 8 до 72 байт (ограничение bcrypt);
- `title` от 1 до 200 символов после обрезки пробелов;
- `description` до 2000 символов.

Формат задачи:

```json
{
  "id": "6f1c2b9e-...",
  "title": "Buy milk",
  "description": "2 liters",
  "completed": false,
  "created_at": "2026-10-12T10:00:00Z",
  "updated_at": "2026-10-12T10:00:00Z"
}
```

`GET /todos` возвращает задачи текущего пользователя, новые первыми. `PATCH /todos/{id}` частично обновляет задачу: меняются только переданные поля (`title`, `description`, `completed`), а пустое тело `{}` даёт `400`.

### Формат ошибок

Все ошибки приходят в JSON с одним полем, включая `404` и `405` от роутера:

```json
{"error":"task not found"}
```

Ошибки валидации содержат причину, например `{"error":"invalid input: title must be 1-200 characters"}`. Внутренние ошибки не раскрываются: клиент получает `{"error":"internal error"}`, а подробности пишутся в лог.

### X-Request-ID

Каждый ответ содержит заголовок `X-Request-ID` с UUID, который сервер генерирует для запроса. Тот же id попадает в лог запроса и в лог внутренних ошибок, так что по нему можно найти конкретный запрос. Значение `X-Request-ID`, присланное клиентом, не используется: сервер всегда создаёт новое.

## Архитектура

```
HTTP-запрос
   │
   ▼
RequestID → RequestLogger → JSONErrors → ServeMux → [RequireAuth] → handler
                                                                       │
transport (internal/transport/httpapi)   разбор запроса, DTO, коды ответа
   │
   ▼
service   (internal/service)             бизнес-логика, валидация, bcrypt
   │
   ▼
repository (internal/repository)         SQL-запросы
   │
   ▼
model     (internal/model)               модели
```

Конкретные реализации связываются вручную в `cmd/api/main.go`. Слои зависят только от узких интерфейсов, поэтому их можно подменять в тестах.

```
cmd/
└── api/
    └── main.go              точка входа: конфиг, пул БД, роутинг, graceful shutdown
internal/
├── auth/
│   └── token.go             выпуск и проверка JWT
├── config/
│   └── config.go            загрузка YAML, подстановка env, валидация
├── model/
│   ├── errors.go            доменные ошибки
│   ├── todo.go
│   └── user.go
├── repository/
│   ├── postgres.go          создание pgxpool
│   ├── todo.go
│   └── user.go
├── service/
│   ├── todo.go
│   └── user.go
└── transport/
    └── httpapi/
        ├── auth.go          /auth/register, /auth/login
        ├── todo.go          /todos
        ├── middleware.go    RequireAuth, RequestID, /me
        ├── logging.go       логирование запросов
        ├── jsonerrors.go    JSON для 404/405 от ServeMux
        ├── errors.go        маппинг доменных ошибок в HTTP-коды
        ├── request.go       декодирование JSON
        └── response.go      запись JSON-ответов
```

Рядом лежат `migrations/` (SQL-миграции golang-migrate), `docs/` (сгенерированный Swagger) и `config/config.yaml`.

## Что дальше

- Тесты: unit-тесты сервисов на моках репозиториев и интеграционные тесты репозиториев на настоящем PostgreSQL. Сейчас тестов нет.
- Refresh-токены, чтобы токены можно было отзывать и делать срок access-токена коротким, а также rate limit на `/auth/*`. И то и другое можно хранить в Redis.

