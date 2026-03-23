import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import App from "./App";

const sampleResponse = {
  paths: [{ id: "path-1" }, { id: "path-2" }],
  test_cases: [
    {
      id: "tc-happy",
      path_id: "path-1",
      pathId: "path-1",
      strategy: "happy_path",
      severity: "P1",
      title: "happy path :: path-1",
      description: "Happy path login succeeds",
    },
    {
      id: "tc-auth",
      path_id: "path-1",
      pathId: "path-1",
      strategy: "auth_fail",
      severity: "P0",
      title: "auth fail :: path-1",
      description: "Invalid password is rejected",
    },
  ],
  selenium_scripts: [
    {
      provider: "AUTOTEST",
      path_id: "path-1",
      start_url: "/login",
      selenium_script: "driver.Get('/login')",
      command_output: [
        "Successfully parsed 2 test cases",
        "Saved test script: test_scripts/test_login_success.py",
        "Saved test script: test_scripts/test_login_invalid.py",
        "Test report generated: reports/test_report.json",
      ].join("\n"),
      cases: [
        { id: "tc-happy", strategy: "happy_path", severity: "medium", description: "Happy path login succeeds" },
        { id: "tc-auth", strategy: "auth_fail", severity: "high", description: "Invalid password is rejected" },
      ],
    },
  ],
};

describe("App", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders the idle placeholder before generation", () => {
    render(<App />);

    expect(screen.getByText("Scenario2Test Platform")).toBeInTheDocument();
    expect(screen.getByText("等待编排层执行...")).toBeInTheDocument();
    expect(screen.getByText("生成测试资产")).toBeInTheDocument();
  });

  it("posts YAML to the backend and renders AUTOTEST results", async () => {
    global.fetch.mockResolvedValue({
      ok: true,
      json: async () => sampleResponse,
    });

    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: /生成测试资产/i }));

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        "/generate",
        expect.objectContaining({
          method: "POST",
          headers: { "Content-Type": "application/x-yaml" },
          body: expect.stringContaining("scenario:"),
        })
      );
    });

    expect(await screen.findByText(/结构化用例/i)).toBeInTheDocument();
    expect(screen.getByText("P0")).toBeInTheDocument();
    expect(screen.getByText("P1")).toBeInTheDocument();
    expect(screen.getByText(/auth fail :: path-1/i)).toBeInTheDocument();
    expect(screen.getByText(/生成用例：2/)).toBeInTheDocument();
    expect(screen.getAllByText(/test_scripts\/test_login_success.py/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/reports\/test_report.json/).length).toBeGreaterThan(0);
  });

  it("filters disabled strategies out of the rendered payload", async () => {
    global.fetch.mockResolvedValue({
      ok: true,
      json: async () => sampleResponse,
    });

    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: /生成测试资产/i }));
    expect(await screen.findByText(/auth fail :: path-1/i)).toBeInTheDocument();

    await user.click(screen.getByRole("checkbox", { name: "认证失败" }));

    expect(screen.queryByText(/auth fail :: path-1/i)).not.toBeInTheDocument();
    expect(screen.getByText(/happy path :: path-1/i)).toBeInTheDocument();
  });

  it("shows a backend-specific hint when the request cannot reach the server", async () => {
    global.fetch.mockRejectedValue(new TypeError("Failed to fetch"));

    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: /生成测试资产/i }));

    expect(await screen.findByText("流水线执行失败")).toBeInTheDocument();
    expect(screen.getByText(/无法连接后端服务/i)).toBeInTheDocument();
    expect(screen.getByText(/localhost:8080/i)).toBeInTheDocument();
  });
});
