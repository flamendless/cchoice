(function () {
  "use strict";

  function debounce(fn, ms) {
    var timer;
    return function () {
      var args = arguments;
      var self = this;
      clearTimeout(timer);
      timer = setTimeout(function () {
        fn.apply(self, args);
      }, ms);
    };
  }

  function initSearch(config) {
    var searchInput = config.searchInput;
    var hiddenInput = config.hiddenInput;
    var suggestionsEl = config.suggestionsEl;
    if (!searchInput || !hiddenInput || !suggestionsEl || !config.searchUrl) {
      return;
    }

    var minChars = config.minChars || 2;
    var onSelect = config.onSelect || function () {};
    var onClear = config.onClear || function () {};

    function hideSuggestions() {
      suggestionsEl.classList.add("hidden");
      suggestionsEl.innerHTML = "";
    }

    function showSuggestions(items) {
      suggestionsEl.innerHTML = "";
      if (!items || items.length === 0) {
        suggestionsEl.innerHTML = '<div class="px-3 py-2 text-sm text-gray-500">No results</div>';
        suggestionsEl.classList.remove("hidden");
        return;
      }
      items.forEach(function (item) {
        var btn = document.createElement("button");
        btn.type = "button";
        btn.className =
          "admin-entity-suggest-item w-full text-left px-3 py-2 text-sm hover:bg-gray-100 border-b border-gray-100 last:border-b-0";
        btn.textContent = item.label || "";
        btn.addEventListener("click", function (e) {
          e.preventDefault();
          hiddenInput.value = item.id || "";
          searchInput.value = item.label || "";
          hideSuggestions();
          onSelect(item);
        });
        suggestionsEl.appendChild(btn);
      });
      suggestionsEl.classList.remove("hidden");
    }

    var doSearch = debounce(function () {
      var q = searchInput.value.trim();
      if (q.length < minChars) {
        hideSuggestions();
        return;
      }
      fetch(config.searchUrl + "?q=" + encodeURIComponent(q))
        .then(function (resp) {
          return resp.json();
        })
        .then(function (data) {
          showSuggestions((data && data.items) || []);
        })
        .catch(function () {
          hideSuggestions();
        });
    }, 300);

    searchInput.addEventListener("input", function () {
      var q = searchInput.value.trim();
      if (!q) {
        hiddenInput.value = "";
        hideSuggestions();
        onClear();
        return;
      }
      if (hiddenInput.value) {
        hiddenInput.value = "";
      }
      doSearch();
    });

    searchInput.addEventListener("focus", function () {
      if (searchInput.value.trim().length >= minChars) {
        doSearch();
      }
    });

    document.addEventListener("click", function (e) {
      if (!searchInput.contains(e.target) && !suggestionsEl.contains(e.target)) {
        hideSuggestions();
      }
    });
  }

  window.AdminEntitySearch = {
    init: initSearch,
    bind: function (root, options) {
      if (!root) return;
      var searchInput = root.querySelector("#" + options.searchInputId);
      var hiddenInput = root.querySelector("#" + options.hiddenInputId);
      var suggestionsEl = root.querySelector("#" + options.searchInputId + "-suggestions");
      initSearch({
        searchInput: searchInput,
        hiddenInput: hiddenInput,
        suggestionsEl: suggestionsEl,
        searchUrl: options.searchUrl,
        minChars: options.minChars,
        onSelect: options.onSelect,
        onClear: options.onClear,
      });
    },
  };
})();
