import { api, ApiError } from "./api.js";

const packageList = document.querySelector("#package-list");
const packageState = document.querySelector("#package-state");
const templateList = document.querySelector("#template-list");
const templateState = document.querySelector("#template-state");
const packageFilter = document.querySelector("#package-filter");
const orderDialog = document.querySelector("#order-dialog");
const orderForm = document.querySelector("#order-form");
const orderError = document.querySelector("#order-error");
const orderFormView = document.querySelector("#order-form-view");
const orderSuccess = document.querySelector("#order-success");
const previewButton = document.querySelector("#preview-order-button");
const previewFeedback = document.querySelector("#preview-feedback");
const previewPanel = document.querySelector("#preview-panel");
const previewFrame = document.querySelector("#template-preview-frame");

let packages = [];
let orderIdempotencyKey = "";
let selectedTemplate = null;
let activePreviewWindow = null;
let activePreviewOrigin = "";
let activePreviewData = null;

function setState(element, message, type = "") {
  element.textContent = message;
  element.className = `state-message${type ? ` state-${type}` : ""}`;
  element.hidden = false;
}

function formatPrice(amount, currency) {
  if (!Number.isFinite(Number(amount)) || !currency) return "Harga dikonfirmasi server";
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

function make(tag, className, text) {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text !== undefined) element.textContent = text;
  return element;
}

function setButtonLabel(button, label) {
  const arrow = document.createElement("span");
  arrow.setAttribute("aria-hidden", "true");
  arrow.textContent = "↗";
  button.replaceChildren(document.createTextNode(`${label} `), arrow);
}

function displayPackages(items, { admin = false } = {}) {
  packageList.replaceChildren();
  if (!items.length) {
    setState(packageState, "Belum ada paket aktif. Coba lagi nanti.", "empty");
    return;
  }
  packageState.hidden = true;
  items.forEach((pkg, index) => {
    const card = make("article", `package-card${index === 1 ? " package-featured" : ""}`);
    const top = make("div", "package-top");
    top.append(make("p", "package-kicker", pkg.name || pkg.slug));
    if (index === 1) top.append(make("span", "package-badge", "Paling dipilih"));
    card.append(top);
    card.append(make("p", "package-price", formatPrice(pkg.price, pkg.currency)));
    const features = make("ul", "feature-list");
    (Array.isArray(pkg.features) ? pkg.features : []).forEach((feature) => {
      features.append(make("li", "", String(feature).replaceAll("_", " ")));
    });
    if (!features.children.length) features.append(make("li", "", "Fitur paket dikonfigurasi oleh server"));
    card.append(features);
    if (!admin) {
      const choose = make("a", "text-link package-choose", "Lihat desain paket ↗");
      choose.href = `#templates`;
      choose.addEventListener("click", () => {
        packageFilter.value = pkg.slug;
        loadTemplates(pkg.slug);
      });
      card.append(choose);
    }
    packageList.append(card);
  });
}

function appendTemplateCard(template) {
  const card = make("article", "template-card");
  const art = make("div", `template-art art-${(template.package_slug || "silver").toLowerCase()}`);
  const label = make("span", "template-art-label", (template.package_slug || "DESIGN").toUpperCase());
  const monogram = make("span", "template-monogram");
  monogram.append(document.createTextNode("A "), make("em", "", "&"), document.createTextNode(" B"));
  art.append(label, monogram);
  const content = make("div", "template-card-content");
  const title = make("h3", "", template.name || template.slug);
  const subtitle = make("p", "template-package", (template.package_slug || "Paket").toUpperCase());
  content.append(title, subtitle);
  const actions = make("div", "template-actions");
  const demoUrl = resolveDemoUrl(template.demo_url);
  if (demoUrl) {
    const demo = make("a", "button button-outline button-small", "Preview ↗");
    demo.href = demoUrl.href;
    demo.target = "_blank";
    demo.rel = "noopener noreferrer";
    demo.setAttribute("aria-label", `Buka preview ${template.name || template.slug} di tab baru`);
    actions.append(demo);
  } else {
    const unavailable = make("span", "preview-unavailable", "Preview belum tersedia");
    actions.append(unavailable);
  }
  const order = make("button", "button button-dark button-small", "Pilih desain");
  order.type = "button";
  order.disabled = template.available === false;
  if (order.disabled) order.textContent = "Tidak tersedia";
  order.addEventListener("click", () => openOrder(template));
  actions.append(order);
  content.append(actions);
  card.append(art, content);
  templateList.append(card);
}

async function loadPackages() {
  setState(packageState, "Memuat pilihan paket…");
  try {
    const result = await api.packages();
    packages = Array.isArray(result.data) ? result.data : [];
    displayPackages(packages);
    packageFilter.replaceChildren(new Option("Semua paket", ""));
    packages.forEach((pkg) => packageFilter.add(new Option(pkg.name || pkg.slug, pkg.slug)));
  } catch (error) {
    setState(packageState, error.message, "error");
    const retry = make("button", "button button-outline button-small retry-button", "Coba lagi");
    retry.type = "button";
    retry.addEventListener("click", loadPackages);
    packageState.append(" ", retry);
  }
}

async function loadTemplates(packageSlug = "") {
  setState(templateState, "Memuat koleksi template…");
  templateList.replaceChildren();
  try {
    const result = await api.templates(packageSlug);
    const templates = Array.isArray(result.data) ? result.data : [];
    if (!templates.length) {
      setState(templateState, "Belum ada template yang tersedia untuk pilihan ini.", "empty");
      return;
    }
    templateState.hidden = true;
    templates.forEach(appendTemplateCard);
  } catch (error) {
    setState(templateState, error.message, "error");
    const retry = make("button", "button button-outline button-small retry-button", "Coba lagi");
    retry.type = "button";
    retry.addEventListener("click", () => loadTemplates(packageSlug));
    templateState.append(" ", retry);
  }
}

function openOrder(template) {
  selectedTemplate = template;
  orderIdempotencyKey = createIdempotencyKey();
  const form = orderForm.elements;
  orderForm.reset();
  form.package_slug.value = template.package_slug || "";
  form.template_slug.value = template.slug || "";
  const selectedPackage = packages.find((pkg) => pkg.slug === template.package_slug);
  document.querySelector("#selected-design").textContent =
    `${template.name || template.slug} · ${template.package_slug || "Paket"} · ${
      selectedPackage ? formatPrice(selectedPackage.price, selectedPackage.currency) : "Harga dikonfirmasi server"
    }`;
  orderError.hidden = true;
  previewFeedback.hidden = true;
  previewPanel.hidden = true;
  previewFrame.removeAttribute("src");
  activePreviewWindow = null;
  activePreviewOrigin = "";
  activePreviewData = null;
  previewButton.disabled = true;
  document.querySelector("#template-fields-message").hidden = true;
  document.querySelector("#invitation-fields").replaceChildren();
  orderFormView.hidden = false;
  orderSuccess.hidden = true;
  orderDialog.showModal();
  loadTemplateFields(template.slug);
}

function isSafeDemoUrl(value) {
  return Boolean(resolveDemoUrl(value));
}

function resolveDemoUrl(value) {
  if (typeof value !== "string" || !value.trim()) return null;
  try {
    const apiBase = new URL(window.SUKAMOTO_CONFIG?.apiBaseUrl || "/api/v1", window.location.href);
    const url = new URL(value, apiBase);
    if (!["http:", "https:"].includes(url.protocol) || url.origin !== apiBase.origin) return null;
    return url;
  } catch {
    return null;
  }
}

function invitationFieldLabel(field) {
  const labels = {
    bride_name: "Nama mempelai wanita",
    groom_name: "Nama mempelai pria",
    event_date: "Tanggal acara",
    venue: "Lokasi acara",
    quote: "Kutipan",
    gallery: "Galeri (URL gambar)",
    story: "Cerita pasangan",
    events: "Detail acara",
    guest_name: "Nama tamu (opsional)",
  };
  return labels[field] || field.replaceAll("_", " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}

async function loadTemplateFields(slug) {
  const container = document.querySelector("#invitation-fields");
  const message = document.querySelector("#template-fields-message");
  orderForm.querySelector('[type="submit"]').disabled = true;
  message.textContent = "Memuat field undangan template…";
  message.className = "state-message";
  message.hidden = false;
  try {
    const { data } = await api.template(slug);
    if (orderForm.elements.template_slug.value !== slug || !orderDialog.open) return;
    const supported = Array.isArray(data.supported_fields)
      ? [...new Set(data.supported_fields.filter((field) =>
          typeof field === "string" && /^[a-zA-Z][a-zA-Z0-9_]*$/.test(field),
        ))]
      : [];
    if (!supported.length) {
      setState(message, "Template ini belum memiliki metadata field formulir. Pesanan tidak dapat dibuat sampai metadata tersedia.", "error");
      orderForm.querySelector('[type="submit"]').disabled = true;
      return;
    }
    container.replaceChildren();
    for (const field of supported) {
      const label = document.createElement("label");
      label.append(document.createTextNode(invitationFieldLabel(field)));
      const input = ["story", "quote", "events"].includes(field)
        ? document.createElement("textarea")
        : document.createElement("input");
      input.name = `invitation.${field}`;
      input.maxLength = 2000;
      input.required = ["bride_name", "groom_name"].includes(field);
      if (field.endsWith("_date") || field === "event_date") input.type = "date";
      if (field === "gallery" || field === "events") {
        input.removeAttribute("type");
        input.placeholder = field === "gallery"
          ? "https://… (pisahkan beberapa URL dengan baris baru)"
          : "Nama acara, tanggal dan lokasi";
      }
      label.append(input);
      container.append(label);
    }
    orderForm.querySelector('[type="submit"]').disabled = false;
    previewButton.disabled = !selectedTemplate || !isSafeDemoUrl(selectedTemplate.demo_url);
    message.hidden = true;
  } catch (error) {
    if (orderForm.elements.template_slug.value !== slug || !orderDialog.open) return;
    setState(message, error.message, "error");
    orderForm.querySelector('[type="submit"]').disabled = true;
  }
}

function collectInvitation(formData = new FormData(orderForm)) {
  const invitation = {};
  const events = [];
  for (const [key, rawValue] of formData.entries()) {
    if (!key.startsWith("invitation.")) continue;
    const field = key.slice("invitation.".length);
    const value = String(rawValue).trim();
    if (!value) continue;
    if (field === "events") {
      events.push(...value.split(/\r?\n/).map((description) => description.trim()).filter(Boolean)
        .map((description) => ({ description })));
    } else {
      invitation[field] = field === "gallery"
        ? value.split(/\r?\n/).map((item) => item.trim()).filter(Boolean)
        : value;
    }
  }
  const eventDate = invitation.event_date || "";
  const venue = invitation.venue || "";
  if (eventDate || venue) events.unshift({ date: eventDate, venue });
  invitation.events = events;
  return invitation;
}

window.addEventListener("message", (event) => {
  if (
    event.data?.type !== "SUKAMOTO_TEMPLATE_PREVIEW_READY" ||
    event.source !== activePreviewWindow ||
    event.origin !== activePreviewOrigin
  ) return;
  activePreviewWindow.postMessage({ type: "SUKAMOTO_TEMPLATE_PREVIEW_DATA", invitation: activePreviewData }, activePreviewOrigin);
  setState(previewFeedback, "Pratinjau siap. Data hanya dikirim ke template untuk sesi ini.", "success");
});

previewButton.addEventListener("click", () => {
  const previewURL = selectedTemplate && resolveDemoUrl(selectedTemplate.demo_url);
  if (!previewURL) return;
  const requiredFields = document.querySelectorAll(
    "#invitation-fields input[required], #invitation-fields textarea[required]",
  );
  for (const field of requiredFields) {
    if (!field.reportValidity()) return;
  }
  previewURL.searchParams.set("preview", "1");
  activePreviewOrigin = previewURL.origin;
  activePreviewData = collectInvitation(new FormData(orderForm));
  activePreviewWindow = previewFrame.contentWindow;
  previewPanel.hidden = false;
  previewFrame.src = previewURL.href;
  setState(previewFeedback, "Memuat pratinjau personal…", "success");
});

function createIdempotencyKey() {
  if (typeof crypto.randomUUID === "function") return crypto.randomUUID();
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
}

function showReceipt(order, instructions) {
  document.querySelector("#receipt-reference").textContent = order.order_reference || "";
  document.querySelector("#receipt-total").textContent = formatPrice(order.total_amount, order.currency);
  const instructionText = [instructions?.method, instructions?.description].filter(Boolean).join(" — ");
  document.querySelector("#receipt-instructions").textContent =
    instructionText || "Instruksi belum tersedia. Silakan hubungi admin.";
  const statusLink = document.querySelector("#receipt-status-link");
  statusLink.href = `./status.html?reference=${encodeURIComponent(order.order_reference || "")}`;
  orderFormView.hidden = true;
  orderSuccess.hidden = false;
}

orderForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!orderForm.reportValidity()) return;
  const submitButton = orderForm.querySelector('[type="submit"]');
  submitButton.disabled = true;
  submitButton.textContent = "Mengirim pesanan…";
  orderError.hidden = true;
  const form = new FormData(orderForm);
  const invitation = collectInvitation(form);

  try {
    if (!window.confirm("Buat pesanan dengan paket, desain, dan data undangan yang telah diisi?")) return;
    const result = await api.createOrder({
      package_slug: String(form.get("package_slug")),
      template_slug: String(form.get("template_slug")),
      customer: {
        name: String(form.get("customer_name")).trim(),
        whatsapp: String(form.get("customer_whatsapp")).trim(),
      },
      invitation,
      idempotency_key: orderIdempotencyKey,
    });
    orderIdempotencyKey = "";
    showReceipt(result.data, result.data.payment_instructions);
  } catch (error) {
    const fieldMessages = Object.values(error.fields || {}).filter(Boolean);
    orderError.textContent = fieldMessages.join(" ") || error.message;
    orderError.hidden = false;
  } finally {
    submitButton.disabled = false;
    setButtonLabel(submitButton, "Buat pesanan");
  }
});

packageFilter.addEventListener("change", () => loadTemplates(packageFilter.value));
document.querySelector(".dialog-close").addEventListener("click", () => orderDialog.close());
orderDialog.addEventListener("click", (event) => {
  if (event.target === orderDialog) orderDialog.close();
});
orderDialog.addEventListener("close", () => {
  orderForm.querySelector('[type="submit"]').disabled = false;
  previewFrame.removeAttribute("src");
  previewPanel.hidden = true;
  activePreviewWindow = null;
  activePreviewOrigin = "";
  activePreviewData = null;
  if (!orderSuccess.hidden) orderForm.reset();
});

if (orderForm && packageList && templateList) {
  loadPackages();
  loadTemplates();
}
