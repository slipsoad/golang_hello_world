import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.InterfaceAddress;
import java.net.NetworkInterface;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Enumeration;
import java.util.List;

public class Main {

    private static final DateTimeFormatter TIMESTAMP_FORMAT =
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private static final String[] PUBLIC_IP_SERVICES = {
            "https://api.ipify.org",
            "https://checkip.amazonaws.com",
            "https://ipinfo.io/ip",
            "https://icanhazip.com"
    };

    // getLocalIP returns the local IP address of the machine
    private static String getLocalIP() throws IOException {
        Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
        while (interfaces.hasMoreElements()) {
            NetworkInterface iface = interfaces.nextElement();
            for (InterfaceAddress addr : iface.getInterfaceAddresses()) {
                InetAddress ip = addr.getAddress();
                if (ip.getAddress().length == 4 && !ip.isLoopbackAddress() && !ip.isLinkLocalAddress()) {
                    return ip.getHostAddress();
                }
            }
        }
        throw new IOException("не вдалося знайти локальний IP");
    }

    // getPublicIP returns the public IP address by querying external services
    private static String getPublicIP() throws IOException {
        for (String service : PUBLIC_IP_SERVICES) {
            try {
                URL url = new URL(service);
                HttpURLConnection conn = (HttpURLConnection) url.openConnection();
                conn.setConnectTimeout(10_000);
                conn.setReadTimeout(10_000);
                conn.setRequestMethod("GET");

                try (InputStream in = conn.getInputStream()) {
                    String body = new String(in.readAllBytes(), StandardCharsets.UTF_8).trim();
                    if (isValidIp(body)) {
                        return body;
                    }
                } finally {
                    conn.disconnect();
                }
            } catch (IOException e) {
                // спробувати наступний сервіс
            }
        }
        throw new IOException("не вдалося отримати публічний IP");
    }

    private static boolean isValidIp(String ip) {
        try {
            InetAddress.getByName(ip);
            String[] parts = ip.split("\\.");
            if (parts.length != 4) {
                return false;
            }
            for (String part : parts) {
                int n = Integer.parseInt(part);
                if (n < 0 || n > 255) {
                    return false;
                }
            }
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    // getAllLocalIPs returns all local IP addresses
    private static List<String> getAllLocalIPs() {
        List<String> ips = new ArrayList<>();
        try {
            Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
            while (interfaces.hasMoreElements()) {
                NetworkInterface iface = interfaces.nextElement();
                for (InterfaceAddress addr : iface.getInterfaceAddresses()) {
                    InetAddress ip = addr.getAddress();
                    if (ip.getAddress().length == 4 && !ip.isLoopbackAddress()) {
                        ips.add(ip.getHostAddress());
                    }
                }
            }
        } catch (IOException e) {
            // повертаємо те, що встигли зібрати
        }
        return ips;
    }

    private static String now() {
        return LocalDateTime.now().format(TIMESTAMP_FORMAT);
    }

    // logRequest логує HTTP запити
    private static void logRequest(HttpExchange exchange) {
        System.out.printf("[%s] %s %s - User-Agent: %s - Remote: %s%n",
                now(),
                exchange.getRequestMethod(),
                exchange.getRequestURI().getPath(),
                exchange.getRequestHeaders().getFirst("User-Agent"),
                exchange.getRemoteAddress());
    }

    private static void sendJson(HttpExchange exchange, String json) throws IOException {
        byte[] bytes = json.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(200, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    private static void sendHtml(HttpExchange exchange, String html) throws IOException {
        byte[] bytes = html.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "text/html; charset=utf-8");
        exchange.sendResponseHeaders(200, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    private static String jsonString(String value) {
        return "\"" + value.replace("\\", "\\\\").replace("\"", "\\\"") + "\"";
    }

    // ipHandler обробляє запити для отримання IP інформації
    private static class IpHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            logRequest(exchange);

            String localIP;
            try {
                localIP = getLocalIP();
            } catch (IOException e) {
                System.err.println("❌ Помилка отримання локального IP: " + e.getMessage());
                localIP = "невідомий";
            }

            String publicIP;
            try {
                publicIP = getPublicIP();
            } catch (IOException e) {
                System.err.println("❌ Помилка отримання публічного IP: " + e.getMessage());
                publicIP = "невідомий";
            }

            List<String> allLocalIPs = getAllLocalIPs();

            StringBuilder ipsJson = new StringBuilder("[");
            for (int i = 0; i < allLocalIPs.size(); i++) {
                if (i > 0) {
                    ipsJson.append(",");
                }
                ipsJson.append(jsonString(allLocalIPs.get(i)));
            }
            ipsJson.append("]");

            String json = "{"
                    + "\"local_ip\":" + jsonString(localIP) + ","
                    + "\"public_ip\":" + jsonString(publicIP) + ","
                    + "\"all_local_ips\":" + ipsJson + ","
                    + "\"timestamp\":" + jsonString(now())
                    + "}";

            sendJson(exchange, json);
        }
    }

    // homeHandler обробляє головну сторінку
    private static class HomeHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            logRequest(exchange);

            String html = "\n"
                    + "<!DOCTYPE html>\n"
                    + "<html lang=\"uk\">\n"
                    + "<head>\n"
                    + "\t   <meta charset=\"UTF-8\">\n"
                    + "\t   <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n"
                    + "\t   <title>IP Information Server</title>\n"
                    + "\t   <style>\n"
                    + "\t       body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }\n"
                    + "\t       .container { max-width: 600px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }\n"
                    + "\t       h1 { color: #333; text-align: center; }\n"
                    + "\t       .ip-info { background: #e8f4fd; padding: 15px; border-radius: 5px; margin: 10px 0; }\n"
                    + "\t       .btn { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 5px; }\n"
                    + "\t       .btn:hover { background: #0056b3; }\n"
                    + "\t       pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }\n"
                    + "\t   </style>\n"
                    + "</head>\n"
                    + "<body>\n"
                    + "\t   <div class=\"container\">\n"
                    + "\t       <h1>🌐 IP Information Server</h1>\n"
                    + "\t       <div class=\"ip-info\">\n"
                    + "\t           <p><strong>Цей сервер показує інформацію про IP адреси</strong></p>\n"
                    + "\t           <a href=\"/api/ip\" class=\"btn\">🔗 Отримати JSON</a>\n"
                    + "\t           <a href=\"/health\" class=\"btn\">💚 Health Check</a>\n"
                    + "\t       </div>\n"
                    + "\t       \n"
                    + "\t       <h3>📡 API Endpoints:</h3>\n"
                    + "\t       <ul>\n"
                    + "\t           <li><code>GET /</code> - Ця сторінка</li>\n"
                    + "\t           <li><code>GET /api/ip</code> - JSON з IP інформацією</li>\n"
                    + "\t           <li><code>GET /health</code> - Health check</li>\n"
                    + "\t       </ul>\n"
                    + "\t       \n"
                    + "\t       <h3>📄 Приклад відповіді JSON:</h3>\n"
                    + "\t       <pre>{\n"
                    + "\t \"local_ip\": \"192.168.1.100\",\n"
                    + "\t \"public_ip\": \"203.0.113.1\",\n"
                    + "\t \"all_local_ips\": [\"192.168.1.100\", \"10.0.0.1\"],\n"
                    + "\t \"timestamp\": \"2024-01-01 12:00:00\"\n"
                    + "}</pre>\n"
                    + "\t   </div>\n"
                    + "</body>\n"
                    + "</html>";

            sendHtml(exchange, html);
        }
    }

    // healthHandler обробляє health check запити
    private static class HealthHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            logRequest(exchange);

            String json = "{"
                    + "\"status\":" + jsonString("OK") + ","
                    + "\"timestamp\":" + jsonString(now()) + ","
                    + "\"uptime\":" + jsonString("running")
                    + "}";

            sendJson(exchange, json);
        }
    }

    public static void main(String[] args) throws IOException {
        // Отримати порт з змінної середовища або використовувати за замовчуванням
        String portEnv = System.getenv("PORT");
        int port = (portEnv == null || portEnv.isEmpty()) ? 8080 : Integer.parseInt(portEnv);

        System.out.printf("🚀 Запуск IP Information Server на порту %d%n", port);

        // Показати початкову інформацію про IP при запуску
        System.out.println("=== Початкова інформація про IP адреси ===");
        try {
            System.out.printf("🏠 Локальний IP: %s%n", getLocalIP());
        } catch (IOException e) {
            System.err.println("❌ Помилка отримання локального IP: " + e.getMessage());
        }

        try {
            System.out.printf("🌍 Публічний IP: %s%n", getPublicIP());
        } catch (IOException e) {
            System.err.println("❌ Помилка отримання публічного IP: " + e.getMessage());
        }

        // Налаштування роутів
        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/", new HomeHandler());
        server.createContext("/api/ip", new IpHandler());
        server.createContext("/health", new HealthHandler());

        // Додаткова інформація про мережеві інтерфейси при запуску
        System.out.println("📋 Отримання інформації про мережеві інтерфейси...");
        try {
            Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
            while (interfaces.hasMoreElements()) {
                NetworkInterface iface = interfaces.nextElement();
                if (!iface.isUp()) {
                    continue;
                }

                List<InterfaceAddress> addrs = iface.getInterfaceAddresses();
                if (!addrs.isEmpty()) {
                    byte[] mac = iface.getHardwareAddress();
                    System.out.printf("🔌 Інтерфейс: %s (%s)%n", iface.getName(), macToString(mac));
                    for (InterfaceAddress addr : addrs) {
                        InetAddress ip = addr.getAddress();
                        if (ip.getAddress().length == 4) {
                            System.out.printf("   IP: %s%n", ip.getHostAddress());
                        }
                    }
                }
            }
        } catch (IOException e) {
            System.err.println("❌ Помилка отримання інтерфейсів: " + e.getMessage());
        }

        System.out.printf("🌐 Веб-сервер доступний за адресою: http://localhost:%d%n", port);
        System.out.printf("📊 API endpoint: http://localhost:%d/api/ip%n", port);
        System.out.printf("💚 Health check: http://localhost:%d/health%n", port);

        // Запуск сервера
        server.start();
    }

    private static String macToString(byte[] mac) {
        if (mac == null) {
            return "";
        }
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < mac.length; i++) {
            sb.append(String.format("%02X%s", mac[i], (i < mac.length - 1) ? ":" : ""));
        }
        return sb.toString();
    }
}
