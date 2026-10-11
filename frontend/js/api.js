const API_BASE_URL = new URL(
  window.SUKAMOTO_CONFIG?.apiBaseUrl || "/api/v1",
  window.location.origin,
).toString().replace(/\/$/, "");

const REQUEST_TIMEOUT_MS = 15000;

function csrfToken() {
  const cookieName = window.SUKAMOTO_CONFIG?.csrfCookieName || "sukamoto_csrf";
  const prefix = `${encodeURIComponent(cookieName)}=`;
  const cookie = document.cookie.split("; ").find((item) => item.startsWith(prefix));
  return cookie ? decodeURIComponent(cookie.slice(prefix.length)) : "";
}

export class ApiError extends Error {
  constructor(message, { status = 0, code = "NETWORK_ERROR", fields = {} } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

async function request(path, { method = "GET", body, signal, headers = {}, timeoutMs = REQUEST_TIMEOUT_MS } = {}) {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), timeoutMs);
  const abortFromCaller = () => controller.abort();
  signal?.addEventListener("abort", abortFromCaller, { once: true });

  const requestHeaders = new Headers(headers);
  let payload = body;
  if (body instanceof FormData) {
    payload = body;
  } else if (body !== undefined) {
    requestHeaders.set("Content-Type", "application/json");
    payload = JSON.stringify(body);
  }
  requestHeaders.set("Accept", "application/json");
  const csrf = csrfToken();
  if (csrf && !["GET", "HEAD", "OPTIONS"].includes(method.toUpperCase())) {
    requestHeaders.set("X-CSRF-Token", csrf);
  }

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method,
      body: payload,
      headers: requestHeaders,
      credentials: "include",
      cache: "no-store",
      signal: controller.signal,
    });

    if (response.status === 204) return null;
    const contentType = response.headers.get("content-type") || "";
    let result;
    if (contentType.includes("application/json")) {
      try {
        result = await response.json();
      } catch {
        throw new ApiError("Respons server tidak dapat dibaca.", {
          status: response.status,
          code: "INVALID_RESPONSE",
        });
      }
    } else {
      throw new ApiError("Server mengirim respons yang tidak sesuai.", {
        status: response.status,
        code: "INVALID_RESPONSE",
      });
    }

    if (!response.ok || result?.success === false) {
      throw new ApiError(result?.error?.message || "Permintaan gagal. Silakan coba lagi.", {
        status: response.status,
        code: result?.error?.code || "HTTP_ERROR",
        fields: result?.error?.fields || {},
      });
    }
    if (result?.success !== true || !("data" in result)) {
      throw new ApiError("Format respons server tidak sesuai kontrak.", {
        status: response.status,
        code: "INVALID_RESPONSE",
      });
    }
    return { data: result.data, meta: result.meta ?? null };
  } catch (error) {
    if (error instanceof ApiError) throw error;
    if (error?.name === "AbortError") {
      throw new ApiError(
        signal?.aborted ? "Permintaan dibatalkan." : "Server terlalu lama merespons. Coba lagi.",
        { code: signal?.aborted ? "REQUEST_ABORTED" : "TIMEOUT" },
      );
    }
    throw new ApiError("Tidak dapat terhubung ke server. Periksa koneksi lalu coba lagi.");
  } finally {
    window.clearTimeout(timeout);
    signal?.removeEventListener("abort", abortFromCaller);
  }
}

const segment = (value) => encodeURIComponent(String(value));

export const api = Object.freeze({
  packages: () => request("/packages"),
  templates: (packageSlug = "") =>
    request(`/templates${packageSlug ? `?package=${segment(packageSlug)}` : ""}`),
  template: (slug) => request(`/templates/${segment(slug)}`),
  createOrder: (data) => request("/orders", { method: "POST", body: data }),
  orderStatus: (reference) => request(`/orders/${segment(reference)}`),
  uploadPayment: (reference, formData) =>
    request(`/orders/${segment(reference)}/payment`, { method: "POST", body: formData }),
  uploadGalleryPhotos: (reference, formData) =>
    request(`/orders/${segment(reference)}/photos`, { method: "POST", body: formData, timeoutMs: 120000 }),
  adminLogin: (data) => request("/admin/auth/login", { method: "POST", body: data }),
  adminLogout: () => request("/admin/auth/logout", { method: "POST" }),
  adminOrders: ({ status = "", page = 1, limit = 20 } = {}) => {
    const query = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (status) query.set("status", status);
    return request(`/admin/orders?${query}`);
  },
  adminOrder: (id) => request(`/admin/orders/${segment(id)}`),
  verifyPayment: (id, data) =>
    request(`/admin/orders/${segment(id)}/verify-payment`, { method: "POST", body: data }),
  publishOrder: (id, data) =>
    request(`/admin/orders/${segment(id)}/publish`, { method: "POST", body: data }),
  rejectOrder: (id, data) =>
    request(`/admin/orders/${segment(id)}/reject`, { method: "POST", body: data }),
  adminTemplates: () => request("/admin/templates"),
  setTemplateAvailability: (id, available) =>
    request(`/admin/templates/${segment(id)}`, { method: "PATCH", body: { available } }),
});
