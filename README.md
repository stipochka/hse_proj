# edu-platform

Серверная часть платформы учёта студенческой активности. Студенты загружают PDF-материалы об активностях за год, кураторы групп оценивают их во внутренней валюте и зачётных единицах, администратор видит сводную статистику.

Стек: Go, PostgreSQL, Keycloak (аутентификация), MinIO (хранилище файлов).

## Запуск

```bash
docker compose up --build
```

Стартуют четыре сервиса:

| Сервис | Адрес |
|---|---|
| Приложение | `http://localhost:8080` |
| Keycloak | `http://localhost:8180` |
| MinIO API | `http://localhost:9000` |
| MinIO консоль | `http://localhost:9001` |

### Миграции

Применять по порядку:

```bash
psql "$DATABASE_URL" -f migrations/001_init.sql
psql "$DATABASE_URL" -f migrations/002_activities.sql
psql "$DATABASE_URL" -f migrations/003_evals_transactions.sql
psql "$DATABASE_URL" -f migrations/004_add_roles.sql
psql "$DATABASE_URL" -f migrations/005_groups_courses.sql
psql "$DATABASE_URL" -f migrations/006_activities_extend.sql
psql "$DATABASE_URL" -f migrations/007_activity_files_s3.sql
psql "$DATABASE_URL" -f migrations/008_keycloak.sql
```

## Настройка Keycloak

1. Открыть `http://localhost:8180`, войти `admin / admin`.
2. Создать realm с именем `edu`.
3. Создать роли realm: `student`, `group_admin`, `super_admin`.
4. Для пользователей с ролью `group_admin` добавить атрибут `group_id` — числовой ID группы из нашей БД.
5. В настройках клиента добавить **User Attribute Mapper**: атрибут `group_id` → JWT claim `group_id`.
6. Создать клиент для фронтенда (тип `public`, Standard flow).

После этого фронтенд получает токены напрямую от Keycloak и передаёт их бэкенду через `Authorization: Bearer <token>`. Бэкенд только верифицирует подпись через JWKS и читает claims — своей базы сессий нет.

При первом запросе от нового пользователя бэкенд автоматически создаёт запись в таблице `users` (JIT-провизионинг).

## Переменные окружения

| Переменная | Описание | По умолчанию |
|---|---|---|
| `DATABASE_URL` | Строка подключения к PostgreSQL | — |
| `PORT` | Порт HTTP-сервера | `8080` |
| `KEYCLOAK_JWKS_URL` | JWKS-эндпоинт реалма Keycloak | `http://keycloak:8080/realms/edu/...` |
| `S3_ENDPOINT` | Адрес MinIO | `minio:9000` |
| `S3_ACCESS_KEY` | Access key | `minioadmin` |
| `S3_SECRET_KEY` | Secret key | `minioadmin` |
| `S3_BUCKET` | Бакет для файлов | `edu-files` |
| `S3_USE_SSL` | TLS для S3 | `false` |

## Роли

Роли задаются в Keycloak (`realm_access.roles` в JWT).

| Роль | Что может |
|---|---|
| `student` | создаёт активности, загружает PDF, смотрит свою статистику |
| `group_admin` | видит ленту активностей своей группы, выставляет оценки, смотрит отчёты по своей группе |
| `super_admin` | всё то же, что group_admin, но для любой группы; управляет группами и курсами |

`group_admin` привязан к группе через claim `group_id` в токене. Бэкенд не даст ему видеть чужую группу, даже если передать другой `group_id` в параметрах запроса.

## API

Все ручки требуют заголовок `Authorization: Bearer <token>`.

### Профиль студента

| Метод | Путь | Описание |
|---|---|---|
| GET | `/me` | Профиль: user_id, email, role, group_id, balance |
| GET | `/me/balance` | Текущий баланс внутренней валюты |
| GET | `/me/transactions` | История начислений |
| GET | `/me/evaluations` | Оценки по своим активностям |

### Активности

| Метод | Путь | Описание |
|---|---|---|
| POST | `/activities` | Создать активность |
| GET | `/activities` | Список своих активностей |
| GET | `/activities/{id}` | Детали активности |
| DELETE | `/activities/{id}` | Удалить свою активность |

Тело `POST /activities`:
```json
{
  "title": "Олимпиада по математике",
  "description": "Призовое место, региональный этап",
  "category": "olympiad",
  "activity_date": "2025-04-10",
  "draft": false
}
```

Поле `status` жизненный цикл: `draft` → `submitted` → `approved` / `rejected`.

### Файлы

| Метод | Путь | Описание |
|---|---|---|
| POST | `/files` | Загрузить PDF |
| GET | `/files/{id}` | Скачать файл |

`POST /files` — multipart/form-data, поля `file` (PDF, макс. 20 МБ) и `activity_id`. Файлы хранятся в MinIO по ключу `activities/{activity_id}/{timestamp}_{random}.pdf`.

### Оценивание — group_admin, super_admin

`POST /evaluate`:
```json
{
  "activity_id": 7,
  "student_id": 42,
  "score": 8,
  "comment": "Призовое место на региональном уровне"
}
```

`score` от 0 до 10. Автоматически: `currency = score × 10`, `credits = score / 2.5`. Активность переходит в статус `approved`.

### Лента активностей — group_admin, super_admin

`GET /admin/activities`

`group_admin` видит только свою группу. `super_admin` может добавить `?group_id=`.

| Параметр | Описание |
|---|---|
| `status` | фильтр по статусу (`submitted`, `approved`, ...) |
| `student_id` | конкретный студент |
| `category` | вид активности |
| `limit` | записей на страницу (макс. 100, по умолчанию 50) |
| `offset` | смещение |

### Отчёты — group_admin, super_admin

`GET /admin/reports`

`group_admin` — только своя группа. `super_admin` — фильтры `group_id`, `user_id`, `course_id`, `stream`. Добавить `?format=csv` для выгрузки файла.

### Группы и курсы — super_admin

| Метод | Путь | Тело |
|---|---|---|
| GET | `/admin/groups` | — |
| POST | `/admin/groups` | `{ name, stream, course_year }` |
| POST | `/admin/groups/assign` | `{ user_id, group_id }` |
| POST | `/admin/groups/remove` | `{ user_id, group_id }` |
| GET | `/admin/courses` | — |
| POST | `/admin/courses` | `{ name }` |
| POST | `/admin/courses/assign` | `{ user_id, course_id }` |

## Структура проекта

```
cmd/server/           точка входа
internal/
  handlers/
    handler.go        Handler, конструктор, константы
    middleware.go     Auth / AuthGroupAdmin / AuthSuperAdmin, JIT-провизионинг
    me.go             Me, MyBalance, MyTransactions, MyEvaluations
    activities.go     CRUD активностей
    files.go          загрузка и скачивание PDF
    evaluate.go       Evaluate
    admin.go          лента, отчёты, группы, курсы
  server/             маршруты и инициализация зависимостей
  jwks/               верификация токенов Keycloak через JWKS
  s3/                 клиент MinIO
  store/              запросы к БД
  domain/             модели данных
  db/                 пул соединений
  policy/             расчёт currency и credits из score
migrations/           SQL-миграции 001–008
```
