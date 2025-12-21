# API Документация masque-vpn

VPN сервер предоставляет REST HTTP API для управления и мониторинга.

**Базовый URL**: `http://<server_ip>:8080/api/v1`

## Обзор

API сервер работает на отдельном порту (по умолчанию 8080) и предоставляет JSON endpoints для:
- Мониторинга состояния сервера
- Управления клиентскими соединениями  
- Получения статистики и логов
- Просмотра конфигурации

## Endpoints

### Проверка состояния

#### Health Check

`GET /health`

Проверка работоспособности API сервера.

**Ответ:**
```json
{
  "status": "healthy",
  "service": "masque-vpn-server",
  "time": "2025-12-21T00:59:38.7596593Z"
}
```

#### Статус сервера

`GET /api/v1/status`

Возвращает текущий статус VPN сервера.

**Ответ:**
```json
{
  "status": "running",
  "active_connections": 0,
  "listen_addr": "127.0.0.1:4433",
  "network_cidr": "10.0.0.0/24",
  "server_name": "masque-vpn-server",
  "tun_device": "disabled"
}
```

### Управление клиентами

#### Получить список клиентов

`GET /api/v1/clients`

Возвращает список всех подключенных клиентов.

**Ответ:**
```json
{
  "clients": [
    {
      "id": "client-uuid-123",
      "assigned_ip": "10.0.0.2",
      "connected_at": "2025-12-21T00:30:00Z",
      "bytes_sent": 1024,
      "bytes_received": 2048,
      "status": "connected"
    }
  ],
  "total": 1
}
```

#### Получить информацию о клиенте

`GET /api/v1/clients/{id}`

Возвращает детальную информацию о конкретном клиенте.

**Ответ:**
```json
{
  "id": "client-uuid-123",
  "assigned_ip": "10.0.0.2",
  "connected_at": "2025-12-21T00:30:00Z",
  "bytes_sent": 1024,
  "bytes_received": 2048,
  "status": "connected"
}
```

#### Отключить клиента

`DELETE /api/v1/clients/{id}`

Принудительно отключает клиента от VPN.

**Ответ:**
```json
{
  "message": "Client disconnected",
  "client_id": "client-uuid-123"
}
```

### Статистика и мониторинг

#### Статистика сервера

`GET /api/v1/stats`

Возвращает детальную статистику работы сервера.

**Ответ:**
```json
{
  "active_connections": 5,
  "total_connections": 150,
  "network_cidr": "10.0.0.0/24",
  "tun_device": "tun0",
  "uptime": 3600000000000,
  "packets_forwarded": 10000,
  "bytes_forwarded": 5242880
}
```

#### Логи соединений

`GET /api/v1/logs`

Возвращает логи соединений клиентов.

**Ответ:**
```json
{
  "logs": [
    {
      "id": 1,
      "client_id": "client-uuid-123",
      "event_type": "connect",
      "timestamp": "2025-12-21T00:30:00Z",
      "details": "Client connected successfully"
    }
  ],
  "total": 1
}
```

#### Конфигурация сервера

`GET /api/v1/config`

Возвращает текущую конфигурацию сервера (без секретных данных).

**Ответ:**
```json
{
  "listen_addr": "127.0.0.1:4433",
  "assign_cidr": "10.0.0.0/24",
  "advertise_routes": ["0.0.0.0/0"],
  "server_name": "masque-vpn-server",
  "mtu": 1413,
  "log_level": "debug",
  "enable_ipv6": false,
  "fec_enabled": false,
  "metrics_enabled": true
}
```

### Метрики Prometheus

#### Метрики в формате Prometheus

`GET /metrics`

Возвращает метрики в формате Prometheus для мониторинга.

**Пример метрик:**
```text
# HELP masque_vpn_active_connections Current number of active connections
# TYPE masque_vpn_active_connections gauge
masque_vpn_active_connections 5

# HELP masque_vpn_bytes_sent_total Total bytes sent to clients
# TYPE masque_vpn_bytes_sent_total counter
masque_vpn_bytes_sent_total{client_id="client1"} 1024
```

## Использование с curl

### Примеры запросов

Проверка состояния:
```bash
curl http://localhost:8080/health
```

Получение статуса сервера:
```bash
curl http://localhost:8080/api/v1/status
```

Список клиентов:
```bash
curl http://localhost:8080/api/v1/clients
```

Отключение клиента:
```bash
curl -X DELETE http://localhost:8080/api/v1/clients/client-uuid-123
```

Получение метрик:
```bash
curl http://localhost:8080/metrics
```

## Коды ошибок

- `200 OK` - Успешный запрос
- `404 Not Found` - Клиент не найден
- `500 Internal Server Error` - Внутренняя ошибка сервера

## Примечания для разработчиков

- API использует in-memory хранилище для логов (до 1000 записей)
- Все timestamps возвращаются в формате RFC3339 (UTC)
- Метрики bytes_sent/bytes_received пока возвращают 0 (заглушка для будущей реализации)
- API не требует аутентификации в текущей учебной версии
