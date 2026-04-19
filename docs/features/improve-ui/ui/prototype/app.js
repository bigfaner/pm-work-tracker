/* PM Tracker — Shared Interactions */

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

function switchFilter(el) {
  el.parentElement.querySelectorAll('.tab').forEach(function(t) { t.classList.remove('active'); });
  el.classList.add('active');
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

function toggleExpand(card) {
  var content = card.querySelector('.expand-content');
  var icon = card.querySelector('.expand-icon');
  if (content) {
    var open = content.style.display === 'block';
    content.style.display = open ? 'none' : 'block';
    if (icon) icon.style.transform = open ? '' : 'rotate(90deg)';
  }
}

function toggleGanttSub(label) {
  var next = label.nextElementSibling;
  if (next && next.classList.contains('gantt-sub-group')) {
    var open = next.style.display === 'block';
    next.style.display = open ? 'none' : 'block';
    var icon = label.querySelector('.expand-icon');
    if (icon) icon.style.transform = open ? '' : 'rotate(90deg)';
    var row = label.getAttribute('data-row');
    var tlRows = document.querySelectorAll('.gantt-timeline .gantt-row[data-parent="' + row + '"]');
    tlRows.forEach(function(r) { r.style.display = open ? 'none' : 'block'; });
  }
}

document.addEventListener('DOMContentLoaded', function() {
  var path = window.location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.sidebar-nav a').forEach(function(a) {
    if (a.getAttribute('href') === path) a.classList.add('active');
  });
});
