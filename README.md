# Subscription Aggregator

REST-сервис для агрегации данных об онлайн подписках пользователей.

## Запуск

```bash
docker compose up --build
```

## Остановка и удаление контейнеров:

```bash
docker compose down -v
```

Сервис поднимается на `http://localhost:8080`

## Swagger UI

```
http://localhost:8080/swagger/index.html
```

## API

| Метод  | Путь                       | Описание                      |
|--------|----------------------------|-------------------------------|
| POST   | /api/v1/subscriptions      | Создать подписку              |
| GET    | /api/v1/subscriptions      | Список подписок               |
| GET    | /api/v1/subscriptions/{id} | Получить подписку             |
| PUT    | /api/v1/subscriptions/{id} | Обновить подписку             |
| DELETE | /api/v1/subscriptions/{id} | Удалить подписку              |
| GET    | /api/v1/subscriptions/cost | Суммарная стоимость за период |

## Пример запроса

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```
