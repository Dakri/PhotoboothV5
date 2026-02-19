(function() {
    var state = 'idle';
    var lastState = '';
    var POLL_INTERVAL = 500;
    
    var elStatus = document.getElementById('status');
    var elStart = document.getElementById('start-screen');
    var elCountdown = document.getElementById('countdown');
    var elMessage = document.getElementById('message');
    var elPreview = document.getElementById('preview');
    var btnTrigger = document.getElementById('trigger-btn');

    function updateUI(data) {
        state = data.state;
        elStatus.textContent = 'State: ' + state + ' | ' + new Date().toLocaleTimeString();

        // Hide all first
        elStart.style.display = 'none';
        elCountdown.style.display = 'none';
        elMessage.style.display = 'none';
        elPreview.style.display = 'none';

        if (state === 'idle') {
            elStart.style.display = 'flex';
            elStart.style.flexDirection = 'column';
            elStart.style.alignItems = 'center';
            btnTrigger.disabled = false;
        } else if (state === 'countdown') {
            elCountdown.textContent = '...'; // We don't get exact seconds via poll reliably
            elCountdown.style.display = 'block';
        } else if (state === 'capturing') {
            elMessage.textContent = 'CHEESE!';
            elMessage.style.display = 'block';
        } else if (state === 'processing') {
            elMessage.textContent = 'Processing...';
            elMessage.style.display = 'block';
        } else if (state === 'preview') {
            if (data.lastPhoto) {
                elPreview.src = data.lastPhoto.url + '?t=' + new Date().getTime();
                elPreview.style.display = 'block';
            }
        } else if (state === 'error') {
            elMessage.textContent = 'ERROR';
            elMessage.style.color = 'red';
            elMessage.style.display = 'block';
        }

        lastState = state;
    }

    function poll() {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', '/api/legacy/poll', true);
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                if (xhr.status === 200) {
                    try {
                        var data = JSON.parse(xhr.responseText);
                        updateUI(data);
                    } catch (e) {
                        console.error('Parse error', e);
                    }
                }
                setTimeout(poll, POLL_INTERVAL);
            }
        };
        xhr.onerror = function() {
            elStatus.textContent = 'Connection Error';
            setTimeout(poll, POLL_INTERVAL * 2);
        };
        xhr.send();
    }

    window.trigger = function() {
        btnTrigger.disabled = true;
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/trigger', true);
        xhr.send();
    };

    // Start polling
    poll();
})();
