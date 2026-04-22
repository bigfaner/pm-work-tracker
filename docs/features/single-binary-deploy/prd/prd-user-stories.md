---
feature: single-binary-deploy
---

# User Stories: single-binary-deploy

## Story 1: 一键构建可部署二进制

**As a** 开发者
**I want to** 通过一条命令构建出包含前后端的单体二进制文件
**So that** 我可以快速将应用部署到目标环境进行自测

**Acceptance Criteria:**
- Given 我在正确的 git 分支上
- When 我执行构建脚本并指定环境参数
- Then 脚本自动完成前端构建和后端构建，输出一个可运行的二进制文件

---

## Story 2: 二进制独立运行无需额外依赖

**As a** 开发者
**I want to** 在目标机器上只运行一个二进制文件就能使用完整应用
**So that** 我不需要在部署目标上安装 Web 服务器或前端运行时环境

**Acceptance Criteria:**
- Given 我已将构建好的二进制文件和配置文件放到目标机器
- When 我启动二进制并指定配置文件
- Then 应用同时提供 API 接口和前端页面，浏览器访问根路径可以看到完整的应用界面

---

## Story 3: 前端路由直接访问正常工作

**As a** 开发者
**I want to** 在浏览器中直接访问任意前端路由路径（如 /teams/1/items）
**So that** 我可以通过书签或分享链接直达特定页面，而不会遇到 404 错误

**Acceptance Criteria:**
- Given 应用已启动
- When 我在浏览器地址栏直接输入前端路由路径
- Then 页面正常加载并渲染对应的路由内容，而非显示 404 错误

---

## Story 4: 启动时发现配置错误并给出明确提示

**As a** 开发者
**I want to** 在启动应用时，如果配置有问题能立即看到明确的错误信息
**So that** 我不需要在运行中排查因配置错误导致的问题

**Acceptance Criteria:**
- Given 配置文件中必填字段仍为占位符
- When 我启动应用
- Then 应用拒绝启动，退出并打印包含具体字段名称的错误信息

---

## Story 5: 本地开发流程不受影响

**As a** 开发者
**I want to** 在本地开发时继续使用现有的开发流程
**So that** 构建能力的引入不会降低日常开发效率

**Acceptance Criteria:**
- Given 我正在进行日常功能开发
- When 我启动本地开发环境
- Then 现有的前后端开发服务器正常启动，所有现有测试通过，无需执行构建脚本
