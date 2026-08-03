(function () {
  "use strict";

  function modal() {
    return document.getElementById("collection-receipt-generate-modal");
  }

  function toCentavos(str) {
    if (!str) return 0;
    var cleaned = ("" + str).replace(/,/g, "").trim();
    var f = parseFloat(cleaned);
    if (isNaN(f)) return 0;
    return Math.round(f * 100);
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

  function makeSettlementRow(invoiceId, invoiceNumber, amount) {
    var tr = document.createElement("tr");
    tr.className = "collection-receipt-settlement-row border-t border-gray-100";
    if (invoiceId) tr.dataset.invoiceId = invoiceId;
    tr.innerHTML =
      '<td class="px-3 py-2"><input type="text" class="settlement-invoice px-2 py-1 border border-gray-300 rounded w-full" placeholder="Invoice number"></td>' +
      '<td class="px-3 py-2"><input type="text" class="settlement-amount px-2 py-1 border border-gray-300 rounded w-full text-right" placeholder="0.00"></td>' +
      '<td class="px-3 py-2 text-center"><button type="button" class="settlement-remove text-red-600 hover:text-red-800 font-bold">&times;</button></td>';
    tr.querySelector(".settlement-invoice").value = invoiceNumber || "";
    tr.querySelector(".settlement-amount").value = amount || "";
    return tr;
  }

  function recomputeAmountFromSettlements() {
    var m = modal();
    if (!m) return;
    var currency = m.dataset.currency || "PHP";
    var total = 0;
    m.querySelectorAll(".collection-receipt-settlement-row").forEach(function (row) {
      total += toCentavos(row.querySelector(".settlement-amount").value);
    });
    if (total > 0) {
      var amountInput = m.querySelector("#collection-receipt-amount");
      if (amountInput) amountInput.value = formatMoney(total, currency).replace(currency + " ", "");
    } else {
      var amountInputZero = m.querySelector("#collection-receipt-amount");
      if (amountInputZero) amountInputZero.value = "";
    }
  }

  function syncAmountToSingleSettlement() {
    var m = modal();
    if (!m) return;
    var rows = m.querySelectorAll(".collection-receipt-settlement-row");
    if (rows.length !== 1) return;
    var amountInput = m.querySelector("#collection-receipt-amount");
    if (!amountInput) return;
    var settlementAmount = rows[0].querySelector(".settlement-amount");
    if (settlementAmount) settlementAmount.value = amountInput.value;
  }

  function showError(msg) {
    var m = modal();
    if (!m) return;
    var box = m.querySelector("#collection-receipt-generate-error");
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
    var invoiceSelect = m.querySelector("#collection-receipt-invoice-id");
    var payload = {
      invoice_id: invoiceSelect ? invoiceSelect.value : "",
      received_from: ((m.querySelector("#collection-receipt-received-from") || {}).value || "").trim(),
      recipient_email: (m.querySelector("#collection-receipt-email") || {}).value || "",
      tin: (m.querySelector("#collection-receipt-tin") || {}).value || "",
      address: (m.querySelector("#collection-receipt-address") || {}).value || "",
      receipt_date: (m.querySelector("#collection-receipt-date") || {}).value || "",
      amount: (m.querySelector("#collection-receipt-amount") || {}).value || "",
      payment_for: (m.querySelector("#collection-receipt-payment-for") || {}).value || "",
      payment_form: (m.querySelector("#collection-receipt-payment-form") || {}).value || "CASH",
      sc_citizen_tin: (m.querySelector("#collection-receipt-sc-tin") || {}).value || "",
      osca_pwd_id_no: (m.querySelector("#collection-receipt-osca-pwd") || {}).value || "",
      action: action,
      settlements: [],
    };

    m.querySelectorAll(".collection-receipt-settlement-row").forEach(function (row) {
      var invoiceNumber = row.querySelector(".settlement-invoice").value.trim();
      var amount = row.querySelector(".settlement-amount").value.trim();
      if (!invoiceNumber && !amount) return;
      payload.settlements.push({
        invoice_id: row.dataset.invoiceId || "",
        invoice_number: invoiceNumber,
        amount: amount,
      });
    });

    return payload;
  }

  function validate(payload) {
    if (!payload.received_from) {
      return "Please enter the received-from name.";
    }
    if (!payload.amount) {
      return "Please enter the receipt amount.";
    }
    return "";
  }

  function closeGenerateModal() {
    var container = document.getElementById("collection-receipt-generate-modal-container");
    if (container) container.innerHTML = "";
  }

  function refreshTable(tableUrl) {
    if (window.htmx && tableUrl) {
      window.htmx.ajax("GET", tableUrl, { target: "#collection-receipts-table-content", swap: "innerHTML" });
    }
  }

  function openPreviewModal(previewUrl) {
    if (!previewUrl || !window.htmx) return;
    window.htmx.ajax("GET", previewUrl, {
      target: "#collection-receipt-preview-modal-container",
      swap: "innerHTML",
    });
  }

  function reloadModalWithInvoice(invoiceId) {
    var m = modal();
    if (!m || !window.htmx) return;
    var baseUrl = m.dataset.generateUrl;
    if (!baseUrl) return;
    var url = invoiceId ? baseUrl + "?invoice_id=" + encodeURIComponent(invoiceId) : baseUrl;
    window.htmx.ajax("GET", url, {
      target: "#collection-receipt-generate-modal-container",
      swap: "innerHTML",
    });
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

    fetch(m.dataset.createUrl, {
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
          showError((res.data && res.data.error) || "Failed to create collection receipt.");
          return;
        }
        var data = res.data;
        closeGenerateModal();
        refreshTable(m.dataset.tableUrl);
        if (action === "preview" && data.preview_url) {
          openPreviewModal(data.preview_url);
        }
      })
      .catch(function () {
        showError("Network error while creating collection receipt.");
      });
  }

  document.addEventListener("click", function (e) {
    var m = modal();
    if (!m) return;

    if (e.target.closest("#collection-receipt-add-settlement")) {
      e.preventDefault();
      m.querySelector("#collection-receipt-settlements").appendChild(makeSettlementRow("", "", ""));
      return;
    }

    var removeBtn = e.target.closest(".settlement-remove");
    if (removeBtn && m.contains(removeBtn)) {
      e.preventDefault();
      var row = removeBtn.closest(".collection-receipt-settlement-row");
      if (row) row.remove();
      recomputeAmountFromSettlements();
      return;
    }

    var actionBtn = e.target.closest("[data-collection-receipt-action]");
    if (actionBtn && m.contains(actionBtn)) {
      e.preventDefault();
      submit(actionBtn.getAttribute("data-collection-receipt-action"));
    }
  });

  document.addEventListener("input", function (e) {
    var m = modal();
    if (!m || !m.contains(e.target)) return;
    if (e.target.classList && e.target.classList.contains("settlement-amount")) {
      recomputeAmountFromSettlements();
      return;
    }
    if (e.target.id === "collection-receipt-amount") {
      syncAmountToSingleSettlement();
    }
  });

  function initInvoiceSearch() {
    var m = modal();
    if (!m || !window.AdminEntitySearch) return;
    window.AdminEntitySearch.bind(m, {
      searchInputId: "collection-receipt-invoice-search",
      hiddenInputId: "collection-receipt-invoice-id",
      searchUrl: m.dataset.invoiceSearchUrl,
      onSelect: function (item) {
        reloadModalWithInvoice(item.id);
      },
      onClear: function () {
        reloadModalWithInvoice("");
      },
    });
  }

  document.body.addEventListener("htmx:afterSwap", function (evt) {
    if (evt.detail && evt.detail.target && evt.detail.target.id === "collection-receipt-generate-modal-container") {
      var m = modal();
      if (m && m.querySelector("#collection-receipt-settlements") && m.querySelectorAll(".collection-receipt-settlement-row").length === 0) {
        m.querySelector("#collection-receipt-settlements").appendChild(makeSettlementRow("", "", ""));
      }
      initInvoiceSearch();
    }
  });
})();
