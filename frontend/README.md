# HSE Admin Frontend

Административная панель для управления активностями студентов на платформе HSE.

## Технологический стек

- **React** 18.2 - фреймворк UI
- **TypeScript** - типизация
- **Vite** - сборщик проекта
- **React Router** - маршрутизация
- **TailwindCSS** - стили
- **Ant Design** - компоненты UI
- **Axios** - HTTP клиент
- **Keycloak** - аутентификация
- **Zustand** - управление состоянием

## Структура проекта

```
frontend/
├── src/
│   ├── app/
│   │   ├── layout/
│   │   │   └── MainLayout.tsx
│   │   └── store/
│   │       └── authStore.ts
│   ├── pages/
│   │   ├── dashboard/
│   │   ├── activities/
│   │   ├── activity-evaluation/
│   │   ├── group-students/
│   │   └── export/
│   ├── shared/
│   │   ├── api/
│   │   │   ├── activities.ts
│   │   │   ├── dashboard.ts
│   │   │   ├── export.ts
│   │   │   ├── references.ts
│   │   │   └── students.ts
│   │   ├── components/
│   │   │   └── FilterBar.tsx
│   │   ├── lib/
│   │   │   ├── api.ts
│   │   │   └── keycloak.ts
│   │   └── types/
│   │       └── index.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.ts
├── postcss.config.js
├── package.json
├── .env
└── README.md
```

## Функциональность

### 1. Аутентификация
- Интеграция с Keycloak
- Автоматическое перенаправление на страницу логина
- Управление токеном доступа
- Выход из системы

### 2. Дашборд (сводный)
- Агрегированная статистика
- Фильтры: студент, группа, поток, курс, категория, период
- Топ-студенты
- Распределение баллов

### 3. Активности на проверку
- Список заявок со статусом SUBMITTED
- Фильтрация и сортировка
- Просмотр PDF
- Переход к оценке активности

### 4. Проверка активности
- Просмотр информации об активности
- Просмотр PDF документа
- Форма оценки: баллы, опциональные з.е., комментарий
- Кнопки оценить/отклонить

### 5. Студенты группы
- Список студентов выбранной группы
- Суммарные баллы
- Количество активностей
- Фильтрация и сортировка

### 6. Экспорт
- Выгрузка данных в CSV
- Выгрузка данных в XLSX
- Применение фильтров к экспорту

## Установка и запуск

### Предварительные требования
- Node.js >= 16
- npm или yarn

### Установка зависимостей
```bash
npm install
```

### Конфигурация окружения
Создайте файл `.env` в корне проекта на основе `.env.example`:

```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_KEYCLOAK_URL=http://localhost:8080/auth
VITE_KEYCLOAK_REALM=master
VITE_KEYCLOAK_CLIENT_ID=admin-frontend
```

### Запуск в режиме разработки
```bash
npm run dev
```

Приложение будет доступно по адресу `http://localhost:5173`

### Сборка для production
```bash
npm run build
```

Собранное приложение будет находиться в папке `dist`

### Preview production сборки
```bash
npm run preview
```

## API Endpoints

Приложение использует следующие API endpoints:

- `GET /api/activities` - получить список активностей
- `GET /api/activities/:id` - получить активность по ID
- `POST /api/activities/:id/evaluate` - оценить активность
- `POST /api/activities/:id/reject` - отклонить активность
- `GET /api/activities/:id/file` - получить файл активности
- `GET /api/dashboard/aggregates` - получить агрегированные данные
- `GET /api/dashboard/top-students` - получить топ-студентов
- `GET /api/dashboard/score-distribution` - получить распределение баллов
- `GET /api/students` - получить список студентов
- `GET /api/groups/:groupId/students` - получить студентов группы
- `GET /api/groups` - получить список групп
- `GET /api/courses` - получить список курсов
- `GET /api/categories` - получить список категорий
- `GET /api/streams` - получить список потоков
- `GET /api/export/csv` - экспортировать в CSV
- `GET /api/export/xlsx` - экспортировать в XLSX

## FSD (Feature-Sliced Design)

Проект следует архитектуре FSD с разделением на слои:

- **app** - инициализация приложения, layout, глобальное состояние
- **pages** - страницы приложения
- **shared** - общие компоненты, утилиты, типы, API

## Стили

Проект использует TailwindCSS для утилитарных стилей и Ant Design для компонентов UI.

## Развертывание

### Docker

```dockerfile
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

