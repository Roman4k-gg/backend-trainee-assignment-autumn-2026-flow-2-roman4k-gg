## CJM

### CJM Пользователя
```mermaid
sequenceDiagram
    actor User
    participant API as Kitchen API
    participant DB as PostgreSQL

    User->>API: GET /api/v1/restaurants
    API->>DB: SELECT restaurants
    API-->>User: List of restaurants
    User->>API: GET /api/v1/restaurants/{id}/menu
    API->>DB: SELECT menu_items
    API-->>User: Menu for restaurant
    User->>API: POST /api/v1/orders (user_id, restaurant_id, items)
    API->>DB: Validate items & Create Order (Status: CREATED)
    API-->>User: 201 Created (order_id)
    User->>API: GET /api/v1/orders/{id}
    API-->>User: Order Status (e.g. PREPARING, DELIVERED)
    User->>API: GET /api/v1/orders/{id}/items
    API-->>User: Order Items Details
```

### CJM Ресторана (Через Mock-Сервис)
```mermaid
sequenceDiagram
    participant API as Kitchen API
    participant Mock as Mock Restaurant
    participant DB as PostgreSQL

    API->>Mock: POST /webhook (order_id, total_amount)
    Mock-->>API: 200 OK
    Note over Mock: Start processing order
    Mock->>API: PATCH /api/v1/orders/{id}/status {"status":"ACCEPTED"}
    API->>DB: UPDATE order status
    API-->>Mock: 200 OK
    Mock->>API: PATCH /api/v1/orders/{id}/status {"status":"PREPARING"}
    API->>DB: UPDATE order status
    API-->>Mock: 200 OK
```

## Архитектура (C4 Container Diagram)

```mermaid
flowchart TD
    %% Определяем стили в духе C4
    classDef person fill:#08427b,color:#fff,stroke:#052e56,stroke-width:2px
    classDef container fill:#438dd5,color:#fff,stroke:#2e6295,stroke-width:2px
    classDef external fill:#999999,color:#fff,stroke:#6b6b6b,stroke-width:2px
    classDef boundary fill:none,stroke:#444,stroke-width:2px,stroke-dasharray: 5 5

    user((" User\n(Customer)")):::person
    
    subgraph System_Boundary ["Avito Kitchen System"]
        api[" Kitchen API\n(Go, go-chi)"]:::container
        db[(" Database\n(PostgreSQL 16)")]:::container
    end
    class System_Boundary boundary

    mock[" Mock Restaurant\n(External System)"]:::external
    
    user -- "Uses (HTTPS/JSON)" --> api
    api -- "Reads / Writes" --> db
    api -- "Sends Webhook\n(HTTP POST)" --> mock
    mock -- "Updates Status\n(HTTP PATCH)" --> api
```

## Схема Базы Данных
В проекте используются 5 основных таблиц:
- `users`: Пользователи системы.
- `restaurants`: Рестораны с указанием `webhook_url`.
- `menu_items`: Элементы меню ресторана с ценой и доступностью.
- `orders`: Заказы (связаны с пользователем и рестораном, содержат итоговую сумму и статус).
- `order_items`: Связь "один-ко-многим" для конкретных позиций в заказе.

## Запуск
Для запуска всех сервисов выполните команду:
```bash
docker-compose up --build
```
API будет доступно на порту `8080`, Mock-ресторан на порту `8081`. 
База данных Postgres проброшена на порт `15432` (чтобы не конфликтовать с моими локальными серверами).
При первом запуске база автоматически заполняется таблицами и тестовыми данными (ресторанами и меню) через сиды в `/migrations`.

Спецификация API (OpenAPI) находится в `api/openapi.yaml`. 
Конфигурация линтера — `.golangci.yml`. 
