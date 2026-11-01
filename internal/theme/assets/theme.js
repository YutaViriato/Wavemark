// Runs before first paint: marks JS support and applies the stored theme.
(function () {
  var d = document.documentElement;
  d.classList.add("wm-js");
  try {
    var t = localStorage.getItem("wm-theme");
    if (t === "light" || t === "dark") d.setAttribute("data-theme", t);
  } catch (e) {}
})();
