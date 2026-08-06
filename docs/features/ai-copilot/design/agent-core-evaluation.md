---
created: "2026-08-06"
updated: "2026-08-06"
parent: tech-design.md
---

# Agent 核心逻辑评估：意图切换 / 问答 agent / Path 驱动执行

> 返回 [`tech-design.md`](./tech-design.md)
>
> 本文是对现有 Planner & Executors 核心逻辑的一次评估，基准是三条新增诉求：(1) 会话中途意图切换；(2) 必有的"问答"意图；(3) Path 驱动的动态执行。评估基于 [`agent-architecture.md`](./agent-architecture.md)、[`state-machines.md`](./state-machines.md)、[`state-model.md`](./state-model.md)、[`request-model.md`](./request-model.md)、[`llm-integration.md`](./llm-integration.md)、[`api-handbook.md`](./api-handbook.md)。**本文是评估而非定稿**——三条决策已对齐，但若干开放问题仍待拍板，落地时再据此修改对应设计文档。

## 0. 决策对齐（评估基准）

经逐条对齐，三条诉求的落地方向确定为：

1. **意图切换**：新 turn（保留 superseded）+ Planner 显式承接旧 turn 的可复用字段。
2. **问答 agent**：新增独立的"问答/采访 agent"，Planner 判定不清时路由过去，做 ReAct 多次只读查询 + 多轮追问，信息够了再出意图。
3. **Path 驱动执行**：意图识别产出带 `httpMethod + path`（模板 path）的键，作为 Card ↔ API 映射；**读**由后端按此键直接执行、无须用户确认；**写**由前端按此键直调实体 API，再把结果回传后端，由后端落"已提交"消息与跟进文案。

下面逐条评估现有 agent 核心逻辑，指出差距、必须改动之处，以及其中真正的张力。

## 1. 意图切换：承接不是加一句 prompt 就够

现有设计处理"中途改主意"靠的是 Turn 状态机里的 superseded（[state-machines.md](./state-machines.md) §3 规则 #20、流程 E）。用户在 `awaiting_confirm_intent`、`awaiting_clarify`、`awaiting_commit` 这些非终态下发了新的 free_text，系统就把旧 turn 标记为 superseded，把它关联的 intent / form / disambig 卡片相应地置为 cancelled 或 discarded，然后开一个全新的 turn，从头跑一遍 Planner。这个机制本身是自洽的——"换话题"在系统看来就是"作废一轮、重开一轮"。

问题出在"承接"这一层。按决策 1，新 turn 的 Planner 应当复用旧 turn 里还能用的信息（比如已经说好的标题、优先级），但现有 Planner 的 system prompt（[agent-architecture.md](./agent-architecture.md) §3.2）里没有任何关于"承接上一未完成 turn"的规则。也就是说，承不承接全看 LLM 自己有没有"想起来"翻历史，这既不可控，也没法写测试去验证。要让承接变成确定性行为，prompt 必须新增一条显式规则：Planner 在新建 turn 时先检查是否存在刚被替代的前序 turn，若有则读取其 intent，把与新意图兼容的字段带上，只有当用户明确表示放弃时才丢弃。

但光改 prompt 还不够稳，因为承接所依赖的 History 本身是会被裁剪的。ContextBuilder 的裁剪策略（[llm-integration.md](./llm-integration.md) §5.4）是 group-aware 的：先按 intent 分组，token 超预算时从最旧的完整组开始丢。被 supersede 的旧 turn 的 intent 组，在时间序上恰好是最旧的那一组，于是在上下文紧张时它会被最先丢掉——而这正好是我们最想承接的那条信息。所以承接不能只指望 History，更稳妥的做法是模仿 pageContext（[state-model.md](./state-model.md) §4）的请求级注入：在 supersede 发生的那个事务里，把被替代 turn 的 intent 摘要显式写进新 turn 的 RequestState 或 Environment，让它绕过裁剪、确定性地出现在 Planner 的输入里。这样承接就从"LLM 自觉"变成了"系统保证"。

还有一层是现有 IntentSpec 结构表达不了的：典型例子里 SubItem 需要一个父 MainItem，而那个 MainItem 在用户改主意时根本还没落库（写操作要等 commit_card）。就算把标题承接过来，也定不了"挂在哪个主事项下"。这其实不是意图切换机制能单独解决的，它天然要落到第 2 节的问答 agent 身上——由它在采访中追问"这个子事项挂在哪里"。所以意图切换这一条，改动面其实很小（prompt + 上下文注入，不动状态机和并发），但它把一个明确的接力棒交给了问答能力。

## 2. 问答 agent：能力是现成的，缺的是一个一等公民的角色

先纠正一个前提。原始诉求里提到"read executor 没采用 ReAct、一次只调一个工具"——但从 agent 循环的设计看（[llm-integration.md](./llm-integration.md) §2），任何 agent 在一次 StreamRun 内都可以"调只读工具 → 结果回灌 → 再调"，循环上限是 8 轮。也就是说，"单次调用内多次 read"这个 ReAct 能力，机制层面是现成的，并不存在"一次只能调一个工具"的硬限制。真正的差距不在循环能力，而在角色设计：系统里没有一个一等公民的"问答/采访"角色。现在的澄清完全挂在 Planner 身上，而且是 schema 驱动的 missingInfo——只有当某个必填字段缺失时才问，一轮问完为止（[agent-architecture.md](./agent-architecture.md) §3.2、[request-model.md](./request-model.md) §6.2）。它做不了开放式采访，比如"你是想新建一个，还是改已有的？""你说的'那个'是不是上周建的认证模块？"这种需要先查一查再问的主动追问。

按决策 2，要新增一个独立的 Interviewer agent。它在结构上很轻：复用现有的 baseAgent（天然就带 ReAct loop），工具集是若干只读查询工具加上一个 `ask_question` 的 emission（调用即终止流、turn 进入等用户回答的状态）。多轮采访靠 DraftState（[state-model.md](./state-model.md) §2.2）来累积——DraftState 本来就是为 clarify 多轮回答设计的累加器，每轮用户的回答以 answer_clarify 形式进来、在新请求里重跑 Interviewer 时被重放。这条链路几乎是为这个场景准备的，契合度很高。

需要新增的是 Planner 的一个路由出口。现有 Planner 的 decision 枚举（[agent-architecture.md](./agent-architecture.md) §3.4）是 `confirm / show_candidates / cannot_understand`，都是"识别完成"导向的。要接入问答，得加一个 `needs_interview`：Planner 一旦判定当前信息不足以形成清晰意图，就返回这个 decision，Orchestrator 据此把控制权交给 Interviewer，而不是去 persist 一条半吊子的 intent。turn 状态机上相应地要有一个"等用户回答采访问题"的状态——可以新设 `awaiting_answer`，也可以直接复用 `awaiting_clarify` 的语义，区别只在于是否在审计上区分"schema 缺字段"和"开放式采访"两种问法。

这里有两个必须正视的张力。**第一是 Interviewer 与 Planner 的边界。** Interviewer 跑完、信息够了之后，是它自己直接产出 IntentSpec，还是把整理好的"澄清简报"交回 Planner 再生成意图？本文倾向前者：Interviewer 在采访过程中已经把上下文吃得最透，让它直接出意图能省掉一次 Planner 重跑；相应地 Planner 的定位可以收窄成"清晰意图的快路径"——能直接识别就直接 confirm，识别不了就甩给 Interviewer。这个分工需要拍板，因为它决定了两个 agent 各自的 prompt 和交接协议。**第二是成本。** 现有设计非常强调 LLM 调用次数的可预测（[llm-integration.md](./llm-integration.md) §7.2 甚至列了每类场景的调用数），而开放式采访天然不确定：可能一两轮就问清，也可能三四轮还得多查几次。问答 agent 一旦上线，单 turn 的 LLM 调用数会从"可数"变成"上不封顶"，这与配额管理、成本预估的设计初衷直接冲突。落地时必须给它配轮次上限和配额占用规则，否则一个说不清的用户就能把成本打爆。

## 3. Path 驱动执行：收益清晰，但写侧的安全机制要整体搬家

这是三条里和现状分歧最大、改动面也最大的一条。先看现状：IntentSpec 里只有 opType 和 entityType，没有任何 API 路径信息；写操作统一走后端的 `commit_card → Dispatcher → entity service`，这条路径上集中了幂等（`copilot_idempotency_keys`）、RBAC、孤儿防护和跟进文案生成（[api-handbook.md](./api-handbook.md) §2.3、[request-model.md](./request-model.md) §6.1、[security.md](./security.md) §7.3）；读操作则是 Reader executor 调一次 query_entities 工具、emit 一张 query_result 卡片。

Path 驱动的核心，是引入一个 **API Registry** 作为新的单一事实源。它以 `httpMethod + 模板 path`（比如 `GET /api/v1/teams/{teamId}/items`）为键，映射到每个 API 的描述符：参数 schema、是读还是写、以及怎么调用。用 method 加 path 而不是单用 path，正是因为系统是 RESTful 的，同一个 path 上 GET 和 POST 是两个完全不同的操作。这份注册表要被前后端共享：后端拿它来执行读、做提交前的校验；前端拿它来执行写。需要特别说明的是，现有的 [api-handbook.md](./api-handbook.md) 记录的是 Copilot 自身的端点（`/sessions`、`/messages` 这些），而 Path 驱动需要的是另一份"实体 API 注册表"，两者不是一回事，不能混用。

在数据模型上，IntentSpec 要新增一个 `apiRef` 字段（method + path），各 executor 的 system prompt 也从"按 opType/entityType 拼模板"改成"按 apiRef 从 registry 取描述符来拼"（[agent-architecture.md](./agent-architecture.md) §5 的动态拼装机制本身可以沿用，只是数据源换了）。

**读路径**相对简单：后端拿到意图里的 apiRef，查 registry 找到对应的读 API 并执行，结果照旧以 query_result 卡片展示给用户。这里的"无须用户确认"是指读是安全的、不用先征得用户同意就能跑，并不是说结果不再展示。这条可以替换掉 Reader executor 里那把 query_entities 工具。但要区分清楚两类 read：实体的 list/get 有自然的 method+path，走 apiRef；而像 `fuzzy_match_member`、`query_team_schema` 这类是 Copilot 内部为了把"张三"解析成 user_bizkey 而造的语义辅助工具，它们没有对应的实体 REST 端点，应当继续作为内部工具保留——第 2 节里问答 agent 的多次 read，主要靠的就是这一类语义工具加实体查询的组合。

**写路径**是真正要小心的地方，因为它把现有"AI 不直接写"的安全模型从后端搬到了前端。新的流程是：form/action 卡片带上 apiRef 和预填参数，用户在前端确认后，前端按 apiRef 直接调用实体 API（带用户自己的 JWT），调用成功后再把结果回传给 Copilot 后端，后端据此落"已提交"消息和跟进文案。这意味着 commit_card handler 的职责要从"做这次写"变成"记录前端已经做过的写"。从原则上看，"AI 不直接写库"依然成立——真正发起写的是前端、是用户的确认动作，AI 只提供了 apiRef 和参数；但这套论证需要从"后端 broker、集中把关"改写成"前端发起、后端记录"，[security.md](./security.md) §7.3 那一段得相应重述。

随之而来的是几样原本集中在后端的安全机制要重新安家，这是这条决策最大的工程风险。

**第一是幂等。** 现在 `copilot_idempotency_keys` 把守在后端 commit_card，网络重试不会重复建实体。一旦写挪到前端直调，这层把关就失效了，幂等必须上移到实体 API 本身——给实体 API 加 Idempotency-Key 头的支持。这是对既有 app API 的一次跨切面改造，不是 Copilot 模块内部能闭环的，得当成一项前置依赖来对待；如果做不到，就只能退而求其次，靠前端弱幂等加上"回传可重试"来兜底，可靠性会打折扣。

**第二是孤儿防护模型的反转。** 现在的设计是"预填不写库、用户提交才写"，所以用户中途放弃只会留下一条 `status=prefilled` 的卡片，没有任何 DB 实体副作用，这是一个很干净的不变量。改成前端直调后，实体是在前端那次调用成功时就创建了的；如果紧接着的回传失败（网络断、后端挂），就会出现"实体已存在、但 Copilot 这边没有记录"的孤儿。应对的办法是让回传本身按实体身份幂等——回传失败可以安全重试，后端收到回传时若发现实体已存在就只补记消息、不重复创建。换句话说，幂等的关注点从"别重复建实体"转移到了"别重复记账"。

**第三是提交前的预校验。** 现在 Updater 和 Mover 在 emit_form_card 之前会调 `validate_transition` / `validate_source_target` 这类只读预校验，把可能的错误预先塞进卡片（[agent-architecture.md](./agent-architecture.md) §2.2）。前端直调之后，这些预校验可以留在"展示卡片之前的后端 read"这一步（行为不变，只是读的来源可能从专门工具改成 apiRef），也可以考虑挪一部分到前端。本文倾向保留在后端、展示前做，因为校验逻辑（状态机迁移是否合法）在后端更权威，前端重复实现既容易漂移也不好测。

最后，form 卡片的状态机（[state-machines.md](./state-machines.md) §4.4）大体仍然适用，只是 `submitting → submitted` 这个迁移的触发者，从"后端 entity service 返回"变成了"前端回传结果"。Reader executor 是否保留也顺带有了答案：读走 apiRef 之后，Reader 可以瘦成一个"执行 apiRef、emit query_result"的薄壳，纯查询意图仍然需要它，不必并入问答 agent。

## 4. 三条之间的耦合

这三条不能各改各的，它们在几个点上咬合在一起。意图切换里那个"SubItem 找不到父 MainItem"的关系推理问题，本身不是切换机制能解决的，它正好是问答 agent 要追问的内容——这是第一条把接力棒交给第二条。问答 agent 要做多次只读查询，而只读查询的执行方式由 Path 驱动决定（实体类走 apiRef、语义类走内部工具）——这是第二条依赖第三条。再者，意图切换的"承接"和问答的"采访"都需要把额外上下文稳定地喂给 agent，这两件事可以共用同一套请求级上下文注入机制（复用 [state-model.md](./state-model.md) §4 的 pageContext 模式），没必要各搞一套。把它们当成一组耦合的改动来规划，比逐条独立改要省事得多。

## 5. 待拍板的开放问题

在动手改设计文档之前，下列各点需要定调。本文给出的倾向仅作参考。

| # | 开放问题 | 倾向 |
|---|---------|------|
| 1 | Interviewer 跑完是直接出 IntentSpec，还是交回 Planner？ | 直接出。它上下文最全，能少一次 Planner 重跑；Planner 收窄为"清晰意图快路径"。 |
| 2 | API registry 放哪、前后端怎么共享？ | 后端 Go registry 为单一事实源，生成前端 client map（或 OpenAPI 片段），避免两边漂移。 |
| 3 | 幂等怎么落地？ | 实体 API 加 Idempotency-Key 头（最稳）；做不到再退到前端弱幂等 + 回传重试。 |
| 4 | 预校验（validate_transition 等）放哪？ | 留后端、展示卡片前做；errors 进卡片但不阻塞流，与现状对齐。 |
| 5 | Reader executor 保留吗？ | 保留，瘦成"执行 apiRef"薄壳。 |

## 6. 后续

本文为评估，落地需修改的设计文档（待开放问题拍板后进行）：

- [agent-architecture.md](./agent-architecture.md)：新增 Interviewer 角色 + `needs_interview` decision；IntentSpec 加 `apiRef`；executor prompt 改为按 apiRef 拼装；承接规则进 Planner prompt。
- [state-machines.md](./state-machines.md)：新增 `awaiting_answer`（或复用 `awaiting_clarify`）；form 卡片 `submitting → submitted` 触发者改为前端回传。
- [request-model.md](./request-model.md)：commit_card 从"做写"改为"记录前端已做的写"；新增写结果回传请求类型。
- [api-handbook.md](./api-handbook.md)：新增实体 API 注册表（method+path 为键）；commit_card 响应语义调整。
- [security.md](./security.md)：§7.3 "AI 不直接写"论证从"后端 broker"改写为"前端发起、后端记录"；幂等/孤儿防护新模型。
- [state-model.md](./state-model.md)：承接所需的请求级上下文注入（复用 pageContext 模式）。
