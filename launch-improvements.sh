#!/bin/bash

API_URL="http://localhost:3000/api/v1"
PROJECT_DIR="/Users/maxlanies/Git/2GC/cloudbridge-relay-installer/oss-repositories/cloudbridge/masque-vpn"

echo "🚀 Запуск миссии: MASQUE VPN Improvements"
echo "📋 Workflow: TASK-002 (Research → Arch → Plan → Dev)"
echo ""

# Создаём задачу
cat > /tmp/masque-task.md << 'EOF'
# Задача: Улучшение MASQUE VPN Client

## Цель
Модернизировать клиентскую часть `masque-vpn` (`vpn_client`), внедрив современные практики наблюдаемости и обновив зависимости.

## Требования

### 1. Обновление зависимостей
- Обновить `github.com/quic-go/quic-go` до версии `v0.57.1` (или последней стабильной) во всех модулях:
  - `vpn_client/go.mod`
  - `vpn_server/go.mod`
  - `common/go.mod` (если есть)
- Убедиться, что код компилируется с новой версией (возможно потребуются правки API).

### 2. Структурированное логирование
- Заменить стандартный пакет `log` на `go.uber.org/zap` в `vpn_client`.
- Настроить формат логов (JSON для production, Console для dev).
- Логировать важные события: подключение, ошибки, изменение IP/маршрутов.

### 3. Метрики (Observability)
- Добавить Prometheus metrics server в `vpn_client` (например, на порту 9090 или :8081/metrics).
- Реализовать метрики:
  - `vpn_client_bytes_sent_total` (Counter)
  - `vpn_client_bytes_received_total` (Counter)
  - `vpn_client_connection_status` (Gauge: 0=Disconnected, 1=Connected)
  - `vpn_client_errors_total` (Counter)

## Ожидаемый результат
- Клиент успешно собирается и запускается.
- Логи пишутся через Zap.
- Метрики доступны по HTTP endpoint.
- Зависимости обновлены.
EOF

RESPONSE=$(curl -s -X POST "${API_URL}/missions/create" \
  -H "Content-Type: application/json" \
  -d "{
    \"task\": \"$(cat /tmp/masque-task.md | sed 's/"/\\"/g' | tr '\n' ' ')\",
    \"targetDirectory\": \"${PROJECT_DIR}\",
    \"workflowId\": \"TASK-002\",
    \"agentLevel\": \"senior\",
    \"autoContinue\": true,
    \"metadata\": {
      \"source\": \"cli-script\",
      \"project\": \"masque-vpn\",
      \"type\": \"refactoring\"
    }
  }")

echo "$RESPONSE" | jq '.'
echo ""
echo "✅ Миссия MASQUE VPN запущена! ID: $(echo "$RESPONSE" | jq -r '.missionId')"
