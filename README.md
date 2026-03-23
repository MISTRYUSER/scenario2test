# Scenario2Test Platform

Scenario2Test Platform is an orchestrator for scenario-driven automated test generation.

It turns a scenario DSL into:

- a control-flow graph
- enumerated execution paths
- strategy-expanded test cases
- AUTOTEST-ready end-to-end payloads and scripts

Detailed code analysis and verification notes:

- [2026-03-23 Scenario2Test Code Analysis Wiki](/Users/wentao.xue/Project/scenario2test/docs/wiki/2026-03-23-scenario2test-code-analysis.md)

## Architecture

```text
Scenario DSL -> Parser -> Flow Graph -> Path Generator -> Strategy Engine
                                                     -> E2E Adapter
                                          -> Aggregator -> JSON
```

## Project Layout

```text
scenario2test/
├── cmd/server
├── internal/parser
├── internal/graph
├── internal/path
├── internal/strategy
├── internal/generator/e2e
├── internal/aggregator
├── api
├── configs
└── examples
```

## Quick Start

1. Install Go 1.22+.
2. Run `go mod tidy`.
3. Run `go run ./cmd/server --scenario ./examples/login.yaml`.
4. Or run `go run ./cmd/server --listen :8080` to expose the HTTP API.
5. If `./web/dist` exists, the backend also serves the built frontend at `/`.
6. Run `./scripts/verify_all.sh` for a full local verification pass.

## Web UI

The repo also includes a React + Vite frontend in [web](/Users/wentao.xue/Project/scenario2test/web).

1. Start the backend with `go run ./cmd/server --listen :8080`.
2. In `/Users/wentao.xue/Project/scenario2test/web`, run `npm install`.
3. Run `npm run dev`.
4. For a production bundle, run `npm run build`, then start the Go server with `--listen :8080`.

The frontend proxies `/generate` and `/healthz` to `http://localhost:8080`.

The current implementation exposes a single AUTOTEST channel.

## Adapter Modes

Adapters are configuration-driven via [config.example.yaml](/Users/wentao.xue/Project/scenario2test/configs/config.example.yaml).

- `e2e.mode: mock | cli`
  `cli` mode executes the configured AUTOTEST command with placeholder substitution such as `{{url}}`, `{{signature}}`, and `{{auth_data_file}}`.

When an external adapter is not configured or fails, Scenario2Test keeps returning the internal draft output instead of failing the entire generation pipeline.

## Local Integrations

This workspace already includes local clones under `/Users/wentao.xue/Project/integrations` for:

- `AUTOTEST`

Recommended local config:
- [config.local.yaml](/Users/wentao.xue/Project/scenario2test/configs/config.local.yaml)
- [config.integration.yaml](/Users/wentao.xue/Project/scenario2test/configs/config.integration.yaml)

Helper scripts:
- [setup_integrations.sh](/Users/wentao.xue/Project/scenario2test/scripts/setup_integrations.sh)
- [run_local_stack.sh](/Users/wentao.xue/Project/scenario2test/scripts/run_local_stack.sh)
- [verify_all.sh](/Users/wentao.xue/Project/scenario2test/scripts/verify_all.sh)
- [run_full_integration.sh](/Users/wentao.xue/Project/scenario2test/scripts/run_full_integration.sh)

Current environment note:
- `AUTOTEST` can be installed with Python 3.9 and its CLI integration is wired.

## Full Integration Mode

`config.integration.yaml` and `run_full_integration.sh` wire the AUTOTEST generator in this workspace:

- AUTOTEST runs against the local mock LLM endpoint and the static example login page.

Use `./scripts/run_full_integration.sh 8094` to start the local mock LLM, static example page, Go server, and then POST the example scenario through the full orchestrator.
