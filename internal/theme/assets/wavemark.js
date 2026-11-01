// Wavemark progressive enhancements. The page is complete without this file.
(function () {
  "use strict";
  var root = document.documentElement;
  var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // ---- Tabs (Tabs and CodeGroup) ----
  function initTabs(container) {
    var list = container.querySelector(":scope > .wm-tabs__list, :scope > .wm-code__header > .wm-tabs__list");
    if (!list) return;
    var tabs = Array.prototype.slice.call(list.querySelectorAll('[role="tab"]'));
    var indicator = list.querySelector(".wm-tabs__indicator");
    var current = 0;
    function place() {
      if (!indicator || !tabs[current] || !tabs[current].offsetWidth) return;
      var t = tabs[current];
      indicator.style.width = t.offsetWidth + "px";
      indicator.style.height = t.offsetHeight + "px";
      indicator.style.top = t.offsetTop + "px";
      indicator.style.setProperty("--wm-x", t.offsetLeft + "px");
      list.setAttribute("data-indicator", "");
    }
    function select(i, focus) {
      current = i;
      tabs.forEach(function (tab, j) {
        var on = i === j;
        tab.setAttribute("aria-selected", on ? "true" : "false");
        tab.tabIndex = on ? 0 : -1;
        var panel = document.getElementById(tab.getAttribute("aria-controls"));
        if (panel) panel.hidden = !on;
      });
      place();
      if (focus) tabs[i].focus();
    }
    if (indicator) {
      indicator.style.transition = "none";
      requestAnimationFrame(function () { place(); requestAnimationFrame(function () { indicator.style.transition = ""; }); });
      window.addEventListener("resize", place);
      if (document.fonts && document.fonts.ready) document.fonts.ready.then(place);
    }
    tabs.forEach(function (tab, i) {
      tab.addEventListener("click", function () { select(i, false); });
      tab.addEventListener("keydown", function (e) {
        var n = tabs.length, k = e.key, next = -1;
        if (k === "ArrowRight") next = (i + 1) % n;
        else if (k === "ArrowLeft") next = (i - 1 + n) % n;
        else if (k === "Home") next = 0;
        else if (k === "End") next = n - 1;
        if (next >= 0) { e.preventDefault(); select(next, true); }
      });
    });
    select(0, false);
  }
  document.querySelectorAll("[data-wm-tabs]").forEach(initTabs);

  // ---- Copy code ----
  function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text);
    return new Promise(function (resolve, reject) {
      var ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      try { document.execCommand("copy") ? resolve() : reject(); } catch (err) { reject(err); }
      document.body.removeChild(ta);
    });
  }
  document.addEventListener("click", function (e) {
    var btn = e.target.closest("[data-wm-copy]");
    if (!btn) return;
    var box = btn.closest(".wm-code");
    var pres = box ? box.querySelectorAll("pre") : [];
    var pre = null;
    for (var i = 0; i < pres.length; i++) {
      var panel = pres[i].closest(".wm-tabs__panel");
      if (!panel || !panel.hidden) { pre = pres[i]; break; }
    }
    if (!pre) return;
    copyText(pre.innerText.replace(/\n$/, "")).then(function () {
      btn.setAttribute("data-state", "copied");
      var label = btn.getAttribute("aria-label");
      btn.setAttribute("aria-label", btn.getAttribute("data-copied") || label);
      setTimeout(function () { btn.removeAttribute("data-state"); btn.setAttribute("aria-label", label); }, 1600);
    }, function () {});
  });

  // ---- Theme toggle: system → light → dark ----
  var themeBtn = document.querySelector("[data-wm-theme]");
  if (themeBtn) {
    var modes = ["system", "light", "dark"];
    var current = root.getAttribute("data-theme") || "system";
    function apply(mode) {
      current = mode;
      themeBtn.setAttribute("data-mode", mode);
      themeBtn.setAttribute("title", themeBtn.getAttribute("data-label-" + mode) || mode);
      if (mode === "system") root.removeAttribute("data-theme"); else root.setAttribute("data-theme", mode);
      try { mode === "system" ? localStorage.removeItem("wm-theme") : localStorage.setItem("wm-theme", mode); } catch (e) {}
    }
    apply(current);
    themeBtn.hidden = false;
    themeBtn.addEventListener("click", function () {
      var next = modes[(modes.indexOf(current) + 1) % modes.length];
      themeBtn.removeAttribute("data-animate");
      void themeBtn.offsetWidth;
      themeBtn.setAttribute("data-animate", "");
      if (document.startViewTransition && !reduceMotion) document.startViewTransition(function () { apply(next); });
      else apply(next);
    });
  }

  // ---- Mobile sidebar ----
  var menuBtn = document.querySelector("[data-wm-menu]");
  var sidebar = document.querySelector("[data-wm-sidebar]");
  if (menuBtn && sidebar) {
    function setOpen(open) {
      sidebar.setAttribute("data-open", open ? "true" : "false");
      menuBtn.setAttribute("aria-expanded", open ? "true" : "false");
    }
    menuBtn.addEventListener("click", function () { setOpen(sidebar.getAttribute("data-open") !== "true"); });
    document.addEventListener("keydown", function (e) { if (e.key === "Escape") setOpen(false); });
    document.addEventListener("click", function (e) {
      if (sidebar.getAttribute("data-open") === "true" && !sidebar.contains(e.target) && !menuBtn.contains(e.target)) setOpen(false);
    });
    var active = sidebar.querySelector('[aria-current="page"]');
    if (active && active.scrollIntoView) active.scrollIntoView({ block: "center" });
  }

  // ---- Scroll reveal ----
  // Blocks below the fold start "pending" (hidden, offscreen) and appear as
  // they enter the viewport; blocks visible on load are never hidden.
  if (!reduceMotion && "IntersectionObserver" in window) {
    var layout = root.getAttribute("data-layout");
    var selectors = {
      landing: ".wm-prose--landing > :not(.wm-hero, .wm-columns, .wm-steps, .wm-section, .wm-pagelist, .wm-flow, hr), .wm-section__inner > :not(.wm-columns, .wm-steps, .wm-pagelist, .wm-flow), .wm-columns > *, .wm-steps > *, .wm-pagelist > *, .wm-flow__step",
      article: ".wm-prose--reading .wm-flow__step, .wm-prose--reading > .wm-figure, .wm-prose--reading > .wm-quote, .wm-prose--reading .wm-columns > *, .wm-prose--reading > .wm-steps > *",
      docs: ".wm-prose .wm-flow__step, .wm-prose .wm-columns > *, .wm-prose .wm-steps > *, .wm-prose > .wm-figure, .wm-prose .wm-pagelist > *"
    };
    var els = document.querySelectorAll(selectors[layout] || selectors.docs);
    var vh = window.innerHeight;
    var io2 = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) {
        if (!en.isIntersecting) return;
        en.target.setAttribute("data-wm-reveal", "shown");
        io2.unobserve(en.target);
      });
    }, { rootMargin: "0px 0px -8% 0px", threshold: 0.08 });
    Array.prototype.forEach.call(els, function (el) {
      var r = el.getBoundingClientRect();
      if (r.top < vh && r.bottom > 0) {
        el.setAttribute("data-wm-reveal", "shown-now");
      } else {
        el.setAttribute("data-wm-reveal", "pending");
        io2.observe(el);
      }
    });
  }

  // ---- Table of contents: highlight the current section ----
  var toc = document.querySelector("[data-wm-toc]");
  if (toc && "IntersectionObserver" in window) {
    var links = Array.prototype.slice.call(toc.querySelectorAll("a"));
    var targets = links.map(function (a) { return document.getElementById(decodeURIComponent(a.hash.slice(1))); });
    var visible = new Set();
    function update() {
      var idx = -1;
      for (var i = 0; i < targets.length; i++) if (targets[i] && visible.has(targets[i])) { idx = i; break; }
      if (idx < 0) {
        // Above the first visible heading: pick the last heading scrolled past.
        for (var j = targets.length - 1; j >= 0; j--) if (targets[j] && targets[j].getBoundingClientRect().top < 100) { idx = j; break; }
      }
      links.forEach(function (a, k) { k === idx ? a.setAttribute("aria-current", "true") : a.removeAttribute("aria-current"); });
    }
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) { en.isIntersecting ? visible.add(en.target) : visible.delete(en.target); });
      update();
    }, { rootMargin: "-56px 0px -55% 0px" });
    targets.forEach(function (t) { if (t) io.observe(t); });
  }
})();
