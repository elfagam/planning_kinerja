export function readUrlState(defaults) {
  const params = new URLSearchParams(window.location.search);
  const next = { ...defaults };

  Object.keys(defaults).forEach((key) => {
    const v = params.get(key);
    if (v != null && v !== "") {
      next[key] = v;
    }
  });

  return next;
}

export function writeUrlState(state, { replace = false } = {}) {
  const params = new URLSearchParams();
  Object.entries(state).forEach(([key, value]) => {
    if (value == null || value === "") {
      return;
    }
    params.set(key, String(value));
  });

  const url = `${window.location.pathname}?${params.toString()}`;
  if (replace) {
    window.history.replaceState(state, "", url);
    return;
  }
  window.history.pushState(state, "", url);
}
