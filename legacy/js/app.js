(function() {
    "use strict";

    var POLL_INTERVAL = 400;
    var CIRCUMFERENCE = 2 * Math.PI * 90;

    // ─── Animation Options (must match CSS classes) ───
    var ANIM_OPTIONS = [
        'anim-pulse-glow', 'anim-float', 'anim-wiggle', 'anim-bounce-in',
        'anim-slide-up', 'anim-slide-down', 'anim-zoom-pulse', 'anim-swing',
        'anim-rubber', 'anim-jello', 'anim-heartbeat', 'anim-flash',
        'anim-shake-x', 'anim-blur-in', 'anim-tracking-expand', 'anim-wave'
    ];
    var SMILE_DURATION = 1200; // ms to show "Bitte lächeln!"

    // ─── Elements ───
    var screens = {
        idle:       document.getElementById('screen-idle'),
        countdown:  document.getElementById('screen-countdown'),
        capturing:  document.getElementById('screen-capturing'),
        preview:    document.getElementById('screen-preview'),
        error:      document.getElementById('screen-error')
    };

    var elStatusBar    = document.getElementById('status-bar');
    var elStatusDot    = document.getElementById('status-dot');
    var elStatusText   = document.getElementById('status-text');
    var elStatusState  = document.getElementById('status-state');
    var elCountdownNum = document.getElementById('countdown-number');
    var elRingProgress = document.getElementById('ring-progress');
    var elPreviewImg   = document.getElementById('preview-img');
    var elErrorMsg     = document.getElementById('error-msg');

    // Capturing elements
    var elSpinner       = document.getElementById('capture-spinner');
    var elSmileText     = document.getElementById('smile-text');
    var elProcessingText = document.getElementById('processing-text');

    var currentState = '';
    var pollOk = false;
    var smileTimeout = null;
    var currentAnimClass = '';

    // Init SVG ring
    elRingProgress.setAttribute('stroke-dasharray', CIRCUMFERENCE);
    elRingProgress.setAttribute('stroke-dashoffset', '0');

    function randomAnim() {
        return ANIM_OPTIONS[Math.floor(Math.random() * ANIM_OPTIONS.length)];
    }

    function showScreen(name) {
        var key;
        for (key in screens) {
            if (screens.hasOwnProperty(key)) {
                screens[key].className = (key === name) ? 'screen active' : 'screen';
            }
        }
    }

    function updateRing(remaining, total) {
        if (total <= 0) {
            elRingProgress.setAttribute('stroke-dashoffset', CIRCUMFERENCE);
            return;
        }
        var progress = remaining / total;
        var offset = CIRCUMFERENCE * (1 - progress);
        elRingProgress.setAttribute('stroke-dashoffset', offset);
    }

    function showSmile() {
        // Show "Bitte lächeln!" with tada, hide spinner and processing text
        elSpinner.classList.add('hidden');
        elProcessingText.classList.add('hidden');
        elSmileText.className = 'smile-text visible';

        if (smileTimeout) clearTimeout(smileTimeout);
        smileTimeout = setTimeout(function() {
            showProcessing();
        }, SMILE_DURATION);
    }

    function showProcessing() {
        // Switch to spinner + random animation text
        elSmileText.className = 'smile-text';

        // Pick a random animation class
        if (currentAnimClass) {
            elProcessingText.classList.remove(currentAnimClass);
        }
        currentAnimClass = randomAnim();
        elProcessingText.classList.add(currentAnimClass);

        elSpinner.classList.remove('hidden');
        elProcessingText.classList.remove('hidden');
    }

    function updateUI(data) {
        var state = data.state || 'idle';
        var countdown = data.countdown || {};

        // Status bar
        elStatusBar.style.display = 'flex';
        elStatusDot.className = 'online';
        elStatusText.textContent = 'Verbunden';
        elStatusState.textContent = state;
        pollOk = true;

        if (state === currentState && state !== 'countdown') return;

        if (state === 'idle') {
            showScreen('idle');
            // Reset capturing elements
            elSmileText.className = 'smile-text';
            elProcessingText.classList.add('hidden');
            elSpinner.classList.add('hidden');
        } else if (state === 'countdown') {
            showScreen('countdown');
            var remaining = countdown.remaining || 0;
            var total = countdown.total || 3;
            elCountdownNum.textContent = remaining > 0 ? remaining : '';
            updateRing(remaining, total);
        } else if (state === 'capturing') {
            showScreen('capturing');
            showSmile();
        } else if (state === 'processing') {
            showScreen('capturing'); // same screen
            showProcessing();
        } else if (state === 'preview') {
            showScreen('preview');
            if (data.lastPhoto && data.lastPhoto.url) {
                elPreviewImg.src = data.lastPhoto.url + '?t=' + Date.now();
            }
        } else if (state === 'error') {
            showScreen('error');
            if (data.error) {
                elErrorMsg.textContent = data.error;
            }
        }

        currentState = state;
    }

    function poll() {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', '/api/legacy/poll', true);
        xhr.timeout = 3000;

        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                if (xhr.status === 200) {
                    try {
                        var data = JSON.parse(xhr.responseText);
                        updateUI(data);
                    } catch (e) {
                        // Parse error
                    }
                }
                setTimeout(poll, POLL_INTERVAL);
            }
        };

        xhr.onerror = function() {
            if (pollOk) {
                elStatusDot.className = '';
                elStatusText.textContent = 'Verbindung verloren';
                pollOk = false;
            }
            setTimeout(poll, POLL_INTERVAL * 3);
        };

        xhr.ontimeout = function() {
            setTimeout(poll, POLL_INTERVAL * 2);
        };

        xhr.send();
    }

    // Trigger (global for onclick)
    window.doTrigger = function() {
        if (currentState !== 'idle' && currentState !== 'preview') return;

        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/trigger', true);
        xhr.send();
    };

    // Start
    poll();
})();
