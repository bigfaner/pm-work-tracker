/* ============================================
   Milestone Map Prototype - Shared JS
   ============================================ */

// --- Sidebar Active Highlight ---
(function () {
  const currentPage = location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.sidebar-link').forEach(function (link) {
    const href = link.getAttribute('href');
    if (href === currentPage) {
      link.classList.add('active');
    } else {
      link.classList.remove('active');
    }
  });
})();

// --- Modal Open/Close ---
function openModal(id) {
  var overlay = document.getElementById(id);
  if (!overlay) return;
  overlay.classList.add('show');
  document.body.style.overflow = 'hidden';
  // Focus first input
  var firstInput = overlay.querySelector('input, select, textarea');
  if (firstInput) setTimeout(function () { firstInput.focus(); }, 100);
}

function closeModal(id) {
  var overlay = document.getElementById(id);
  if (!overlay) return;
  overlay.classList.remove('show');
  document.body.style.overflow = '';
}

// Close modal on overlay click
document.addEventListener('click', function (e) {
  if (e.target.classList.contains('modal-overlay') && e.target.classList.contains('show')) {
    e.target.classList.remove('show');
    document.body.style.overflow = '';
  }
});

// Close modal on Escape
document.addEventListener('keydown', function (e) {
  if (e.key === 'Escape') {
    document.querySelectorAll('.modal-overlay.show').forEach(function (overlay) {
      overlay.classList.remove('show');
    });
    // Also close detail panel
    var panel = document.querySelector('.detail-panel.show');
    if (panel) {
      panel.classList.remove('show');
      var panelOverlay = document.querySelector('.panel-overlay.show');
      if (panelOverlay) panelOverlay.classList.remove('show');
    }
    document.body.style.overflow = '';
  }
});

// --- Detail Panel ---
function openPanel(panelId, overlayId) {
  var panel = document.getElementById(panelId);
  var overlay = document.getElementById(overlayId);
  if (panel) panel.classList.add('show');
  if (overlay) overlay.classList.add('show');
  document.body.style.overflow = 'hidden';
}

function closePanel(panelId, overlayId) {
  var panel = document.getElementById(panelId);
  var overlay = document.getElementById(overlayId);
  if (panel) panel.classList.remove('show');
  if (overlay) overlay.classList.remove('show');
  document.body.style.overflow = '';
}

// --- Dropdown Toggle ---
function toggleDropdown(id) {
  var menu = document.getElementById(id);
  if (!menu) return;

  // Close other dropdowns
  document.querySelectorAll('.dropdown-menu.show').forEach(function (other) {
    if (other.id !== id) other.classList.remove('show');
  });

  menu.classList.toggle('show');

  // Update trigger arrow
  var trigger = menu.parentElement.querySelector('.dropdown-trigger');
  if (trigger) {
    trigger.classList.toggle('open', menu.classList.contains('show'));
  }
}

// Close dropdowns on outside click
document.addEventListener('click', function (e) {
  if (!e.target.closest('.dropdown')) {
    document.querySelectorAll('.dropdown-menu.show').forEach(function (menu) {
      menu.classList.remove('show');
      var trigger = menu.parentElement.querySelector('.dropdown-trigger');
      if (trigger) trigger.classList.remove('open');
    });
  }
});

// --- Toast Notifications ---
function showToast(message, actionText, onAction, duration) {
  duration = duration || 5000;
  var container = document.getElementById('toast-container');
  if (!container) return;

  var toast = document.createElement('div');
  toast.className = 'toast';
  toast.innerHTML = '<span class="toast-message">' + message + '</span>';

  if (actionText && onAction) {
    var actionEl = document.createElement('span');
    actionEl.className = 'toast-action';
    actionEl.textContent = actionText;
    actionEl.addEventListener('click', function () {
      onAction();
      removeToast(toast);
    });
    toast.appendChild(actionEl);
  }

  container.appendChild(toast);

  var timer = setTimeout(function () {
    removeToast(toast);
  }, duration);

  toast._timer = timer;
}

function removeToast(toast) {
  if (toast._timer) clearTimeout(toast._timer);
  toast.classList.add('toast-out');
  setTimeout(function () {
    if (toast.parentNode) toast.parentNode.removeChild(toast);
  }, 200);
}

// --- Zoom Controls ---
function setZoom(level) {
  document.querySelectorAll('.zoom-btn').forEach(function (btn) {
    btn.classList.toggle('active', btn.dataset.zoom === level);
  });
}

// --- State Toggle (for prototype preview) ---
function switchState(stateName) {
  document.querySelectorAll('[data-state]').forEach(function (el) {
    el.style.display = 'none';
  });
  var target = document.querySelector('[data-state="' + stateName + '"]');
  if (target) target.style.display = '';

  // Update state toggle buttons
  document.querySelectorAll('.state-toggle-btn').forEach(function (btn) {
    btn.classList.toggle('active', btn.dataset.state === stateName);
  });
}
