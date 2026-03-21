export function getAccessToken() {
  return (
    localStorage.getItem("AUTH_TOKEN") ||
    localStorage.getItem("authToken") ||
    ""
  );
}

export function redirectToLogin() {
  const next = encodeURIComponent(
    window.location.pathname + window.location.search,
  );
  window.location.href = `/ui/login?next=${next}`;
}

export async function fetchJSON(url, options = {}) {
  const token = getAccessToken();
  const headers = {
    Accept: "application/json",
    ...(options.headers || {}),
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(url, {
    ...options,
    headers,
  });

  if (res.status === 401) {
    redirectToLogin();
    throw new Error("Sesi login berakhir, silakan login ulang");
  }

  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }

  const body = await res.json();
  if (!body?.success) {
    throw new Error(body?.error || "respons API tidak valid");
  }

  return body.data;
}
