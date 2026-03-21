(() => {
  const root = document.body;
  const endpoint = root.dataset.apiEndpoint;
  const moduleName = root.dataset.moduleName;

  const form = document.getElementById("crud-form");
  const idInput = document.getElementById("entity-id");
  const codeInput = document.getElementById("code");
  const nameInput = document.getElementById("name");
  const descriptionInput = document.getElementById("description");
  const attrsInput = document.getElementById("attributes");
  const tableBody = document.getElementById("table-body");
  const title = document.getElementById("module-title");
  const status = document.getElementById("status-text");
  const queryInput = document.getElementById("query");
  const limitInput = document.getElementById("limit");
  const pageInput = document.getElementById("page");
  const metaText = document.getElementById("meta-text");
  const btnSearch = document.getElementById("btn-search");
  const btnPrev = document.getElementById("btn-prev");
  const btnNext = document.getElementById("btn-next");

  function getAuthToken() {
    const tokenFromDataAttr = root.dataset.authToken;
    if (tokenFromDataAttr) {
      return tokenFromDataAttr;
    }

    return (
      localStorage.getItem("AUTH_TOKEN") ||
      localStorage.getItem("authToken") ||
      ""
    );
  }

  function redirectToLogin() {
    const next = encodeURIComponent(
      window.location.pathname + window.location.search,
    );
    window.location.href = `/ui/login?next=${next}`;
  }

  async function apiFetch(url, options = {}) {
    const token = getAuthToken();
    const headers = {
      ...(options.headers || {}),
    };

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    return fetch(url, {
      ...options,
      headers,
    });
  }

  if (!endpoint) {
    return;
  }

  title.textContent = `CRUD ${moduleName}`;

  function setStatus(message, isError = false) {
    status.textContent = message;
    status.className = isError ? "text-danger small" : "text-success small";
  }

  function canWrite() {
    return true;
  }

  function applyReadOnlyAccess() {
    // No-op: OPERATOR is no longer forced into read-only mode.
  }

  async function resolveRoleAccess() {
    try {
      const token = getAuthToken();
      const res = await fetch("/api/v1/auth/me", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) {
        return;
      }

      const body = await res.json();
      if (!body?.success) {
        return;
      }

      applyReadOnlyAccess();
    } catch (_) {
      // Keep default behavior and rely on server-side authorization.
    }
  }

  function handleUnauthorized(res, body) {
    if (res.status !== 401) {
      return false;
    }

    const authError = String(body?.error || "").toLowerCase();
    if (
      authError.includes("authorization") ||
      authError.includes("token") ||
      authError.includes("unauthorized")
    ) {
      setStatus("Sesi login berakhir, silakan login ulang", true);
      redirectToLogin();
      return true;
    }

    return false;
  }

  function parseAttributes() {
    const raw = attrsInput.value.trim();
    if (!raw) {
      return {};
    }
    try {
      return JSON.parse(raw);
    } catch {
      throw new Error("Attributes harus format JSON valid");
    }
  }

  function resetForm() {
    idInput.value = "";
    codeInput.value = "";
    nameInput.value = "";
    descriptionInput.value = "";
    attrsInput.value = "{}";
  }

  function currentParams() {
    const page = Number(pageInput.value || 1);
    const limit = Number(limitInput.value || 10);
    const q = queryInput.value.trim();
    return { page: page < 1 ? 1 : page, limit: limit < 1 ? 10 : limit, q };
  }

  function rowTemplate(item) {
    const tr = document.createElement("tr");
    const actionHTML = canWrite()
      ? '<button class="btn btn-sm btn-outline-primary me-1" data-action="edit">Edit</button><button class="btn btn-sm btn-outline-danger" data-action="delete">Hapus</button>'
      : '<span class="text-muted">Read-only</span>';
    tr.innerHTML = `
      <td>${item.id}</td>
      <td>${item.code || "-"}</td>
      <td>${item.name || "-"}</td>
      <td>${item.description || "-"}</td>
      <td><code>${JSON.stringify(item.attributes || {})}</code></td>
      <td class="text-nowrap">
        ${actionHTML}
      </td>
    `;

    if (!canWrite()) {
      return tr;
    }

    tr.querySelector('[data-action="edit"]').addEventListener("click", () => {
      idInput.value = item.id;
      codeInput.value = item.code || "";
      nameInput.value = item.name || "";
      descriptionInput.value = item.description || "";
      attrsInput.value = JSON.stringify(item.attributes || {}, null, 2);
      setStatus(`Mode edit ID ${item.id}`);
    });

    tr.querySelector('[data-action="delete"]').addEventListener(
      "click",
      async () => {
        if (!confirm(`Hapus data ID ${item.id}?`)) {
          return;
        }
        const res = await apiFetch(`${endpoint}/${item.id}`, {
          method: "DELETE",
        });
        const body = await res.json();
        if (!res.ok || !body.success) {
          if (handleUnauthorized(res, body)) {
            return;
          }
          setStatus(body.error || "Gagal hapus data", true);
          return;
        }
        setStatus(`Data ID ${item.id} dihapus`);
        await loadList();
      },
    );

    return tr;
  }

  async function loadList() {
    const { page, limit, q } = currentParams();
    const params = new URLSearchParams({
      page: String(page),
      limit: String(limit),
    });
    if (q !== "") {
      params.set("q", q);
    }

    const res = await apiFetch(`${endpoint}?${params.toString()}`);
    const body = await res.json();
    if (!res.ok || !body.success) {
      if (handleUnauthorized(res, body)) {
        return;
      }
      setStatus(body.error || "Gagal memuat daftar", true);
      return;
    }

    const items = body.data.items || [];
    const meta = body.data.meta || { page, limit, total: items.length };
    metaText.textContent = `Page ${meta.page} | Limit ${meta.limit} | Total ${meta.total}`;

    btnPrev.disabled = Number(meta.page) <= 1;
    btnNext.disabled =
      Number(meta.page) * Number(meta.limit) >= Number(meta.total);

    tableBody.innerHTML = "";
    if (items.length === 0) {
      const tr = document.createElement("tr");
      tr.innerHTML =
        '<td colspan="6" class="text-center text-muted">Belum ada data</td>';
      tableBody.appendChild(tr);
      return;
    }

    items
      .sort((a, b) => a.id - b.id)
      .forEach((item) => tableBody.appendChild(rowTemplate(item)));
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    if (!canWrite()) {
      setStatus("Role OPERATOR hanya dapat melihat data", true);
      return;
    }

    let attributes;
    try {
      attributes = parseAttributes();
    } catch (err) {
      setStatus(err.message, true);
      return;
    }

    const payload = {
      code: codeInput.value.trim(),
      name: nameInput.value.trim(),
      description: descriptionInput.value.trim(),
      attributes,
    };

    const id = idInput.value.trim();
    const isEdit = id !== "";
    const url = isEdit ? `${endpoint}/${id}` : endpoint;
    const method = isEdit ? "PUT" : "POST";

    const res = await apiFetch(url, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const body = await res.json();
    if (!res.ok || !body.success) {
      if (handleUnauthorized(res, body)) {
        return;
      }
      setStatus(body.error || "Gagal simpan data", true);
      return;
    }

    setStatus(
      isEdit ? `Data ID ${id} diperbarui` : "Data baru berhasil disimpan",
    );
    resetForm();
    await loadList();
  });

  document.getElementById("btn-reset").addEventListener("click", () => {
    if (!canWrite()) {
      setStatus("Role OPERATOR hanya dapat melihat data", true);
      return;
    }
    resetForm();
    setStatus("Form direset");
  });

  btnSearch.addEventListener("click", () => {
    pageInput.value = "1";
    loadList();
  });

  btnPrev.addEventListener("click", () => {
    const page = Number(pageInput.value || 1);
    if (page > 1) {
      pageInput.value = String(page - 1);
      loadList();
    }
  });

  btnNext.addEventListener("click", () => {
    const page = Number(pageInput.value || 1);
    pageInput.value = String(page + 1);
    loadList();
  });

  limitInput.addEventListener("change", () => {
    pageInput.value = "1";
    loadList();
  });

  if (!getAuthToken()) {
    redirectToLogin();
    return;
  }

  resetForm();
  resolveRoleAccess().finally(() => {
    loadList();
  });
})();
