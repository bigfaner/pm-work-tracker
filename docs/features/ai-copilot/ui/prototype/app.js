/* ============================================================
   AI Copilot Prototype — Shared Interactions (revised)
   - Resizable panel (drag left edge)
   - Auto-grow textarea (1000 chars)
   - Input dual mode: text (textarea) ↔ options (arrow-key group)
   - No secondary popups (diff inlined)
   ============================================================ */

const CHAR_MAX = 1000;

/* ---- Panel open/close (UF-1 <-> UF-2) ---- */
function openPanel() {
  const panel = document.getElementById('copilotPanel');
  const bubble = document.getElementById('copilotBubble');
  if (!panel || !bubble) return;
  restorePanelWidth();
  panel.classList.add('open');
  bubble.classList.add('hidden');
  const dot = bubble.querySelector('.badge-dot');
  if (dot) dot.remove();
  clearBubbleActivity();   // opening acknowledges pending activity
  focusInputOrOptions();
  const list = document.getElementById('messageList');
  if (list) list.scrollTop = list.scrollHeight;   // surface any background message
}
function closePanel() {
  const panel = document.getElementById('copilotPanel');
  const bubble = document.getElementById('copilotBubble');
  if (!panel) return;
  panel.classList.remove('open');
  if (bubble) { bubble.classList.remove('hidden'); bubble.focus(); }
  // closing does NOT interrupt the session; if a confirmation is pending, flag the bubble
  if (hasPendingConfirmation()) showBubbleActivity(1);
}
function togglePanel() {
  const panel = document.getElementById('copilotPanel');
  if (!panel) return;
  panel.classList.contains('open') ? closePanel() : openPanel();
}

/* ---- Bubble activity notification (session runs in background while closed) ---- */
function hasPendingConfirmation() {
  const om = document.getElementById('inputOptionsMode');
  return !!(om && !om.hidden);
}
function showBubbleActivity(count) {
  const bubble = document.getElementById('copilotBubble');
  if (!bubble) return;
  let badge = bubble.querySelector('.activity-badge');
  if (!badge) { badge = document.createElement('span'); badge.className = 'activity-badge'; bubble.appendChild(badge); }
  badge.textContent = count > 0 ? count : '';
  badge.classList.add('show');
}
function clearBubbleActivity() {
  const badge = document.getElementById('copilotBubble')?.querySelector('.activity-badge');
  if (badge) badge.classList.remove('show');
}
/* Demo: simulate the agent returning a message while the panel is closed. */
function simulateBackgroundReturn() {
  const panel = document.getElementById('copilotPanel');
  const list = document.getElementById('messageList');
  if (!list) return;
  const m = document.createElement('div');
  m.className = 'msg ai';
  m.innerHTML = '<div class="ts">' + nowTs() + '</div><div class="msg-bubble">已为你创建 P1 事项「完成用户认证模块」，需要你确认里程碑。</div>';
  list.appendChild(m);
  if (panel && !panel.classList.contains('open')) {
    showBubbleActivity(1);   // panel closed → notify the bubble
  } else {
    list.scrollTop = list.scrollHeight;
  }
}

/* ---- Panel resize (drag left edge) ---- */
const PANEL_MIN = 420, PANEL_MAX = 960;
function restorePanelWidth() {
  const panel = document.getElementById('copilotPanel');
  if (!panel) return;
  let w = parseInt(sessionStorage.getItem('copilotPanelW'), 10);
  if (!w || isNaN(w)) w = 630;
  w = Math.max(PANEL_MIN, Math.min(PANEL_MAX, w));
  panel.style.width = w + 'px';
}
function initResizer() {
  const resizer = document.getElementById('panelResizer');
  const panel = document.getElementById('copilotPanel');
  if (!resizer || !panel) return;
  let dragging = false, startX = 0, startW = 0;
  resizer.addEventListener('mousedown', (e) => {
    dragging = true; startX = e.clientX; startW = panel.offsetWidth;
    resizer.classList.add('active'); document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    e.preventDefault();
  });
  document.addEventListener('mousemove', (e) => {
    if (!dragging) return;
    let w = startW + (startX - e.clientX);   // drag left → wider
    w = Math.max(PANEL_MIN, Math.min(PANEL_MAX, w));
    panel.style.width = w + 'px';
  });
  document.addEventListener('mouseup', () => {
    if (!dragging) return;
    dragging = false; resizer.classList.remove('active');
    document.body.style.cursor = ''; document.body.style.userSelect = '';
    sessionStorage.setItem('copilotPanelW', panel.offsetWidth);
  });
}

/* ---- Auto-grow textarea + 1000-char count ---- */
function handleInput() {
  const ta = document.getElementById('chatInput');
  const count = document.getElementById('charCount');
  if (!ta) return;
  if (ta.value.length > CHAR_MAX) {
    ta.value = ta.value.slice(0, CHAR_MAX);
    if (count) count.classList.add('over');
    showTruncationNotice();
  } else if (count) {
    count.classList.remove('over');
  }
  if (count) count.textContent = ta.value.length + '/' + CHAR_MAX;
  // auto-grow
  ta.style.height = 'auto';
  ta.style.height = Math.min(ta.scrollHeight, parseInt(getComputedStyle(document.documentElement).getPropertyValue('--input-max-h'))) + 'px';
}
let truncationTimer = null;
function showTruncationNotice() {
  const notice = document.getElementById('truncationNotice');
  if (!notice) return;
  notice.style.display = 'flex';
  clearTimeout(truncationTimer);
  truncationTimer = setTimeout(() => { notice.style.display = 'none'; }, 3000);
}

function sendMessage() {
  const ta = document.getElementById('chatInput');
  const list = document.getElementById('messageList');
  if (!ta || !list || !ta.value.trim()) return;
  if (typeof CURRENT_SESSION_IS_NEW !== 'undefined' && CURRENT_SESSION_IS_NEW) {
    const t = ta.value.trim();
    setSessionTitle(t.slice(0, 16) + (t.length > 16 ? '…' : ''));
    CURRENT_SESSION_IS_NEW = false;        // first message makes the session non-empty
  }
  const msg = document.createElement('div');
  msg.className = 'msg user';
  msg.innerHTML = '<div class="ts">' + nowTs() + '</div><div class="msg-bubble"></div>';
  msg.querySelector('.msg-bubble').textContent = ta.value;
  list.appendChild(msg);
  ta.value = ''; handleInput();
  list.scrollTop = list.scrollHeight;
}
function nowTs() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const MM = (d.getMonth() + 1).toString().padStart(2, '0');
  const dd = d.getDate().toString().padStart(2, '0');
  const HH = d.getHours().toString().padStart(2, '0');
  const mm = d.getMinutes().toString().padStart(2, '0');
  return yyyy + '/' + MM + '/' + dd + ' ' + HH + ':' + mm;
}

/* ============================================================
   Entity Schemas + Renderer
   驱动 UF-3 表单卡片与 UF-4 结果卡片的字段渲染。
   Schema 契约见 docs/features/ai-copilot/prd/entity-schemas.md。
   - 一份 schema 同时驱动 form 与 result 两种模式
   - 渲染器只认 role，不认 entity type；新增实体加 schema 即可
   ============================================================ */

const ENTITY_SCHEMAS = {
  MainItem: {
    label: '主事项',
    fields: [
      { name: 'title',           role: 'title',    control: 'input',     required: true,  label: '标题' },
      { name: 'description',     role: 'text',     control: 'textarea',                   label: '描述' },
      { name: 'priority',        role: 'priority', control: 'select',    options: ['P1','P2','P3'], required: true, label: '优先级', derived: true },
      { name: 'status',          role: 'status',   control: 'select',    options: ['todo','in_progress','paused','completed','cancelled'], label: '状态' },
      { name: 'assignee',        role: 'assignee', control: 'select',    source: 'teamMembers', label: '负责人', derived: true },
      { name: 'planStartDate',   role: 'date',     control: 'datepicker',                  label: '计划开始' },
      { name: 'expectedEndDate', role: 'date',     control: 'datepicker', required: true,   label: '预期截止' },
      { name: 'milestoneKey',    role: 'parent',   control: 'select',    source: 'teamMilestones', required: true, label: '里程碑' },
    ],
    result_slots: {
      head: ['title', 'code'],
      fields: ['priority', 'status', 'assignee'],
      meta: [
        { field: 'expectedEndDate', icon: '📅', label: '截止' },
        { field: 'milestoneKey',    icon: '🏁', label: '里程碑' },
      ],
      progress: { source: 'subitems', tpl: (d) => `${d.percent}% · ${d.completed}/${d.total} 子任务` },
      text: ['description'],
    },
  },

  SubItem: {
    label: '子事项',
    fields: [
      { name: 'title',       role: 'title',    control: 'input',     required: true, label: '标题' },
      { name: 'parent',      role: 'parent',   control: 'select',    source: 'teamMainItems', required: true, label: '父事项' },
      { name: 'description', role: 'text',     control: 'textarea',                   label: '描述' },
      { name: 'status',      role: 'status',   control: 'select',    options: ['todo','in_progress','paused','completed','cancelled'], label: '状态' },
      { name: 'assignee',    role: 'assignee', control: 'select',    source: 'teamMembers', label: '负责人', derived: true },
      { name: 'completion',  role: 'progress', control: 'number',    min: 0, max: 100, label: '完成度' },
      { name: 'achievement', role: 'text',     control: 'textarea',                   label: '达成说明' },
    ],
    result_slots: {
      head: ['title', 'code'],
      fields: ['status', 'assignee'],
      meta: [
        { field: 'parent',          icon: '📁', label: '父事项' },
        { field: 'expectedEndDate', icon: '📅', label: '截止' },
      ],
      progress: { source: 'direct', field: 'completion', tpl: (d) => `${d.value}%` },
      text: ['achievement'],
    },
  },

  Milestone: {
    label: '里程碑',
    fields: [
      { name: 'title',           role: 'title',    control: 'input',     required: true, label: '标题' },
      { name: 'parent',          role: 'parent',   control: 'select',    source: 'teamMilestoneMaps', required: true, label: '里程碑图' },
      { name: 'status',          role: 'status',   control: 'select',    options: ['planned','in_progress','completed','cancelled'], label: '状态' },
      { name: 'expectedEndDate', role: 'date',     control: 'datepicker',                  label: '预期截止' },
      { name: 'description',     role: 'text',     control: 'textarea',                   label: '描述' },
    ],
    result_slots: {
      head: ['title', 'code'],
      fields: ['status'],
      meta: [
        { field: 'parent',          icon: '📁', label: '里程碑图' },
        { field: 'expectedEndDate', icon: '📅', label: '截止' },
      ],
      progress: { source: 'subitems_of_milestone', tpl: (d) => `${d.percent}% · ${d.completed}/${d.total} 子事项` },
      text: ['description'],
    },
  },

  MilestoneMap: {
    label: '里程碑图',
    fields: [
      { name: 'title',           role: 'title',    control: 'input',     required: true, label: '标题' },
      { name: 'team',            role: 'team',     control: 'select',    source: 'userTeams', required: true, label: 'Team' },
      { name: 'status',          role: 'status',   control: 'select',    options: ['planned','active','completed','cancelled'], label: '状态' },
      { name: 'expectedEndDate', role: 'date',     control: 'datepicker',                  label: '预期截止' },
      { name: 'description',     role: 'text',     control: 'textarea',                   label: '描述' },
    ],
    result_slots: {
      head: ['title', 'code'],
      fields: ['status', 'team'],
      meta: [
        { field: 'expectedEndDate', icon: '📅', label: '截止' },
        { field: 'milestoneCount',  icon: '🏆', label: (v) => `${v} 个里程碑` },
      ],
      progress: { source: 'milestones', tpl: (d) => `${d.percent}% · ${d.completed}/${d.total} 里程碑` },
      text: ['description'],
    },
  },

  ProgressRecord: {
    label: '进度记录',
    fields: [
      { name: 'subItem',    role: 'parent',   control: 'select',    source: 'teamSubItems', required: true, label: '子任务' },
      { name: 'completion', role: 'progress', control: 'number',    min: 0, max: 100, required: true, label: '完成度' },
      { name: 'achievement', role: 'text',    control: 'textarea',                   label: '达成说明' },
      { name: 'createdAt',  role: 'date',     control: 'readonly',                   label: '记录时间' },
    ],
    result_slots: {
      head: [{ field: 'title', value_from: 'subItem.title' }, 'code'],
      fields: [],
      meta: [{ field: 'createdAt', icon: '📅', label: '记录时间' }],
      progress: { source: 'direct', field: 'completion', tpl: (d) => `${d.value}%` },
      text: ['achievement'],
    },
  },

  ItemPool: {
    label: '待办事项池',
    fields: [
      { name: 'title',          role: 'title',     control: 'input',     required: true, label: '标题' },
      { name: 'background',     role: 'text',      control: 'textarea',  required: true, label: '背景' },
      { name: 'expectedOutput', role: 'text',      control: 'textarea',  required: true, label: '预期产出' },
      { name: 'priority',       role: 'priority',  control: 'select',    options: ['P1','P2','P3'], label: '优先级' },
      { name: 'submitter',      role: 'submitter', control: 'select',    source: 'currentUser', label: '提交人' },
      { name: 'status',         role: 'status',    control: 'select',    options: ['pending','triaged','accepted','rejected'], label: '状态' },
      { name: 'createdAt',      role: 'date',      control: 'readonly',                   label: '提交时间' },
    ],
    result_slots: {
      head: ['title', 'code'],
      fields: ['priority', 'status', 'submitter'],
      meta: [{ field: 'createdAt', icon: '📅', label: '提交时间' }],
      progress: null,
      text: ['background', 'expectedOutput'],
    },
  },
};

/* status → badge class 映射（按状态语义） */
const STATUS_BADGE = {
  todo: ['待开始', 'badge-neutral'],
  in_progress: ['进行中', 'badge-success'],
  paused: ['暂停', 'badge-warning'],
  completed: ['已完成', 'badge-success'],
  cancelled: ['已取消', 'badge-neutral'],
  planned: ['已规划', 'badge-neutral'],
  active: ['进行中', 'badge-success'],
  pending: ['待处理', 'badge-neutral'],
  triaged: ['已分诊', 'badge-warning'],
  accepted: ['已接受', 'badge-success'],
  rejected: ['已拒绝', 'badge-neutral'],
};
/* priority → badge class 映射 */
const PRIORITY_BADGE = {
  P0: ['P0', 'badge-error'],
  P1: ['P1', 'badge-warning'],
  P2: ['P2', 'badge-neutral'],
  P3: ['P3', 'badge-neutral'],
};

/* Mock 数据源（form select 选项）；真实实现由后端预加载 */
const SOURCE_MOCK = {
  teamMembers: ['张三', '李四', '王五'],
  teamMilestones: ['第一阶段（MM-0012）', '第二阶段（MM-0018）'],
  teamMainItems: ['用户认证模块（MI-0023）', '订单导出功能（MI-0024）'],
  teamMilestoneMaps: ['第一阶段（MM-0012）', '第二阶段（MM-0018）'],
  teamSubItems: ['接口联调（SI-0408）', 'API 设计（SI-0410）'],
  userTeams: ['平台组', '业务组'],
  currentUser: ['张三（你）'],
};

/* ---- 渲染入口 ---- */
function renderEntityCard(entityType, mode, data) {
  const schema = ENTITY_SCHEMAS[entityType];
  if (!schema) return '';
  if (mode === 'form') return renderFormFields(schema, data || {});
  if (mode === 'result') return renderResultCard(schema, data || {});
  return '';
}

/* ---- form 模式：渲染字段区（不含外层 chat-card 包装；调用方负责卡片头/底） ---- */
function renderFormFields(schema, data) {
  return schema.fields.map(f => renderFormField(f, data[f.name])).join('');
}
function renderFormField(field, value) {
  const requiredMark = field.required ? '<span class="req">*</span>' : '';
  const derivedTag = field.derived ? '<span class="derived-tag">AI 推断</span>' : '';
  const labelHtml = `<label>${field.label}${requiredMark}${derivedTag}</label>`;
  const cls = ['field'];
  if (field.derived) cls.push('derived');
  if (field.required) cls.push('required');
  let control = '';
  switch (field.control) {
    case 'input':
      control = `<input class="input" value="${value || ''}">`; break;
    case 'textarea':
      control = `<textarea class="textarea">${value || ''}</textarea>`; break;
    case 'datepicker':
      control = `<input type="date" class="input" value="${value || ''}">`; break;
    case 'number':
      control = `<input type="number" class="input" min="${field.min ?? 0}" max="${field.max ?? 100}" value="${value ?? ''}">`; break;
    case 'readonly':
      control = `<input class="input" value="${value || ''}" readonly>`; break;
    case 'select':
      const opts = (field.options || SOURCE_MOCK[field.source] || []).map(o =>
        `<option${o === value ? ' selected' : ''}>${o}</option>`).join('');
      control = `<select class="select">${opts}</select>`; break;
    default:
      control = `<input class="input" value="${value || ''}">`;
  }
  return `<div class="${cls.join(' ')}">${labelHtml}${control}</div>`;
}

/* ---- result 模式：渲染完整 result-card ---- */
function renderResultCard(schema, data) {
  const slots = schema.result_slots;
  const head = renderResultHead(slots.head, data);
  const fields = renderResultFields(slots.fields, data);
  const meta = renderResultMeta(slots.meta, data);
  const progress = renderResultProgress(slots.progress, data);
  const text = renderResultText(slots.text, schema.fields, data);
  const route = data._route || '';
  return `<div class="result-card"${route ? ` onclick="alert('跳转 ${route}')"` : ''}>
    ${head}${fields}${meta}${progress}${text}
  </div>`;
}
function renderResultHead(headSlots, data) {
  if (!headSlots || !headSlots.length) return '';
  const parts = headSlots.map(s => {
    if (typeof s === 'string') {
      if (s === 'title') return `<span class="r-title">${data.title || ''}</span>`;
      if (s === 'code')  return `<span class="r-code">${data.code || ''}</span>`;
      return '';
    }
    // detailed: { field, value_from }
    if (s.value_from) {
      const path = s.value_from.split('.');
      let v = data; for (const p of path) v = v?.[p];
      return `<span class="r-title">${v || ''}</span>`;
    }
    return '';
  });
  return `<div class="rc-head">${parts.join('')}<svg class="rc-chevron" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg></div>`;
}
function renderResultFields(roleList, data) {
  if (!roleList || !roleList.length) return '';
  const parts = roleList.map(role => {
    if (role === 'priority') {
      const v = data.priority;
      const [text, cls] = PRIORITY_BADGE[v] || [v, 'badge-neutral'];
      return `<span class="badge ${cls}">${text}</span>`;
    }
    if (role === 'status') {
      const v = data.status;
      const [text, cls] = STATUS_BADGE[v] || [v, 'badge-neutral'];
      return `<span class="badge ${cls}">${text}</span>`;
    }
    if (role === 'assignee' || role === 'submitter') {
      return `<span>👤 ${data[role] || ''}</span>`;
    }
    if (role === 'team') {
      return `<span>👥 ${data.team || ''}</span>`;
    }
    return '';
  }).filter(Boolean);
  if (!parts.length) return '';
  return `<div class="rc-fields">${parts.join('')}</div>`;
}
function renderResultMeta(metaList, data) {
  if (!metaList || !metaList.length) return '';
  const parts = metaList.map(m => {
    const v = data[m.field];
    if (v == null || v === '') return '';
    let label;
    if (typeof m.label === 'function') label = m.label(v);
    else label = `${m.label} ${v}`;
    return `<span class="rc-meta-item">${m.icon} ${label}</span>`;
  }).filter(Boolean);
  if (!parts.length) return '';
  return `<div class="rc-meta">${parts.join('')}</div>`;
}
function renderResultProgress(progSlot, data) {
  if (!progSlot) return '';
  let label = '';
  let percent = 0;
  if (progSlot.source === 'direct') {
    const v = data[progSlot.field];
    percent = typeof v === 'number' ? v : 0;
    label = progSlot.tpl({ value: v });
  } else {
    // computed: data._progress = { percent, completed, total }
    const p = data._progress;
    if (!p) return '';
    percent = p.percent;
    label = progSlot.tpl(p);
  }
  return `<div class="rc-progress"><div class="progress-track"><div class="progress-fill" style="width:${percent}%"></div></div><span class="progress-label">${label}</span></div>`;
}
function renderResultText(textFields, schemaFields, data) {
  if (!textFields || !textFields.length) return '';
  const labelOf = (name) => {
    const f = schemaFields.find(x => x.name === name);
    return f ? f.label : name;
  };
  const parts = textFields.map(f => {
    const v = data[f];
    if (!v) return '';
    return `<div class="rc-text"><span class="rc-text-label">${labelOf(f)}:</span> ${v}</div>`;
  }).filter(Boolean);
  if (!parts.length) return '';
  return parts.join('');
}

/* ---- Onboarding example chips ---- */
function fillExample(text) {
  const ta = document.getElementById('chatInput');
  if (!ta) return;
  setTextMode();
  ta.value = text; handleInput(); ta.focus();
  const ob = document.getElementById('onboardingCard');
  if (ob) ob.style.display = 'none';
}

/* ---- Form fold/expand (UF-3 lifecycle: 已提交/已丢弃 折叠态) ---- */
function toggleFormFold(el) {
  if (!el || !el.classList) return;
  el.classList.toggle('expanded');
  const body = el.nextElementSibling;
  if (body && body.classList.contains('form-readonly-body')) body.hidden = !body.hidden;
}

/* ---- Input dual mode: text ↔ options ---- */
let optIndex = 0, optList = [];
function setTextMode() {
  const tm = document.getElementById('inputTextMode');
  const om = document.getElementById('inputOptionsMode');
  if (tm) tm.hidden = false;
  if (om) om.hidden = true;
  const ta = document.getElementById('chatInput');
  if (ta) setTimeout(()=>ta.focus(), 0);
}
function setOptionsMode(options) {
  const tm = document.getElementById('inputTextMode');
  const om = document.getElementById('inputOptionsMode');
  if (tm) tm.hidden = true;
  if (om) om.hidden = false;
  optList = options; optIndex = 0;
  renderOptions();
}
function renderOptions() {
  const list = document.getElementById('optionList');
  if (!list) return;
  list.innerHTML = '';
  optList.forEach((o, i) => {
    const el = document.createElement('div');
    el.className = 'option' + (i === optIndex ? ' active' : '');
    el.dataset.action = o.action;
    el.innerHTML = (o.icon || '') + '<span class="opt-label"></span>' + (o.hint ? '<span class="opt-key">'+o.hint+'</span>' : '');
    el.querySelector('.opt-label').textContent = o.label;
    el.addEventListener('click', (ev) => { ev.stopPropagation(); optIndex = i; renderOptions(); triggerOption(); });
    list.appendChild(el);
  });
}
function moveOpt(delta) {
  if (!optList.length) return;
  optIndex = (optIndex + delta + optList.length) % optList.length;
  renderOptions();
  const list = document.getElementById('optionList');
  if (list) { const a = list.querySelector('.active'); if (a) a.scrollIntoView({block:'nearest'}); }
}
function triggerOption() {
  const o = optList[optIndex];
  if (!o || !o.action) return;
  handleOptionAction(o.action);
}
function handleOptionAction(action) {
  // Generic handling for the demo; real product would dispatch to backend.
  if (action === 'confirm-intent') {
    // UF-9 理解正确 → 切到 write-prefilled 状态展示 UF-3 表单卡片
    switchState('write-prefilled');
    return;
  }
  if (action === 'adjust') {
    // UF-9 我要调整 → 切回文本模式、聚焦输入框、预填用户上一条指令原文
    setTextMode();
    const list = document.getElementById('messageList');
    const lastUser = list ? list.querySelector('.msg.user .msg-bubble') : null;
    const ta = document.getElementById('chatInput');
    if (lastUser && ta) { ta.value = lastUser.textContent; handleInput(); setTimeout(() => ta.focus(), 0); }
    return;
  }
  if (action === 'apply' || action === 'discard' || action === 'cancel' || action === 'done' || action === 'retry' || action === 'manual' || action === 'edit' || action === 'submit') {
    // demo: flash a system confirmation then return to text mode
    const list = document.getElementById('messageList');
    if (list) {
      const m = document.createElement('div');
      m.className = 'msg system';
      m.innerHTML = '<div class="msg-bubble">已' + actionLabel(action) + '（原型演示）</div>';
      list.appendChild(m); list.scrollTop = list.scrollHeight;
    }
    setTextMode();
    return;
  }
  if (action && action.indexOf('pick:') === 0) {
    const idx = parseInt(action.slice(5), 10);
    const c = (CURRENT_CANDIDATES && CURRENT_CANDIDATES[idx]);
    const list = document.getElementById('messageList');
    if (list && c) {
      const m = document.createElement('div');
      m.className = 'msg system';
      m.innerHTML = '<div class="msg-bubble">已选择「' + c.title + ' · ' + c.code + '」，继续后续流程…（原型演示）</div>';
      list.appendChild(m); list.scrollTop = list.scrollHeight;
    }
    setTextMode();
    return;
  }
}
function actionLabel(a) {
  return ({apply:'应用变更', discard:'丢弃变更', cancel:'取消', done:'完成', retry:'重试', manual:'跳转手动入口', edit:'进入编辑', submit:'提交'})[a] || a;
}
function focusInputOrOptions() {
  const om = document.getElementById('inputOptionsMode');
  if (om && !om.hidden) { renderOptions(); }
  else { const ta = document.getElementById('chatInput'); if (ta) setTimeout(()=>ta.focus(),0); }
}

/* ---- Candidate selection (UF-5) — kept for click-on-card affordance ---- */
function selectCandidate(el) {
  document.querySelectorAll('.candidate').forEach(c => c.classList.remove('selected'));
  el.classList.add('selected');
}

/* ---- State switcher + per-state input mode ---- */
let CURRENT_CANDIDATES = null;
function switchState(state) {
  const list = document.getElementById('messageList');
  if (!list) return;
  document.querySelectorAll('.state-switcher button[data-state]').forEach(b => b.classList.remove('active'));
  const active = document.querySelector('.state-switcher button[data-state="' + state + '"]');
  if (active) active.classList.add('active');
  list.innerHTML = TEMPLATES[state] || '';
  list.scrollTop = 0;
  if (state === 'agent-trace') playAgentTrace();
  applyStateInput(state);
}
function applyStateInput(state) {
  const cfg = STATE_INPUT[state] || { mode: 'text' };
  CURRENT_CANDIDATES = cfg.candidates || null;
  if (cfg.mode === 'options') setOptionsMode(cfg.options);
  else {
    setTextMode();
    const ta = document.getElementById('chatInput');
    if (ta) ta.disabled = !!cfg.disabled;
    const sendBtn = document.getElementById('sendBtn');
    if (sendBtn) sendBtn.disabled = !!cfg.disabled;   // AI 处理期间禁用发送，避免一次发多条
  }
}

/* ---- Keyboard wiring ---- */
document.addEventListener('keydown', (e) => {
  const om = document.getElementById('inputOptionsMode');
  const inOptions = om && !om.hidden;
  if (inOptions) {
    if (e.key === 'ArrowDown') { e.preventDefault(); moveOpt(1); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); moveOpt(-1); }
    else if (e.key === 'Enter') { e.preventDefault(); triggerOption(); }
    else if (e.key === 'ArrowLeft') { e.preventDefault(); setTextMode(); }   // ← 返回文本输入
    else if (e.key === 'Escape') { e.preventDefault(); closePanel(); }       // Esc 仅收起面板
    return;
  }
  if (e.key === 'Escape') {
    const panel = document.getElementById('copilotPanel');
    if (panel && panel.classList.contains('open')) closePanel();
  }
});

/* ---- Agent process trace (collapse/expand) ---- */
function toggleTrace(el) {
  if (el && el.classList) el.classList.toggle('collapsed');
}
/* Stream the trace (思考 → 计划 → 操作) then auto-collapse on completion. */
let traceTimer = null;
function playAgentTrace() {
  const trace = document.getElementById('agentTrace');
  if (!trace) return;
  clearTimeout(traceTimer);
  trace.classList.remove('collapsed');
  const sections = trace.querySelectorAll('.at-section');
  const queue = [];
  sections.forEach(sec => {
    const label = sec.querySelector('.at-section-label');
    const steps = sec.querySelectorAll('.at-step');
    if (label) { label.style.display = 'none'; queue.push({el: label, isStep: false}); }
    steps.forEach(s => { s.style.display = 'none'; s.classList.remove('streaming'); queue.push({el: s, isStep: true}); });
  });
  let i = 0;
  const clearPrevStreaming = () => { for (let j = i - 1; j >= 0; j--) { if (queue[j].isStep) { queue[j].el.classList.remove('streaming'); return; } } };
  const tick = () => {
    if (i < queue.length) {
      const item = queue[i];
      item.el.style.display = '';
      if (item.isStep) { clearPrevStreaming(); item.el.classList.add('streaming'); }
      i++;
      traceTimer = setTimeout(tick, item.isStep ? 240 : 120);
    } else {
      clearPrevStreaming();                                  // last step done
      traceTimer = setTimeout(() => trace.classList.add('collapsed'), 600);  // 思考+计划完成 → 自动折叠
    }
  };
  tick();
}

/* ---- Session list / history view ----
   Single active session in the panel; history lets you switch (no parallel multi-session). */
function toggleSessionList() {
  const chat = document.getElementById('chatView');
  const list = document.getElementById('sessionListView');
  if (!chat || !list) return;
  if (list.hidden) { showSessionListView(); } else { showChatView(); }
}
function showChatView() {
  const chat = document.getElementById('chatView');
  const list = document.getElementById('sessionListView');
  if (chat) chat.hidden = false;
  if (list) list.hidden = true;
  const listBtn = document.getElementById('historyBtn');
  if (listBtn) listBtn.classList.remove('active');
  renderSessionTitleEl();
  focusInputOrOptions();
}
function showSessionListView() {
  const chat = document.getElementById('chatView');
  const list = document.getElementById('sessionListView');
  if (chat) chat.hidden = true;
  if (list) list.hidden = false;
  const listBtn = document.getElementById('historyBtn');
  if (listBtn) listBtn.classList.add('active');
  renderSessionTitleEl();
  renderSessionList();
}
function renderSessionList() {
  const wrap = document.getElementById('sessionList');
  if (!wrap) return;
  const sessions = (typeof SESSIONS !== 'undefined') ? SESSIONS : [];
  if (!sessions.length) { wrap.innerHTML = '<div class="session-empty">暂无历史会话</div>'; return; }
  wrap.innerHTML = '';
  sessions.forEach(s => {
    const el = document.createElement('div');
    el.className = 'session-item' + (s.id === CURRENT_SESSION_ID ? ' active' : '');
    el.innerHTML = '<span class="s-title"></span><span class="s-preview"></span><div class="s-meta"><span></span><span></span></div>';
    el.querySelector('.s-title').textContent = s.title;
    el.querySelector('.s-preview').textContent = s.preview;
    const meta = el.querySelectorAll('.s-meta span');
    meta[0].textContent = s.time;
    meta[1].textContent = s.messageCount + ' 条消息' + (s.pending ? ' · 待确认' : '');
    el.addEventListener('click', () => openSession(s.id));
    wrap.appendChild(el);
  });
}
function openSession(id) {
  const s = (typeof SESSIONS !== 'undefined') ? SESSIONS.find(x => x.id === id) : null;
  if (!s) return;
  CURRENT_SESSION_ID = id;
  CURRENT_SESSION_IS_NEW = false;            // history session is not "new"
  setSessionTitle(s.title);
  const list = document.getElementById('messageList');
  if (list && s.transcript) { list.innerHTML = s.transcript; list.scrollTop = list.scrollHeight; }
  applyStateInput(s.inputState || 'onboarding');
  showChatView();
}
function newSession() {
  if (CURRENT_SESSION_IS_NEW) { showChatView(); return; }   // already on a fresh empty session → no-op
  CURRENT_SESSION_ID = 's' + Date.now();
  CURRENT_SESSION_IS_NEW = true;
  setSessionTitle('新会话');
  const list = document.getElementById('messageList');
  if (list) list.innerHTML = '';
  switchState('onboarding');
  showChatView();
}
function setSessionTitle(title) {
  CURRENT_SESSION_TITLE = title;
  renderSessionTitleEl();
}
function renderSessionTitleEl() {
  const el = document.getElementById('sessionTitle');
  const lv = document.getElementById('sessionListView');
  if (!el) return;
  el.textContent = (lv && !lv.hidden) ? '历史会话' : CURRENT_SESSION_TITLE;
}

/* ---- Click outside the panel → collapse (session continues in background) ---- */
document.addEventListener('click', (e) => {
  const panel = document.getElementById('copilotPanel');
  if (!panel || !panel.classList.contains('open')) return;
  if (panel.contains(e.target)) return;
  const bubble = document.getElementById('copilotBubble');
  if (bubble && bubble.contains(e.target)) return;          // the opening click
  const switcher = document.querySelector('.state-switcher');
  if (switcher && switcher.contains(e.target)) return;       // demo affordance
  closePanel();
});

/* ---- Init ---- */
window.addEventListener('DOMContentLoaded', () => {
  initResizer();
  openPanel();
  switchState('onboarding');
  const ta = document.getElementById('chatInput');
  if (ta) {
    ta.addEventListener('input', handleInput);
    ta.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); }
    });
  }
});
