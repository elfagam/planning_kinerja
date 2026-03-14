(() => {
  const ACCESS_KEY = "AUTH_TOKEN";
  const REFRESH_KEY = "REFRESH_TOKEN";
  const LOGIN_PATH = "/ui/login";

  let authEnabledCache = null;
  let refreshPromise = null;
  let refreshTimer = null;

  function decodeJwtPayload(token) {
    try {
      const parts = token.split(".");
      if (parts.length < 2) {
        return null;
      }

      const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
      const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
      const json = atob(padded);
      return JSON.parse(json);
    } catch (_) {
      return null;
    }
  }

  function getTokenExpiryMs(token) {
    const payload = decodeJwtPayload(token);
    const exp = Number(payload?.exp || 0);
    if (!exp) {
      return 0;
    }
    return exp * 1000;
  }

  function getSessionState() {
    const token = getAccessToken();
    if (!token) {
      return {
        status: "expired",
        expiresAtMs: 0,
        secondsRemaining: 0,
      };
    }

    const expiresAtMs = getTokenExpiryMs(token);
    if (!expiresAtMs) {
      return {
        status: "active",
        expiresAtMs: 0,
        secondsRemaining: 0,
      };
    }

    const secondsRemaining = Math.floor((expiresAtMs - Date.now()) / 1000);
    if (secondsRemaining <= 0) {
      return {
        status: "expired",
        expiresAtMs,
        secondsRemaining: 0,
      };
    }

    if (secondsRemaining <= 60) {
      return {
        status: "refresh_soon",
        expiresAtMs,
        secondsRemaining,
      };
    }

    return {
      status: "active",
      expiresAtMs,
      secondsRemaining,
    };
  }

  function clearRefreshTimer() {
    if (refreshTimer) {
      clearTimeout(refreshTimer);
      refreshTimer = null;
    }
  }

  function scheduleProactiveRefresh() {
    clearRefreshTimer();

    const accessToken = getAccessToken();
    if (!accessToken) {
      return;
    }

    const expiryMs = getTokenExpiryMs(accessToken);
    if (!expiryMs) {
      return;
    }

    const nowMs = Date.now();
    const refreshBeforeMs = 60 * 1000;
    const waitMs = Math.max(0, expiryMs - nowMs - refreshBeforeMs);

    refreshTimer = setTimeout(async () => {
      const enabled = await fetchAuthEnabled();
      if (!enabled) {
        return;
      }

      const refreshed = await refreshAccessToken();
      if (!refreshed) {
        clearTokens();
        redirectToLogin();
        return;
      }

      scheduleProactiveRefresh();
    }, waitMs);
  }

  function getAccessToken() {
    return (
      localStorage.getItem(ACCESS_KEY) ||
      localStorage.getItem("authToken") ||
      ""
    );
  }

  function getRefreshToken() {
    return localStorage.getItem(REFRESH_KEY) || "";
  }

  function setTokens(accessToken, refreshToken) {
    if (accessToken) {
      localStorage.setItem(ACCESS_KEY, accessToken);
      localStorage.setItem("authToken", accessToken);
    }
    if (refreshToken) {
      localStorage.setItem(REFRESH_KEY, refreshToken);
    }

    scheduleProactiveRefresh();
  }

  function clearTokens() {
    clearRefreshTimer();
    localStorage.removeItem(ACCESS_KEY);
    localStorage.removeItem("authToken");
    localStorage.removeItem(REFRESH_KEY);
  }

  function redirectToLogin() {
    const current = window.location.pathname;
    if (current === LOGIN_PATH) {
      return;
    }
    const next = encodeURIComponent(current + window.location.search);
    window.location.href = `${LOGIN_PATH}?next=${next}`;
  }

  async function fetchAuthEnabled() {
    if (authEnabledCache != null) {
      return authEnabledCache;
    }

    try {
      const res = await window.fetch("/health", {
        method: "GET",
        cache: "no-store",
      });
      if (!res.ok) {
        throw new Error(`health check failed: ${res.status}`);
      }
      const body = await res.json();
      authEnabledCache = Boolean(body?.auth_enabled);
    } catch (_) {
      // Avoid false-negative read-only mode when /health temporarily fails.
      // We treat this as unknown and default to enabled behavior.
      authEnabledCache = true;
    }

    return authEnabledCache;
  }

  async function refreshAccessToken() {
    if (refreshPromise) {
      return refreshPromise;
    }

    const refreshToken = getRefreshToken();
    if (!refreshToken) {
      return false;
    }

    refreshPromise = (async () => {
      try {
        const res = await window.fetch("/api/v1/auth/refresh", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (!res.ok) {
          return false;
        }

        const body = await res.json();
        if (!body?.success || !body?.data?.access_token) {
          return false;
        }

        setTokens(body.data.access_token, body.data.refresh_token || "");
        return true;
      } catch (_) {
        return false;
      } finally {
        refreshPromise = null;
      }
    })();

    return refreshPromise;
  }

  async function verifySession() {
    const enabled = await fetchAuthEnabled();
    if (!enabled) {
      return;
    }

    const current = window.location.pathname;
    const accessToken = getAccessToken();
    if (!accessToken) {
      redirectToLogin();
      return;
    }

    if (current === LOGIN_PATH) {
      window.location.href = "/ui/dashboard";
      return;
    }

    try {
      const res = await window.fetch("/api/v1/auth/me", {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      if (res.ok) {
        scheduleProactiveRefresh();
        return;
      }
    } catch (_) {
      // ignore, continue refresh flow
    }

    const refreshed = await refreshAccessToken();
    if (!refreshed) {
      clearTokens();
      redirectToLogin();
    }
  }

  async function login(username, password) {
    const res = await window.fetch("/api/v1/auth/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ username, password }),
    });

    const body = await res.json();
    if (!res.ok || !body?.success || !body?.data?.access_token) {
      throw new Error(body?.error || "Login gagal");
    }

    setTokens(body.data.access_token, body.data.refresh_token || "");
    return body.data;
  }

  async function logout() {
    const token = getAccessToken();
    try {
      if (token) {
        await window.fetch("/api/v1/auth/logout", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
      }
    } catch (_) {
      // ignore network issues when logging out client-side
    }

    clearTokens();
    window.location.href = LOGIN_PATH;
  }

  window.__AUTH__ = {
    login,
    logout,
    getAccessToken,
    getRefreshToken,
    setTokens,
    clearTokens,
    getSessionState,
    verifySession,
    isAuthEnabled: fetchAuthEnabled,
  };

  document.addEventListener("DOMContentLoaded", () => {
    document.querySelectorAll("[data-auth-logout]").forEach((el) => {
      el.addEventListener("click", (event) => {
        event.preventDefault();
        logout();
      });
    });

    verifySession();
  });
})();
