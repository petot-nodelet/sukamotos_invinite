import { api } from "./api.js";

const form = document.querySelector("#admin-login-form");
const errorMessage = document.querySelector("#login-error");

function restoreButton(button) {
  const arrow = document.createElement("span");
  arrow.setAttribute("aria-hidden", "true");
  arrow.textContent = "↗";
  button.replaceChildren(document.createTextNode("Masuk ke dashboard "), arrow);
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!form.reportValidity()) return;
  const button = form.querySelector('[type="submit"]');
  button.disabled = true;
  button.textContent = "Memeriksa sesi…";
  errorMessage.hidden = true;
  try {
    await api.adminLogin({
      username: form.elements.username.value.trim(),
      password: form.elements.password.value,
    });
    window.location.assign("./index.html");
  } catch (error) {
    errorMessage.textContent = error.message;
    errorMessage.hidden = false;
  } finally {
    button.disabled = false;
    restoreButton(button);
    form.elements.password.value = "";
  }
});
