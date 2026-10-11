import { api, ApiError } from "./api.js";

const pageSize = 20;
let currentPage = 1;
let totalPages = 1;
let pendingAction = null;

const orderBody = document.querySelector("#orders-body");
const orderMessage = document.querySelector("#admin-message");
const orderFilter = document.querySelector("#order-filter");
const detailDrawer = document.querySelector("#order-detail");

const orderLabels = {
  pending_payment: "Menunggu pembayaran",
  payment_submitted: "Bukti dikirim",
  payment_verified: "Terverifikasi",
  in_progress: "Diproses",
  published: "Terbit",
  rejected: "Ditolak",
  cancelled: "Dibatalkan",
};

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function setMessage(node, text, kind = "") {
  node.textContent = text;
  node.className = `state-message${kind ? ` state-${kind}` : ""}`;
  node.hidden = false;
}

function requireLogin(error) {
  if (error instanceof ApiError && error.status === 401) {
    window.location.replace("./login.html");
    return true;
  }
  return false;
}

function money(amount, currency) {
  if (!Number.isFinite(Number(amount)) || !currency) return "—";
  try {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency,
      maximumFractionDigits: 0,
    }).format(Number(amount));
  } catch {
    return `${currency} ${Number(amount).toLocaleString("id-ID")}`;
  }
}

function actionButton(label, action, variant = "button-outline") {
  const button = element("button", `button ${variant} button-small`, label);
  button.type = "button";
  button.addEventListener("click", action);
  return button;
}

async function loadOrders() {
  setMessage(orderMessage, "Memuat pesanan…");
  document.querySelector("#orders-empty").hidden = true;
  orderBody.replaceChildren();
  try {
    const { data, meta } = await api.adminOrders({
      status: orderFilter.value,
      page: currentPage,
      limit: pageSize,
    });
    const orders = Array.isArray(data) ? data : [];
    totalPages = Math.max(1, Number(meta?.total_pages) || 1);
    currentPage = Number(meta?.page) || currentPage;
    orderMessage.hidden = true;
    orders.forEach((order) => {
      const row = document.createElement("tr");
      const reference = element("td", "order-ref", order.order_reference || "—");
      const customer = element("td", "", order.customer_name || "—");
      const design = element(
        "td",
        "order-design",
        `${order.package_slug || "—"} · ${order.template_slug || "—"}`,
      );
      const total = element("td", "", money(order.total_amount, order.currency));
      const status = element("td");
      status.append(element("span", `status-pill status-${order.status || "unknown"}`, orderLabels[order.status] || order.status || "—"));
      const actions = element("td", "table-actions");
      actions.append(actionButton("Detail", () => openOrder(order.id)));
      row.append(reference, customer, design, total, status, actions);
      orderBody.append(row);
    });
    document.querySelector("#orders-empty").hidden = orders.length > 0;
    document.querySelector("#page-label").textContent =
      `Halaman ${currentPage} dari ${totalPages} · ${Number(meta?.total) || 0} pesanan`;
    document.querySelector("#previous-page").disabled = currentPage <= 1;
    document.querySelector("#next-page").disabled = currentPage >= totalPages;
  } catch (error) {
    if (requireLogin(error)) return;
    setMessage(orderMessage, error.message, "error");
    const retry = actionButton("Coba lagi", loadOrders);
    orderMessage.append(" ", retry);
  }
}

function addDetailLine(parent, label, value) {
  const row = element("div", "detail-line");
  row.append(element("span", "", label), element("strong", "", value || "—"));
  parent.append(row);
}

function renderOrderDetail(order, payment) {
  const content = document.querySelector("#detail-content");
  content.replaceChildren();
  const heading = element("div", "detail-order-heading");
  heading.append(
    element("span", "status-pill", orderLabels[order.status] || order.status || "—"),
    element("strong", "detail-amount", money(order.total_amount, order.currency)),
  );
  content.append(heading);

  const customer = element("section", "detail-section");
  customer.append(element("h3", "", "Pelanggan & desain"));
  addDetailLine(customer, "Referensi", order.order_reference);
  addDetailLine(customer, "Nama", order.customer_name);
  addDetailLine(customer, "WhatsApp", order.customer_whatsapp);
  addDetailLine(customer, "Paket", order.package_slug);
  addDetailLine(customer, "Template", order.template_slug);
  content.append(customer);

  const paymentSection = element("section", "detail-section");
  paymentSection.append(element("h3", "", "Informasi pembayaran"));
  addDetailLine(paymentSection, "Metode", payment?.method);
  addDetailLine(paymentSection, "Status", payment?.status);
  addDetailLine(
    paymentSection,
    "Diverifikasi pada",
    payment?.verified_at ? new Date(payment.verified_at).toLocaleString("id-ID") : "Belum diverifikasi",
  );
  if (payment?.rejection_reason) addDetailLine(paymentSection, "Catatan penolakan", payment.rejection_reason);
  paymentSection.append(element("p", "detail-note", "Bukti pembayaran tidak dapat diakses dari kontrak API saat ini; tidak ada URL bukti yang ditampilkan."));
  content.append(paymentSection);

  const invitation = order.invitation && typeof order.invitation === "object" ? order.invitation : {};
  const invitationSection = element("section", "detail-section");
  invitationSection.append(element("h3", "", "Data undangan"));
  Object.entries(invitation).forEach(([key, value]) => {
    const text = typeof value === "string" || typeof value === "number" ? String(value) : JSON.stringify(value);
    addDetailLine(invitationSection, key.replaceAll("_", " "), text);
  });
  if (!invitationSection.querySelector(".detail-line")) {
    invitationSection.append(element("p", "detail-note", "Tidak ada detail undangan yang dikembalikan."));
  }
  content.append(invitationSection);

  const actions = element("section", "detail-actions");
  actions.append(element("h3", "", "Tindakan"));
  if (order.status === "payment_submitted") {
    actions.append(
      actionButton("Setujui pembayaran", () => confirmAction("approve-payment", order), "button-dark"),
      actionButton("Tolak bukti pembayaran", () => confirmAction("reject-payment", order)),
    );
  } else if (order.status === "pending_payment") {
    actions.append(actionButton("Tolak pesanan", () => confirmAction("reject-order", order)));
  } else if (["payment_verified", "in_progress"].includes(order.status)) {
    actions.append(actionButton("Terbitkan undangan", () => confirmAction("publish", order), "button-dark"));
  } else if (order.status === "published") {
    actions.append(element("p", "detail-note", "Undangan sudah diterbitkan."));
  } else {
    actions.append(element("p", "detail-note", "Tidak ada tindakan yang tersedia untuk status ini."));
  }
  content.append(actions);
}

async function openOrder(id) {
  detailDrawer.hidden = false;
  document.querySelector("#detail-title").textContent = "Memuat detail…";
  document.querySelector("#detail-content").replaceChildren();
  try {
    const { data } = await api.adminOrder(id);
    document.querySelector("#detail-title").textContent = data.order?.order_reference || `Pesanan ${id}`;
    renderOrderDetail(data.order || {}, data.payment || null);
  } catch (error) {
    if (requireLogin(error)) return;
    const content = document.querySelector("#detail-content");
    content.replaceChildren(element("p", "state-message state-error", error.message));
  }
}

function slugify(value) {
  return String(value || "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 80);
}

function confirmAction(action, order) {
  pendingAction = { action, order };
  const dialog = document.querySelector("#confirm-dialog");
  const title = document.querySelector("#confirm-title");
  const copy = document.querySelector("#confirm-copy");
  const noteLabel = document.querySelector("#confirm-note-label");
  const slugLabel = document.querySelector("#confirm-slug-label");
  const note = document.querySelector("#confirm-note");
  const slug = document.querySelector("#confirm-slug");
  note.value = "";
  slug.value = "";
  noteLabel.hidden = !["reject-payment", "reject-order"].includes(action);
  slugLabel.hidden = action !== "publish";
  const invitation = order.invitation && typeof order.invitation === "object" ? order.invitation : {};
  slug.value = slugify(`${invitation.bride_name || ""}-${invitation.groom_name || ""}`);
  const titles = {
    "approve-payment": ["Setujui pembayaran?", "Pastikan pembayaran telah diperiksa sebelum menyetujui."],
    "reject-payment": ["Tolak bukti pembayaran?", "Pesanan akan kembali menunggu pembayaran. Catatan wajib diisi."],
    "reject-order": ["Tolak pesanan?", "Tindakan ini mengubah status pesanan. Catatan wajib diisi."],
    publish: ["Terbitkan undangan?", "Undangan akan dapat diakses publik setelah berhasil diterbitkan."],
  };
  [title.textContent, copy.textContent] = titles[action];
  dialog.hidden = false;
  (action === "publish" ? slug : action.includes("reject") ? note : document.querySelector("#confirm-submit")).focus();
}

async function submitConfirmedAction() {
  if (!pendingAction) return;
  const { action, order } = pendingAction;
  const submit = document.querySelector("#confirm-submit");
  const note = document.querySelector("#confirm-note").value.trim();
  const slug = document.querySelector("#confirm-slug").value.trim();
  if (action.includes("reject") && !note) {
    document.querySelector("#confirm-note").setCustomValidity("Catatan wajib diisi untuk penolakan.");
    document.querySelector("#confirm-note").reportValidity();
    return;
  }
  document.querySelector("#confirm-note").setCustomValidity("");
  if (action === "publish" && !/^[a-z0-9_-]{3,80}$/.test(slug)) {
    document.querySelector("#confirm-slug").setCustomValidity("Gunakan minimal 3 karakter: huruf kecil, angka, tanda hubung, atau garis bawah.");
    document.querySelector("#confirm-slug").reportValidity();
    return;
  }
  document.querySelector("#confirm-slug").setCustomValidity("");
  submit.disabled = true;
  submit.textContent = "Menyimpan…";
  try {
    if (action === "approve-payment") {
      await api.verifyPayment(order.id, { decision: "approve", note: "" });
    } else if (action === "reject-payment") {
      await api.verifyPayment(order.id, { decision: "reject", note });
    } else if (action === "reject-order") {
      await api.rejectOrder(order.id, { note });
    } else {
      await api.publishOrder(order.id, { slug });
    }
    document.querySelector("#confirm-dialog").hidden = true;
    pendingAction = null;
    setMessage(orderMessage, "Perubahan berhasil disimpan oleh server.", "success");
    await loadOrders();
    await openOrder(order.id);
  } catch (error) {
    if (requireLogin(error)) return;
    setMessage(orderMessage, error.message, "error");
  } finally {
    submit.disabled = false;
    submit.textContent = "Konfirmasi";
  }
}

async function loadTemplates() {
  const message = document.querySelector("#templates-message");
  const list = document.querySelector("#admin-template-list");
  list.replaceChildren();
  setMessage(message, "Memuat metadata template…");
  try {
    const { data } = await api.adminTemplates();
    const templates = Array.isArray(data) ? data : [];
    message.hidden = true;
    document.querySelector("#templates-empty").hidden = templates.length > 0;
    templates.forEach((template) => {
      const row = element("article", "admin-template-row");
      const info = element("div");
      info.append(
        element("strong", "", template.name || template.slug),
        element("span", "template-package", `${template.package_slug || "—"} · ${template.slug || ""}`),
      );
      const availability = element(
        "span",
        `availability ${template.available ? "is-available" : "is-unavailable"}`,
        template.available ? "Tersedia" : "Tidak tersedia",
      );
      const toggle = actionButton(template.available ? "Nonaktifkan" : "Aktifkan", async () => {
        if (!window.confirm(`${template.available ? "Nonaktifkan" : "Aktifkan"} template ${template.name || template.slug}?`)) return;
        toggle.disabled = true;
        try {
          await api.setTemplateAvailability(template.id, !template.available);
          await loadTemplates();
        } catch (error) {
          if (requireLogin(error)) return;
          setMessage(message, error.message, "error");
        } finally {
          toggle.disabled = false;
        }
      });
      row.append(info, availability, toggle);
      list.append(row);
    });
  } catch (error) {
    if (requireLogin(error)) return;
    setMessage(message, error.message, "error");
    message.append(" ", actionButton("Coba lagi", loadTemplates));
  }
}

async function loadPackages() {
  const message = document.querySelector("#packages-message");
  const list = document.querySelector("#admin-package-list");
  list.replaceChildren();
  setMessage(message, "Memuat paket…");
  try {
    const { data } = await api.packages();
    const packages = Array.isArray(data) ? data : [];
    message.hidden = true;
    packages.forEach((pkg) => {
      const card = element("article", "package-card admin-package-card");
      card.append(element("p", "package-kicker", pkg.name || pkg.slug));
      card.append(element("p", "package-price", money(pkg.price, pkg.currency)));
      const features = element("ul", "feature-list");
      (Array.isArray(pkg.features) ? pkg.features : []).forEach((feature) => {
        features.append(element("li", "", String(feature).replaceAll("_", " ")));
      });
      card.append(features);
      list.append(card);
    });
    if (!packages.length) setMessage(message, "Belum ada paket aktif.", "empty");
  } catch (error) {
    if (requireLogin(error)) return;
    setMessage(message, error.message, "error");
  }
}

document.querySelector("#order-filter").addEventListener("change", () => {
  currentPage = 1;
  loadOrders();
});
document.querySelector("#previous-page").addEventListener("click", () => {
  if (currentPage > 1) {
    currentPage -= 1;
    loadOrders();
  }
});
document.querySelector("#next-page").addEventListener("click", () => {
  if (currentPage < totalPages) {
    currentPage += 1;
    loadOrders();
  }
});
document.querySelector("#refresh-button").addEventListener("click", () => {
  loadOrders();
  if (!document.querySelector("#templates-panel").hidden) loadTemplates();
});
document.querySelector("#close-detail").addEventListener("click", () => {
  detailDrawer.hidden = true;
});
document.querySelector("#confirm-cancel").addEventListener("click", () => {
  document.querySelector("#confirm-dialog").hidden = true;
  pendingAction = null;
});
document.querySelector("#confirm-submit").addEventListener("click", submitConfirmedAction);
document.querySelectorAll(".admin-tab").forEach((tab) => {
  tab.addEventListener("click", () => {
    document.querySelectorAll(".admin-tab").forEach((item) => item.classList.toggle("is-active", item === tab));
    ["orders", "templates", "packages"].forEach((name) => {
      document.querySelector(`#${name}-panel`).hidden = tab.dataset.tab !== name;
    });
    if (tab.dataset.tab === "templates") loadTemplates();
    if (tab.dataset.tab === "packages") loadPackages();
  });
});
document.querySelector("#logout-button").addEventListener("click", async (event) => {
  const button = event.currentTarget;
  button.disabled = true;
  try {
    await api.adminLogout();
    window.location.replace("./login.html");
  } catch (error) {
    if (requireLogin(error)) return;
    setMessage(orderMessage, error.message, "error");
  } finally {
    button.disabled = false;
  }
});

loadOrders();
