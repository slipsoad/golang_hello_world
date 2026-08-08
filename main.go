package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
		// defer resp.Body.Close() // БАГ 1: Закоментовано закриття Body - витік ресурсів

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

// IPInfo структура для відповіді веб-сервера
type IPInfo struct {
	LocalIP    string   `json:"local_ip"`
	PublicIP   string   `json:"public_ip"`
	AllLocalIPs []string `json:"all_local_ips"`
	Timestamp  string   `json:"timestamp"`
}

// logRequest логує HTTP запити
func logRequest(r *http.Request) {
	log.Printf("[%s] %s %s - User-Agent: %s - Remote: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		r.Method,
		r.URL.Path,
		r.UserAgent(),
		r.RemoteAddr)
}

// ipHandler обробляє запити для отримання IP інформації
func ipHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	
	// Отримати локальний IP
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("❌ Помилка отримання локального IP: %v", err)
		localIP = "невідомий"
	}
	
	// Отримати публічний IP
	publicIP, err := getPublicIP()
	if err != nil {
		log.Printf("❌ Помилка отримання публічного IP: %v", err)
		publicIP = "невідомий"
	}
	
	// Отримати всі локальні IP
	allLocalIPs := getAllLocalIPs()
	
	ipInfo := IPInfo{
		LocalIP:     localIP,
		PublicIP:    publicIP,
		AllLocalIPs: allLocalIPs,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ipInfo)
}

// homeHandler обробляє головну сторінку
func homeHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	
	html := `
<!DOCTYPE html>
<html lang="uk">
<head>
	   <meta charset="UTF-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1.0">
	   <title>IP Information Server</title>
	   <style>
	       body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
	       .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
	       h1 { color: #333; text-align: center; }
	       .ip-info { background: #e8f4fd; padding: 15px; border-radius: 5px; margin: 10px 0; }
	       .btn { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 5px; }
	       .btn:hover { background: #0056b3; }
	       pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
	   </style>
</head>
<body>
	   <div class="container">
	       <h1>🌐 IP Information Server</h1>
	       <div class="ip-info">
	           <p><strong>Цей сервер показує інформацію про IP адреси</strong></p>
	           <a href="/api/ip" class="btn">🔗 Отримати JSON</a>
	           <a href="/health" class="btn">💚 Health Check</a>
	       </div>
	       
	       <h3>📡 API Endpoints:</h3>
	       <ul>
	           <li><code>GET /</code> - Ця сторінка</li>
	           <li><code>GET /api/ip</code> - JSON з IP інформацією</li>
	           <li><code>GET /health</code> - Health check</li>
	       </ul>
	       
	       <h3>📄 Приклад відповіді JSON:</h3>
	       <pre>{
	 "local_ip": "192.168.1.100",
	 "public_ip": "203.0.113.1",
	 "all_local_ips": ["192.168.1.100", "10.0.0.1"],
	 "timestamp": "2024-01-01 12:00:00"
}</pre>
	   </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// healthHandler обробляє health check запити
func healthHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	
	health := map[string]interface{}{
		"status":    "OK",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"uptime":    "running",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
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
	// Отримати порт з змінної середовища або використовувати за замовчуванням
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Запуск IP Information Server на порту %s", port)
	
	// Показати початкову інформацію про IP при запуску
	fmt.Println("=== Початкова інформація про IP адреси ===")
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("❌ Помилка отримання локального IP: %v", err)
	} else {
		log.Printf("🏠 Локальний IP: %s", localIP)
	}
	
	publicIP, err := getPublicIP()
	if err != nil {
		log.Printf("❌ Помилка отримання публічного IP: %v", err)
	} else {
		log.Printf("🌍 Публічний IP: %s", publicIP)
	}
	
	// Налаштування роутів
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/ip", ipHandler)
	http.HandleFunc("/health", healthHandler)
	
	// Додаткова інформація про мережеві інтерфейси при запуску
	log.Println("📋 Отримання інформації про мережеві інтерфейси...")
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("❌ Помилка отримання інтерфейсів: %v", err)
	} else {
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
				log.Printf("🔌 Інтерфейс: %s (%s)", iface.Name, iface.HardwareAddr)
				for _, addr := range addrs {
					if ipNet, ok := addr.(*net.IPNet); ok {
						if ipNet.IP.To4() != nil {
							log.Printf("   IP: %s", ipNet.IP.String())
						}
					}
				}
			}
		}
	}
	
	log.Printf("🌐 Веб-сервер доступний за адресою: http://localhost:%s", port)
	log.Printf("📊 API endpoint: http://localhost:%s/api/ip", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	
	// Запуск сервера
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
