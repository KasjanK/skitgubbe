/* game.js — Skitgubbe game logic */

const gameId = gameIdFromPath();
let state = null;        // VisibleState
let prevState = null;
let selectedCards = [];  // indices into you.hand (or faceup/facedown depending on phase)
let selectedHandIdx = null;  // For setup phase
let isAnimating = false;

const UI = {
  hand: qs('#your-hand'),
  pile: qs('#pile'),
  deck: qs('#deck-area'),
  others: qs('#other-hands'),
  actions: qs('#action-buttons')
};

async function makeMove(move) {
  try {
    const needsAnim = ['play_card', 'play_many', 'play_face_up', 'play_face_down'].includes(move.type);
    
    await API.post(`/api/games/${gameId}/move`, move);
    // Don't clear selectedCards here - wait for new state
    
    if (needsAnim) {
      await animateCardsToPile(move);
    }
    selectedCards = [];
    await fetchState();
    // Clear after state updates
  } catch (e) {
    console.error('Move failed:', e);
  }
    render();
}

async function animateCardsToPile(move) {
    isAnimating = true;
  const pileArea = UI.pile;
  const pileRect = pileArea.getBoundingClientRect();
  
  let cardsToAnimate = [];
  
  if (move.type === 'play_card') {
    const cardEl = qs(`[data-key="hand-${move.index}-${move.card.rank}-${move.card.suit}"]`);
    if (cardEl) cardsToAnimate.push(cardEl);
  } else if (move.type === 'play_many') {
    move.indices.forEach(idx => {
      const cardEl = document.querySelector(`#your-hand [data-index="${idx}"]`);
      if (cardEl) cardsToAnimate.push(cardEl);
    });
  } else if (move.type === 'play_face_up') {
    const cardEl = document.querySelector(`#your-faceup [data-index="${move.index}"]`);
    if (cardEl) cardsToAnimate.push(cardEl);
  } else if (move.type === 'play_face_down') {
      const cardEl = document.querySelector(`#your-facedown [data-index="${move.index}"]`);
      if (cardEl) cardsToAnimate.push(cardEl);
  }  
  if (cardsToAnimate.length === 0) return;
  
  // Animate each card
  const animations = cardsToAnimate.map(cardEl => {
    const cardRect = cardEl.getBoundingClientRect();
    const flyX = pileRect.left - cardRect.left + (pileRect.width - cardRect.width) / 2;
    const flyY = pileRect.top - cardRect.top + (pileRect.height - cardRect.height) / 2;
    
    cardEl.style.setProperty('--fly-x', `${flyX}px`);
    cardEl.style.setProperty('--fly-y', `${flyY}px`);
    cardEl.style.animation = 'cardFlyToPile 0.4s ease-out forwards';
    
    return new Promise(resolve => {
      setTimeout(resolve, 400);
    });
  });
  
  await Promise.all(animations);
  isAnimating = false;
    render();
}

async function fetchState() {
  try {
    const next = await API.get(`/api/games/${gameId}`);
    if (!next) return;

      if (state && JSON.stringify(state) === JSON.stringify(next)) {
          return;
      }

    prevState = state;
    state = next;
    render();
  } catch (e) {
    console.error(e);
  }
}

function render() {
  if (!state || isAnimating) return;
  
  if (state.phase === 'setup') {
    renderSetup();
  } else {
    renderPlay();
  }
  
  renderWinners();
}

function renderSetup() {
  const you = state.you;
  if (!you) return;

  console.log('renderSetup called, hand size:', you.hand.length);

  // Hide action buttons during setup
  UI.actions.innerHTML = '';

  // Render deck and pile (usually empty during setup)
  renderDeck();
  renderPile();

  // Render opponents (they're also setting up)
  renderOthers();

  // Render your cards with swap functionality
  renderYourHandSetup(you);
  renderYourTableSetup(you);

  // Show ready button
  const actionsEl = UI.actions;
  if (!you.ready) {
    const readyBtn = document.createElement('button');
    readyBtn.className = 'btn btn-gold';
    readyBtn.textContent = '✔ READY UP';
    readyBtn.onclick = readySetup;
    actionsEl.appendChild(readyBtn);

    const hint = document.createElement('div');
    hint.className = 'status-bar';
    hint.innerHTML = '<span style="color:var(--cream-dark);">Click a hand card, then a face-up slot to swap.</span>';
    actionsEl.appendChild(hint);
  } else {
    const waiting = document.createElement('div');
    waiting.className = 'status-bar';
    waiting.innerHTML = '<span class="blink">⏳</span> Waiting for other players to ready up...';
    actionsEl.appendChild(waiting);
  }
}

function renderYourHandSetup(you) {
  const el = UI.hand;
  el.innerHTML = '';

  if (!you.hand.length) {
    el.innerHTML = `<span class="no-cards">No cards</span>`;
    return;
  }

  you.hand.forEach((card, idx) => {
    const isSelected = selectedHandIdx === idx;
    const el2 = buildCardEl(card, {
      selected: isSelected,
      disabled: you.ready,
      onClick: you.ready ? null : () => selectHandCardSetup(idx),
    });
    el.appendChild(el2);
  });
}

function selectHandCardSetup(idx) {
  if (selectedHandIdx === idx) {
    selectedHandIdx = null;
  } else {
    selectedHandIdx = idx;
  }
  renderYourHandSetup(state.you);
  console.log('Hand card selected:', idx, 'selectedHandIdx=', selectedHandIdx);
}

function renderYourTableSetup(you) {
  const fuEl = qs('#your-faceup');
  const fdEl = qs('#your-facedown');
  fuEl.innerHTML = '';
  fdEl.innerHTML = '';

  // Face-up slots (clickable to swap with selected hand card)
  (you.faceup_table_cards ?? []).forEach((card, idx) => {
    const el = buildCardEl(card, {
      disabled: you.ready,
      onClick: you.ready ? null : () => swapFaceUp(idx),
    });
    fuEl.appendChild(el);
  });

  // Add empty slots if less than 3
  for (let i = (you.faceup_table_cards ?? []).length; i < 3; i++) {
    const slot = document.createElement('div');
    slot.className = 'card';
    slot.style.cssText = 'background:transparent; border-style:dashed; border-color:#3a6b4a; cursor:pointer;';
    slot.innerHTML = `<span style="color:#3a6b4a; font-size:11px; font-family:'Press Start 2P',monospace;">+</span>`;
    if (!you.ready) slot.addEventListener('click', () => swapFaceUp(i));
    fuEl.appendChild(slot);
  }

  // Face-down (non-interactive during setup)
  const fdCount = (you.facedown_table_cards ?? []).length;
  for (let i = 0; i < fdCount; i++) {
    fdEl.appendChild(buildCardEl({ rank: 0, suit: 0 }, { facedown: true, disabled: true }));
  }
}

async function swapFaceUp(slotIndex) {
  console.log('faceup slot clicked', slotIndex, 'selectedHandIdx=', selectedHandIdx);
  if (selectedHandIdx == null) return;

  try {
    await API.post(`/api/games/${gameId}/move`, {
      type: 'swap_face_up',
      index: selectedHandIdx,
      indices: [slotIndex],
    });
    selectedHandIdx = null;
    await fetchState();
  } catch (e) {
    console.error('Swap failed:', e);
  }
}

async function readySetup() {
  try {
    await API.post(`/api/games/${gameId}/move`, { type: 'ready_setup' });
    await fetchState();
  } catch (e) {
    console.error('Ready setup failed:', e);
  }
}

// ─── Play Phase ────────────────────────────────────────────────────────────

function renderPlay() {
  renderPile();
  renderDeck();
  renderOthers();
  renderYourArea();
}

// ─── Pile ──────────────────────────────────────────────────────────────────

function renderPile() {
  const el = UI.pile;
  el.innerHTML = '';

  const pile = state.pile ?? [];
  const count = pile.length;

  if (count === 0) {
    el.innerHTML = `<div class="pile-empty"><span>PILE</span><br/><span style="font-size:13px;opacity:0.5;">empty</span></div>`;
    return;
  }

  // Show up to 5 fanned cards, top card fully visible
  const show = pile.slice(Math.max(0, count - 5));
  show.forEach((card, i) => {
    const isTop = i === show.length - 1;
    const offset = i * 3;
    const rotation = (i - show.length / 2) * 2;
    const el2 = buildCardEl(card, { disabled: true });
    el2.style.cssText = `position:absolute; left:${offset}px; top:${offset}px; transform:rotate(${rotation}deg);`;
    if (!isTop) el2.style.filter = 'brightness(0.7)';
    el.appendChild(el2);
  });

  // Pile count badge
  const badge = document.createElement('div');
  badge.className = 'pile-count';
  badge.textContent = count;
  el.appendChild(badge);
}
// ─── Deck ──────────────────────────────────────────────────────────────────
function renderDeck() {
  const el = UI.deck;
  el.innerHTML = '';
  const deckSize = state.deck_size;

  if (deckSize === 0) {
    el.innerHTML = `<div class="pile-empty"><span>DECK</span><br/><span style="font-size:13px;opacity:0.5;">empty</span></div>`;
    return;
  }

  // Stack all cards on top of each other (slight offset for depth)
  const stackDepth = Math.min(deckSize, 10); // Show up to 10 layers for visual depth
  for (let i = 0; i < stackDepth; i++) {
    const c = buildCardEl({ rank: 0, suit: 0 }, { facedown: true });
    c.style.cssText = `position:absolute; left:${i*0.5}px; top:${i*0.5}px; cursor:default;`;
    el.appendChild(c);
  }

  const badge = document.createElement('div');
  badge.className = 'pile-count';
  badge.textContent = deckSize;
  el.appendChild(badge);
}

// ─── Opponents ─────────────────────────────────────────────────────────────

function renderOthers() {
  const container = UI.others;
  container.innerHTML = '';

  const others = state.others ?? [];
  const POSITIONS = ['player-top', 'player-left', 'player-right'];

  others.forEach((p, i) => {
    const isCurrent = p.id === state.current_player;
    const posClass = POSITIONS[i % 3];

    const div = document.createElement('div');
    div.className = `player-area ${posClass}`;

    // Name badge
    const badge = document.createElement('div');
    badge.className = `player-name-badge ${isCurrent ? 'active' : 'inactive'}`;
    badge.textContent = `${isCurrent ? '▶ ' : ''}${p.id}`;
    div.appendChild(badge);

    // Hand row (separate, on top)
    if (p.hand_size > 0) {
      const handRow = document.createElement('div');
      handRow.className = 'opp-hand-row';
      for (let j = 0; j < Math.min(p.hand_size, 8); j++) {
        const c = buildCardEl({ rank: 0, suit: 0 }, { facedown: true, disabled: true });
        handRow.appendChild(c);
      }
      div.appendChild(handRow);
    }

    // Table row - 3 stacks (faceup on top of facedown)
    const tableRow = document.createElement('div');
    tableRow.className = 'opp-table-row';

    const faceupCards = p.others_faceup_table_cards ?? [];
    const facedownCount = p.facedown_table_cards_size ?? 0;

    for (let j = 0; j < 3; j++) {
      const stack = document.createElement('div');
      stack.className = 'opp-table-stack';

      // Facedown card (bottom layer) - just a generic card back, no value
      if (j < facedownCount) {
        const fd = buildCardEl({ rank: 0, suit: 0 }, { facedown: true, disabled: true });
        stack.appendChild(fd);
      }

      // Faceup card (top layer)
      if (j < faceupCards.length) {
        const fu = buildCardEl(faceupCards[j], { disabled: true });
        stack.appendChild(fu);
      }

      tableRow.appendChild(stack);
    }

    div.appendChild(tableRow);
    container.appendChild(div);
  });
}

// ─── Your area ─────────────────────────────────────────────────────────────

function renderYourArea() {
  const you = state.you;
  if (!you) return;
  const isMyTurn = state.current_player === you.id;
  const phase = state.phase;

  qs('#your-name-label').textContent = you.id;

  renderYourHand(you, isMyTurn, phase);
  renderYourTable(you, isMyTurn, phase);
  renderActionButtons(you, isMyTurn, phase);
}

function renderYourHand(you, isMyTurn, phase) {
  const container = UI.hand;

    const existing = new Map();
    qsa('.card', container).forEach(el => existing.set(el.dataset.key, el));

    if (!you || !you.hand) {
        return;
    }

  if (!you.hand.length) {
    el.innerHTML = `<span class="no-cards">Hand empty</span>`;
    return;
  }

    const newHandKeys = new Set();

    you.hand.forEach((card, idx) => {
        const key = `hand-${idx}-${card.rank}-${card.suit}`;
        newHandKeys.add(key);
        let el = existing.get(key);
        const isSelected = selectedCards.some(sc => sc.key === key);
        if (el) {
            el.classList.toggle('selected', isSelected);
            el.dataset.index = idx;
        } else {
            el = buildCardEl(card, {
                selected: isSelected,
                onClick: (c, cardEl) => toggleCardSelection(cardEl, card, idx, key),
            });
            el.dataset.key = key;
            container.appendChild(el);
        }
    });
    existing.forEach((el, key) => {
        if (!newHandKeys.has(key)) el.remove();
    });
}

function renderYourTable(you, isMyTurn, phase) {
  const fuContainer = qs('#your-faceup');
  const fdContainer = qs('#your-facedown');

    const existingFU = new Map();
    qsa('.card', fuContainer).forEach(el => existingFU.set(el.dataset.key, el));
    const newFUKeys = new Set();

    (you.faceup_table_cards ?? []).forEach((card, idx) => {
        const key = `faceup-${idx}-${card.rank}-${card.suit}`;
        newFUKeys.add(key);
        let el = existingFU.get(key);
        const isSelected = selectedCards.some(sc => sc.key === key);
        if (el) {
            el.classList.toggle('selected', isSelected);
            el.dataset.index = idx;
        } else {
            el = buildCardEl(card, {
                selected: isSelected,
                onClick: (c, cardEl) => toggleCardSelection(cardEl, card, idx, key),
            });
            el.dataset.key = key;
            fuContainer.appendChild(el);
        }
    });
    existingFU.forEach((el, key) => {
        if (!newFUKeys.has(key)) el.remove();
    });

    const existingFD = new Map();
    qsa('.card', fdContainer).forEach(el => existingFD.set(el.dataset.key, el));
    const newFDKeys = new Set();

    const fdCount = (you.facedown_table_cards ?? []).length;
    for (let i = 0; i < fdCount; i++) {
        const key = `facedown-${i}`;
        const isSelected = selectedCards.some(sc => sc.key === key);
        newFDKeys.add(key);
        let el = existingFD.get(key);
        if (el) {
            el.classList.toggle('selected', isSelected);
            el.dataset.index = i;
        } else {
            el = buildCardEl({ rank: 0, suit: 0 }, {
                facedown: true,
                selected: isSelected,
                onClick: (c, cardEl) => toggleCardSelection(cardEl, { rank: null, suit: null }, i, key),
            });
            el.dataset.key = key;
            fdContainer.appendChild(el);
        }
    }
    existingFD.forEach((el, key) => {
        if (!newFDKeys.has(key)) el.remove();
    });
}

function toggleCardSelection(li, card, idx, key) {
  const existing = selectedCards.findIndex(c => c.key === key);
  if (existing >= 0) {
    selectedCards.splice(existing, 1);
    li.classList.remove('selected');
  } else {
    selectedCards.push({ key, card, index: idx });
    li.classList.add('selected');
  }

  renderActionButtons(state.you, state.current_player === state.you.id, state.phase);
}

function renderActionButtons(you, isMyTurn, phase) {
  const container = UI.actions;
  container.innerHTML = '';

    if (state.finished || phase !== 'play') return;

  // Check what types of cards are selected
  const handSelected = selectedCards.filter(sc => sc.key.startsWith('hand-'));
  const faceupSelected = selectedCards.filter(sc => sc.key.startsWith('faceup-'));
  const facedownSelected = selectedCards.filter(sc => sc.key.startsWith('facedown-'));
  const zonesSelected = [handSelected.length > 0, faceupSelected.length > 0, facedownSelected.length > 0].filter(Boolean).length;

  console.log('renderActionButtons:', {
    selectedCards,
    handSelected,
    faceupSelected, 
    facedownSelected,
    zonesSelected
  });

  // Play selected cards button - always show if cards are selected
  if (selectedCards.length > 0 && zonesSelected === 1) {
    if (handSelected.length > 0) {
      const sorted = handSelected.map(sc => sc.index).sort((a, b) => a - b);
      
      if (handSelected.length === 1) {
        const card = handSelected[0].card;
        const btn = document.createElement('button');
        btn.className = 'btn btn-gold';
        btn.textContent = `▶ PLAY ${cardRankLabel(card.rank)}${cardSuitSymbol(card.suit)}`;
        btn.onclick = () => makeMove({ type: 'play_card', card, index: sorted[0] });
        container.appendChild(btn);
      } else {
        // Check all same rank for play_many
        const ranks = handSelected.map(sc => sc.card.rank);
        const allSame = ranks.every(r => r === ranks[0]);
        if (allSame) {
          const btn = document.createElement('button');
          btn.className = 'btn btn-gold';
          btn.textContent = `▶ PLAY ${handSelected.length}× ${cardRankLabel(ranks[0])}`;
          btn.onclick = () => makeMove({ type: 'play_many', indices: sorted });
          container.appendChild(btn);
        }
      }
    } else if (faceupSelected.length === 1) {
      const card = faceupSelected[0].card;
      const idx = faceupSelected[0].index;
      const btn = document.createElement('button');
      btn.className = 'btn btn-gold';
      btn.textContent = `▶ PLAY ${cardRankLabel(card.rank)}${cardSuitSymbol(card.suit)}`;
      btn.onclick = () => makeMove({ type: 'play_face_up', index: idx });
      container.appendChild(btn);
    } else if (facedownSelected.length === 1) {
      const idx = facedownSelected[0].index;
      const btn = document.createElement('button');
      btn.className = 'btn btn-gold';
      btn.textContent = `▶ PLAY FACE-DOWN`;
      btn.onclick = () => makeMove({ type: 'play_face_down', index: idx });
      container.appendChild(btn);
    }
  }

  // Chance button - always available
  const deckSize = state.deck_size ?? 0;
  const pileSize = state.pile?.length ?? 0;
  const canChance = deckSize > 0 && pileSize > 0;
  
  const chanceBtn = document.createElement('button');
  chanceBtn.className = `btn ${canChance ? 'btn-green' : 'btn-gray'}`;
  chanceBtn.textContent = '🎲 TAKE CHANCE';
  chanceBtn.disabled = !canChance;
  chanceBtn.title = !canChance 
    ? (deckSize === 0 ? 'Deck is empty' : 'Pile is empty') 
    : 'Draw a card from the deck and try to play it';
  chanceBtn.onclick = () => makeMove({ type: 'chance' });
  container.appendChild(chanceBtn);

  // Pickup button - always available
  const pileEmpty = !state.pile?.length;
  const pickupBtn = document.createElement('button');
  pickupBtn.className = `btn ${pileEmpty ? 'btn-gray' : 'btn-red'}`;
  pickupBtn.textContent = '✋ PICK UP PILE';
  pickupBtn.disabled = pileEmpty;
  pickupBtn.title = pileEmpty ? 'Pile is empty' : 'Take all pile cards into your hand';
  pickupBtn.onclick = () => makeMove({ type: 'pickup' });
  container.appendChild(pickupBtn);

  // Deselect - show if cards are selected
  if (selectedCards.length > 0) {
    const clearBtn = document.createElement('button');
    clearBtn.className = 'btn btn-green';
    clearBtn.textContent = '✕ CLEAR';
    clearBtn.onclick = () => { 
      // Manually remove selected class from all cards
      selectedCards.forEach(sc => {
        const cardEl = document.querySelector(`[data-key="${sc.key}"]`);
        if (cardEl) cardEl.classList.remove('selected');
      });
      selectedCards = [];
      renderActionButtons(state.you, true, state.phase); 
    };
    container.appendChild(clearBtn);
  }
}
// ─── Winner overlay ────────────────────────────────────────────────────────

function renderWinners() {
  if (!state.finished) return;
  const overlay = qs('#winner-overlay');
  if (!overlay) return;
  overlay.style.display = 'flex';

  const winners = state.winners ?? [];
  const you = state.you?.id;
  const allPlayers = [you, ...(state.others?.map(p => p.id) ?? [])];
  const loser = allPlayers.find(p => !winners.includes(p));

  const isWinner = winners.includes(you);
  const isLoser = loser === you;

  qs('#winner-title').textContent = isWinner ? '🏆 YOU WIN!' : isLoser ? '💀 SKITGUBBE!' : '🎮 GAME OVER';
  qs('#winner-title').style.color = isWinner ? 'var(--gold)' : isLoser ? 'var(--red-bright)' : 'var(--cream)';

  const list = qs('#winner-list');
  list.innerHTML = '';
  
  winners.forEach((playerId, index) => {
    const div = document.createElement('div');
    div.className = 'winner-entry' + (playerId === you ? ' winner-anim' : '');
    const place = index + 1;
    let placeText = place + 'th';
    if (place === 1) placeText = '1st';
    else if (place === 2) placeText = '2nd';
    else if (place === 3) placeText = '3rd';
    
    div.textContent = `${placeText}: ${playerId}`;
    if (playerId === you) {
      div.textContent += ' (you)';
    }
    list.appendChild(div);
  });
  
  if (loser) {
    const d = document.createElement('div');
    d.style.color = 'var(--red-bright)';
    d.style.marginTop = '8px';
    d.textContent = `💀 Skitgubbe: ${loser}`;
    list.appendChild(d);
  }
}

// ─── Boot ──────────────────────────────────────────────────────────────────

fetchState();
const poller = new Poller(fetchState, 2000);
poller.start();
