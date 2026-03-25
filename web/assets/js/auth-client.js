(() => {
  const ACCESS_KEY = "AUTH_TOKEN";
  const REFRESH_KEY = "REFRESH_TOKEN";
  const LOGIN_PATH = "/ui/login";

  let authEnabledCache = null;
  let actorContextCache = null;
  let healthInfoCache = null;
  let refreshPromise = null;
  let refreshTimer = null;

  async function fetchJSON(url, options = {}) {
    const res = await window.fetch(url, options);
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data?.error || "Request failed");
    }
    return data;
  }

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

      // If refresh succeeded but expiry is still too close (e.g. clock sync issue),
      // we must avoid a 0ms tight loop. Add a small intentional delay.
      const newState = getSessionState();
      if (newState.status === "refresh_soon" && newState.secondsRemaining <= 5) {
        setTimeout(scheduleProactiveRefresh, 5000);
      } else {
        scheduleProactiveRefresh();
      }
    }, Math.max(waitMs, 1000)); // Ensure at least 1s wait
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
    const currentPath = window.location.pathname;
    if (currentPath === LOGIN_PATH) {
      return; // Already here, stop recursion
    }
    const next = encodeURIComponent(currentPath + window.location.search);
    window.location.href = `${LOGIN_PATH}?next=${next}`;
  }

  async function fetchHealthInfo() {
    if (healthInfoCache != null) {
      return healthInfoCache;
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
      healthInfoCache = {
        authEnabled: Boolean(body?.auth_enabled),
        actorContextEnabled: Boolean(
          body?.auth_enabled || body?.dev_auth_actor_enabled,
        ),
      };
    } catch (_) {
      healthInfoCache = {
        authEnabled: true,
        actorContextEnabled: true,
      };
    }

    return healthInfoCache;
  }

  async function fetchAuthEnabled() {
    if (authEnabledCache != null) {
      return authEnabledCache;
    }

    const info = await fetchHealthInfo();
    authEnabledCache = info.authEnabled;

    return authEnabledCache;
  }

  async function fetchActorContextEnabled() {
    if (actorContextCache != null) {
      return actorContextCache;
    }

    const info = await fetchHealthInfo();
    actorContextCache = info.actorContextEnabled;
    return actorContextCache;
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

  const ALLOWED_OPERATOR_PAGES = [
    "/ui/rencana-kerja",
    "/ui/indikator-kinerja",
    "/ui/clients",
    "/ui/target-realisasi",
    "/ui/dokumen_pdf",
    "/ui/dashboard",
    "/ui/unit-pengusul"
  ];

  function applyRoleBasedAccessControl(rawRole) {
    const role = String(rawRole || "").trim().toUpperCase();
    
    // Admins and Pimpinan see everything
    if (role === "ADMIN" || role === "PIMPINAN") {
      return;
    }

    // Treat empty/unknown roles as restricted
    const restrictedRoles = ["OPERATOR", "PERENCANA", "VERIFIKATOR", ""];
    if (restrictedRoles.includes(role)) {
      const currentPath = window.location.pathname.replace(/\/$/, ""); // Strip trailing slash for matching
      const loginPath = LOGIN_PATH.replace(/\/$/, "");
      
      const allowedPaths = ALLOWED_OPERATOR_PAGES.map(p => p.replace(/\/$/, ""));

      // 1. Route Guard
      const isAllowed = allowedPaths.some((p) => currentPath.endsWith(p));
      const isLogin = currentPath.endsWith(loginPath);
      if (!isAllowed && !isLogin) {
        window.location.href = "dashboard";
        return;
      }

      // 2. Menu Guard
      document.querySelectorAll('.admin-link').forEach(link => {
        const rawHref = link.getAttribute("href");
        if (rawHref) {
          const href = rawHref.split('?')[0].replace(/\/$/, "");
          const isAllowedLink = allowedPaths.some((p) => href.endsWith(p));
          if (!isAllowedLink) {
            link.style.display = 'none';
          }
        }
      });
      
      // 3. Hide Empty Group Labels
      document.querySelectorAll(".menu-group-label").forEach((label) => {
        let nextEl = label.nextElementSibling;
        let hasVisibleLinks = false;
        while(nextEl && !nextEl.classList.contains("menu-group-label")) {
          if (nextEl.classList.contains("admin-link") && nextEl.style.display !== 'none') {
            hasVisibleLinks = true;
            break;
          }
          nextEl = nextEl.nextElementSibling;
        }
        if (!hasVisibleLinks) {
          label.style.display = 'none';
        }
      });
    }
  }

  async function performRoleCheck() {
    try {
      const token = getAccessToken();
      const res = await window.fetch("/api/v1/auth/me", {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const userData = await res.json();
        const role = userData?.data?.role || "";
        applyRoleBasedAccessControl(role);
        return true;
      }
    } catch (_) {}
    return false;
  }

  async function verifySession() {
    const enabled = await fetchAuthEnabled();
    if (!enabled) {
      return;
    }

    const current = window.location.pathname;
    const accessToken = getAccessToken();
    if (!accessToken) {
      if (current !== LOGIN_PATH) {
        redirectToLogin();
      }
      return;
    }

    if (current === LOGIN_PATH) {
      const url = new URL(window.location.href);
      const next = url.searchParams.get("next");
      if (next && next.startsWith("/") && next !== LOGIN_PATH) {
        window.location.href = next;
      } else {
        window.location.href = "/ui/dashboard";
      }
      return;
    }

    const checkSuccess = await performRoleCheck();
    if (checkSuccess) {
      scheduleProactiveRefresh();
      return;
    }

    const refreshed = await refreshAccessToken();
    if (!refreshed) {
      clearTokens();
      redirectToLogin();
    } else {
      await performRoleCheck();
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
    fetchJSON,
    getAccessToken,
    getRefreshToken,
    setTokens,
    clearTokens,
    getSessionState,
    verifySession,
    isAuthEnabled: fetchAuthEnabled,
    hasActorContext: fetchActorContextEnabled,
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
