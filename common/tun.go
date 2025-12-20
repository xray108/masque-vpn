package common

import (
	"fmt"
	"net"
	"net/netip"
	"sync"

	"golang.zx2c4.com/wireguard/tun"
)

// TUNDevice represents a cross-platform TUN device
type TUNDevice struct {
	device tun.Device
	name   string
}

// Name returns the device name
func (t *TUNDevice) Name() string {
	return t.name
}

// Close closes the TUN device
func (t *TUNDevice) Close() error {
	return t.device.Close()
}

// SetIP sets the IP address for the TUN device (platform-specific implementation)
func (t *TUNDevice) SetIP(ipNet net.IPNet) error {
	return setPlatformIP(t, ipNet)
}

// AddRoute adds a route through the TUN device (platform-specific implementation)
func (t *TUNDevice) AddRoute(ipNet net.IPNet) error {
	return addPlatformRoute(t, ipNet)
}

// getDefaultTunName returns the default TUN device name for the platform
func getDefaultTunName() string {
	return getDefaultPlatformTunName()
}

// createPlatformTunDevice creates a platform-specific TUN device
func createPlatformTunDevice(name string, mtu int) (tun.Device, error) {
	return tun.CreateTUN(name, mtu)
}

// Предварительно выделенные буферы для TUN устройства для повышения производительности
type tunBuffers struct {
	bufs  [][]byte // Массив буферов для пакетов
	sizes []int    // Массив размеров пакетов
}

// Пул объектов для переиспользования tunBuffers структур
var tunBuffersPool = sync.Pool{
	New: func() interface{} {
		return &tunBuffers{
			bufs:  make([][]byte, 1),
			sizes: make([]int, 1),
		}
	},
}

// ReadPacket читает один пакет из TUN устройства с оптимизацией через пул объектов
func (t *TUNDevice) ReadPacket(packet []byte, offset int) (int, error) {
	// Получаем структуру из пула
	tb := tunBuffersPool.Get().(*tunBuffers)

	// Устанавливаем ссылку на переданный буфер
	tb.bufs[0] = packet
	tb.sizes[0] = 0

	// Вызываем базовый метод чтения
	n, err := t.device.Read(tb.bufs, tb.sizes, offset)

	// Сохраняем результат
	var size int
	if n > 0 {
		size = tb.sizes[0]
	}

	// Очищаем ссылку для предотвращения утечек памяти
	tb.bufs[0] = nil

	// Возвращаем структуру в пул
	tunBuffersPool.Put(tb)

	if err != nil {
		return 0, err
	}

	if n == 0 {
		return 0, nil
	}

	return size, nil
}

// Write записывает данные в TUN устройство, используя нативный интерфейс
func (t *TUNDevice) Write(bufs [][]byte, offset int) (int, error) {
	return t.device.Write(bufs, offset)
}

// WritePacket записывает один пакет в TUN устройство с оптимизацией через пул объектов
func (t *TUNDevice) WritePacket(packet []byte, offset int) error {
	// Получаем структуру из пула
	tb := tunBuffersPool.Get().(*tunBuffers)

	// Устанавливаем ссылку на переданный буфер
	tb.bufs[0] = packet

	// Вызываем базовый метод записи
	_, err := t.device.Write(tb.bufs, offset)

	// Очищаем ссылку для предотвращения утечек памяти
	tb.bufs[0] = nil

	// Возвращаем структуру в пул
	tunBuffersPool.Put(tb)

	return err
}

// BatchSize возвращает размер батча для устройства
func (t *TUNDevice) BatchSize() int {
	return t.device.BatchSize()
}

// CreateTunDevice creates and configures a TUN device
func CreateTunDevice(name string, ipNet net.IPNet, mtu int) (*TUNDevice, error) {
	// Use default name if empty
	if name == "" {
		name = getDefaultTunName()
	}

	// Create the TUN device
	device, err := createPlatformTunDevice(name, mtu)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Get the actual device name (may be different from requested)
	actualName, err := device.Name()
	if err != nil {
		device.Close()
		return nil, fmt.Errorf("failed to get device name: %w", err)
	}

	tunDev := &TUNDevice{
		device: device,
		name:   actualName,
	}

	// Set IP address
	if err := tunDev.SetIP(ipNet); err != nil {
		device.Close()
		return nil, fmt.Errorf("failed to set IP address: %w", err)
	}

	return tunDev, nil
}

// AddRoute добавляет маршрут для указанного TUN устройства (глобальная функция для совместимости)
func AddRoute(tunDevice *TUNDevice, prefix netip.Prefix) error {
	// Конвертируем netip.Prefix в net.IPNet для совместимости
	ipNet := net.IPNet{
		IP:   prefix.Addr().AsSlice(),
		Mask: net.CIDRMask(prefix.Bits(), prefix.Addr().BitLen()),
	}
	return tunDevice.AddRoute(ipNet)
}
