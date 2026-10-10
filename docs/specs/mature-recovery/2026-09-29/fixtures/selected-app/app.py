import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path != "/__byteport_probe":
            self.send_response(404)
            self.end_headers()
            return
        body = json.dumps({
            "source_commit": os.environ["SOURCE_COMMIT"],
            "manifest_digest": os.environ["MANIFEST_DIGEST"],
        }).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_):
        pass

HTTPServer(("0.0.0.0", 8080), Handler).serve_forever()
