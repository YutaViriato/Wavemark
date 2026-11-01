// wavemark dev: reload the page when the content changes, keeping the scroll position.
(function () {
  "use strict";
  var key = "wm-dev-scroll:" + location.pathname;
  try {
    var y = sessionStorage.getItem(key);
    if (y !== null) {
      sessionStorage.removeItem(key);
      window.scrollTo(0, parseInt(y, 10) || 0);
    }
  } catch (e) {}
  if (!window.EventSource) return;
  function reload() {
    try { sessionStorage.setItem(key, String(window.scrollY)); } catch (e) {}
    location.reload();
  }
  var boot = null;
  var es = new EventSource("/_wavemark/events");
  es.addEventListener("reload", reload);
  // A different boot id means the server restarted (new binary): reload too.
  es.addEventListener("hello", function (e) {
    if (boot !== null && boot !== e.data) reload();
    boot = e.data;
  });
})();
