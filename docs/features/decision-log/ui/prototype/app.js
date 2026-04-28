// Shared interactions for prototype

// Dialog open/close
function openDialog(id) {
  const overlay = document.getElementById(id);
  if (overlay) {
    overlay.classList.add('open');
    document.body.style.overflow = 'hidden';
  }
}
function closeDialog(id) {
  const overlay = document.getElementById(id);
  if (overlay) {
    overlay.classList.remove('open');
    document.body.style.overflow = '';
  }
}

// Close dialog on overlay click
document.addEventListener('click', function(e) {
  if (e.target.classList.contains('dialog-overlay') && e.target.classList.contains('open')) {
    e.target.classList.remove('open');
    document.body.style.overflow = '';
  }
});

// Close dialog on Escape
document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') {
    const open = document.querySelector('.dialog-overlay.open');
    if (open) {
      open.classList.remove('open');
      document.body.style.overflow = '';
    }
  }
});

// Toast notifications
function showToast(message, type = 'success', duration = 3000) {
  let container = document.querySelector('.toast-container');
  if (!container) {
    container = document.createElement('div');
    container.className = 'toast-container';
    document.body.appendChild(container);
  }
  const toast = document.createElement('div');
  toast.className = 'toast ' + type;
  toast.textContent = message;
  container.appendChild(toast);
  setTimeout(() => {
    toast.remove();
  }, duration);
}

// Content expand/collapse
document.addEventListener('click', function(e) {
  const textEl = e.target.closest('.timeline-text');
  if (textEl) {
    textEl.classList.toggle('expanded');
  }
});

// Tag input handling
function initTagInput(wrapper) {
  const input = wrapper.querySelector('input');
  const dropdown = wrapper.querySelector('.recent-tags-dropdown');

  if (!input) return;

  input.addEventListener('keydown', function(e) {
    if (e.key === 'Enter' || (e.key === ',' && input.value.trim())) {
      e.preventDefault();
      addTag(wrapper, input.value.trim().replace(/,$/, ''));
      input.value = '';
      if (dropdown) dropdown.classList.remove('open');
    }
    if (e.key === 'Backspace' && !input.value) {
      const pills = wrapper.querySelectorAll('.tag-pill');
      if (pills.length > 0) {
        pills[pills.length - 1].remove();
      }
    }
  });

  input.addEventListener('focus', function() {
    if (dropdown && dropdown.children.length > 0) {
      dropdown.classList.add('open');
    }
  });

  input.addEventListener('blur', function() {
    setTimeout(() => {
      if (dropdown) dropdown.classList.remove('open');
    }, 150);
  });

  if (dropdown) {
    dropdown.addEventListener('click', function(e) {
      const option = e.target.closest('.tag-option');
      if (option) {
        addTag(wrapper, option.textContent.trim());
        input.value = '';
        dropdown.classList.remove('open');
      }
    });
  }
}

function addTag(wrapper, text) {
  if (!text) return;
  // Check duplicate
  const existing = wrapper.querySelectorAll('.tag-pill .tag-text');
  for (const el of existing) {
    if (el.textContent === text) return;
  }
  if (text.length > 20) {
    showToast('标签不能超过 20 字符', 'error');
    return;
  }
  const pill = document.createElement('span');
  pill.className = 'tag-pill';
  pill.innerHTML = '<span class="tag-text">' + text + '</span><span class="remove-tag">&times;</span>';
  pill.querySelector('.remove-tag').addEventListener('click', function() {
    pill.remove();
  });
  const input = wrapper.querySelector('input');
  wrapper.insertBefore(pill, input);
}

// Initialize all tag inputs on page
document.addEventListener('DOMContentLoaded', function() {
  document.querySelectorAll('.tag-input-wrapper').forEach(initTagInput);
});

// Character counter
function initCharCounter(textarea, counterEl) {
  if (!textarea || !counterEl) return;
  const max = parseInt(counterEl.dataset.max) || 2000;
  textarea.addEventListener('input', function() {
    const len = textarea.value.length;
    counterEl.textContent = len + '/' + max;
    counterEl.className = 'char-counter';
    if (len > max) counterEl.classList.add('error');
    else if (len > max * 0.9) counterEl.classList.add('warning');
  });
}

// Active sidebar highlight
document.addEventListener('DOMContentLoaded', function() {
  const path = window.location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.sidebar-nav a').forEach(a => {
    if (a.getAttribute('href') === path) a.classList.add('active');
  });
});

// State toggle helper (prototype only)
function showState(stateName, container) {
  const states = container.querySelectorAll('[data-state]');
  states.forEach(s => s.style.display = 'none');
  const target = container.querySelector('[data-state="' + stateName + '"]');
  if (target) target.style.display = '';
}
