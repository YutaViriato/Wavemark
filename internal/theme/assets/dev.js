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
  var es = new EventSource("/_wavemark/events");
  es.addEventListener("reload", function () {
    try { sessionStorage.setItem(key, String(window.scrollY)); } catch (e) {}
    location.reload();
  });
})();
