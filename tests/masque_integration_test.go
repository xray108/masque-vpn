package tests

import (
	"testing"
	"time"

	common "github.com/iselt/masque-vpn/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestMASQUEConnBasicFunctionality(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Создаем MASQUE соединение для сервера
	serverConn := common.NewMASQUEConnForServer(logger)
	assert.NotNil(t, serverConn)
	
	// Тестируем запись и чтение через каналы
	testPacket := []byte{0x45, 0x00, 0x00, 0x1c, 0x00, 0x01, 0x00, 0x00, 0x40, 0x01}
	
	// Записываем пакет
	err := serverConn.WritePacket(testPacket)
	assert.NoError(t, err)
	
	// Читаем пакет
	buffer := make([]byte, 100)
	n, err := serverConn.ReadPacket(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testPacket), n)
	assert.Equal(t, testPacket, buffer[:n])
	
	// Закрываем соединение
	err = serverConn.Close()
	assert.NoError(t, err)
	
	// Проверяем, что после закрытия операции возвращают ошибки
	err = serverConn.WritePacket(testPacket)
	assert.Error(t, err)
	
	n, err = serverConn.ReadPacket(buffer)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestMASQUEConnTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	serverConn := common.NewMASQUEConnForServer(logger)
	
	// Тестируем таймаут при чтении из пустого канала
	buffer := make([]byte, 100)
	
	start := time.Now()
	n, err := serverConn.ReadPacket(buffer)
	duration := time.Since(start)
	
	assert.Error(t, err)
	assert.Equal(t, 0, n)
	assert.Contains(t, err.Error(), "timeout")
	// Проверяем, что таймаут сработал примерно через ожидаемое время
	assert.True(t, duration >= 100*time.Millisecond)
	assert.True(t, duration < 35*time.Second) // Должно быть намного меньше полного таймаута
	
	serverConn.Close()
}

func TestMASQUEProxyFunctions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Создаем mock TUN устройство и MASQUE соединение
	// Это базовый тест структуры - полная интеграция требует реального TUN устройства
	
	masqueConn := common.NewMASQUEConnForServer(logger)
	assert.NotNil(t, masqueConn)
	
	// Тестируем, что соединение создается с правильными каналами
	testPacket := []byte{0x45, 0x00, 0x00, 0x1c}
	
	err := masqueConn.WritePacket(testPacket)
	assert.NoError(t, err)
	
	buffer := make([]byte, 100)
	n, err := masqueConn.ReadPacket(buffer)
	assert.NoError(t, err)
	assert.Equal(t, len(testPacket), n)
	
	masqueConn.Close()
}