const isLocalFrontendDev =
  ["localhost", "127.0.0.1"].includes(window.location.hostname) &&
  window.location.port === "3000";

window.SUKAMOTO_CONFIG = Object.freeze({
  apiBaseUrl: isLocalFrontendDev
    ? `${window.location.protocol}//${window.location.hostname}:8080/api/v1`
    : "/api/v1",
  demoBaseUrl: isLocalFrontendDev
    ? `${window.location.protocol}//${window.location.hostname}:8080`
    : window.location.origin,
  csrfCookieName: "sukamoto_csrf",
});
