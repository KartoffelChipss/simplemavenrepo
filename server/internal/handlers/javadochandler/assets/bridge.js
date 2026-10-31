(function () {
    'use strict';
    if (window.parent === window) return;

    var root = document.documentElement;

    function report() {
        window.parent.postMessage(
            { type: 'smr-javadoc', path: location.pathname, hash: location.hash, title: document.title },
            '*'
        );
    }

    window.addEventListener('message', function (e) {
        var data = e.data;
        if (e.source !== window.parent || !data || data.type !== 'smr-javadoc') return;
        if (data.theme === 'light' || data.theme === 'dark') root.setAttribute('data-smr-theme', data.theme);
        else if (data.theme === 'system') root.removeAttribute('data-smr-theme');
    });

    document.addEventListener(
        'click',
        function (e) {
            var a = e.target instanceof Element ? e.target.closest('a[href]') : null;
            if (!a) return;
            var url = new URL(a.href, location.href);
            if (url.host !== location.host || (url.protocol !== 'http:' && url.protocol !== 'https:')) {
                a.target = '_blank';
                a.rel = 'noopener noreferrer';
            } else if (a.target) {
                a.removeAttribute('target');
            }
        },
        true
    );

    window.addEventListener('hashchange', report);
    report();
})();
