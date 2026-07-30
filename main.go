package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// getLocalIP returns the local IP address of the machine
func getLocalIP() (string, error) {
	// Get all network interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// Check if it's a valid IP and not a loopback
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			// Check if it's IPv4 and not a local link address
			if ipNet.IP.To4() != nil && !ipNet.IP.IsLinkLocalUnicast() {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("не вдалося знайти локальний IP")
}

// getPublicIP returns the public IP address by querying external services
func getPublicIP() (string, error) {
	// List of services to try for getting public IP
	services := []string{
		"https://api.ipify.org",
		"https://checkip.amazonaws.com",
		"https://ipinfo.io/ip",
		"https://icanhazip.com",
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		// Validate that we got a proper IP
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	return "", fmt.Errorf("не вдалося отримати публічний IP")
}

// getAllLocalIPs returns all local IP addresses
func getAllLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

func main() {
	fmt.Println("=== Інформація про IP адреси ===")
	fmt.Println()

	// Отримати локальний IP
	localIP, err := getLocalIP()
	if err != nil {
		fmt.Printf("❌ Помилка отримання локального IP: %v\n", err)
	} else {
		fmt.Printf("🏠 Основний локальний IP: %s\n", localIP)
	}

	// Отримати всі локальні IP адреси
	allLocalIPs := getAllLocalIPs()
	if len(allLocalIPs) > 0 {
		fmt.Printf("📍 Всі локальні IP адреси:\n")
		for _, ip := range allLocalIPs {
			fmt.Printf("   - %s\n", ip)
		}
	}

	fmt.Println()

	// Отримати публічний IP
	fmt.Print("🌐 Отримання публічного IP... ")
	publicIP, err := getPublicIP()
	if err != nil {
		fmt.Printf("\n❌ Помилка отримання публічного IP: %v\n", err)
	} else {
		fmt.Printf("\n🌍 Публічний IP: %s\n", publicIP)
	}

	fmt.Println()
	fmt.Println("=== Завершено ===")

	// Додаткова інформація про мережеві інтерфейси
	fmt.Println()
	fmt.Println("📋 Детальна інформація про мережеві інтерфейси:")
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("❌ Помилка отримання інтерфейсів: %v\n", err)
		return
	}

	for _, iface := range interfaces {
		// Пропускаємо неактивні інтерфейси
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		if len(addrs) > 0 {
			fmt.Printf("\n🔌 Інтерфейс: %s (%s)\n", iface.Name, iface.HardwareAddr)
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok {
					if ipNet.IP.To4() != nil {
						fmt.Printf("   IP: %s\n", ipNet.IP.String())
					}
				}
			}
		}
	}
}
