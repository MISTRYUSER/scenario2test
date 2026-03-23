import { useMemo, useState } from "react";
import {
  CheckCircle2,
  ChevronRight,
  Code,
  GitMerge,
  Layers,
  Play,
  Radar,
  Settings,
  ShieldAlert,
  Terminal,
} from "lucide-react";

const defaultYaml = `scenario:
  name: login flow
  description: Validate login behavior for valid and invalid user credentials.
  metadata:
    base_url: http://127.0.0.1:8001
  steps:
    - id: open_login
      action: open_page
      target: /login.html
      expected: login page is visible
    - id: input_username
      action: input
      field: username
      value: valid_user
    - id: input_password
      action: input
      field: password
      value: invalid_password
    - id: click_login
      action: click
      target: login_button
      expected: authentication result is returned
  branches:
    - from: click_login
      condition: credentials_invalid
      to: end
  assertions:
    - type: response
      target: status_code
      value: "401"`;

const initialStrategies = {
  happyPath: true,
  boundary: true,
  authFail: true,
  invalidInput: true,
  rateLimit: true,
};

export default function App() {
  const [yamlCode, setYamlCode] = useState(defaultYaml);
  const [isGenerating, setIsGenerating] = useState(false);
  const [results, setResults] = useState(null);
  const [error, setError] = useState("");
  const [strategies, setStrategies] = useState(initialStrategies);

  const rendered = useMemo(() => {
    if (!results) {
      return {
        cases: [],
        suites: [],
        stats: [
          { label: "路径数", value: "--" },
          { label: "策略数", value: String(enabledCount(strategies)) },
          { label: "用例数", value: "--" },
          { label: "适配器", value: "1" },
        ],
      };
    }

    const filteredCases = filterCases(results.test_cases ?? [], strategies);
    const suites = filterSuiteCases(results.selenium_scripts ?? [], filteredCases);
    const stats = [
      { label: "路径数", value: String(results.paths?.length ?? 0) },
      { label: "策略数", value: String(enabledCount(strategies)) },
      { label: "用例数", value: String(filteredCases.length) },
      { label: "适配器", value: "1" },
    ];

    return { cases: filteredCases, stats, suites };
  }, [results, strategies]);

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError("");
    setResults(null);

    try {
      const response = await fetch("/generate", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-yaml",
        },
        body: yamlCode,
      });

      if (!response.ok) {
        const message = await response.text();
        throw new Error(message || "Generation failed");
      }

      const payload = await response.json();
      payload.test_cases = filterCases(payload.test_cases ?? [], strategies);
      setResults(payload);
    } catch (err) {
      setError(formatError(err));
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(240,90,40,0.16),_transparent_28%),linear-gradient(160deg,#f5f1e8_0%,#f8fafc_52%,#e9f0f6_100%)] px-4 py-6 text-stone-900 md:px-8 lg:px-10">
      <div className="mx-auto max-w-7xl">
        <header className="mb-6 overflow-hidden rounded-[28px] border border-white/70 bg-white/75 p-5 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-7">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-start gap-4">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-orange-500 via-red-500 to-amber-500 text-white shadow-lg shadow-orange-500/20">
                <GitMerge className="h-7 w-7" />
              </div>
              <div>
                <p className="mb-2 inline-flex items-center gap-2 rounded-full border border-orange-200 bg-orange-50 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.24em] text-orange-700">
                  Scenario 路径编排器
                </p>
                <h1 className="text-3xl font-semibold tracking-tight text-stone-950 md:text-4xl">
                  Scenario2Test Platform
                </h1>
                <p className="mt-2 max-w-2xl text-sm leading-6 text-stone-600 md:text-base">
                  当前版本聚焦 AUTOTEST 生成链路：系统将 Scenario DSL 解析为 CFG，完成路径枚举后，生成 AUTOTEST
                  端到端测试资产。
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {rendered.stats.map((item) => (
                <div key={item.label} className="rounded-2xl border border-stone-200/80 bg-stone-50/90 px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.2em] text-stone-500">{item.label}</p>
                  <p className="mt-1 text-2xl font-semibold text-stone-900">{item.value}</p>
                </div>
              ))}
            </div>
          </div>
        </header>

        <main className="grid grid-cols-1 gap-6 lg:grid-cols-12">
          <section className="lg:col-span-4">
            <Panel
              title="Scenario DSL 输入"
              subtitle="在这里编辑送入编排层的 YAML scenario。"
              badge="YAML"
              icon={<FileIcon className="h-4 w-4 text-orange-600" />}
            >
              <div className="overflow-hidden rounded-3xl border border-slate-900/90 bg-[#0b1020] shadow-[inset_0_1px_0_rgba(255,255,255,0.03)]">
                <div className="flex items-center justify-between border-b border-slate-800 bg-[#12182a] px-4 py-3 text-xs text-slate-400">
                  <span>scenario.yaml</span>
                  <span>{yamlCode.split("\n").length} lines</span>
                </div>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 flex w-12 flex-col items-center border-r border-slate-800 bg-[#12182a] py-4 font-mono text-xs text-slate-500">
                    {yamlCode.split("\n").map((_, index) => (
                      <span key={index}>{index + 1}</span>
                    ))}
                  </div>
                  <textarea
                    value={yamlCode}
                    onChange={(event) => setYamlCode(event.target.value)}
                    className="min-h-[560px] w-full resize-none bg-transparent px-4 py-4 pl-16 font-mono text-sm leading-6 text-slate-200 outline-none"
                    spellCheck="false"
                  />
                </div>
              </div>
            </Panel>
          </section>

          <section className="lg:col-span-3">
            <Panel
              title="流水线引擎"
              subtitle="当前仅生成 AUTOTEST 端到端结果。"
              icon={<Settings className="h-4 w-4 text-orange-600" />}
            >
              <div className="space-y-6">
                <div className="rounded-3xl border border-orange-100 bg-gradient-to-br from-orange-50 via-white to-amber-50 p-4">
                  <h3 className="mb-4 text-xs font-semibold uppercase tracking-[0.24em] text-stone-500">
                    执行流程
                  </h3>
                  <div className="relative space-y-3">
                    <div className="absolute bottom-2 left-4 top-4 w-px bg-gradient-to-b from-orange-300 via-stone-200 to-transparent" />
                    <PipelineStep icon={<Code size={16} />} title="1. 解析 YAML DSL" active />
                    <PipelineStep icon={<GitMerge size={16} />} title="2. 构建 CFG" active />
                    <PipelineStep icon={<Layers size={16} />} title="3. 枚举路径" active />
                    <PipelineStep icon={<ShieldAlert size={16} />} title="4. 应用测试策略" active />
                    <PipelineStep icon={<Radar size={16} />} title="5. 输出 AUTOTEST" active />
                  </div>
                </div>

                <ControlBlock title="用例覆盖维度">
                  <Checkbox
                    label="正常路径"
                    checked={strategies.happyPath}
                    onChange={() => toggleFlag(setStrategies, "happyPath")}
                  />
                  <Checkbox
                    label="非法输入"
                    checked={strategies.invalidInput}
                    onChange={() => toggleFlag(setStrategies, "invalidInput")}
                  />
                  <Checkbox
                    label="边界值"
                    checked={strategies.boundary}
                    onChange={() => toggleFlag(setStrategies, "boundary")}
                  />
                  <Checkbox
                    label="认证失败"
                    checked={strategies.authFail}
                    onChange={() => toggleFlag(setStrategies, "authFail")}
                  />
                  <Checkbox
                    label="限流"
                    checked={strategies.rateLimit}
                    onChange={() => toggleFlag(setStrategies, "rateLimit")}
                  />
                </ControlBlock>

                <ControlBlock title="目标适配器">
                  <div className="rounded-2xl border border-orange-200 bg-orange-50/80 px-3 py-3 text-sm text-orange-800">
                    当前仅启用 AUTOTEST 端到端测试适配器。
                  </div>
                </ControlBlock>

                <button
                  onClick={handleGenerate}
                  disabled={isGenerating}
                  className={`flex w-full items-center justify-center gap-2 rounded-2xl px-4 py-3 text-sm font-semibold text-white shadow-lg transition ${
                    isGenerating
                      ? "cursor-not-allowed bg-stone-400 shadow-none"
                      : "bg-gradient-to-r from-orange-600 via-red-500 to-amber-500 shadow-orange-500/25 hover:translate-y-[-1px]"
                  }`}
                >
                  {isGenerating ? <Spinner /> : <Play className="h-4 w-4" />}
                  {isGenerating ? "流水线执行中..." : "生成测试资产"}
                </button>

                <p className="text-xs leading-5 text-stone-500">
                  页面当前仅展示 AUTOTEST 输出；上方开关用于按覆盖维度筛选生成的测试用例。
                </p>
              </div>
            </Panel>
          </section>

          <section className="lg:col-span-5">
            <Panel
              title="AUTOTEST 输出"
              subtitle="这里展示结构化测试用例和 AUTOTEST 产物摘要。"
              icon={<Terminal className="h-4 w-4 text-orange-600" />}
            >
              <div className="mb-4 flex flex-wrap gap-2">
                <div className="inline-flex items-center rounded-2xl border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-medium text-orange-700">
                  <Terminal className="mr-2 h-4 w-4" />
                  AUTOTEST
                </div>
              </div>

              <div className="overflow-hidden rounded-3xl border border-slate-900/90 bg-[#0b1020]">
                <div className="flex items-center justify-between border-b border-slate-800 bg-[#12182a] px-4 py-3">
                  <div className="flex items-center gap-2 text-xs uppercase tracking-[0.22em] text-slate-400">
                    <span>适配器输出</span>
                    <ChevronRight className="h-3 w-3" />
                    <span>AUTOTEST 端到端</span>
                  </div>
                  <span className="rounded-full border border-slate-700 px-2 py-1 text-[11px] text-slate-400">
                    {results ? "后端实时响应" : "空闲"}
                  </span>
                </div>

                <div className="relative min-h-[560px] p-4">
                  {!results && !isGenerating && !error && <Placeholder />}

                  {isGenerating && <LoadingState />}

                  {error && !isGenerating && (
                    <div className="rounded-2xl border border-red-500/30 bg-red-500/10 p-4 text-sm text-red-200">
                      <p className="font-semibold">流水线执行失败</p>
                      <p className="mt-2 whitespace-pre-wrap text-red-100/80">{error}</p>
                    </div>
                  )}

                  {results && !isGenerating && !error && (
                    <div className="space-y-4">
                      <TestCaseSummary testCases={rendered.cases} />
                      <SuiteSummary suites={rendered.suites} />
                    </div>
                  )}
                </div>
              </div>
            </Panel>
          </section>
        </main>
      </div>
    </div>
  );
}

function extractSavedScripts(commandOutput) {
  return (commandOutput ?? "")
    .split("\n")
    .filter((line) => line.includes("Saved test script:"))
    .map((line) => line.split("Saved test script:")[1]?.trim())
    .filter(Boolean);
}

function summarizeCommandOutput(commandOutput) {
  return (commandOutput ?? "")
    .split("\n")
    .filter((line) => {
      const trimmed = line.trim();
      return (
        trimmed.includes("Successfully parsed") ||
        trimmed.includes("Saved test script:") ||
        trimmed.includes("Test report generated:")
      );
    });
}

function extractGeneratedReport(commandOutput) {
  return (commandOutput ?? "")
    .split("\n")
    .find((line) => line.includes("Test report generated:"))
    ?.split("Test report generated:")[1]
    ?.trim();
}

function filterCases(testCases, strategies) {
  const enabled = new Set(
    Object.entries(strategies)
      .filter(([, value]) => value)
      .map(([key]) => strategyKeyToBackend(key))
  );

  return testCases.filter((testCase) => enabled.has(testCase.strategy));
}

function strategyKeyToBackend(key) {
  const map = {
    happyPath: "happy_path",
    invalidInput: "invalid_input",
    authFail: "auth_fail",
    boundary: "boundary",
    rateLimit: "rate_limit",
  };
  return map[key];
}

function enabledCount(flags) {
  return Object.values(flags).filter(Boolean).length;
}

function toggleFlag(setter, key) {
  setter((current) => ({ ...current, [key]: !current[key] }));
}

function filterSuiteCases(suites, filteredCases) {
  const allowedIds = new Set(filteredCases.map((item) => item.id));
  return suites.map((suite) => ({
    ...suite,
    cases: (suite.cases ?? []).filter((item) => allowedIds.has(item.id)),
  }));
}

function normalizePriority(severity) {
  const value = String(severity ?? "").toUpperCase();
  if (value.startsWith("P")) {
    return value;
  }
  if (value === "HIGH") return "P0";
  if (value === "MEDIUM") return "P1";
  if (value === "LOW") return "P2";
  return "P1";
}

function strategyLabel(strategy) {
  const map = {
    happy_path: "正常路径",
    invalid_input: "非法输入",
    auth_fail: "认证失败",
    boundary: "边界值",
    rate_limit: "限流",
  };
  return map[strategy] ?? strategy;
}

function formatError(error) {
  const message = error instanceof Error ? error.message : "Unknown error";
  const normalized = message.trim().toLowerCase();

  if (
    normalized.includes("failed to fetch") ||
    normalized.includes("networkerror") ||
    normalized.includes("load failed") ||
    normalized === "generation failed"
  ) {
    return "无法连接后端服务。请确认 Go 服务已启动，并监听 http://localhost:8080。";
  }

  return message || "Unknown error";
}

function TestCaseSummary({ testCases }) {
  if (!testCases?.length) {
    return (
      <div className="rounded-2xl border border-stone-700/60 bg-slate-950/50 p-4 text-sm text-slate-400">
        当前没有可展示的测试用例。
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 text-xs uppercase tracking-[0.22em] text-orange-300">
        <span>结构化用例</span>
        <ChevronRight className="h-3 w-3" />
        <span>{testCases.length} 条</span>
      </div>
      <div className="grid gap-3">
        {testCases.map((testCase) => (
          <div key={testCase.id} className="rounded-2xl border border-orange-500/20 bg-orange-500/5 p-4 text-sm text-slate-100">
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full border border-red-400/20 bg-red-500/10 px-2 py-1 text-xs font-semibold text-red-200">
                {normalizePriority(testCase.severity)}
              </span>
              <span className="rounded-full border border-slate-700 bg-slate-900/70 px-2 py-1 text-xs text-slate-300">
                {strategyLabel(testCase.strategy)}
              </span>
              <span className="rounded-full border border-slate-700 bg-slate-900/70 px-2 py-1 text-xs text-slate-400">
                {testCase.path_id || testCase.pathId}
              </span>
            </div>
            <p className="mt-3 font-medium text-orange-100">{testCase.title || testCase.id}</p>
            <p className="mt-2 text-sm leading-6 text-slate-300">{testCase.description || "暂无描述"}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function SuiteSummary({ suites }) {
  if (!suites?.length) {
    return null;
  }

  return (
    <div className="space-y-3">
      {suites.map((suite) => {
        const savedScripts = extractSavedScripts(suite.command_output);
        const summaryLines = summarizeCommandOutput(suite.command_output);

        return (
          <div key={suite.path_id} className="rounded-2xl border border-emerald-500/20 bg-emerald-500/5 p-4 text-sm text-slate-200">
            <div className="flex flex-wrap items-center gap-2 text-xs uppercase tracking-[0.2em] text-emerald-300">
              <span>{suite.provider}</span>
              <span>/</span>
              <span>{suite.path_id}</span>
            </div>
            <p className="mt-2 text-sm text-slate-300">起始地址：{suite.start_url || "n/a"}</p>
            <p className="mt-2 font-medium text-emerald-200">生成用例：{suite.cases?.length ?? 0}</p>
            <div className="mt-3">
              <p className="font-medium text-emerald-200">保存脚本</p>
              {savedScripts.length ? (
                <ul className="mt-2 space-y-1 text-xs text-slate-300">
                  {savedScripts.map((scriptPath) => (
                    <li key={scriptPath}>{scriptPath}</li>
                  ))}
                </ul>
              ) : (
                <p className="mt-2 text-xs text-slate-400">当前没有脚本保存记录。</p>
              )}
            </div>
            <div className="mt-3">
              <p className="font-medium text-emerald-200">测试报告</p>
              {extractGeneratedReport(suite.command_output) ? (
                <p className="mt-2 text-xs text-slate-300">{extractGeneratedReport(suite.command_output)}</p>
              ) : (
                <p className="mt-2 text-xs text-slate-400">当前没有测试报告记录。</p>
              )}
            </div>
            <div className="mt-3">
              <p className="font-medium text-emerald-200">执行摘要</p>
              {summaryLines.length ? (
                <ul className="mt-2 space-y-1 text-xs text-slate-300">
                  {summaryLines.map((line) => (
                    <li key={line}>{line}</li>
                  ))}
                </ul>
              ) : (
                <p className="mt-2 text-xs text-slate-400">当前没有可提取的执行摘要。</p>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

function Panel({ title, subtitle, badge, icon, children }) {
  return (
    <div className="rounded-[28px] border border-white/70 bg-white/75 p-5 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-6">
      <div className="mb-4 flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-orange-50">{icon}</span>
            <h2 className="text-lg font-semibold text-stone-900">{title}</h2>
          </div>
          <p className="mt-2 text-sm leading-6 text-stone-600">{subtitle}</p>
        </div>
        {badge ? (
          <span className="rounded-full border border-stone-200 bg-stone-50 px-2 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-stone-500">
            {badge}
          </span>
        ) : null}
      </div>
      {children}
    </div>
  );
}

function ControlBlock({ title, children }) {
  return (
    <div>
      <h3 className="mb-3 text-xs font-semibold uppercase tracking-[0.24em] text-stone-500">{title}</h3>
      <div className="space-y-2">{children}</div>
    </div>
  );
}

function FileIcon({ className }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
      />
    </svg>
  );
}

function PipelineStep({ icon, title, active }) {
  return (
    <div className="relative z-10 flex items-center bg-transparent py-1">
      <div
        className={`flex h-8 w-8 items-center justify-center rounded-full border-2 ${
          active
            ? "border-orange-500 bg-orange-50 text-orange-600"
            : "border-stone-300 bg-white text-stone-400"
        }`}
      >
        {icon}
      </div>
      <span className={`ml-3 text-sm ${active ? "font-medium text-stone-800" : "text-stone-500"}`}>{title}</span>
    </div>
  );
}

function Checkbox({ label, checked, onChange }) {
  return (
    <label className="flex cursor-pointer items-center rounded-2xl border border-stone-200 bg-stone-50/60 px-3 py-2 transition hover:border-orange-200 hover:bg-orange-50/50">
      <input checked={checked} onChange={onChange} type="checkbox" className="sr-only" />
      <div
        className={`mr-3 flex h-5 w-5 items-center justify-center rounded-md border ${
          checked ? "border-orange-600 bg-orange-600 text-white" : "border-stone-300 bg-white text-transparent"
        }`}
      >
        <CheckCircle2 className="h-3.5 w-3.5" />
      </div>
      <span className="text-sm text-stone-700">{label}</span>
    </label>
  );
}

function Placeholder() {
  return (
    <div className="absolute inset-0 flex flex-col items-center justify-center text-slate-500">
      <Terminal className="mb-4 h-12 w-12 opacity-20" />
      <p className="text-sm">等待编排层执行...</p>
      <p className="mt-1 text-xs opacity-60">点击“生成测试资产”后会把 YAML 发送到 `/generate`。</p>
    </div>
  );
}

function LoadingState() {
  return (
    <div className="absolute inset-0 flex flex-col items-center justify-center text-indigo-300">
      <div className="w-full max-w-md space-y-3 px-6">
        <div className="h-2 w-3/4 animate-pulse rounded bg-slate-800" />
        <div className="h-2 w-full animate-pulse rounded bg-slate-800" />
        <div className="h-2 w-5/6 animate-pulse rounded bg-slate-800" />
        <div className="h-2 w-1/2 animate-pulse rounded bg-slate-800" />
        <p className="mt-4 text-center font-mono text-xs text-slate-500">
          &gt; 正在解析 YAML ... 构建 CFG ... 枚举路径 ...
        </p>
      </div>
    </div>
  );
}

function Spinner() {
  return (
    <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="10" className="opacity-20" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-90" fill="currentColor" d="M12 2a10 10 0 0 1 10 10h-4a6 6 0 0 0-6-6V2Z" />
    </svg>
  );
}
