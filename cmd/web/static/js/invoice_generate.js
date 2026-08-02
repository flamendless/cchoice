(function () {
  "use strict";

  function modal() {
    return document.getElementById("invoice-generate-modal");
  }

  function formatMoney(centavos, currency) {
    var negative = centavos < 0;
    var v = Math.abs(centavos);
    var pesos = Math.floor(v / 100);
    var frac = v % 100;
    var intStr = pesos.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
    var cur = currency ? currency + " " : "";
    return cur + (negative ? "-" : "") + intStr + "." + (frac < 10 ? "0" + frac : frac);
  }

  // Parse a decimal peso string into integer centavos.
  function toCentavos(str) {
    if (!str) return 0;
    var cleaned = ("" + str).replace(/,/g, "").trim();
    var f = parseFloat(cleaned);
    if (isNaN(f)) return 0;
    return Math.round(f * 100);
  }

  function makeRow(name, pricePesos, productId) {
    var tr = document.createElement("tr");
    tr.className = "invoice-line-row border-t border-gray-100";
    if (productId) tr.dataset.productId = productId;
    tr.innerHTML =
      '<td class="px-3 py-2"><input type="text" class="line-desc px-2 py-1 border border-gray-300 rounded w-full" placeholder="Item description"></td>' +
      '<td class="px-3 py-2"><input type="number" min="1" step="1" value="1" class="line-qty px-2 py-1 border border-gray-300 rounded w-full text-right"></td>' +
      '<td class="px-3 py-2"><input type="text" class="line-price px-2 py-1 border border-gray-300 rounded w-full text-right" placeholder="0.00"></td>' +
      '<td class="px-3 py-2 text-right line-amount text-gray-700">—</td>' +
      '<td class="px-3 py-2 text-center"><button type="button" class="line-remove text-red-600 hover:text-red-800 font-bold">&times;</button></td>';
    tr.querySelector(".line-desc").value = name || "";
    if (pricePesos !== null && pricePesos !== undefined) {
      tr.querySelector(".line-price").value = pricePesos;
    }
    return tr;
  }

  function recompute() {
    var m = modal();
    if (!m) return;
    var currency = m.dataset.currency || "PHP";
    var vatPct = parseFloat(m.dataset.vat || "0") || 0;
    var subtotal = 0;
    m.querySelectorAll(".invoice-line-row").forEach(function (row) {
      var qty = parseInt(row.querySelector(".line-qty").value, 10) || 0;
      var price = toCentavos(row.querySelector(".line-price").value);
      var amount = qty * price;
      subtotal += amount;
      row.querySelector(".line-amount").textContent = formatMoney(amount, currency);
    });
    var vat = Math.round((subtotal * vatPct) / 100);
    var total = subtotal + vat;
    var vatLabel = m.querySelector("#invoice-vat-label");
    if (vatLabel) vatLabel.textContent = vatPct ? "VAT (" + vatPct + "%)" : "VAT";
    setText(m, "#invoice-subtotal", formatMoney(subtotal, currency));
    setText(m, "#invoice-vat", formatMoney(vat, currency));
    setText(m, "#invoice-total", formatMoney(total, currency));
  }

  function setText(root, sel, text) {
    var el = root.querySelector(sel);
    if (el) el.textContent = text;
  }

  function showError(msg) {
    var m = modal();
    if (!m) return;
    var box = m.querySelector("#invoice-generate-error");
    if (!box) return;
    if (!msg) {
      box.classList.add("hidden");
      box.textContent = "";
      return;
    }
    box.textContent = msg;
    box.classList.remove("hidden");
  }

  function collectPayload(action) {
    var m = modal();
    var recipientSelect = m.querySelector("#invoice-recipient-select");
    var recipientId = recipientSelect ? recipientSelect.value : "";
    var payload = {
      recipient_id: recipientId === "__new__" ? "" : recipientId,
      new_recipient: null,
      notes: (m.querySelector("#invoice-notes") || {}).value || "",
      due_date: (m.querySelector("#invoice-due-date") || {}).value || "",
      action: action,
      lines: [],
    };

    if (recipientId === "__new__") {
      var nr = {};
      m.querySelectorAll("[data-new-recipient]").forEach(function (input) {
        nr[input.getAttribute("data-new-recipient")] = input.value || "";
      });
      payload.new_recipient = nr;
    }

    m.querySelectorAll(".invoice-line-row").forEach(function (row) {
      var desc = row.querySelector(".line-desc").value.trim();
      var qty = parseInt(row.querySelector(".line-qty").value, 10) || 0;
      var price = row.querySelector(".line-price").value.trim();
      if (!desc && !price) return;
      payload.lines.push({
        product_id: row.dataset.productId || "",
        description: desc,
        unit_price: price,
        quantity: qty,
      });
    });

    return payload;
  }

  function validate(payload) {
    var m = modal();
    var recipientSelect = m.querySelector("#invoice-recipient-select");
    if (recipientSelect && recipientSelect.value === "__new__") {
      if (!payload.new_recipient || !payload.new_recipient.name) {
        return "Please enter a name for the new recipient.";
      }
    } else if (!payload.recipient_id) {
      return "Please select a recipient.";
    }
    if (!payload.lines.length) {
      return "Please add at least one line item.";
    }
    return "";
  }

  function submit(action) {
    var m = modal();
    if (!m) return;
    showError("");
    var payload = collectPayload(action);
    var err = validate(payload);
    if (err) {
      showError(err);
      return;
    }

    var url = m.dataset.createUrl;
    var tableUrl = m.dataset.tableUrl;
    fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then(function (resp) {
        return resp.json().then(function (data) {
          return { ok: resp.ok, data: data };
        });
      })
      .then(function (res) {
        if (!res.ok) {
          showError((res.data && res.data.error) || "Failed to create invoice.");
          return;
        }
        var data = res.data;
        if (action === "download" && data.pdf_url) {
          window.open(data.pdf_url, "_blank");
        }
        if (action === "email" && data.email_error) {
          showError("Invoice saved, but email failed: " + data.email_error);
          refreshTable(tableUrl);
          return;
        }
        // Close modal and refresh the invoices table.
        var container = document.getElementById("invoice-generate-modal-container");
        if (container) container.innerHTML = "";
        refreshTable(tableUrl);
      })
      .catch(function () {
        showError("Network error while creating invoice.");
      });
  }

  function refreshTable(tableUrl) {
    if (window.htmx && tableUrl) {
      window.htmx.ajax("GET", tableUrl, { target: "#invoices-table", swap: "innerHTML" });
    }
  }

  document.addEventListener("click", function (e) {
    var m = modal();
    if (!m) return;

    if (e.target.closest("#invoice-add-manual")) {
      e.preventDefault();
      m.querySelector("#invoice-lines").appendChild(makeRow("", "", ""));
      recompute();
      return;
    }

    if (e.target.closest("#invoice-add-product")) {
      e.preventDefault();
      var sel = m.querySelector("#invoice-product-select");
      if (!sel || !sel.value) return;
      var opt = sel.options[sel.selectedIndex];
      var name = opt.getAttribute("data-name") || opt.textContent.trim();
      var priceCentavos = parseInt(opt.getAttribute("data-price") || "0", 10) || 0;
      var pricePesos = (priceCentavos / 100).toFixed(2);
      m.querySelector("#invoice-lines").appendChild(makeRow(name, pricePesos, sel.value));
      recompute();
      return;
    }

    var removeBtn = e.target.closest(".line-remove");
    if (removeBtn && m.contains(removeBtn)) {
      e.preventDefault();
      var row = removeBtn.closest(".invoice-line-row");
      if (row) row.remove();
      recompute();
      return;
    }

    var actionBtn = e.target.closest("[data-invoice-action]");
    if (actionBtn && m.contains(actionBtn)) {
      e.preventDefault();
      submit(actionBtn.getAttribute("data-invoice-action"));
    }
  });

  document.addEventListener("input", function (e) {
    var m = modal();
    if (!m || !m.contains(e.target)) return;
    if (e.target.closest(".invoice-line-row")) {
      recompute();
    }
  });

  document.addEventListener("change", function (e) {
    var m = modal();
    if (!m || !m.contains(e.target)) return;
    if (e.target.id === "invoice-recipient-select") {
      var newBox = m.querySelector("#invoice-new-recipient");
      if (newBox) {
        newBox.classList.toggle("hidden", e.target.value !== "__new__");
      }
    }
  });

  // Seed the modal with one empty line when it is swapped in.
  document.body.addEventListener("htmx:afterSwap", function (evt) {
    if (evt.detail && evt.detail.target && evt.detail.target.id === "invoice-generate-modal-container") {
      var m = modal();
      if (m && m.querySelector("#invoice-lines") && m.querySelectorAll(".invoice-line-row").length === 0) {
        m.querySelector("#invoice-lines").appendChild(makeRow("", "", ""));
        recompute();
      }
    }
  });
})();
