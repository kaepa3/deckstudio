let ws = null;
let config = null;
let activeProfileId = null;
let autoSwitch = true;
let currentAppInfo = { app_name: '', title: '' };

const statusBadge = document.getElementById('statusBadge');
const statusText = document.getElementById('statusText');
const activeAppName = document.getElementById('activeAppName');
const activeAppIcon = document.getElementById('activeAppIcon');
const autoSwitchBtn = document.getElementById('autoSwitchBtn');
const globalGrid = document.getElementById('globalGrid');
const profileTabs = document.getElementById('profileTabs');
const appGrid = document.getElementById('appGrid');
const currentProfileName = document.getElementById('currentProfileName');
const currentProfileIcon = document.getElementById('currentProfileIcon');

// Auto switch toggle
autoSwitchBtn.addEventListener('click', () => {
  autoSwitch = !autoSwitch;
  if (autoSwitch) {
    autoSwitchBtn.classList.add('active');
    syncProfileWithApp();
  } else {
    autoSwitchBtn.classList.remove('active');
  }
});

function connectWebSocket() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    statusBadge.classList.add('connected');
    statusText.textContent = 'Connected';
  };

  ws.onclose = () => {
    statusBadge.classList.remove('connected');
    statusText.textContent = 'Reconnecting...';
    setTimeout(connectWebSocket, 2000);
  };

  ws.onerror = (err) => {
    console.error('WebSocket error:', err);
    ws.close();
  };

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      handleServerMessage(data);
    } catch (e) {
      console.error('Failed to parse server message:', e);
    }
  };
}

function handleServerMessage(data) {
  if (data.type === 'init') {
    config = data.config;
    renderGlobalButtons();
    renderProfileTabs();
    if (config.profiles && config.profiles.length > 0 && !activeProfileId) {
      activeProfileId = config.profiles[0].id;
    }
    syncProfileWithApp();
  } else if (data.type === 'active_app') {
    currentAppInfo = data;
    updateActiveAppUI(data);
    if (autoSwitch) {
      syncProfileWithApp();
    }
  }
}

function updateActiveAppUI(info) {
  if (info.app_name) {
    activeAppName.textContent = info.app_name;
  } else {
    activeAppName.textContent = 'Desktop';
  }
}

function syncProfileWithApp() {
  if (!config || !config.profiles || !currentAppInfo.app_name) return;

  const appLower = currentAppInfo.app_name.toLowerCase();
  
  for (const profile of config.profiles) {
    if (profile.app_names) {
      for (const name of profile.app_names) {
        if (appLower.includes(name.toLowerCase())) {
          switchProfile(profile.id);
          return;
        }
      }
    }
  }
}

function switchProfile(profileId) {
  activeProfileId = profileId;

  // Update tabs UI
  document.querySelectorAll('.profile-tab').forEach(tab => {
    if (tab.dataset.id === profileId) {
      tab.classList.add('active');
    } else {
      tab.classList.remove('active');
    }
  });

  renderAppButtons();
}

function renderGlobalButtons() {
  if (!config || !config.global_buttons) return;

  globalGrid.innerHTML = '';
  config.global_buttons.forEach(btn => {
    const btnEl = createDeckButton(btn);
    globalGrid.appendChild(btnEl);
  });
  lucide.createIcons();
}

function renderProfileTabs() {
  if (!config || !config.profiles) return;

  profileTabs.innerHTML = '';
  config.profiles.forEach(profile => {
    const tab = document.createElement('button');
    tab.className = `profile-tab ${profile.id === activeProfileId ? 'active' : ''}`;
    tab.dataset.id = profile.id;
    tab.innerHTML = `
      <i data-lucide="${profile.icon || 'layers'}"></i>
      <span>${profile.name}</span>
    `;

    tab.addEventListener('click', () => {
      // Manual click disables auto switch temporarily or stays on tab
      switchProfile(profile.id);
    });

    profileTabs.appendChild(tab);
  });
  lucide.createIcons();
}

function renderAppButtons() {
  if (!config || !config.profiles) return;

  const currentProfile = config.profiles.find(p => p.id === activeProfileId);
  if (!currentProfile) return;

  currentProfileName.textContent = currentProfile.name.toUpperCase();
  currentProfileIcon.setAttribute('data-lucide', currentProfile.icon || 'layers');

  appGrid.innerHTML = '';
  if (currentProfile.buttons) {
    currentProfile.buttons.forEach(btn => {
      const btnEl = createDeckButton(btn);
      appGrid.appendChild(btnEl);
    });
  }

  lucide.createIcons();
}

function createDeckButton(btn) {
  const button = document.createElement('div');
  button.className = 'deck-btn';
  button.dataset.id = btn.id;

  if (btn.color) {
    button.style.setProperty('--btn-color', btn.color);
  }

  button.innerHTML = `
    <i data-lucide="${btn.icon || 'command'}"></i>
    <span>${btn.Label || btn.label}</span>
  `;

  const handlePress = (e) => {
    e.preventDefault();

    // Haptic feedback
    if (navigator.vibrate) {
      navigator.vibrate(35);
    }

    // Visual animation
    button.classList.add('pressed');
    setTimeout(() => button.classList.remove('pressed'), 150);

    // Send WebSocket command
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'trigger',
        button_id: btn.id
      }));
    }
  };

  button.addEventListener('pointerdown', handlePress);

  return button;
}

// Initial connection
connectWebSocket();
