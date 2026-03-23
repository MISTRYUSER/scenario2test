package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "scenario2test/internal/config"
)

func TestGenerateEndpoint(t *testing.T) {
	mux := newMux("", appconfig.Default())

	body, err := os.ReadFile(filepath.Join("..", "..", "examples", "login.yaml"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/x-yaml")
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	payload := recorder.Body.String()
	for _, expected := range []string{`"scenario":"login flow"`, `"test_cases":[`, `"selenium_scripts":[`} {
		if !strings.Contains(payload, expected) {
			t.Fatalf("response missing %q in %s", expected, payload)
		}
	}
	for _, unexpected := range []string{`"ui_test_cases":[`, `"unit_tests":[`} {
		if strings.Contains(payload, unexpected) {
			t.Fatalf("response unexpectedly contains %q in %s", unexpected, payload)
		}
	}
}

func TestStaticFrontendFallback(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html><body>scenario2test</body></html>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	mux := newMux(tempDir, appconfig.Default())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body, _ := io.ReadAll(recorder.Body)
	if !strings.Contains(string(body), "scenario2test") {
		t.Fatalf("unexpected response body: %s", string(body))
	}
}

func TestGenerateEndpointSupportsFlexibleDSL(t *testing.T) {
	mux := newMux("", appconfig.Default())

	body := `scenario:
  name: "E-Commerce Checkout Path Discovery"
  steps:
    - action: "open_url"
      name: "访问首页"
      params: { url: "https://mall.com" }
      next: "check_auth"
    - action: "conditional_branch"
      name: "检查登录状态"
      branches:
        - condition: "auth == 'unlogged'"
          next: "user_login"
        - condition: "auth == 'logged'"
          next: "search_item"
    - action: "user_login"
      name: "登录流程"
      params: { user: "demo", pwd: "secret" }
      next: "search_item"
    - action: "search_item"
      name: "搜索商品"
      params: { keyword: "MacBook" }
      next: "add_to_cart"
    - action: "add_to_cart"
      name: "加入购物车"
      next: "terminal_confirm"
    - action: "terminal_confirm"
      name: "结算并断言"
      type: "end"
`

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-yaml")
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", recorder.Code, recorder.Body.String())
	}

	payload := recorder.Body.String()
	for _, expected := range []string{
		`"scenario":"E-Commerce Checkout Path Discovery"`,
		`"path_01"`,
		`"path_02"`,
		`"start_url":"https://mall.com"`,
	} {
		if !strings.Contains(payload, expected) {
			t.Fatalf("response missing %q in %s", expected, payload)
		}
	}
}
