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
  m.innerHTML = '<div class="msg-bubble">已为你创建 P1 事项「完成用户认证模块」，需要你确认里程碑。</div><div class="ts">' + nowTs() + '</div>';
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
  msg.innerHTML = '<div class="msg-bubble"></div><div class="ts">' + nowTs() + '</div>';
  msg.querySelector('.msg-bubble').textContent = ta.value;
  list.appendChild(msg);
  ta.value = ''; handleInput();
  list.scrollTop = list.scrollHeight;
}
function nowTs() {
  const d = new Date();
  return d.getHours().toString().padStart(2,'0') + ':' + d.getMinutes().toString().padStart(2,'0');
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

/* ---- Undo countdown (UF-3 success) ---- */
let undoTimer = null, undoRemaining = 0;
function startUndoCountdown(seconds) {
  undoRemaining = seconds;
  updateUndoLabel();
  clearInterval(undoTimer);
  undoTimer = setInterval(() => {
    undoRemaining--;
    if (undoRemaining <= 0) { clearInterval(undoTimer); expireUndo(); refreshOptionsForState('success-undo'); }
    else { updateUndoLabel(); }
  }, 1000);
}
function updateUndoLabel() {
  const mm = Math.floor(undoRemaining/60).toString().padStart(2,'0');
  const ss = (undoRemaining%60).toString().padStart(2,'0');
  const opt = document.querySelector('.option[data-action="undo"] .opt-label');
  if (opt) opt.textContent = '↩ 撤回 (' + mm + ':' + ss + ')';
  const btn = document.getElementById('undoBtn');
  if (btn) btn.querySelector('.undo-countdown').textContent = '撤回 (' + mm + ':' + ss + ')';
}
function expireUndo() {
  const slot = document.getElementById('undoSlot');
  if (slot) slot.innerHTML = '<div class="success-row" style="color:var(--text-tertiary)"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>撤回窗口已过期</div>';
}
function performUndo() {
  clearInterval(undoTimer);
  const card = document.getElementById('writeCardSuccess');
  if (card) {
    const footer = card.querySelector('.card-footer');
    if (footer) footer.innerHTML = '<div class="success-row"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 1 0 9-9"/><path d="M3 4v5h5"/></svg>已撤回</div>';
  }
  setTextMode();
}

/* ---- Leave-confirm dialog (route guard only) ---- */
function showLeaveDialog() { const dlg = document.getElementById('leaveDialog'); if (dlg) dlg.classList.add('open'); const c = dlg && dlg.querySelector('[data-action="cancel"]'); if (c) setTimeout(()=>c.focus(),50); }
function closeLeaveDialog() { const dlg = document.getElementById('leaveDialog'); if (dlg) dlg.classList.remove('open'); }

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
  if (action === 'undo') { performUndo(); return; }
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
  if (state === 'success-undo') startUndoCountdown(300);
  if (state === 'agent-trace') playAgentTrace();
  applyStateInput(state);
}
function applyStateInput(state) {
  const cfg = STATE_INPUT[state] || { mode: 'text' };
  CURRENT_CANDIDATES = cfg.candidates || null;
  if (cfg.mode === 'options') setOptionsMode(cfg.options);
  else { setTextMode(); const ta = document.getElementById('chatInput'); if (ta && cfg.disabled) ta.disabled = true; else if (ta) ta.disabled = false; }
}

/* ---- Keyboard wiring ---- */
document.addEventListener('keydown', (e) => {
  const dlg = document.getElementById('leaveDialog');
  if (dlg && dlg.classList.contains('open')) {
    if (e.key === 'Escape') { closeLeaveDialog(); }
    return;
  }
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
  const dlg = document.getElementById('leaveDialog');
  if (dlg && dlg.contains(e.target)) return;
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
