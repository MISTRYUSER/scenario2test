#!/usr/bin/env python3
import json
import re
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


def json_analysis_response():
    return json.dumps(
        {
            "auth_requirements": {
                "auth_required": True,
                "auth_type": "login",
                "auth_fields": [
                    {"name": "username", "type": "text", "required": True},
                    {"name": "password", "type": "password", "required": True},
                ],
            },
            "contact_form_fields": [],
            "main_content": "A login experience with username, password, and submit action.",
            "key_actions": ["login", "submit credentials"],
            "content_hierarchy": {
                "primary_sections": ["authentication form"],
                "subsections": ["username input", "password input", "login button"],
            },
            "interactive_patterns": {
                "forms": ["login"],
                "dynamic_elements": [],
            },
            "security_indicators": ["https"],
        }
    )


def autotest_cases_response():
    return json.dumps(
        {
            "test_cases": [
                {
                    "name": "Login with invalid password",
                    "type": "auth-negative",
                    "steps": [
                        "Open the login page",
                        "Fill username with valid_user",
                        "Fill password with invalid_password",
                        "Click the login button",
                    ],
                    "selectors": {
                        "username": "[name='username']",
                        "password": "[name='password']",
                        "submit": "#login_button",
                    },
                    "validation": "An invalid credentials error is shown.",
                    "test_data": {
                        "username": "valid_user",
                        "password": "invalid_password",
                    },
                }
            ]
        }
    )


def extract_current_url(messages):
    combined = "\n".join((m.get("content") or "") for m in messages)
    match = re.search(r"Current page URL:\s*(\S+)", combined)
    if match:
        return match.group(1)
    match = re.search(r'"url"\s*:\s*"([^"]+)"', combined)
    if match:
        return match.group(1)
    return "http://127.0.0.1:8096/login.html"


def autotest_script_response(messages):
    url = extract_current_url(messages)
    return f"""```python
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from webdriver_manager.chrome import ChromeDriverManager
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait

options = Options()
options.binary_location = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
options.add_argument("--headless")
options.add_argument("--no-sandbox")
options.add_argument("--disable-gpu")
service = Service(ChromeDriverManager().install())
driver = webdriver.Chrome(service=service, options=options)
driver.get("{url}")
WebDriverWait(driver, 10).until(lambda d: d.execute_script("return document.readyState") == "complete")
driver.find_element(By.NAME, "username").send_keys("valid_user")
driver.find_element(By.NAME, "password").send_keys("invalid_password")
driver.find_element(By.ID, "login_button").click()
print("mock selenium script generated")
driver.quit()
```"""

def choose_chat_content(messages):
    combined = "\n".join((m.get("content") or "") for m in messages)
    lowered = combined.lower()
    if "analyze this web page structure and return json metadata" in lowered:
        return json_analysis_response()
    if "output only valid json using this exact structure" in lowered and '"selectors"' in lowered:
        return autotest_cases_response()
    if "return only executable" in lowered and "selenium" in lowered:
        return autotest_script_response(messages)
    return json.dumps({"message": "mock"})


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/healthz":
            self.send_response(200)
            self.send_header("Content-Type", "text/plain; charset=utf-8")
            self.end_headers()
            self.wfile.write(b"ok")
            return
        self.send_response(404)
        self.end_headers()

    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length) if length else b"{}"
        try:
            payload = json.loads(raw.decode("utf-8"))
        except Exception:
            payload = {}

        if self.path.endswith("/chat/completions"):
            content = choose_chat_content(payload.get("messages", []))
            body = {
                "id": "chatcmpl-mock",
                "object": "chat.completion",
                "choices": [
                    {
                        "index": 0,
                        "message": {"role": "assistant", "content": content},
                        "finish_reason": "stop",
                    }
                ],
            }
            self.send_json(body)
            return

        self.send_response(404)
        self.end_headers()

    def log_message(self, fmt, *args):
        return

    def send_json(self, payload):
        raw = json.dumps(payload).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", 11434), Handler)
    print("mock llm listening on http://127.0.0.1:11434")
    server.serve_forever()
