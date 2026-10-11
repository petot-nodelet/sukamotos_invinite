import { api } from "./api.js";

const statusForm = document.querySelector("#status-form");
const paymentForm = document.querySelector("#payment-form");
const message = document.querySelector("#status-message");
const resultCard = document.querySelector("#status-result");
let currentReference = "";
let currentOrderStatus = null;

const orderLabels = {
  pending_payment: "Menunggu pembayaran",
  payment_submitted: "Bukti pembayaran dikirim",
  payment_verified: "Pembayaran terverifikasi",
  in_progress: "Sedang diproses",
  published: "Undangan diterbitkan",
  rejected: "Pesanan ditolak",
  cancelled: "Pesanan dibatalkan",
};
const paymentLabels = {
  unpaid: "Belum dibayar",
  submitted: "Menunggu verifikasi",
  verified: "Terverifikasi",
  rejected: "Ditolak — bukti dapat dikirim ulang",
};

function setMessage(text, type = "") {
  message.textContent = text;
  message.className = `state-message${type ? ` state-${type}` : ""}`;
  message.hidden = false;
}

function restoreButton(button, label) {
  const arrow = document.createElement("span");
  arrow.setAttribute("aria-hidden", "true");
  arrow.textContent = "↗";
  button.replaceChildren(document.createTextNode(`${label} `), arrow);
}

function canUpload(order) {
  return order.status === "pending_payment" && ["unpaid", "rejected"].includes(order.payment_status);
}

function isSafeInvitationUrl(value) {
  if (typeof value !== "string" || !value) return false;
  try {
    return ["http:", "https:"].includes(new URL(value, window.location.origin).protocol);
  } catch {
    return false;
  }
}

async function checkStatus(reference) {
  currentReference = reference.trim();
  if (!currentReference) {
    setMessage("Masukkan referensi pesanan untuk melanjutkan.", "error");
    return;
  }
  resultCard.hidden = true;
  setMessage("Memeriksa status terbaru…");
  const checkButton = statusForm.querySelector('[type="submit"]');
  checkButton.disabled = true;
  try {
    const { data } = await api.orderStatus(currentReference);
    currentOrderStatus = data;
    document.querySelector("#status-reference").textContent = data.order_reference || currentReference;
    const orderStatus = data.status || "unknown";
    const paymentStatus = data.payment_status || "unknown";
    document.querySelector("#status-pill").textContent = orderLabels[orderStatus] || "Status diperbarui";
    document.querySelector("#status-pill").className = `status-pill status-${orderStatus}`;
    document.querySelector("#payment-status").textContent = paymentLabels[paymentStatus] || "Informasi tidak tersedia";
    document.querySelector("#invitation-status").textContent =
      data.published === true ? "Sudah diterbitkan" : "Belum diterbitkan";
    document.querySelector("#status-explanation").textContent =
      data.published === true
        ? "Undangan Anda sudah siap dibagikan."
        : orderStatus === "payment_submitted"
          ? "Bukti pembayaran telah diterima. Admin akan memeriksanya."
          : orderStatus === "payment_verified"
            ? "Pembayaran terverifikasi. Undangan menunggu proses penerbitan."
            : orderStatus === "pending_payment"
              ? "Pesanan tercatat. Selesaikan pembayaran dan unggah bukti melalui formulir di bawah."
              : "Hubungi admin jika Anda membutuhkan bantuan terkait pesanan ini.";
    const invitationLink = document.querySelector("#invitation-link");
    invitationLink.hidden = !(data.published === true && isSafeInvitationUrl(data.invitation_url));
    if (!invitationLink.hidden) {
      invitationLink.href = data.invitation_url;
      invitationLink.target = "_blank";
      invitationLink.rel = "noopener noreferrer";
    }
    document.querySelector("#payment-panel").hidden = !canUpload(data);
    paymentForm.reset();
    resultCard.hidden = false;
    message.hidden = true;
    return true;
  } catch (error) {
    setMessage(error.message, "error");
    return false;
  } finally {
    checkButton.disabled = false;
    restoreButton(checkButton, "Periksa status");
  }
}

statusForm.addEventListener("submit", (event) => {
  event.preventDefault();
  if (!statusForm.reportValidity()) return;
  checkStatus(statusForm.elements.reference.value);
});

paymentForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!currentOrderStatus || !canUpload(currentOrderStatus) || !paymentForm.reportValidity()) return;
  const button = paymentForm.querySelector('[type="submit"]');
  const error = document.querySelector("#payment-error");
  if (!window.confirm("Kirim bukti pembayaran ini untuk diperiksa admin?")) return;
  button.disabled = true;
  button.textContent = "Mengirim bukti…";
  error.hidden = true;
  const body = new FormData();
  body.set("method", String(paymentForm.elements.method.value).trim());
  body.set("proof", paymentForm.elements.proof.files[0]);
  try {
    await api.uploadPayment(currentReference, body);
    setMessage("Bukti pembayaran berhasil dikirim. Status akan diperbarui setelah diperiksa admin.", "success");
    paymentForm.reset();
    const refreshed = await checkStatus(currentReference);
    if (refreshed) {
      setMessage("Bukti pembayaran berhasil dikirim. Status akan diperbarui setelah diperiksa admin.", "success");
    }
  } catch (uploadError) {
    error.textContent = uploadError.message;
    error.hidden = false;
  } finally {
    button.disabled = false;
    restoreButton(button, "Kirim bukti");
  }
});

const initialReference = new URLSearchParams(window.location.search).get("reference");
if (initialReference) {
  statusForm.elements.reference.value = initialReference;
  checkStatus(initialReference);
}
