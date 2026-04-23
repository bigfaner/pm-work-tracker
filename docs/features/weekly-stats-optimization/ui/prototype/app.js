// Tooltip logic
const TOOLTIP_DELAY = 300;
let tooltipTimer = null;
let activeTooltip = null;
let isTouchDevice = false;

function isMobile() {
  return window.matchMedia('(max-width: 767px)').matches || isTouchDevice;
}

function showTooltip(card, immediate) {
  const tooltip = card.querySelector('.tooltip-content');
  if (!tooltip) return;

  // Check flip
  const rect = card.getBoundingClientRect();
  const tooltipHeight = 60; // approximate
  if (rect.top < tooltipHeight + 8) {
    tooltip.classList.add('flipped');
  } else {
    tooltip.classList.remove('flipped');
  }

  tooltip.removeAttribute('aria-hidden');
  tooltip.classList.add('visible');
  activeTooltip = tooltip;
}

function hideTooltip(card) {
  const tooltip = card.querySelector('.tooltip-content');
  if (!tooltip) return;
  tooltip.classList.remove('visible');
  tooltip.setAttribute('aria-hidden', 'true');
  if (activeTooltip === tooltip) activeTooltip = null;
}

function hideAllTooltips() {
  document.querySelectorAll('.tooltip-content.visible').forEach(t => {
    t.classList.remove('visible');
    t.setAttribute('aria-hidden', 'true');
  });
  activeTooltip = null;
}

function initCards() {
  const cards = document.querySelectorAll('.stat-card');

  cards.forEach(card => {
    // Desktop hover
    card.addEventListener('mouseenter', () => {
      if (isMobile()) return;
      tooltipTimer = setTimeout(() => showTooltip(card, false), TOOLTIP_DELAY);
    });
    card.addEventListener('mouseleave', () => {
      if (isMobile()) return;
      clearTimeout(tooltipTimer);
      hideTooltip(card);
    });

    // Mobile click toggle
    card.addEventListener('click', (e) => {
      if (!isMobile()) return;
      const tooltip = card.querySelector('.tooltip-content');
      if (!tooltip) return;
      const isVisible = tooltip.classList.contains('visible');
      hideAllTooltips();
      if (!isVisible) showTooltip(card, true);
      e.stopPropagation();
    });

    // Keyboard focus
    card.addEventListener('focus', () => {
      clearTimeout(tooltipTimer);
      showTooltip(card, true);
    });
    card.addEventListener('blur', () => {
      hideTooltip(card);
    });

    // Escape key
    card.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') hideTooltip(card);
    });
  });

  // Click outside (mobile)
  document.addEventListener('click', () => {
    if (isMobile()) hideAllTooltips();
  });

  // Touch detection
  window.addEventListener('touchstart', () => { isTouchDevice = true; }, { once: true });
}

// State switcher
function initStateSwitcher() {
  const btns = document.querySelectorAll('.state-btn');
  btns.forEach(btn => {
    btn.addEventListener('click', () => {
      btns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      const state = btn.dataset.state;
      applyState(state);
    });
  });
}

const STATS_DATA = {
  loaded: { activeSubItems: 42, newlyCompleted: 8, inProgress: 15, blocked: 3, pending: 12, pausing: 4, overdue: 5 },
  empty:  { activeSubItems: 0,  newlyCompleted: 0,  inProgress: 0,  blocked: 0,  pending: 0,  pausing: 0,  overdue: 0 },
};

const CARD_FIELDS = ['activeSubItems','newlyCompleted','inProgress','blocked','pending','pausing','overdue'];

function applyState(state) {
  const wrapper = document.getElementById('stats-wrapper');
  const errorBanner = document.getElementById('error-banner');
  const cards = document.querySelectorAll('.stat-card');

  errorBanner.style.display = 'none';

  if (state === 'loading') {
    wrapper.setAttribute('aria-busy', 'true');
    cards.forEach(card => {
      card.querySelector('.stat-number').style.display = 'none';
      card.querySelector('.stat-label').style.display = 'none';
      card.querySelector('.skeleton-number').style.display = 'block';
      card.querySelector('.skeleton-label').style.display = 'block';
    });
  } else if (state === 'error') {
    wrapper.setAttribute('aria-busy', 'false');
    errorBanner.style.display = 'flex';
    cards.forEach(card => {
      card.querySelector('.skeleton-number').style.display = 'none';
      card.querySelector('.skeleton-label').style.display = 'none';
      const num = card.querySelector('.stat-number');
      num.style.display = 'block';
      num.textContent = '—';
      num.classList.add('error-dash');
      card.querySelector('.stat-label').style.display = 'block';
    });
  } else {
    wrapper.setAttribute('aria-busy', 'false');
    const data = STATS_DATA[state] || STATS_DATA.loaded;
    cards.forEach((card, i) => {
      card.querySelector('.skeleton-number').style.display = 'none';
      card.querySelector('.skeleton-label').style.display = 'none';
      const num = card.querySelector('.stat-number');
      num.style.display = 'block';
      num.classList.remove('error-dash');
      const val = data[CARD_FIELDS[i]];
      num.textContent = val > 9999 ? '9999+' : val;
      card.querySelector('.stat-label').style.display = 'block';
    });
  }
}

// Week selector simulation (AbortController pattern demo)
let currentAbortController = null;
function initWeekSelector() {
  const sel = document.getElementById('week-select');
  if (!sel) return;
  sel.addEventListener('change', () => {
    if (currentAbortController) {
      currentAbortController.abort();
    }
    currentAbortController = new AbortController();
    applyState('loading');
    // Simulate network delay
    const signal = currentAbortController.signal;
    setTimeout(() => {
      if (signal.aborted) return;
      applyState('loaded');
    }, 1200);
  });
}

document.addEventListener('DOMContentLoaded', () => {
  initCards();
  initStateSwitcher();
  initWeekSelector();
});
