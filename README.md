# edu-platform

Серверная часть платформы для учёта студенческой активности. Студенты загружают материалы о своих активностях в течение года, преподаватели и администраторы оценивают их во внутренней валюте и академических зачётных единицах. Каждый студент видит собственную статистику; администратор получает сводные отчёты с фильтрацией и выгрузкой в CSV.

Написано на Go, хранилище — PostgreSQL, маршрутизация — chi.

## Запуск

```bash
docker compose up --build
```

Поднимает три сервиса: PostgreSQL, MinIO (S3-совместимое хранилище файлов) и само приложение. Веб-консоль MinIO доступна на `http://localhost:9001`.

Миграции применяются вручную в указанном порядке:

```bash
psql "$DATABASE_URL" -f migrations/001_init.sql
psql "$DATABASE_URL" -f migrations/002_activities.sql
psql "$DATABASE_URL" -f migrations/003_evals_transactions.sql
psql "$DATABASE_URL" -f migrations/004_add_roles.sql
psql "$DATABASE_URL" -f migrations/005_groups_courses.sql
psql "$DATABASE_URL" -f migrations/006_activities_extend.sql
```

## Переменные окружения

| Переменная | Описание | По умолчанию |
|---|---|---|
| `DATABASE_URL` | Строка подключения к PostgreSQL | — |
| `JWT_SECRET` | Секрет для подписи JWT | `dev-secret` |
| `ACCESS_TOKEN_EXP` | Время жизни access-токена | `15m` |
| `REFRESH_TOKEN_EXP` | Время жизни refresh-токена | `168h` |
| `PORT` | Порт HTTP-сервера | `8080` |
| `S3_ENDPOINT` | Адрес S3-совместимого хранилища | `minio:9000` |
| `S3_ACCESS_KEY` | Access key | `minioadmin` |
| `S3_SECRET_KEY` | Secret key | `minioadmin` |
| `S3_BUCKET` | Имя бакета | `edu-files` |
| `S3_USE_SSL` | Использовать TLS | `false` |

## Роли

- `student` — создаёт активности, загружает файлы, видит свою статистику. Роль назначается автоматически при регистрации.
- `teacher` — оценивает активности студентов.
- `admin` — все права teacher плюс управление группами, курсами и доступ к отчётам.

Роль кодируется в JWT-токене (claim `role`) и проверяется middleware на каждом маршруте.

## API

### Аутентификация

```
POST /signup              Регистрация (всегда role=student)
POST /login               Вход, возвращает access_token и refresh_token
POST /refresh             Обновление access_token по refresh_token
POST /logout              Инвалидация refresh_token
```

### Профиль студента

```
GET  /me                  Профиль текущего пользователя
GET  /me/balance          Текущий баланс внутренней валюты
GET  /me/transactions     История начислений
GET  /me/evaluations      Все оценки по своим активностям
```

### Активности

```
POST   /activities        Создать активность (поля: title, description, category, activity_date, draft)
GET    /activities        Список своих активностей
GET    /activities/{id}   Детали активности (teacher/admin видят любую)
DELETE /activities/{id}   Удалить свою активность
```

Поле `status` принимает значения: `draft`, `submitted`, `under_review`, `approved`, `rejected`. При оценке активность автоматически переходит в `approved`.

### Файлы

```
POST /files               Загрузить файл (form-data: file + activity_id). Лимит 20 МБ.
GET  /files/{id}          Скачать файл
```

Файлы хранятся в MinIO по ключу `activities/{activity_id}/{timestamp}_{random}{ext}`. Разрешённые расширения: `.pdf`, `.doc`, `.docx`, `.png`, `.jpg`, `.jpeg`, `.zip`, `.txt`. Лимит — 20 МБ.

### Оценивание (teacher, admin)

```
POST /evaluate
```

Тело запроса:

```json
{
  "activity_id": 1,
  "student_id":  2,
  "score":       8,
  "comment":     "Хорошая работа"
}
```

`score` — целое от 0 до 10. Если `currency` не передан, рассчитывается автоматически: `currency = score * 10`, `credits = score / 2.5`.

### Отчёты (admin)

```
GET /admin/reports
```

Параметры фильтрации (все необязательны):

| Параметр | Описание |
|---|---|
| `user_id` | Конкретный студент |
| `group_id` | Группа |
| `course_id` | Курс |
| `stream` | Поток |
| `format=csv` | Выгрузка в CSV вместо JSON |

### Группы (admin)

```
GET  /admin/groups                Список групп
POST /admin/groups                Создать группу (name, stream, course_year)
POST /admin/groups/assign         Добавить студента в группу (user_id, group_id)
POST /admin/groups/remove         Убрать студента из группы (user_id, group_id)
```

### Курсы (admin)

```
GET  /admin/courses               Список курсов
POST /admin/courses               Создать курс (name)
POST /admin/courses/assign        Записать студента на курс (user_id, course_id)
```

## Структура проекта

```
cmd/server/          — точка входа
internal/
  handlers/
    handler.go       — Handler, конструктор, вспомогательные функции
    middleware.go    — Auth, AuthTeacher, AuthAdmin, parseJWT
    auth.go          — SignUp, Login, Refresh, Logout
    me.go            — Me, MyBalance, MyTransactions, MyEvaluations
    activities.go    — CRUD активностей
    files.go         — загрузка и скачивание файлов
    evaluate.go      — Evaluate
    admin.go         — отчёты, группы, курсы
  server/            — регистрация маршрутов
  store/             — запросы к БД
  domain/            — модели данных
  db/                — пул соединений
  policy/            — расчёт вознаграждения
migrations/          — SQL-миграции (применять по порядку 001..006)
```
