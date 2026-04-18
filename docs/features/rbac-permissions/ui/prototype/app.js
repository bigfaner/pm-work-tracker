/* PM Tracker — Shared Interactions (RBAC) */

function toggleSidebar() {
  document.querySelector('.sidebar').classList.toggle('mobile-open');
}

function openModal(id) {
  var el = document.getElementById(id);
  if (el) el.classList.add('active');
}
function closeModal(id) {
  var el = document.getElementById(id);
  if (el) el.classList.remove('active');
}

document.addEventListener('click', function(e) {
  if (e.target.classList.contains('modal-overlay')) e.target.classList.remove('active');
});
document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') {
    document.querySelectorAll('.modal-overlay.active').forEach(function(m) { m.classList.remove('active'); });
    document.querySelectorAll('.dropdown-menu.open').forEach(function(d) { d.classList.remove('open'); });
  }
});

document.addEventListener('click', function(e) {
  var trigger = e.target.closest('[data-dropdown]');
  if (trigger) {
    e.stopPropagation();
    var menu = trigger.nextElementSibling;
    document.querySelectorAll('.dropdown-menu.open').forEach(function(d) { if (d !== menu) d.classList.remove('open'); });
    if (menu) menu.classList.toggle('open');
    return;
  }
  if (!e.target.closest('.dropdown')) {
    document.querySelectorAll('.dropdown-menu.open').forEach(function(d) { d.classList.remove('open'); });
  }
});

function switchTab(tabEl) {
  var group = tabEl.closest('.tabs');
  group.querySelectorAll('.tab').forEach(function(t) { t.classList.remove('active'); });
  tabEl.classList.add('active');
  var target = tabEl.getAttribute('data-tab');
  group.parentElement.querySelectorAll('[data-tab-panel]').forEach(function(p) {
    p.style.display = p.getAttribute('data-tab-panel') === target ? '' : 'none';
  });
}

function showToast(msg, type) {
  type = type || 'success';
  var c = document.querySelector('.toast-container');
  if (!c) { c = document.createElement('div'); c.className = 'toast-container'; document.body.appendChild(c); }
  var t = document.createElement('div');
  t.className = 'toast toast-' + type;
  t.textContent = msg;
  c.appendChild(t);
  setTimeout(function() { t.style.opacity = '0'; setTimeout(function() { t.remove(); }, 300); }, 3000);
}

function togglePermGroup(header) {
  var body = header.nextElementSibling;
  var arrow = header.querySelector('.arrow');
  var open = body.style.display !== 'none';
  body.style.display = open ? 'none' : 'block';
  if (arrow) arrow.classList.toggle('open', !open);
}

function updatePermCount(header) {
  var body = header.nextElementSibling;
  var checks = body.querySelectorAll('.perm-check');
  var checked = body.querySelectorAll('.perm-check:checked');
  var countEl = header.querySelector('.perm-count');
  if (countEl) countEl.textContent = '(' + checked.length + '/' + checks.length + ')';
}

document.addEventListener('change', function(e) {
  if (e.target.classList.contains('perm-check')) {
    var group = e.target.closest('.perm-group');
    if (group) updatePermCount(group.querySelector('.perm-group-header'));
  }
});

function switchView(view, hideId, showId) {
  document.getElementById(hideId).style.display = 'none';
  document.getElementById(showId).style.display = '';
}

document.addEventListener('DOMContentLoaded', function() {
  var path = window.location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.sidebar-nav a').forEach(function(a) {
    if (a.getAttribute('href') === path) a.classList.add('active');
  });
});
