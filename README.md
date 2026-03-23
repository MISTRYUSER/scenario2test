# Scenario2Test Platform

Scenario2Test 是一个基于 `Scenario DSL` 的测试编排平台。当前版本只保留 `AUTOTEST` 通道，核心目标不是“直接把 YAML 丢给大模型”，而是先把业务场景结构化，再把结构化路径交给下游测试生成器。

核心流水线：

`Scenario DSL -> Parser -> CFG -> Path Enumeration -> Strategy Engine -> AUTOTEST Adapter -> Structured Result`

最终产物分两层：

- `test_cases`：结构化测试用例
- `selenium_scripts`：AUTOTEST 生成的脚本摘要、报告路径、执行摘要

## 当前能力

- 支持 YAML 场景输入
- 支持标准 DSL 和宽松 DSL 兼容解析
- 支持 CFG 构建和 DFS 路径枚举
- 内置 5 类测试策略：
  - `happy_path`
  - `invalid_input`
  - `auth_fail`
  - `boundary`
  - `rate_limit`
- 支持 `AUTOTEST` 的 `mock` / `cli` 两种模式
- 提供 Go HTTP API 和 React 前端
- 前端展示结构化结果，不直接暴露原始调试日志

## 项目结构

```text
scenario2test/
├── cmd/server                 # Go 服务入口
├── internal/parser            # DSL 解析与归一化
├── internal/graph             # CFG 数据结构
├── internal/path              # 路径枚举
├── internal/strategy          # 测试策略引擎
├── internal/generator/e2e     # AUTOTEST 适配器
├── internal/aggregator        # 统一结果聚合
├── api                        # OpenAPI 草案
├── configs                    # 运行配置
├── examples                   # 示例页面和示例 YAML
├── scripts                    # 本地联调和验证脚本
├── web                        # React + Vite 前端
└── docs/wiki                  # 代码解析与说明文档
```

## 工作原理

### 1. Scenario DSL

使用 YAML 描述业务流程，例如打开页面、输入、点击、分支、断言。

### 2. Parser

解析 YAML，并做归一化处理，例如：

- `open_url -> open_page`
- `params.url -> target`
- `next` / step 级 `branches` -> 图结构信息

### 3. CFG

把场景步骤和分支关系构造成控制流图。

### 4. Path Enumeration

从 CFG 中枚举所有可能执行路径，生成 `path_01`、`path_02` 等可测试对象。

### 5. Strategy Engine

对每条路径套测试策略，扩展成结构化测试用例。

公式：

`Path × Strategy = Test Cases`

### 6. AUTOTEST Adapter

把结构化路径和测试用例交给 AUTOTEST，进一步生成：

- 页面分析结果
- Selenium Python 脚本
- 测试报告

## 给组员看的逐层解释

这一节不是讲代码细节，而是讲“系统到底一层一层在做什么”。

### 1. Scenario 是什么

`Scenario` 就是“你想测试的业务流程描述”。

它不是测试代码，也不是网页本身，而是一份结构化场景说明，告诉系统：

- 用户先做什么
- 接着做什么
- 哪些地方会分支
- 最终要验证什么

例如登录场景里，`scenario` 描述的是：

- 打开登录页
- 输入用户名
- 输入密码
- 点击登录
- 可能成功，也可能失败

所以 `scenario` 本质上就是：

`测试场景的输入`

### 2. Parser 是什么

`Parser` 就是解析器。

它的作用是把 YAML 文本变成程序内部能处理的数据结构，并做归一化。

它解决两个问题：

- 读取 YAML，转成 Go 结构体
- 兼容不同写法，统一成标准模型

例如：

- `open_url -> open_page`
- `params.url -> target`
- `next -> 连边信息`
- step 里的 `branches -> 分支节点信息`

所以 `Parser` 的职责不是单纯“读取”，而是：

`把外部输入整理成标准化的内部场景模型`

### 3. CFG 是什么

`CFG` 是控制流图，`Control Flow Graph`。

它把“顺序步骤 + 分支条件”的场景表示成图结构：

- 节点：动作或状态
- 边：步骤之间的流转关系
- 条件边：满足某个条件才会走的边

例如：

```text
start -> open_home -> check_auth
                         |--- auth == unlogged ---> login -> search
                         |--- auth == logged -----> search
```

之所以要画成图，是因为真实业务流程不是一条直线，而是会分叉。

### 4. Path Enumeration 是什么

`Path Enumeration` 就是路径枚举。

有了 CFG 之后，系统知道“这是一个带分支的图”，但还不知道：

- 一共有几条可执行路径
- 每条路径经过哪些节点

所以需要从图里枚举出所有可执行路线，例如：

- `path_01: start -> open_home -> check_auth[unlogged] -> login -> search -> end`
- `path_02: start -> open_home -> check_auth[logged] -> search -> end`

这一步的意义是：

- `CFG` 是抽象流程图
- `Path Enumeration` 才是具体可测试路径

### 5. Strategy Engine 是什么

`Strategy Engine` 是测试策略引擎。

有了路径之后，系统已经知道“流程怎么走”，但还没有真正的“测试用例”。

因为同一条路径，可以从不同测试视角去测，例如：

- 正常路径
- 错误输入
- 鉴权失败
- 边界值
- 限流

所以策略引擎会对每条路径做扩展：

`Path × Strategy = Test Cases`

例如同一条登录路径，会扩展出：

- `happy_path_path_01`
- `invalid_input_path_01`
- `auth_fail_path_01`
- `boundary_path_01`
- `rate_limit_path_01`

这里已经生成测试用例了，但它们还是结构化测试设计结果，不是最终脚本。

### 6. 为什么这里已经算生成测试用例

到 `Strategy Engine` 这一层，系统已经生成了真正的 `test_cases`。

例如一条测试用例会长这样：

- `id: auth_fail_path_01`
- `strategy: auth_fail`
- `title: auth fail :: path_01`
- `severity: P0`
- `description: Unauthorized or invalid credentials are handled safely`

所以要区分两层：

- 第一层：结构化测试用例
  - 来源：路径 + 策略
  - 输出：`test_cases`
- 第二层：可执行测试脚本
  - 来源：页面 + 路径上下文 + 测试意图
  - 输出：`selenium_scripts`

### 7. 为什么还要接 AUTOTEST

因为 `Strategy Engine` 生成的只是“抽象测试用例”，还不是“可执行测试”。

例如：

- `auth_fail_path_01`
- `P0`
- `Unauthorized or invalid credentials are handled safely`

这说明了“要测什么”，但还没说明：

- 去哪个页面测
- 用什么 selector 找元素
- 输入什么值
- Selenium 脚本怎么写
- 测试报告怎么生成

所以还需要 `AUTOTEST Adapter`：

- 把路径和测试用例翻译成 AUTOTEST 可消费的输入
- 调用 AUTOTEST
- 收回脚本、报告和执行摘要

可以理解成：

- `Strategy Engine` 解决“测什么”
- `AUTOTEST` 解决“怎么执行”

### 8. 大模型在哪里被调用

大模型不在 `Parser / CFG / Path / Strategy` 这几层里调用。

这些步骤主要是规则和结构化处理。

真正调用大模型的是 `AUTOTEST`，主要发生在三步：

1. 页面分析
2. 测试用例生成
3. Selenium 脚本生成

也就是说，大模型主要负责：

- 理解页面
- 补充测试内容
- 生成执行脚本

而你自己的平台负责：

- 把业务场景结构化
- 把路径和策略组织清楚
- 再把这些上下文交给 AUTOTEST

### 9. 前端点一次“生成测试资产”后发生了什么

完整链路是：

1. 前端把 YAML 发给 `/generate`
2. 后端解析 YAML
3. 后端构建 CFG
4. 后端枚举路径
5. 后端应用测试策略
6. 后端调用 AUTOTEST Adapter
7. AUTOTEST 调大模型生成脚本和报告
8. 后端聚合结果并返回 JSON
9. 前端展示结构化测试用例和 AUTOTEST 摘要

也就是说，前端点一次按钮，后端背后实际经历的是：

`YAML -> 解析 -> 图 -> 路径 -> 测试策略 -> AUTOTEST -> 聚合 -> 返回结果`

### 10. 最终输出是什么

最终输出不是只有代码，也不是只有测试用例，而是两层结果：

- 上层：结构化测试用例
- 下层：AUTOTEST 生成的 Python Selenium 脚本与测试报告

更准确地说：

`Scenario -> 路径 -> 测试用例 -> Python 测试脚本`

所以这个项目不是“直接让模型瞎生成测试”，而是：

- 先把业务流程结构化
- 再把逻辑图化
- 然后枚举路径
- 再做策略扩展
- 最后才让 AUTOTEST 生成脚本

## 输入示例

```yaml
scenario:
  name: login flow
  metadata:
    base_url: http://127.0.0.1:8001
  steps:
    - id: open_login
      action: open_page
      target: /login.html
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
  branches:
    - from: click_login
      condition: credentials_invalid
      to: end
```

也支持更宽松的写法，例如：

- `open_url`
- `params.url`
- step 级 `next`
- step 级 `branches`

## 输出示例

后端 `/generate` 会返回统一 JSON，核心字段包括：

```json
{
  "scenario": "login flow",
  "paths": [],
  "test_cases": [],
  "selenium_scripts": []
}
```

其中：

- `test_cases` 用于前端展示 `P0 / P1 / P2` 测试用例
- `selenium_scripts` 用于展示 AUTOTEST 脚本、报告和执行摘要

## 运行方式

### 1. 只跑后端

```bash
cd /Users/wentao.xue/Project/scenario2test
/Users/wentao.xue/sdk/go/bin/go run ./cmd/server --listen :8080 --config ./configs/config.example.yaml
```

接口：

- `GET /healthz`
- `POST /generate`

### 2. 跑前端开发环境

```bash
cd /Users/wentao.xue/Project/scenario2test/web
PATH=/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin:$PATH npm run dev
```

前端默认访问：

- [http://localhost:5173/](http://localhost:5173/)

Vite 会把：

- `/generate`
- `/healthz`

代理到：

- `http://localhost:8080`

### 3. 跑本地真实 AUTOTEST 栈

如果要让前端页面看到真实脚本和报告摘要，推荐使用本地真实模式：

```bash
cd /Users/wentao.xue/Project/scenario2test
./scripts/run_local_stack.sh
```

这会：

- 构建前端
- 启动 Go 服务
- 使用 [config.local.yaml](/Users/wentao.xue/Project/scenario2test/configs/config.local.yaml)

当前本地真实模式特点：

- 真实调用 AUTOTEST CLI
- 真实生成脚本和报告
- 默认不在一次前端请求里逐个执行所有生成脚本
- 目的是让页面联调更稳定、更快返回

### 4. 跑全量集成验证

```bash
cd /Users/wentao.xue/Project/scenario2test
./scripts/run_full_integration.sh 8093
```

这会启动：

- mock LLM
- 示例静态页面
- Go 服务
- AUTOTEST CLI 联调

## 配置说明

### `config.example.yaml`

默认示例配置，适合：

- 只跑后端主链路
- 不依赖真实 AUTOTEST 外部执行

### `config.local.yaml`

本地真实 AUTOTEST 配置，适合：

- 前端联调
- 真实生成脚本和报告摘要

### `config.integration.yaml`

集成验证配置，适合：

- 跑完整本地联调脚本

## 常用脚本

- [run_local_stack.sh](/Users/wentao.xue/Project/scenario2test/scripts/run_local_stack.sh)
- [run_full_integration.sh](/Users/wentao.xue/Project/scenario2test/scripts/run_full_integration.sh)
- [verify_all.sh](/Users/wentao.xue/Project/scenario2test/scripts/verify_all.sh)
- [setup_integrations.sh](/Users/wentao.xue/Project/scenario2test/scripts/setup_integrations.sh)

## 验证

### Go 测试

```bash
cd /Users/wentao.xue/Project/scenario2test
/Users/wentao.xue/sdk/go/bin/go test ./...
```

### 前端测试

```bash
cd /Users/wentao.xue/Project/scenario2test/web
PATH=/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin:$PATH npm test -- --run
```

### 前端构建

```bash
cd /Users/wentao.xue/Project/scenario2test/web
PATH=/Users/wentao.xue/.nvm/versions/node/v24.14.0/bin:$PATH npm run build
```

### 一键验证

```bash
cd /Users/wentao.xue/Project/scenario2test
./scripts/verify_all.sh 8095
```

## 常见问题

### 前端点“生成测试资产”后提示 `Generation failed`

通常是后端没启动。

请确认：

- `http://localhost:8080/healthz` 返回 `ok`
- 前端开发环境仍在使用默认代理配置

### 为什么前端只看到结构化测试用例，没有脚本和报告

通常是因为当前后端跑的是 `mock` 配置。

如果你要看到真实 AUTOTEST 脚本和报告摘要，请使用：

- [config.local.yaml](/Users/wentao.xue/Project/scenario2test/configs/config.local.yaml)
- 或直接运行 [run_local_stack.sh](/Users/wentao.xue/Project/scenario2test/scripts/run_local_stack.sh)

### 为什么起始地址有时是相对路径，有时是绝对路径

系统会优先检测路径中的入口动作，再结合 `scenario.metadata.base_url` 解析成实际访问地址。

当前实现已经优先在结果中展示解析后的地址。

## 文档

- [代码解析报告 Wiki](/Users/wentao.xue/Project/scenario2test/docs/wiki/2026-03-23-scenario2test-code-analysis.md)
- [OpenAPI 草案](/Users/wentao.xue/Project/scenario2test/api/openapi.yaml)

## 当前限制

- 当前版本只保留 `AUTOTEST` 通道
- 策略层是规则驱动，不是 LLM 驱动
- 宽松 YAML 是“语义兼容”，不是任意字段自动推断
- 页面联调模式优先保证“生成脚本和报告摘要可见”，不是最大化执行耗时
