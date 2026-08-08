import http.server
import os
import socket
import urllib.request
import urllib.error

PORT = int(os.environ.get("AGENT_LOCAL_PORT", os.environ.get("PORT", 5000)))


def get_local_ip():
    """Get the local IP address of the server"""
    try:
        # Create a socket to connect to a remote server
        # This doesn't actually connect, just determines local IP
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
            s.connect(("8.8.8.8", 80))
            local_ip = s.getsockname()[0]
            return local_ip
    except Exception:
        return "127.0.0.1"


def get_public_ip():
    """Get the public IP address of the server"""
    try:
        # Try HTTP services to avoid SSL certificate issues on macOS
        services = [
            "http://icanhazip.com",
            "http://ipecho.net/plain",
            "http://checkip.amazonaws.com"
        ]
        
        for service in services:
            try:
                with urllib.request.urlopen(service, timeout=5) as response:
                    return response.read().decode('utf-8').strip()
            except (urllib.error.URLError, urllib.error.HTTPError):
                continue
        
        return "Unable to detect"
    except Exception:
        return "Unable to detect"


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.end_headers()
        
        # Get both local and public IP addresses
        local_ip = get_local_ip()
        public_ip = get_public_ip()
        
        # Create response message with both IP addresses
        env_lines = "\n".join(f"{key}={value}" for key, value in sorted(os.environ.items()))

        message = f"""Hello world from World8!
Local IP address: {local_ip}
Public IP address: {public_ip}
Server port: {PORT}
This is environment {os.environ.get("ENV")}

Environment variables:
{env_lines}
"""
        self.wfile.write(message.encode('utf-8'))

    def log_message(self, format, *args):
        print("%s - %s" % (self.address_string(), format % args), flush=True)


if __name__ == "__main__":
    local_ip = get_local_ip()
    public_ip = get_public_ip()
    
    print(f"Starting hello-world agent on {local_ip}:{PORT}", flush=True)
    print(f"Local IP address: {local_ip}", flush=True)
    print(f"Public IP address: {public_ip}", flush=True)
    http.server.HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()
