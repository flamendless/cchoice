(function () {
  "use strict";

  function modal() {
    return document.getElementById("invoice-generate-modal");
  }

  function canManage() {
    var m = modal();
    return m && m.dataset.canManage === "true";
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
    var taxCell = canManage()
      ? '<td class="px-3 py-2"><select class="line-tax px-2 py-1 border border-gray-300 rounded w-full text-xs"><option value="VATABLE">VAT</option><option value="VAT_EXEMPT">Exempt</option><option value="ZERO_RATED">Zero</option></select></td>'
      : "";
    tr.innerHTML =
      '<td class="px-3 py-2"><input type="text" class="line-desc px-2 py-1 border border-gray-300 rounded w-full" placeholder="Item description"></td>' +
      '<td class="px-3 py-2"><input type="number" min="1" step="1" value="1" class="line-qty px-2 py-1 border border-gray-300 rounded w-full text-right"></td>' +
      '<td class="px-3 py-2"><input type="text" class="line-price px-2 py-1 border border-gray-300 rounded w-full text-right" placeholder="0.00"></td>' +
      taxCell +
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
    var vatAmount = 0;
    m.querySelectorAll(".invoice-line-row").forEach(function (row) {
      var qty = parseInt(row.querySelector(".line-qty").value, 10) || 0;
      var price = toCentavos(row.querySelector(".line-price").value);
      var amount = qty * price;
      subtotal += amount;
      var taxSel = row.querySelector(".line-tax");
      var taxType = taxSel ? taxSel.value : "VATABLE";
      if (taxType === "VATABLE" && vatPct > 0) {
        var net = Math.round(amount / (1 + vatPct / 100));
        vatAmount += amount - net;
      }
      row.querySelector(".line-amount").textContent = formatMoney(amount, currency);
    });
    var withholding = canManage() ? toCentavos((m.querySelector("#invoice-withholding-tax") || {}).value) : 0;
    var scPwd = canManage() ? toCentavos((m.querySelector("#invoice-sc-pwd-discount") || {}).value) : 0;
    var addVAT = canManage() ? toCentavos((m.querySelector("#invoice-add-vat") || {}).value) : 0;
    var total = subtotal - withholding - scPwd + addVAT;
    var vatLabel = m.querySelector("#invoice-vat-label");
    if (vatLabel) vatLabel.textContent = vatPct ? "VAT (" + vatPct + "%)" : "VAT";
    setText(m, "#invoice-subtotal", formatMoney(subtotal, currency));
    setText(m, "#invoice-vat", formatMoney(vatAmount, currency));
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

  function collectNewRecipientFields(root) {
    var nr = {};
    root.querySelectorAll("[data-new-recipient]").forEach(function (input) {
      nr[input.getAttribute("data-new-recipient")] = input.value || "";
    });
    return nr;
  }

  function collectPayload(action) {
    var m = modal();
    var recipientSelect = m.querySelector("#invoice-recipient-select");
    var recipientId = recipientSelect ? recipientSelect.value : "";
    var registeredName = ((m.querySelector("#invoice-registered-name") || {}).value || "").trim();
    var isNewRecipient = recipientId === "__new__" || (!recipientId && registeredName);
    var paymentTermsValue = parseInt((m.querySelector("#invoice-payment-terms-value") || {}).value, 10) || 0;
    var payload = {
      recipient_id: isNewRecipient ? "" : recipientId,
      new_recipient: null,
      transaction_type: (m.querySelector("#invoice-transaction-type") || {}).value || "CASH_SALES",
      recipient_registered_name: registeredName,
      notes: (m.querySelector("#invoice-notes") || {}).value || "",
      issue_date: (m.querySelector("#invoice-issue-date") || {}).value || "",
      delivery_date: (m.querySelector("#invoice-delivery-date") || {}).value || "",
      due_date: (m.querySelector("#invoice-due-date") || {}).value || "",
      payment_terms_value: paymentTermsValue,
      payment_terms_unit: (m.querySelector("#invoice-payment-terms-unit") || {}).value || "",
      withholding_tax: canManage() ? (m.querySelector("#invoice-withholding-tax") || {}).value || "" : "",
      sc_pwd_discount: canManage() ? (m.querySelector("#invoice-sc-pwd-discount") || {}).value || "" : "",
      add_vat: canManage() ? (m.querySelector("#invoice-add-vat") || {}).value || "" : "",
      action: action,
      lines: [],
    };

    if (isNewRecipient) {
      var nr = collectNewRecipientFields(m);
      if (registeredName && !((nr.name || "").trim())) {
        nr.name = registeredName;
      }
      if (registeredName) {
        nr.registered_name = registeredName;
      }
      payload.new_recipient = nr;
    }

    m.querySelectorAll(".invoice-line-row").forEach(function (row) {
      var desc = row.querySelector(".line-desc").value.trim();
      var qty = parseInt(row.querySelector(".line-qty").value, 10) || 0;
      var price = row.querySelector(".line-price").value.trim();
      var taxSel = row.querySelector(".line-tax");
      if (!desc && !price) return;
      payload.lines.push({
        product_id: row.dataset.productId || "",
        description: desc,
        unit_price: price,
        quantity: qty,
        tax_type: taxSel ? taxSel.value : "VATABLE",
      });
    });

    return payload;
  }

  function validate(payload) {
    var m = modal();
    var recipientSelect = m.querySelector("#invoice-recipient-select");
    var registeredName = ((m.querySelector("#invoice-registered-name") || {}).value || "").trim();
    if (recipientSelect && recipientSelect.value === "__new__") {
      if (!payload.new_recipient || !((payload.new_recipient.name || "").trim())) {
        return "Please enter a name for the new recipient.";
      }
    } else if (!payload.recipient_id) {
      if (!registeredName && (!payload.new_recipient || !((payload.new_recipient.name || "").trim()))) {
        return "Please select a recipient.";
      }
    }
    if (!payload.lines.length) {
      return "Please add at least one line item.";
    }
    return "";
  }

  function openPreviewModal(previewUrl) {
    if (!previewUrl || !window.htmx) return;
    window.htmx.ajax("GET", previewUrl, {
      target: "#invoice-preview-modal-container",
      swap: "innerHTML",
    });
  }

  function closeGenerateModal() {
    var container = document.getElementById("invoice-generate-modal-container");
    if (container) container.innerHTML = "";
  }

  function refreshTable(tableUrl) {
    if (window.htmx && tableUrl) {
      window.htmx.ajax("GET", tableUrl, { target: "#invoices-table-content", swap: "innerHTML" });
    }
  }

  function refreshRecipientsTable(tableUrl) {
    if (window.htmx && tableUrl) {
      window.htmx.ajax("GET", tableUrl, { target: "#invoice-recipients-table-content", swap: "innerHTML" });
    }
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
    var recipientsTableUrl = m.dataset.recipientsTableUrl;
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
        closeGenerateModal();
        refreshTable(tableUrl);
        if (data.created_recipient) {
          refreshRecipientsTable(recipientsTableUrl);
        }
        if (action === "preview" && data.preview_url) {
          openPreviewModal(data.preview_url);
        }
      })
      .catch(function () {
        showError("Network error while creating invoice.");
      });
  }

  function todayPH() {
    var m = modal();
    if (m && m.dataset.todayPh) {
      return m.dataset.todayPh;
    }
    return todayISO();
  }

  function todayISO() {
    var d = new Date();
    var m = String(d.getMonth() + 1).padStart(2, "0");
    var day = String(d.getDate()).padStart(2, "0");
    return d.getFullYear() + "-" + m + "-" + day;
  }

  function parseISODate(s) {
    if (!s) return null;
    var parts = s.split("-");
    if (parts.length !== 3) return null;
    return new Date(parseInt(parts[0], 10), parseInt(parts[1], 10) - 1, parseInt(parts[2], 10));
  }

  function formatISODate(d) {
    var m = String(d.getMonth() + 1).padStart(2, "0");
    var day = String(d.getDate()).padStart(2, "0");
    return d.getFullYear() + "-" + m + "-" + day;
  }

  function computeDueDate() {
    var m = modal();
    if (!m) return;
    var base = (m.querySelector("#invoice-issue-date") || {}).value || (m.querySelector("#invoice-delivery-date") || {}).value || todayPH();
    var value = parseInt((m.querySelector("#invoice-payment-terms-value") || {}).value, 10) || 0;
    var unit = (m.querySelector("#invoice-payment-terms-unit") || {}).value || "";
    var start = parseISODate(base);
    if (!start || value <= 0 || !unit) return;
    var due = new Date(start.getTime());
    if (unit === "DAYS") {
      due.setMonth(due.getMonth() + Math.floor(value / 30));
      due.setDate(due.getDate() + (value % 30));
    } else if (unit === "MONTHS") {
      due.setMonth(due.getMonth() + value);
    } else if (unit === "YEARS") {
      due.setFullYear(due.getFullYear() + value);
    } else {
      return;
    }
    var dueInput = m.querySelector("#invoice-due-date");
    if (dueInput) dueInput.value = formatISODate(due);
  }

  function isDueDateTrigger(el) {
    if (!el || !el.id) return false;
    return (
      el.id === "invoice-issue-date" ||
      el.id === "invoice-payment-terms-value" ||
      el.id === "invoice-payment-terms-unit"
    );
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
      var productId = (m.querySelector("#invoice-product-id") || {}).value || "";
      if (!productId) return;
      var productName = (m.querySelector("#invoice-product-name") || {}).value || "";
      var priceCentavos = parseInt((m.querySelector("#invoice-product-price-centavos") || {}).value, 10) || 0;
      var pricePesos = (priceCentavos / 100).toFixed(2);
      m.querySelector("#invoice-lines").appendChild(makeRow(productName, pricePesos, productId));
      var searchInput = m.querySelector("#invoice-product-search");
      var hiddenId = m.querySelector("#invoice-product-id");
      var hiddenName = m.querySelector("#invoice-product-name");
      var hiddenPrice = m.querySelector("#invoice-product-price-centavos");
      if (searchInput) searchInput.value = "";
      if (hiddenId) hiddenId.value = "";
      if (hiddenName) hiddenName.value = "";
      if (hiddenPrice) hiddenPrice.value = "";
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
    if (isDueDateTrigger(e.target)) {
      computeDueDate();
      return;
    }
    if (
      e.target.closest(".invoice-line-row") ||
      e.target.id === "invoice-withholding-tax" ||
      e.target.id === "invoice-sc-pwd-discount" ||
      e.target.id === "invoice-add-vat"
    ) {
      recompute();
    }
  });

  document.addEventListener("change", function (e) {
    var m = modal();
    if (!m || !m.contains(e.target)) return;
    if (isDueDateTrigger(e.target)) {
      computeDueDate();
      return;
    }
    if (e.target.id === "invoice-recipient-select") {
      var newBox = m.querySelector("#invoice-new-recipient");
      if (newBox) {
        newBox.classList.toggle("hidden", e.target.value !== "__new__");
      }
    }
    if (e.target.classList && e.target.classList.contains("line-tax")) {
      recompute();
    }
  });

  function initProductSearch() {
    var m = modal();
    if (!m || !window.AdminEntitySearch) return;
    window.AdminEntitySearch.bind(m, {
      searchInputId: "invoice-product-search",
      hiddenInputId: "invoice-product-id",
      searchUrl: m.dataset.productSearchUrl,
      onSelect: function (item) {
        var nameInput = m.querySelector("#invoice-product-name");
        var priceInput = m.querySelector("#invoice-product-price-centavos");
        if (nameInput) nameInput.value = item.name || "";
        if (priceInput) priceInput.value = String(item.price_centavos || 0);
      },
      onClear: function () {
        var nameInput = m.querySelector("#invoice-product-name");
        var priceInput = m.querySelector("#invoice-product-price-centavos");
        if (nameInput) nameInput.value = "";
        if (priceInput) priceInput.value = "";
      },
    });
  }

  document.body.addEventListener("htmx:afterSwap", function (evt) {
    if (evt.detail && evt.detail.target && evt.detail.target.id === "invoice-generate-modal-container") {
      var m = modal();
      if (m) {
        var issueDate = m.querySelector("#invoice-issue-date");
        if (issueDate && !issueDate.value) {
          issueDate.value = todayPH();
        }
        if (m.querySelector("#invoice-lines") && m.querySelectorAll(".invoice-line-row").length === 0) {
          m.querySelector("#invoice-lines").appendChild(makeRow("", "", ""));
          recompute();
        }
        initProductSearch();
      }
    }
  });
})();
