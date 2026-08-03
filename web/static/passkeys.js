(function () {
  function base64urlToBuffer(value) {
    var base64 = String(value).replace(/-/g, "+").replace(/_/g, "/");
    while (base64.length % 4) {
      base64 += "=";
    }
    var binary = window.atob(base64);
    var bytes = new Uint8Array(binary.length);
    for (var i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes.buffer;
  }

  function bufferToBase64url(value) {
    var bytes = new Uint8Array(value);
    var binary = "";
    for (var i = 0; i < bytes.byteLength; i += 1) {
      binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
  }

  function credentialCreationOptions(raw) {
    var options = raw && raw.publicKey ? raw.publicKey : raw;
    if (!options) {
      throw new Error("missing passkey options");
    }
    options.challenge = base64urlToBuffer(options.challenge);
    if (options.user && options.user.id) {
      options.user.id = base64urlToBuffer(options.user.id);
    }
    if (Array.isArray(options.excludeCredentials)) {
      options.excludeCredentials = options.excludeCredentials.map(function (credential) {
        credential.id = base64urlToBuffer(credential.id);
        return credential;
      });
    }
    return { publicKey: options };
  }

  function credentialToJSON(credential) {
    return {
      id: credential.id,
      rawId: bufferToBase64url(credential.rawId),
      type: credential.type,
      response: {
        clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
        attestationObject: bufferToBase64url(credential.response.attestationObject),
        transports: credential.response.getTransports ? credential.response.getTransports() : []
      },
      clientExtensionResults: credential.getClientExtensionResults ? credential.getClientExtensionResults() : {}
    };
  }

  function formValue(form, name) {
    var field = form.querySelector('[name="' + name + '"]');
    return field ? field.value : "";
  }

  function showPasskeyError(container, message) {
    var alert = container.querySelector("[data-passkey-error]");
    var text = container.querySelector("[data-passkey-error-message]");
    if (text) {
      text.textContent = message;
    }
    if (alert) {
      alert.hidden = false;
    }
  }

  function captureBootstrapToken() {
    var entry = document.querySelector("[data-bootstrap-entry]");
    if (!entry || !window.location.hash) {
      return;
    }
    var params = new URLSearchParams(window.location.hash.slice(1));
    var token = params.get("token");
    if (!token) {
      return;
    }
    var input = entry.querySelector('[name="bootstrap_token"]');
    if (input) {
      input.value = token;
    }
    window.history.replaceState(null, document.title, window.location.pathname + window.location.search);
  }

  function installPasskeyForm() {
    var container = document.querySelector("[data-bootstrap-passkey]");
    if (!container) {
      return;
    }
    var form = container.querySelector("[data-passkey-form]");
    if (!form) {
      return;
    }
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      if (!window.PublicKeyCredential || !navigator.credentials || !navigator.credentials.create) {
        showPasskeyError(container, "This browser does not support passkey registration.");
        return;
      }
      var csrf = formValue(form, "csrf_token");
      var profile = {
        token: formValue(form, "bootstrap_token"),
        email: formValue(form, "email"),
        first_name: formValue(form, "first_name"),
        last_name: formValue(form, "last_name"),
        preferred_display_name: formValue(form, "preferred_display_name")
      };
      fetch(container.dataset.beginUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf },
        credentials: "same-origin",
        body: JSON.stringify(profile)
      }).then(function (response) {
        if (!response.ok) {
          throw new Error("begin failed");
        }
        return response.json();
      }).then(function (payload) {
        return navigator.credentials.create(credentialCreationOptions(payload.publicKey)).then(function (credential) {
          return { ceremonyId: payload.ceremonyId, credential: credential };
        });
      }).then(function (created) {
        return fetch(container.dataset.finishUrl, {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf },
          credentials: "same-origin",
          body: JSON.stringify({
            token: profile.token,
            email: profile.email,
            first_name: profile.first_name,
            last_name: profile.last_name,
            preferred_display_name: profile.preferred_display_name,
            ceremonyId: created.ceremonyId,
            credential: credentialToJSON(created.credential)
          })
        });
      }).then(function (response) {
        if (!response.ok) {
          return response.json().then(function (body) {
            throw new Error((body && body.error) || "finish failed");
          }, function () {
            throw new Error("finish failed");
          });
        }
        return response.json();
      }).then(function (body) {
        window.location.assign((body && body.redirectTo) || "/account");
      }).catch(function (error) {
        if (error && error.name === "AbortError") {
          showPasskeyError(container, "Passkey registration was canceled. No account was created.");
          return;
        }
        showPasskeyError(container, (error && error.message) || "Try again when your browser is ready.");
      });
    });
  }

  captureBootstrapToken();
  installPasskeyForm();
}());
