// ============================================
// Shared Interactions — RBAC Prototype
// ============================================

document.addEventListener('DOMContentLoaded', () => {
  initDialogs();
  initCollapsibles();
  initMobileNav();
  initTabs();
  initToast();
  highlightActiveNav();
});

// ---- Dialog ----
function initDialogs() {
  // Close on overlay click
  document.querySelectorAll('.dialog-overlay').forEach(overlay => {
    overlay.addEventListener('click', e => {
      if (e.target === overlay) closeDialog(overlay.id);
    });
  });

  // Close on Escape
  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') {
      document.querySelectorAll('.dialog-overlay.open').forEach(overlay => {
        closeDialog(overlay.id);
      });
    }
  });
}

function openDialog(id) {
  const overlay = document.getElementById(id);
  if (overlay) overlay.classList.add('open');
}

function closeDialog(id) {
  const overlay = document.getElementById(id);
  if (overlay) overlay.classList.remove('open');
}

// ---- Collapsible ----
function initCollapsibles() {
  document.querySelectorAll('.collapsible-header').forEach(header => {
    header.addEventListener('click', () => {
      const body = header.nextElementSibling;
      const arrow = header.querySelector('.collapsible-arrow');
      const isOpen = body.classList.contains('open');

      body.classList.toggle('open');
      if (arrow) arrow.classList.toggle('open');
      header.setAttribute('aria-expanded', !isOpen);
    });
  });
}

// ---- Mobile Nav ----
function initMobileNav() {
  const btn = document.querySelector('.mobile-menu-btn');
  const sidebar = document.querySelector('.sidebar');
  if (btn && sidebar) {
    btn.addEventListener('click', () => {
      sidebar.classList.toggle('mobile-open');
    });
  }
}

// ---- Tabs ----
function initTabs() {
  document.querySelectorAll('[data-tabs]').forEach(tabGroup => {
    const tabs = tabGroup.querySelectorAll('[data-tab]');
    const panels = tabGroup.parentElement.querySelectorAll('[data-tab-panel]');

    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        const target = tab.dataset.tab;
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        panels.forEach(p => {
          p.classList.toggle('hidden', p.dataset.tabPanel !== target);
        });
      });
    });
  });
}

// ---- Toast ----
function initToast() {
  // Container created on demand
}

function showToast(message, type = 'success') {
  let container = document.querySelector('.toast-container');
  if (!container) {
    container = document.createElement('div');
    container.className = 'toast-container';
    document.body.appendChild(container);
  }

  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `
    <span>${message}</span>
    <button class="toast-close" onclick="this.parentElement.remove()">&times;</button>
  `;
  container.appendChild(toast);

  setTimeout(() => toast.remove(), 3000);
}

// ---- Active Nav Highlight ----
function highlightActiveNav() {
  const currentPage = window.location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.sidebar-link').forEach(link => {
    if (link.getAttribute('href') === currentPage) {
      link.classList.add('active');
    }
  });
}

// ---- Search Dropdown ----
function toggleSearchResults(show) {
  const results = document.querySelector('.search-results');
  if (results) {
    results.classList.toggle('open', show);
  }
}

function selectUser(name) {
  const input = document.querySelector('#user-search');
  const display = document.querySelector('#selected-user');
  if (input) input.value = name;
  if (display) {
    display.textContent = name;
    display.classList.remove('hidden');
  }
  toggleSearchResults(false);
}

// ---- Inline Role Change ----
function startRoleChange(btn) {
  const container = btn.closest('.role-display');
  const textEl = container.querySelector('.role-text');
  const selectEl = container.querySelector('.role-select');
  const actionsEl = container.querySelector('.role-actions');

  textEl.classList.add('hidden');
  btn.classList.add('hidden');
  selectEl.classList.remove('hidden');
  actionsEl.classList.remove('hidden');
}

function confirmRoleChange(btn) {
  const container = btn.closest('.role-display');
  const textEl = container.querySelector('.role-text');
  const selectEl = container.querySelector('.role-select');
  const actionsEl = container.querySelector('.role-actions');
  const changeBtn = container.querySelector('.role-change-btn');

  textEl.textContent = selectEl.options[selectEl.selectedIndex].text;
  textEl.classList.remove('hidden');
  changeBtn.classList.remove('hidden');
  selectEl.classList.add('hidden');
  actionsEl.classList.add('hidden');

  showToast('角色已更新');
}

function cancelRoleChange(btn) {
  const container = btn.closest('.role-display');
  const textEl = container.querySelector('.role-text');
  const selectEl = container.querySelector('.role-select');
  const actionsEl = container.querySelector('.role-actions');
  const changeBtn = container.querySelector('.role-change-btn');

  // Reset select to original value
  selectEl.value = selectEl.dataset.original;

  textEl.classList.remove('hidden');
  changeBtn.classList.remove('hidden');
  selectEl.classList.add('hidden');
  actionsEl.classList.add('hidden');
}

// ---- Checkbox Counter ----
function updateCheckboxCount(groupId) {
  const group = document.querySelector(`[data-group="${groupId}"]`);
  if (!group) return;
  const checked = group.querySelectorAll('input[type="checkbox"]:checked').length;
  const total = group.querySelectorAll('input[type="checkbox"]').length;
  const counter = document.querySelector(`[data-counter="${groupId}"]`);
  if (counter) counter.textContent = `(${checked}/${total})`;
}

// ---- Permission Demo ----
function switchRole(role) {
  const permissions = {
    superadmin: ['all'],
    pm: ['team:invite', 'team:remove', 'team:update', 'team:delete', 'team:transfer',
         'main_item:create', 'main_item:update', 'main_item:archive',
         'sub_item:create', 'sub_item:read', 'sub_item:update', 'sub_item:assign', 'sub_item:change_status',
         'progress:create', 'progress:read', 'progress:update',
         'item_pool:submit', 'item_pool:review',
         'view:weekly', 'view:gantt', 'view:table',
         'report:export'],
    member: ['main_item:read', 'sub_item:read', 'sub_item:create', 'sub_item:update',
             'progress:read', 'item_pool:submit',
             'view:weekly', 'view:table', 'report:export']
  };

  const userPerms = permissions[role] || [];
  const isSuperadmin = role === 'superadmin';

  document.querySelectorAll('.demo-element').forEach(el => {
    const required = el.dataset.permission;
    const hasPerm = isSuperadmin || userPerms.includes(required);
    el.classList.toggle('hidden-perm', !hasPerm);
    el.querySelector('.perm-status').textContent = hasPerm ? '✓' : '✗';
  });
}
