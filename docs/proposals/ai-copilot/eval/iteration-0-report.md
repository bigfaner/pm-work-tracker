---
iteration: 0
title: "Pre-Revision (Freeform Findings)"
---

# Iteration 0: Pre-Revision Report

## ATTACK_POINTS (Factual Corrections — Direct Edit)

- **[high]** 卡片与对话双向同步的创新声明与风险缓解中的单向数据流自相矛盾 | quote: "The innovation claims bidirectional real-time sync; the risk mitigation admits it should be unidirectional to avoid complexity." | improvement: resolve contradiction — commit to single-source-of-truth model (card as truth, dialog writes into card), update Innovation Highlights to accurately reflect unidirectional data flow

- **[high]** AI 响应延迟目标在 NFR 和 SC 中不一致（NFR 要求 3 秒，SC 要求 P95 < 5 秒），且未考虑完整调用链 | quote: "These targets are inconsistent -- the NFR says 3 seconds, the success criteria says P95 under 5 seconds." | improvement: unify to single target (P95 < 5s end-to-end), decompose into latency budget, update NFR section accordingly

- **[medium]** 意图分类将"assign"列为独立意图类型，但系统中分配人员是实体字段更新而非独立 API 操作 | quote: "The intent taxonomy conflates field-level updates with domain-specific operations that have their own validation rules." | improvement: clarify that "assign" maps to entity field update, restructure intent taxonomy to distinguish field updates from domain-specific operations (status transitions, progress updates)

- **[medium]** 状态机预校验需要 available-transitions 端点，但提案未确认其存在，也未计入可行性评估 | quote: "The proposal does not confirm whether such endpoints currently exist in the API surface. If they do not, the proposal has introduced an unscoped backend requirement" | improvement: verify endpoint existence in codebase; if missing, add to In Scope or acknowledge as prerequisite dependency

## BORDERLINE_FINDINGS (Structural/Architectural — Deferred to Scorer)

- **[high]** 实体解析策略未定义：模糊引用（如"认证模块的事项"）如何映射到唯一实体未说明
- **[high]** AI 服务调用边界未明确（前端直调 vs 后端代理），影响安全性和架构
- **[high]** RBAC 校验时机未定义（卡片生成时 vs 提交时），外部 AI 无法访问权限层
- **[high]** 双向编辑竞态条件：用户直接编辑卡片后立即对话修改同字段，冲突解决语义未定义
- **[medium]** Prompt 工程边界（静态 vs 动态）未定义，影响架构设计
- **[medium]** 确认操作 UX 生命周期未定义（卡片级提交/取消 vs 字段级）
- **[medium]** 错误提示不足以解释跨实体业务规则拒绝原因
- **[medium]** 准确率目标缺乏测量方法论

## SKIPPED_FINDINGS (Subjective Preference — Not Actionable)

- **[low]** 建议采用分级降级模型替代二元可用/不可用切换 — preference for richer degradation UX
- **[low]** 建议首期仅覆盖 MainItem 和 SubItem — Challenge Override: user chose full scope with reason "希望一次到位"

## Rubric

(all dimensions): N/A
