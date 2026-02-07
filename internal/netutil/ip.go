package netutil

import (
	"errors"
	"net"
)

// GetLocalIP возвращает локальный IP-адрес клиента
func GetLocalIP() (string, error) {
	// Получаем все сетевые интерфейсы
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// Проходим по всем интерфейсам и ищем подходящий IP
	for _, iface := range interfaces {
		// Пропускаем loopback и неактивные интерфейсы
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addresses {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Пропускаем IPv6 и loopback адреса
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String(), nil
		}
	}

	// Если не нашли подходящий IP, возвращаем ошибку
	return "", errors.New("no network interface found")
}
