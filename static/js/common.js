// ─── API ───────────────────────────────────────────────────────────────────

const API = {
  async post(url, body) {
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || `HTTP ${res.status}`);
    }
    const text = await res.text();
    return text ? JSON.parse(text) : null;
  },

  async get(url) {
    const res = await fetch(url, { credentials: 'include' });
    if (res.status === 401) { window.location.href = '/login'; return null; }
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || `HTTP ${res.status}`);
    }
    const text = await res.text();
    return text ? JSON.parse(text) : null;
  },
};

// ─── TOAST ─────────────────────────────────────────────────────────────────

function ensureToastContainer() {
  let c = document.getElementById('toast-container');
  if (!c) {
    c = document.createElement('div');
    c.id = 'toast-container';
    document.body.appendChild(c);
  }
  return c;
}

function toast(msg, type = 'info', duration = 3000) {
  const c = ensureToastContainer();
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.textContent = msg;
  c.appendChild(el);
  setTimeout(() => {
    el.style.animation = 'fadeIn 0.2s ease reverse forwards';
    setTimeout(() => el.remove(), 200);
  }, duration);
}

const SUIT_SYMBOLS = { 0: '♥', 1: '♠', 2: '♦', 3: '♣' };
const SUIT_NAMES   = { 0: 'hearts', 1: 'spades', 2: 'diamonds', 3: 'clubs' };
const RANK_NAMES   = {
  2:'2', 3:'3', 4:'4', 5:'5', 6:'6', 7:'7',
  8:'8', 9:'9', 10:'10', 11:'J', 12:'Q', 13:'K', 14:'A'
};

function cardSuitClass(suit) { return SUIT_NAMES[suit] ?? 'spades'; }
function cardRankLabel(rank) { return RANK_NAMES[rank] ?? '?'; }
function cardSuitSymbol(suit) { return SUIT_SYMBOLS[suit] ?? '?'; }

function buildCardEl(card, opts = {}) {
  const el = document.createElement('div');
  const suitCls = cardSuitClass(card.suit);
  el.className = `card ${suitCls}`;

  if (opts.facedown) {
    el.classList.add('card-facedown');
  } else {
    const rank = cardRankLabel(card.rank);
    const sym  = cardSuitSymbol(card.suit);
    el.innerHTML = `
      <span class="card-rank">${rank}</span>
      <span class="card-suit-center">${sym}</span>
      <span class="card-rank-bottom">${rank}</span>
    `;
  }

  if (opts.selected)  el.classList.add('selected');
  if (opts.disabled)  el.classList.add('card-disabled');

  if (opts.animDelay !== undefined) {
    el.style.opacity = '0';
    el.style.animationDelay = `${opts.animDelay}ms`;
    el.classList.add('deal-anim');
    el.addEventListener('animationend', () => { el.style.opacity = '1'; }, { once: true });
  }

  if (opts.onClick && !opts.disabled) {
    el.addEventListener('click', () => opts.onClick(card, el));
  }

  // card key for identity
  el.dataset.rank = card.rank;
  el.dataset.suit = card.suit;
  return el;
}

function cardKey(card) { return `${card.rank}-${card.suit}`; }

// ─── POLLING ───────────────────────────────────────────────────────────────

class Poller {
  constructor(fn, interval = 2000) {
    this.fn = fn;
    this.interval = interval;
    this._timer = null;
    this._running = false;
  }
  start() {
    if (this._running) return;
    this._running = true;
    this._tick();
  }
  _tick() {
    if (!this._running) return;
    this.fn().finally(() => {
      if (this._running) this._timer = setTimeout(() => this._tick(), this.interval);
    });
  }
  stop() {
    this._running = false;
    clearTimeout(this._timer);
  }
}

// ─── MISC ──────────────────────────────────────────────────────────────────

function roomIdFromPath()  { return window.location.pathname.split('/room/')[1]?.split('/')[0]; }
function gameIdFromPath()  { return window.location.pathname.split('/game/')[1]?.split('/')[0]; }

function qs(sel, root = document) { return root.querySelector(sel); }
function qsa(sel, root = document) { return [...root.querySelectorAll(sel)]; }
