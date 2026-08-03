(function () {
  "use strict";

  function modal() {
    return document.getElementById("delivery-receipt-generate-modal");
  }

  function makeRow(qty, unit, description) {
    var tr = document.createElement("tr");
    tr.className = "delivery-receipt-line-row border-t border-gray-100";
    tr.innerHTML =
      '<td class="px-3 py-2"><input type="text" class="line-qty px-2 py-1 border border-gray-300 rounded w-full text-right" placeholder="1"></td>' +
      '<td class="px-3 py-2"><input type="text" class="line-unit px-2 py-1 border border-gray-300 rounded w-full text-right" placeholder="pc"></td>' +
      '<td class="px-3 py-2"><input type="text" class="line-desc px-2 py-1 border border-gray-300 rounded w-full" placeholder="Item description"></td>' +
      '<td class="px-3 py-2 text-center"><button type="button" class="line-remove text-red-600 hover:text-red-800 font-bold">&times;</button></td>';
    tr.querySelector(".line-qty").value = qty || "1";
    tr.querySelector(".line-unit").value = unit || "";
    tr.querySelector(".line-desc").value = description || "";
    return tr;
  }

  function showError(msg) {
    var m = modal();
    if (!m) return;
    var box = m.querySelector("#delivery-receipt-generate-error");
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
    var payload = {
      invoice_id: (m.querySelector("#delivery-receipt-invoice-id") || {}).value || "",
      delivered_to: ((m.querySelector("#delivery-receipt-delivered-to") || {}).value || "").trim(),
      recipient_email: (m.querySelector("#delivery-receipt-email") || {}).value || "",
      tin: (m.querySelector("#delivery-receipt-tin") || {}).value || "",
      address: (m.querySelector("#delivery-receipt-address") || {}).value || "",
      receipt_date: (m.querySelector("#delivery-receipt-date") || {}).value || "",
      terms: (m.querySelector("#delivery-receipt-terms") || {}).value || "",
      po_number: (m.querySelector("#delivery-receipt-po-number") || {}).value || "",
      action: action,
      lines: [],
    };

    m.querySelectorAll(".delivery-receipt-line-row").forEach(function (row) {
      var desc = row.querySelector(".line-desc").value.trim();
      if (!desc) return;
      payload.lines.push({
        quantity: row.querySelector(".line-qty").value.trim() || "1",
        unit: row.querySelector(".line-unit").value.trim(),
        description: desc,
      });
    });

    return payload;
  }

  function validate(payload) {
    if (!payload.delivered_to) {
      return "Please enter the delivered-to name.";
    }
    if (!payload.lines.length) {
      return "Please add at least one line item.";
    }
    return "";
  }

  function closeGenerateModal() {
    var container = document.getElementById("delivery-receipt-generate-modal-container");
    if (container) container.innerHTML = "";
  }

  function refreshTable(tableUrl) {
    if (window.htmx && tableUrl) {
      window.htmx.ajax("GET", tableUrl, { target: "#delivery-receipts-table-content", swap: "innerHTML" });
    }
  }

  function openPreviewModal(previewUrl) {
    if (!previewUrl || !window.htmx) return;
    window.htmx.ajax("GET", previewUrl, {
      target: "#delivery-receipt-preview-modal-container",
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
      target: "#delivery-receipt-generate-modal-container",
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
          showError((res.data && res.data.error) || "Failed to create delivery receipt.");
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
        showError("Network error while creating delivery receipt.");
      });
  }

  document.addEventListener("click", function (e) {
    var m = modal();
    if (!m) return;

    if (e.target.closest("#delivery-receipt-add-line")) {
      e.preventDefault();
      m.querySelector("#delivery-receipt-lines").appendChild(makeRow("", "pc", ""));
      return;
    }

    var removeBtn = e.target.closest(".line-remove");
    if (removeBtn && m.contains(removeBtn)) {
      e.preventDefault();
      var row = removeBtn.closest(".delivery-receipt-line-row");
      if (row) row.remove();
      return;
    }

    var actionBtn = e.target.closest("[data-delivery-receipt-action]");
    if (actionBtn && m.contains(actionBtn)) {
      e.preventDefault();
      submit(actionBtn.getAttribute("data-delivery-receipt-action"));
    }
  });

  function initInvoiceSearch() {
    var m = modal();
    if (!m || !window.AdminEntitySearch) return;
    window.AdminEntitySearch.bind(m, {
      searchInputId: "delivery-receipt-invoice-search",
      hiddenInputId: "delivery-receipt-invoice-id",
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
    if (evt.detail && evt.detail.target && evt.detail.target.id === "delivery-receipt-generate-modal-container") {
      var m = modal();
      if (m && m.querySelector("#delivery-receipt-lines") && m.querySelectorAll(".delivery-receipt-line-row").length === 0) {
        m.querySelector("#delivery-receipt-lines").appendChild(makeRow("1", "pc", ""));
      }
      initInvoiceSearch();
    }
  });
})();
