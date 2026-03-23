package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"scenario2test/internal/aggregator"
	appconfig "scenario2test/internal/config"
	"scenario2test/internal/generator/e2e"
	"scenario2test/internal/parser"
	"scenario2test/internal/path"
	"scenario2test/internal/strategy"
)

func main() {
	scenarioPath := flag.String("scenario", "./examples/login.yaml", "path to scenario DSL file")
	listenAddr := flag.String("listen", "", "http listen address, for example :8080")
	webDist := flag.String("web-dist", "./web/dist", "optional directory of built frontend assets to serve")
	configPath := flag.String("config", "./configs/config.example.yaml", "path to application config")
	flag.Parse()

	cfg, err := appconfig.Load(*configPath)
	if err != nil {
		exitf("load config: %v", err)
	}

	if *listenAddr != "" {
		runServer(*listenAddr, *webDist, cfg)
		return
	}

	result, err := generateFromFile(*scenarioPath, cfg)
	if err != nil {
		exitf("%v", err)
	}

	payload, err := result.JSON()
	if err != nil {
		exitf("marshal result: %v", err)
	}

	fmt.Println(string(payload))
}

func runServer(listenAddr, webDist string, cfg appconfig.Config) {
	mux := newMux(webDist, cfg)
	log.Printf("Scenario2Test server listening on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		exitf("start server: %v", err)
	}
}

func newMux(webDist string, cfg appconfig.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body failed", http.StatusBadRequest)
			return
		}

		result, err := generateFromBytes(body, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "encode response failed", http.StatusInternalServerError)
		}
	})

	if distFS, ok := frontendFS(webDist); ok {
		mux.Handle("/", spaHandler(distFS))
	}

	return mux
}

func generateFromFile(scenarioPath string, cfg appconfig.Config) (aggregator.Result, error) {
	scenario, err := parser.LoadScenarioFile(scenarioPath)
	if err != nil {
		return aggregator.Result{}, fmt.Errorf("load scenario: %w", err)
	}

	return generate(scenario, cfg)
}

func generateFromBytes(body []byte, cfg appconfig.Config) (aggregator.Result, error) {
	scenario, err := parser.LoadScenarioBytes(body)
	if err != nil {
		return aggregator.Result{}, fmt.Errorf("load scenario: %w", err)
	}

	return generate(scenario, cfg)
}

func generate(scenario parser.Scenario, cfg appconfig.Config) (aggregator.Result, error) {
	graph, err := parser.BuildGraph(scenario)
	if err != nil {
		return aggregator.Result{}, fmt.Errorf("build graph: %w", err)
	}

	paths := path.EnumerateDFS(graph)
	engine := strategy.NewDefaultEngine()
	testCases := engine.Generate(paths)

	e2eGenerator := e2e.NewGenerator(cfg.E2E)

	result := aggregator.New().
		WithScenario(scenario.Name).
		WithGraph(graph).
		WithPaths(paths).
		WithTestCases(testCases).
		WithE2ESuites(e2eGenerator.Generate(scenario, paths, testCases)).
		Build()

	return result, nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func frontendFS(distPath string) (fs.FS, bool) {
	if distPath == "" {
		return nil, false
	}

	info, err := os.Stat(distPath)
	if err != nil || !info.IsDir() {
		return nil, false
	}

	return os.DirFS(distPath), true
}

func spaHandler(root fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := filepath.Clean(r.URL.Path)
		if cleanPath == "." || cleanPath == "/" {
			serveIndex(root, w, r)
			return
		}

		trimmed := cleanPath
		if len(trimmed) > 0 && trimmed[0] == '/' {
			trimmed = trimmed[1:]
		}

		if _, err := fs.Stat(root, trimmed); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		serveIndex(root, w, r)
	})
}

func serveIndex(root fs.FS, w http.ResponseWriter, r *http.Request) {
	content, err := fs.ReadFile(root, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
