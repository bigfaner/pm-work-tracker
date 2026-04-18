/* ============================================
   PM Work Tracker — Prototype Shared JS
   ============================================ */

document.addEventListener('DOMContentLoaded', () => {
  initSidebar();
  initModals();
  initDropdowns();
  initTabs();
  initCollapse();
  initRadioGroups();
  initSwitches();
  initToasts();
  initActiveNav();
});

/* --- Sidebar Toggle --- */
function initSidebar() {
  const toggle = document.querySelector('.sidebar-toggle');
  const sidebar = document.querySelector('.sidebar');
  const main = document.querySelector('.main-content');
  if (!toggle || !sidebar) return;
  toggle.addEventListener('click', () => {
    sidebar.classList.toggle('collapsed');
    if (main) main.classList.toggle('sidebar-collapsed');
  });
}

/* --- Modal --- */
function initModals() {
  // Open
  document.querySelectorAll('[data-modal]').forEach(trigger => {
    trigger.addEventListener('click', (e) => {
      e.preventDefault();
      const id = trigger.getAttribute('data-modal');
      openModal(id);
    });
  });
  // Close button
  document.querySelectorAll('.modal-close').forEach(btn => {
    btn.addEventListener('click', () => {
      const overlay = btn.closest('.modal-overlay');
      if (overlay) closeModal(overlay.id);
    });
  });
  // Backdrop click
  document.querySelectorAll('.modal-overlay').forEach(overlay => {
    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) closeModal(overlay.id);
    });
  });
  // Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal-overlay.active').forEach(o => closeModal(o.id));
    }
  });
}

function openModal(id) {
  const el = document.getElementById(id);
  if (el) el.classList.add('active');
}

function closeModal(id) {
  const el = document.getElementById(id);
  if (el) el.classList.remove('active');
}

/* --- Dropdown --- */
function initDropdowns() {
  document.querySelectorAll('[data-dropdown]').forEach(trigger => {
    trigger.addEventListener('click', (e) => {
      e.stopPropagation();
      const menu = trigger.nextElementSibling;
      if (!menu) return;
      // Close others
      document.querySelectorAll('.dropdown-menu.active').forEach(m => {
        if (m !== menu) m.classList.remove('active');
      });
      menu.classList.toggle('active');
    });
  });
  document.addEventListener('click', () => {
    document.querySelectorAll('.dropdown-menu.active').forEach(m => m.classList.remove('active'));
  });
}

/* --- Tabs --- */
function initTabs() {
  document.querySelectorAll('.tabs').forEach(tabs => {
    const btns = tabs.querySelectorAll('.tab-btn');
    const panels = tabs.querySelectorAll('.tab-panel');
    btns.forEach(btn => {
      btn.addEventListener('click', () => {
        const target = btn.getAttribute('data-tab');
        btns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        panels.forEach(p => {
          p.classList.toggle('active', p.getAttribute('data-panel') === target);
        });
      });
    });
  });
}

/* --- Collapse --- */
function initCollapse() {
  document.querySelectorAll('.collapse-header').forEach(header => {
    header.addEventListener('click', () => {
      const item = header.closest('.collapse');
      if (item) item.classList.toggle('open');
    });
  });
}

/* --- Radio Groups --- */
function initRadioGroups() {
  document.querySelectorAll('.radio-group').forEach(group => {
    const btns = group.querySelectorAll('.radio-btn');
    btns.forEach(btn => {
      btn.addEventListener('click', () => {
        btns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        // Dispatch custom event
        const value = btn.getAttribute('data-value');
        group.dispatchEvent(new CustomEvent('change', { detail: { value } }));
      });
    });
  });
}

/* --- Switches --- */
function initSwitches() {
  document.querySelectorAll('.switch').forEach(sw => {
    sw.addEventListener('click', () => {
      if (sw.classList.contains('disabled')) return;
      sw.classList.toggle('active');
    });
  });
}

/* --- Toasts --- */
function initToasts() {
  if (!document.querySelector('.toast-container')) {
    const c = document.createElement('div');
    c.className = 'toast-container';
    document.body.appendChild(c);
  }
}

function showToast(message, type = 'success') {
  const container = document.querySelector('.toast-container');
  if (!container) return;
  const icons = { success: '\u2713', error: '\u2717', warning: '\u26A0' };
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `<span>${icons[type] || ''}</span><span>${message}</span>`;
  container.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transition = 'opacity 0.2s';
    setTimeout(() => toast.remove(), 200);
  }, 3000);
}

/* --- Active Nav --- */
function initActiveNav() {
  const current = location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.nav-item').forEach(item => {
    const href = item.getAttribute('href');
    if (href === current) {
      item.classList.add('active');
    }
  });
}

/* --- Confirm Dialog --- */
function showConfirm(message, onConfirm) {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay active';
  overlay.innerHTML = `
    <div class="modal modal-sm">
      <div class="modal-header">
        <h3>\u786E\u8BA4</h3>
        <button class="modal-close">&times;</button>
      </div>
      <div class="modal-body">
        <p style="display:flex;align-items:flex-start;gap:12px;">
          <span style="color:#faad14;font-size:22px;line-height:1;">&#9888;</span>
          <span>${message}</span>
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-cancel">\u53D6\u6D88</button>
        <button class="btn btn-primary btn-confirm">\u786E\u8BA4</button>
      </div>
    </div>
  `;
  document.body.appendChild(overlay);

  overlay.querySelector('.modal-close').addEventListener('click', () => overlay.remove());
  overlay.querySelector('.btn-cancel').addEventListener('click', () => overlay.remove());
  overlay.addEventListener('click', (e) => { if (e.target === overlay) overlay.remove(); });
  overlay.querySelector('.btn-confirm').addEventListener('click', () => {
    overlay.remove();
    if (onConfirm) onConfirm();
  });
}

/* --- Utility: format date --- */
function formatDate(dateStr) {
  if (!dateStr) return '-';
  const d = new Date(dateStr);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function relativeTime(dateStr) {
  if (!dateStr) return '-';
  const now = new Date();
  const d = new Date(dateStr);
  const diff = Math.floor((now - d) / 1000);
  if (diff < 60) return '\u521A\u521A';
  if (diff < 3600) return `${Math.floor(diff / 60)}\u5206\u949F\u524D`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}\u5C0F\u65F6\u524D`;
  return `${Math.floor(diff / 86400)}\u5929\u524D`;
}
