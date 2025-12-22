#!/usr/bin/env python3
"""
Proxy stub server for cloudflared tunnel.
- Forwards requests to real server if available
- Returns 200 OK if real server is down (keeps Notion webhook alive)
Run: make stub
"""

import json
import logging
import os
import socket
from http.server import HTTPServer, BaseHTTPRequestHandler
from http.client import HTTPConnection
from datetime import datetime
from pathlib import Path

# Configure logging
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

logger = logging.getLogger(__name__)

# Stub server always runs on port 9999
STUB_SERVER_PORT = 9999


class ProxyHandler(BaseHTTPRequestHandler):
    """Proxy handler that forwards to real server or returns 200 if server is down"""

    real_server_port = None  # Will be set on server start

    def _is_server_available(self):
        """Check if real server is available"""
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(0.5)
            result = sock.connect_ex(('localhost', self.real_server_port))
            sock.close()
            return result == 0
        except Exception as e:
            logger.debug(f"Server check failed: {e}")
            return False

    def _proxy_request(self, method):
        """Proxy request to real server or return 200"""
        logger.debug(f"{method} {self.path} from {self.client_address[0]}")

        # Check if real server is available
        if not self._is_server_available():
            logger.debug("Real server unavailable, returning 200 OK")
            self._send_stub_response()
            return

        # Read request body
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length) if content_length > 0 else b''

        try:
            # Forward to real server
            logger.debug(f"Forwarding to real server on port {self.real_server_port}")
            conn = HTTPConnection('localhost', self.real_server_port, timeout=15)

            # Forward headers (exclude hop-by-hop headers)
            forward_headers = {}
            skip_headers = {'connection', 'keep-alive', 'proxy-authenticate', 'proxy-authorization', 'te', 'trailers', 'transfer-encoding', 'upgrade'}
            for key, value in self.headers.items():
                if key.lower() not in skip_headers:
                    forward_headers[key] = value

            conn.request(method, self.path, body, forward_headers)
            response = conn.getresponse()

            # Forward response
            self.send_response(response.status)
            for key, value in response.getheaders():
                if key.lower() not in skip_headers:
                    self.send_header(key, value)
            self.end_headers()
            self.wfile.write(response.read())

            conn.close()
            logger.debug(f"Forwarded successfully, status: {response.status}")

        except Exception as e:
            logger.error(f"Proxy error: {e}")
            self._send_stub_response()

    def _send_stub_response(self):
        """Send 200 OK stub response"""
        response = {
            'status': 'ok',
            'message': 'Stub proxy response (real server unavailable)',
            'timestamp': datetime.utcnow().isoformat() + 'Z'
        }

        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(response).encode('utf-8'))

    def do_GET(self):
        """Handle GET requests"""
        self._proxy_request('GET')

    def do_POST(self):
        """Handle POST requests"""
        self._proxy_request('POST')

    def do_PUT(self):
        """Handle PUT requests"""
        self._proxy_request('PUT')

    def do_DELETE(self):
        """Handle DELETE requests"""
        self._proxy_request('DELETE')

    def log_message(self, format, *args):
        """Override to use our logger"""
        pass


def load_env_file():
    """Load environment variables from .env file"""
    env_path = Path(__file__).parent.parent / '.env'

    if not env_path.exists():
        logger.warning(f".env file not found at {env_path}")
        return

    logger.debug(f"Loading environment from {env_path}")

    with open(env_path, 'r') as f:
        for line in f:
            line = line.strip()
            # Skip comments and empty lines
            if not line or line.startswith('#'):
                continue

            # Parse KEY=VALUE
            if '=' in line:
                key, value = line.split('=', 1)
                key = key.strip()
                value = value.strip()

                # Remove quotes if present
                if value.startswith('"') and value.endswith('"'):
                    value = value[1:-1]
                elif value.startswith("'") and value.endswith("'"):
                    value = value[1:-1]

                os.environ[key] = value
                logger.debug(f"Loaded: {key}={value}")


def get_server_port():
    """Get server port from environment or use default"""
    # Try different common port environment variables
    port_vars = ['PORT', 'SERVER_PORT', 'HTTP_PORT']

    for var in port_vars:
        port_str = os.environ.get(var)
        if port_str:
            try:
                port = int(port_str)
                logger.debug(f"Using port {port} from {var}")
                return port
            except ValueError:
                logger.warning(f"Invalid port value in {var}: {port_str}")

    logger.debug("No port found in environment, using default 8080")
    return 8080


def run_server():
    """Start the proxy stub server"""
    load_env_file()
    real_server_port = get_server_port()

    # Set real server port on handler class
    ProxyHandler.real_server_port = real_server_port

    server_address = ('localhost', STUB_SERVER_PORT)
    httpd = HTTPServer(server_address, ProxyHandler)

    logger.info(f"Starting proxy stub server on http://localhost:{STUB_SERVER_PORT}")
    logger.info(f"Real server should run on port {real_server_port}")
    logger.info("Behavior:")
    logger.info(f"  - If real server available: forwards requests to port {real_server_port}")
    logger.info("  - If real server down: returns 200 OK (keeps webhook alive)")
    logger.info("Press Ctrl+C to stop")

    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        logger.info("Stopping proxy stub server...")
        httpd.shutdown()
        logger.info("Proxy stub server stopped")


if __name__ == '__main__':
    run_server()
