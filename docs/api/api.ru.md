# API Документация masque-vpn

VPN сервер предоставляет HTTP API для управления и мониторинга.

**Базовый URL**: `http://<server_ip>:8080/api`

## Аутентификация

В текущей версии API защищен базовой аутентификацией (Basic Auth) или токенами, управляемыми через веб-интерфейс.

## Эндпоинты

### Управление клиентами

#### Получить список клиентов

`GET /clients`

Возвращает список всех зарегистрированных клиентов.

**Ответ:**
```json
[
  {
    "id": "client1",
    "ip": "10.99.0.2",
    "enabled": true
  }
]
```

#### Создать клиента

`POST /clients`

Создает нового клиента и генерирует конфигурацию.

**Тело запроса:**
```json
{
  "name": "student_pc_1"
}
```

#### Удалить клиента

`DELETE /clients/{id}`

Удаляет клиента и отзывает его доступ.

### Мониторинг

#### Метрики Prometheus

`GET /metrics`

Возвращает метрики в формате Prometheus. Доступно на порту 9090 (по умолчанию).

**Пример метрик:**
```text
# HELP masque_vpn_bytes_rx_total Total bytes received
# TYPE masque_vpn_bytes_rx_total counter
masque_vpn_bytes_rx_total{client_id="client1"} 1024

# HELP masque_vpn_fec_recovery_rate_percent Percentage of lost packets successfully recovered
# TYPE masque_vpn_fec_recovery_rate_percent gauge
masque_vpn_fec_recovery_rate_percent{client_id="client1"} 5.0
```

## Использование с curl

Пример получения метрик:

```bash
curl http://localhost:9090/metrics
```

Пример управления клиентами (требуется авторизация):

```bash
curl -u admin:admin -X GET http://localhost:8080/api/clients
```
