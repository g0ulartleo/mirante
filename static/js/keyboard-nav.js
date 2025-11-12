function initKeyboardNav() {
    const container = document.getElementById('alarms-container');
    let snapshot = [];
    function tiles() {
        const nodeList = Array.from(document.querySelectorAll('#alarms-container .grid > *[tabindex="0"][role="button"][data-idx]'));
        nodeList.forEach(el => {
            if (!el.dataset.href) {
                const anchor = el.querySelector('a[href]');
                if (anchor) el.dataset.href = anchor.getAttribute('href');
            }
        });
        nodeList.sort((a,b) => parseInt(a.dataset.idx,10) - parseInt(b.dataset.idx,10));
        snapshot = nodeList;
        return snapshot;
    }
    function cols(list) {
        if (list.length === 0) return 1;
        const grid = document.querySelector('#alarms-container .grid');
        const attr = grid ? grid.getAttribute('data-cols') : null;
        const parsed = attr ? parseInt(attr, 10) : NaN;
        if (!Number.isNaN(parsed) && parsed > 0) return parsed;
        const y = list[0].getBoundingClientRect().top;
        return Math.max(1, list.filter(el => Math.abs(el.getBoundingClientRect().top - y) < 2).length);
    }
    function setActive(el, active) {
        if (!el) return;
        el.classList.toggle('ring-2', active);
        el.classList.toggle('ring-black-300', active);
        el.classList.toggle('kbd-active', active);
        el.classList.toggle('show-overlay', active);
        const overlay = el.querySelector('.kbd-overlay');
        if (overlay) {
            if (active) {
                overlay.style.setProperty('opacity', '1', 'important');
            } else {
                overlay.style.removeProperty('opacity');
            }
        }
    }
    function clearSelection(list) {
        for (const el of list) {
            el.classList.remove('ring-2', 'ring-black-300', 'kbd-active', 'show-overlay');
            const overlay = el.querySelector('.kbd-overlay');
            if (overlay) overlay.style.removeProperty('opacity');
        }
    }
    let currentIdx = 0;
    function focusIndex(list, idx) {
        if (idx < 0 || idx >= list.length) return;
        clearSelection(list);
        currentIdx = idx;
        list[idx].focus({preventScroll:true});
        list[idx].scrollIntoView({block:'nearest', inline:'nearest'});
        setActive(list[idx], true);
    }
    function parentHref() {
        const p = window.location.pathname.replace(/\/$/, '');
        const i = p.lastIndexOf('/');
        return i <= 0 ? '/' : p.slice(0, i);
    }
    const style = document.createElement('style');
    style.textContent = '[tabindex="0"][role="button"]:focus{outline: none;} .group:hover .kbd-overlay{opacity:0 !important;} .kbd-active .kbd-overlay{opacity:1 !important;} .show-overlay .kbd-overlay{opacity:1 !important;}';
    document.head.appendChild(style);
    let kbdHideTimer = null;
    function bindMouseHandlers(list) {
        list.forEach((el, i) => {
            el.addEventListener('mouseenter', () => {
                container.classList.remove('suppress-hover');
                clearSelection(snapshot);
                // map mouse to logical index
                const idxAttr = el.getAttribute('data-idx');
                const logicalIdx = idxAttr ? parseInt(idxAttr, 10) : i;
                currentIdx = Math.max(0, Math.min(logicalIdx, snapshot.length - 1));
                setActive(el, true);
                if (kbdHideTimer) { clearTimeout(kbdHideTimer); kbdHideTimer = null; }
            });
        });
    }
    if (window.__kbdKeydownHandler) {
        document.removeEventListener('keydown', window.__kbdKeydownHandler, true);
    }
    window.__kbdKeydownHandler = (e) => {
        const list = tiles();
        if (list.length === 0) return;
        const c = cols(list);
        let idx = currentIdx;
        let handled = false;
        switch (e.key) {  // why not
            case 'ArrowLeft': case 'h': {
                handled = true;
                if (idx % c === 0) break; 
                idx -= 1; break;
            }
            case 'ArrowRight': case 'l': {
                handled = true;
                const L = snapshot.length;
                if ((idx % c) === c - 1 || idx + 1 >= L) break;
                idx += 1; break;
            }
            case 'ArrowUp': case 'k': {
                handled = true;
                if (idx < c) break;
                idx -= c; break;
            }
            case 'ArrowDown': case 'j': {
                handled = true;
                const L = snapshot.length;
                if (idx + c >= L) break;
                idx += c; break;
            }
            case 'Enter': {
                const target = list[idx] || null;
                if (target && target.dataset && target.dataset.href) {
                    window.location.href = target.dataset.href;
                    handled = true;
                }
                break;
            }
            case 'Backspace':
                window.location.href = parentHref();
                handled = true;
                break;
        }
        if (handled) {
            e.preventDefault();
            container.classList.add('suppress-hover');
            clearSelection(snapshot);
            idx = Math.max(0, Math.min(idx, snapshot.length - 1));
            focusIndex(snapshot, idx);
            if (kbdHideTimer) clearTimeout(kbdHideTimer);
            kbdHideTimer = setTimeout(() => {
                clearSelection(snapshot);
                container.classList.remove('suppress-hover');
            }, 5000);
        }
    };
    document.addEventListener('keydown', window.__kbdKeydownHandler, true);
    bindMouseHandlers(tiles());
    container.addEventListener('mouseleave', () => {
        clearSelection(snapshot);
        container.classList.remove('suppress-hover');
        if (kbdHideTimer) { clearTimeout(kbdHideTimer); kbdHideTimer = null; }
    });
    window.addEventListener('blur', () => {
        clearSelection(snapshot);
        container.classList.remove('suppress-hover');
    });
    document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'hidden') {
            clearSelection(snapshot);
            container.classList.remove('suppress-hover');
        }
    });
}


