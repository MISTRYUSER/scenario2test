# Scenario2Test 代码解析报告 Wiki

更新时间：2026-03-23

## 1. 项目定位

Scenario2Test 是一个“Scenario 驱动的测试编排器”，当前版本只保留 `AUTOTEST` 通道。

它不直接把原始 YAML 扔给大模型，而是先在本地做一层确定性的工程编排：

`Scenario DSL -> Parser -> CFG -> Path Enumeration -> Strategy Engine -> AUTOTEST Adapter -> Aggregator -> JSON/API/UI`

这层编排的价值在于：

- 把业务场景转换成结构化路径
- 把测试策略和路径绑定，得到稳定的测试用例层
- 再把路径和用例分发给下游测试生成器
- 前端只展示结构化结果和产物摘要，不展示原始调试日志

## 2. 当前系统边界

当前仓库已经裁剪为单通道架构：

- 保留：`AUTOTEST`
- 删除：`Devzery UI`、`TestPilot`

因此当前输出只有两类：

- `test_cases`：结构化测试用例
- `selenium_scripts`：AUTOTEST 产物摘要与脚本草稿

## 3. 核心调用链

### 3.1 CLI / HTTP 入口

入口文件：

- [main.go](/Users/wentao.xue/Project/scenario2test/cmd/server/main.go)

核心职责：

- 解析命令行参数
- 加载配置
- 启动 HTTP 服务
- 接收 `/generate` YAML 请求
- 调用统一生成链路

生成链主流程：

1. `parser.LoadScenarioBytes`
2. `parser.BuildGraph`
3. `path.EnumerateDFS`
4. `strategy.NewDefaultEngine().Generate`
5. `e2e.NewGenerator(...).Generate`
6. `aggregator.New()...Build()`

### 3.2 聚合结果

结果定义在：

- [aggregator.go](/Users/wentao.xue/Project/scenario2test/internal/aggregator/aggregator.go)

输出字段：

- `scenario`
- `graph`
- `paths`
- `test_cases`
- `selenium_scripts`

这也是前端最终消费的数据结构。

## 4. DSL 解析层

相关文件：

- [scenario.go](/Users/wentao.xue/Project/scenario2test/internal/parser/scenario.go)
- [load.go](/Users/wentao.xue/Project/scenario2test/internal/parser/load.go)
- [graph.go](/Users/wentao.xue/Project/scenario2test/internal/parser/graph.go)

### 4.1 标准 DSL

标准格式仍然是：

```yaml
scenario:
  name: login flow
  steps:
    - action: open_page
      target: /login
```

### 4.2 宽松 DSL 兼容

当前解析器已经支持“宽松输入，规范输出”。以下变体现在都能被接收并归一化：

- `open_url` -> `open_page`
- `params.url` -> `target`
- `params.field` -> `field`
- `params.value` -> `value`
- step 级 `next`
- step 级 `branches`
- `next` 既可以引用 step `id`，也可以引用 step `action`

兼容示例：

```yaml
scenario:
  name: E-Commerce Checkout Path Discovery
  steps:
    - action: open_url
      params:
        url: https://mall.com
      next: check_auth
    - action: conditional_branch
      branches:
        - condition: auth == 'unlogged'
          next: user_login
        - condition: auth == 'logged'
          next: search_item
```

### 4.3 解析规则

解析层不会“任意猜 YAML”，但会尽量识别明确语义：

- 关键词能映射时，自动归一化
- `next` 指向无效时，尝试按 step `action` 回退匹配
- 拓扑定义一旦出现，优先采用显式 `next/branches`
- 终止节点可通过 `type: end` 标记

## 5. CFG 与路径枚举

相关文件：

- [graph.go](/Users/wentao.xue/Project/scenario2test/internal/graph/graph.go)
- [graph.go](/Users/wentao.xue/Project/scenario2test/internal/parser/graph.go)
- [path.go](/Users/wentao.xue/Project/scenario2test/internal/path/path.go)

### 5.1 图构建策略

默认情况下，步骤按顺序连边。

如果 DSL 中出现：

- `step.next`
- `step.branches`

则进入“显式拓扑模式”：

- 旧的顺序边被清空
- 只保留显式 `next/branches`
- 对缺失边再做有限兜底

### 5.2 路径枚举策略

当前使用 `DFS`：

- 每条路径输出为 `ExecutionPath`
- 路径带唯一 ID，例如 `path_01`
- 每个 step 记录：
  - `node_id`
  - `node_type`
  - `condition`
  - `payload`

### 5.3 循环保护

`EnumerateDFS` 对同一节点访问次数设置上限：

- 单节点访问超过 2 次则终止递归

这是为了避免 CFG 中存在回边时卡死。

## 6. 测试策略层

相关文件：

- [strategy.go](/Users/wentao.xue/Project/scenario2test/internal/strategy/strategy.go)

当前内置策略：

- `happy_path`
- `invalid_input`
- `auth_fail`
- `boundary`
- `rate_limit`

### 6.1 策略产物

每条策略输出标准化 `TestCase`：

- `id`
- `strategy`
- `title`
- `description`
- `path_id`
- `severity`
- `inputs`
- `expected`
- `artifacts.path_signature`

### 6.2 严重级别

当前严重级别约定：

- `P0`：高风险认证/安全问题
- `P1`：主流程、边界、非法输入
- `P2`：限流/幂等类风险

前端展示也是基于这个层，而不是直接展示 AUTOTEST 脚本。

## 7. AUTOTEST 适配层

相关文件：

- [generator.go](/Users/wentao.xue/Project/scenario2test/internal/generator/e2e/generator.go)
- [config.go](/Users/wentao.xue/Project/scenario2test/internal/config/config.go)
- [config.example.yaml](/Users/wentao.xue/Project/scenario2test/configs/config.example.yaml)

### 7.1 适配模式

支持两种模式：

- `mock`
- `cli`

默认配置是 `mock`，因此在没有外部 AUTOTEST 环境时，系统也会稳定返回结构化结果。

### 7.2 适配职责

AUTOTEST 适配器当前负责：

- 为每条路径分组测试用例
- 提取起始 URL
- 生成本地 Selenium draft
- 可选执行外部 CLI
- 汇总 `command_output`

### 7.3 URL 解析逻辑

起始 URL 的优先来源：

1. 路径中的 `open_page + target`
2. 如果是相对路径，使用 `scenario.metadata.base_url` 拼接

注意：

- 当前 `start_url` 字段保留“检测到的原始值”
- 外部 AUTOTEST CLI 实际使用的是 `resolveStartURL` 解析后的值

因此如果 `target: /` 且配置了 `base_url: https://mall.com`：

- 返回结果中的 `start_url` 可能仍显示 `/`
- 外部 CLI 会实际收到 `https://mall.com`

这是当前实现的一个可读性缺口，不影响外部执行，但会影响前端显示的一致性。

## 8. 前端展示层

相关文件：

- [App.jsx](/Users/wentao.xue/Project/scenario2test/web/src/App.jsx)
- [App.test.jsx](/Users/wentao.xue/Project/scenario2test/web/src/App.test.jsx)
- [vite.config.js](/Users/wentao.xue/Project/scenario2test/web/vite.config.js)

### 8.1 设计原则

前端当前只展示结构化信息：

- 上层：测试用例摘要
- 下层：AUTOTEST 产物摘要

不会直接展示：

- 原始命令调试日志
- 大段脚本文本
- 非结构化 stderr/stack trace

### 8.2 当前展示内容

结构化用例面板展示：

- `P0/P1/P2`
- 策略中文名
- 路径 ID
- 用例标题
- 用例描述

AUTOTEST 摘要展示：

- provider
- path_id
- start_url
- 生成用例数
- 保存脚本列表
- 测试报告路径
- 执行摘要

### 8.3 常见失败兜底

前端对以下情况会给出明确提示：

- `fetch` 失败
- 网络错误
- 后端未启动
- Vite 代理请求失败返回泛化错误

当前提示文案：

`无法连接后端服务。请确认 Go 服务已启动，并监听 http://localhost:8080。`

## 9. API 说明

OpenAPI 草案在：

- [openapi.yaml](/Users/wentao.xue/Project/scenario2test/api/openapi.yaml)

当前对外暴露：

- `GET /healthz`
- `POST /generate`

`/generate` 输入：

- `Content-Type: application/x-yaml`

输出：

- 聚合后的 JSON 结果

## 10. 已验证测试矩阵

本次交付前已实际运行并通过：

### 10.1 Go 单元测试

命令：

```bash
/Users/wentao.xue/sdk/go/bin/go test ./...
```

结果：

- `cmd/server` 通过
- `internal/parser` 通过
- `internal/path` 通过
- `internal/strategy` 通过
- `internal/generator/e2e` 通过

### 10.2 前端测试

命令：

```bash
cd /Users/wentao.xue/Project/scenario2test/web
PATH=/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin:$PATH npm test -- --run
```

结果：

- `4` 个测试全部通过

覆盖点包括：

- 空闲态渲染
- 正常调用后端并展示结构化结果
- 策略过滤
- 后端不可达时的明确错误提示

### 10.3 前端构建

命令：

```bash
cd /Users/wentao.xue/Project/scenario2test/web
PATH=/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin:$PATH npm run build
```

结果：

- 构建成功

### 10.4 全链路验证

命令：

```bash
cd /Users/wentao.xue/Project/scenario2test
./scripts/verify_all.sh 8095
```

结果：

- Go 单测通过
- 前端测试通过
- 前端构建通过
- Go 服务成功启动
- `/healthz` 成功
- `/` 成功返回前端
- `examples/login.yaml` 成功生成
- 宽松 DSL 样例成功生成并识别 `https://mall.com`

## 11. 已覆盖边界条件

当前已考虑并处理的边界条件：

- YAML 使用别名动作，如 `open_url`
- YAML 使用 `params.url` 而不是 `target`
- step 级 `branches`
- `next` 使用 step `id`
- `next` 使用 step `action`
- CFG 存在显式分支
- CFG 存在潜在回边时的 DFS 循环保护
- 外部 AUTOTEST 未配置时不阻塞主流程
- 前端在后端未启动时给出明确提示
- 前端策略开关对结果实时过滤

## 12. 已知限制

当前系统仍存在这些限制：

1. `start_url` 展示值和外部执行值不总是完全一致
   当 `target` 是相对路径时，展示层可能看到 `/`，但 CLI 实际收到的是拼接后的绝对 URL。

2. “宽松 YAML”不是无限制猜测
   当前支持的是明确语义别名和结构变体，不是任意字段名自动推断。

3. 策略引擎仍然是规则驱动
   当前策略层没有使用大模型推理，只输出规则化测试用例。

4. AUTOTEST draft 仍然是摘要级脚本
   页面展示的是结构化摘要，不是完整 Selenium 工程脚本。

## 13. 关键文件速查

### 后端入口

- [main.go](/Users/wentao.xue/Project/scenario2test/cmd/server/main.go)

### DSL 解析

- [scenario.go](/Users/wentao.xue/Project/scenario2test/internal/parser/scenario.go)
- [load.go](/Users/wentao.xue/Project/scenario2test/internal/parser/load.go)
- [graph.go](/Users/wentao.xue/Project/scenario2test/internal/parser/graph.go)

### 路径生成

- [path.go](/Users/wentao.xue/Project/scenario2test/internal/path/path.go)

### 策略引擎

- [strategy.go](/Users/wentao.xue/Project/scenario2test/internal/strategy/strategy.go)

### AUTOTEST 适配器

- [generator.go](/Users/wentao.xue/Project/scenario2test/internal/generator/e2e/generator.go)

### 前端界面

- [App.jsx](/Users/wentao.xue/Project/scenario2test/web/src/App.jsx)
- [App.test.jsx](/Users/wentao.xue/Project/scenario2test/web/src/App.test.jsx)

### 验证脚本

- [verify_all.sh](/Users/wentao.xue/Project/scenario2test/scripts/verify_all.sh)

## 14. 建议的下一步

如果继续往下做，优先级建议是：

1. 统一 `start_url` 展示值和实际执行值
2. 为更多 DSL 别名补充兼容测试
3. 给 `/generate` 结果增加更细的错误分类
4. 把 OpenAPI 文档补成完整响应 schema
5. 在前端增加“路径视图”和“CFG 视图”
