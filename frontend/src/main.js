import './style.css';
import './app.css';

import { GetFeatureStatuses, ToggleFeature } from '../wailsjs/go/main/App';
import arrowIcon from './assets/icons/arrow-right.svg?raw';
import infoIcon from './assets/icons/info.svg?raw';

const cardList = document.getElementById('card-list');
const statusBadge = document.getElementById('status-badge');
const noticeIcon = document.getElementById('notice-icon');

noticeIcon.innerHTML = infoIcon;

function renderStatus(features) {
    const enabled = features.filter(f => f.enabled).length;
    statusBadge.innerHTML = `<strong>${enabled}</strong> / ${features.length} 已启用`;
}

function renderCard(feature, index) {
    const card = document.createElement('div');
    card.className = 'card';

    const modsHtml = feature.mods.map(mod => `
        <div class="mod-row">
            <span class="mod-dot ${mod.enabled ? 'on' : ''}">■</span>
            <span class="mod-name ${mod.found ? '' : 'missing'}">${mod.displayName}</span>
        </div>
    `).join('');

    card.innerHTML = `
        <div class="card-top">
            <span class="card-arrow">${arrowIcon}</span>
            <span class="card-name">${feature.name}</span>
            <div class="switch ${feature.enabled ? 'on' : ''}"><div class="switch-knob"></div></div>
        </div>
        <div class="card-desc">${feature.description}</div>
        <div class="card-detail">${modsHtml}</div>
    `;

    const cardTop = card.querySelector('.card-top');
    const switchEl = card.querySelector('.switch');

    cardTop.addEventListener('click', (event) => {
        if (event.target.closest('.switch')) return;
        card.classList.toggle('expanded');
    });

    switchEl.addEventListener('click', (event) => {
        event.stopPropagation();
        const nextEnabled = !switchEl.classList.contains('on');
        ToggleFeature(index, nextEnabled).then(renderAll);
    });

    return card;
}

async function renderAll() {
    const features = await GetFeatureStatuses();

    renderStatus(features);

    cardList.innerHTML = '';
    features.forEach((feature, index) => {
        cardList.appendChild(renderCard(feature, index));
    });
}

renderAll();
