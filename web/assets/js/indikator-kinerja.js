(() => {
  const endpoint = document.body.dataset.apiEndpoint;
  const rencanaKerjaEndpoint = "/api/v1/rencana_kerja";

  const form = document.getElementById("crud-form");
  const idInput = document.getElementById("entity-id");
  const rencanaKerjaInput = document.getElementById("rencanaKerjaId");
  const kodeInput = document.getElementById("kode");
  const namaInput = document.getElementById("nama");
  const satuanInput = document.getElementById("satuan");
  const targetTahunanInput = document.getElementById("targetTahunan");
  const hargaSatuanInput = document.getElementById("hargaSatuan");
  const anggaranTahunanInput = document.getElementById("anggaranTahunan");
  const statusText = document.getElementById("status-text");
  const queryInput = document.getElementById("query");
  const btnSearch = document.getElementById("btn-search");
  const btnReset = document.getElementById("btn-reset");
  const tableBody = document.getElementById("table-body");
  const metaText = document.getElementById("meta-text");

  const rencanaKerjaMap = new Map();
  let currentListItems = [];
  let page = 1;
  const pageSize = window.pageSize || 5;
  let totalItems = 0;

  const MUTATION_ROLES = new Set(["ADMIN", "PERENCANA", "OPERATOR"]);
  let currentUserRole = "";

  function normalizeRole(raw) { return String(raw || "").trim().toUpperCase(); }
  function canMutateByRole(role) { return MUTATION_ROLES.has(normalizeRole(role)); }

  function applyRoleAccess() {
    if (canMutateByRole(currentUserRole)) return;
    form.querySelectorAll("input, textarea, select, button[type='submit']").forEach((el) => {
      el.disabled = true;
    });
    const info = document.createElement("p");
    info.className = "text-warning small mt-2 mb-0";
    info.textContent = "Mode baca saja — role Anda tidak memiliki hak untuk mengubah data.";
    form.appendChild(info);
  }

  function authHeader() {
    const token = window.__AUTH__ ? window.__AUTH__.getAccessToken() : (localStorage.getItem("AUTH_TOKEN") || localStorage.getItem("authToken") || "");
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  function setStatus(message, isError = false) {
    statusText.textContent = message;
    statusText.className = isError ? "text-danger small" : "text-success small";
  }

  function normalizeRencanaKerja(raw) {
    return {
      id: Number(raw.id ?? raw.ID ?? 0),
      kode: raw.kode ?? raw.Kode ?? "",
      nama: raw.nama ?? raw.Nama ?? "",
    };
  }

  function normalizeItem(raw) {
    return {
      id: Number(raw.id ?? raw.ID ?? 0),
      rencanaKerjaId: Number(raw.rencana_kerja_id ?? raw.RencanaKerjaID ?? 0),
      kode: raw.kode ?? raw.Kode ?? "",
      nama: raw.nama ?? raw.Nama ?? "",
      satuan: raw.satuan ?? raw.Satuan ?? "",
      targetTahunan: Number(raw.target_tahunan ?? raw.TargetTahunan ?? 0),
      hargaSatuan: Number(raw.harga_satuan ?? raw.HargaSatuan ?? 0),
      anggaranTahunan: Number(raw.anggaran_tahunan ?? raw.AnggaranTahunan ?? 0),
    };
  }

  async function fetchJSON(url, options = {}) {
    const res = await fetch(url, {
      ...options,
      headers: { ...(options.headers || {}), ...authHeader() },
    });
    const body = await res.json();
    if (!res.ok || !body.success) throw new Error(body.error || `HTTP ${res.status}`);
    return body.data;
  }

  function isDuplicateKode(kode, editingID) {
    const normalizedKode = String(kode || "").trim().toLowerCase();
    if (!normalizedKode) return false;
    return currentListItems.some((item) => {
      const itemID = Number(item.id || 0);
      if (editingID > 0 && itemID === editingID) return false;
      return String(item.kode || "").trim().toLowerCase() === normalizedKode;
    });
  }

  function suggestNextKode() {
    let maxNumber = 0;
    let padWidth = 3;
    currentListItems.forEach((item) => {
      const kode = String(item.kode || "").trim();
      const match = /^IRK-(\d+)$/i.exec(kode);
      if (!match) return;
      const parsed = Number(match[1]);
      if (Number.isFinite(parsed) && parsed > maxNumber) {
        maxNumber = parsed;
        padWidth = Math.max(padWidth, match[1].length);
      }
    });
    return `IRK-${String(maxNumber + 1).padStart(padWidth, "0")}`;
  }

  function resetForm() {
    idInput.value = "";
    rencanaKerjaInput.value = "";
    kodeInput.value = suggestNextKode();
    namaInput.value = "";
    satuanInput.value = "";
    hargaSatuanInput.value = "0";
    targetTahunanInput.value = "0";
    anggaranTahunanInput.value = "0";
  }

  function rencanaKerjaLabel(id) {
    return rencanaKerjaMap.get(Number(id)) || `Rencana Kerja #${id}`;
  }

  function rowTemplate(item) {
    const tr = document.createElement("tr");

    const tdId = document.createElement("td");
    tdId.style.width = "1%";
    tdId.style.whiteSpace = "nowrap";
    tdId.textContent = item.id;
    tr.appendChild(tdId);

    const tdRk = document.createElement("td");
    tdRk.textContent = rencanaKerjaLabel(item.rencanaKerjaId);
    tr.appendChild(tdRk);

    const tdKode = document.createElement("td");
    tdKode.textContent = item.kode;
    tr.appendChild(tdKode);

    const tdNama = document.createElement("td");
    tdNama.textContent = item.nama;
    tr.appendChild(tdNama);

    const tdSatuan = document.createElement("td");
    tdSatuan.textContent = item.satuan || "-";
    tr.appendChild(tdSatuan);

    const tdTarget = document.createElement("td");
    tdTarget.textContent = item.targetTahunan.toFixed(2);
    tr.appendChild(tdTarget);

    const tdHarga = document.createElement("td");
    tdHarga.textContent = item.hargaSatuan !== undefined ? item.hargaSatuan.toFixed(2) : "-";
    tr.appendChild(tdHarga);

    const tdAnggaran = document.createElement("td");
    tdAnggaran.textContent = item.anggaranTahunan.toFixed(2);
    tr.appendChild(tdAnggaran);

    const tdAction = document.createElement("td");
    tdAction.className = "text-nowrap";

    if (canMutateByRole(currentUserRole)) {
      const btnEdit = document.createElement("button");
      btnEdit.className = "btn btn-sm btn-outline-primary me-1";
      btnEdit.textContent = "Edit";
      btnEdit.addEventListener("click", () => {
        idInput.value = item.id;
        rencanaKerjaInput.value = String(item.rencanaKerjaId);
        kodeInput.value = item.kode;
        namaInput.value = item.nama;
        satuanInput.value = item.satuan;
        targetTahunanInput.value = String(item.targetTahunan);
        hargaSatuanInput.value = String(item.hargaSatuan !== undefined ? item.hargaSatuan : "0");
        anggaranTahunanInput.value = String(item.anggaranTahunan);
        setStatus(`Mode edit ID ${item.id}`);
      });
      tdAction.appendChild(btnEdit);

      const btnDelete = document.createElement("button");
      btnDelete.className = "btn btn-sm btn-outline-danger";
      btnDelete.textContent = "Hapus";
      btnDelete.addEventListener("click", async () => {
        if (!confirm(`Hapus Rencana Aksi rencana kerja ${item.nama}?`)) return;
        try {
          await fetchJSON(`${endpoint}/${item.id}`, { method: "DELETE" });
          setStatus("Data berhasil dihapus");
          if (Number(idInput.value) === item.id) resetForm();
          await loadList();
        } catch (error) {
          setStatus(error.message, true);
        }
      });
      tdAction.appendChild(btnDelete);
    } else {
      const spanReadOnly = document.createElement("span");
      spanReadOnly.className = "text-muted small";
      spanReadOnly.textContent = "Read-only";
      tdAction.appendChild(spanReadOnly);
    }

    tr.appendChild(tdAction);
    return tr;
  }

  async function loadRencanaKerjaOptions() {
    try {
      const data = await fetchJSON(rencanaKerjaEndpoint);
      const items = (Array.isArray(data?.items) ? data.items : []).map(normalizeRencanaKerja).sort((a, b) => a.id - b.id);
      rencanaKerjaMap.clear();
      rencanaKerjaInput.innerHTML = '<option value="">Pilih rencana kerja</option>';
      items.forEach((item) => {
        const label = `${item.kode} - ${item.nama}`;
        rencanaKerjaMap.set(item.id, label);
        const option = document.createElement("option");
        option.value = String(item.id);
        option.textContent = label;
        rencanaKerjaInput.appendChild(option);
      });
    } catch (error) {
      setStatus(`Pilihan rencana kerja tidak ditemukan: ${error.message}`, true);
    }
  }

  async function loadList() {
    tableBody.innerHTML = '<tr><td colspan="9" class="text-center text-muted py-4">Memuat...</td></tr>';
    try {
      const params = new URLSearchParams();
      const q = queryInput.value.trim();
      const selectedRencanaKerjaID = Number(rencanaKerjaInput.value || 0);
      if (q) params.set("q", q);

      const url = params.size ? `${endpoint}?${params.toString()}` : endpoint;
      const data = await fetchJSON(url);
      const allItems = (Array.isArray(data?.items) ? data.items : []).map(normalizeItem);
      currentListItems = allItems;
      const items = selectedRencanaKerjaID
        ? allItems.filter((item) => item.rencanaKerjaId === selectedRencanaKerjaID)
        : allItems;

      const qLower = queryInput.value.trim().toLowerCase();
      const itemsFiltered = items.filter((item) => {
        const rkLabel = rencanaKerjaLabel(item.rencanaKerjaId).toLowerCase();
        return (!qLower || item.kode.toLowerCase().includes(qLower) || item.nama.toLowerCase().includes(qLower) || rkLabel.includes(qLower));
      });

      totalItems = itemsFiltered.length;
      const ps = window.pageSize || 5;
      const startIdx = (page - 1) * ps;
      const endIdx = startIdx + ps;
      const paginatedItems = itemsFiltered.slice(startIdx, endIdx);

      tableBody.innerHTML = "";
      metaText.textContent = `Total ${totalItems} data | Halaman ${page} dari ${Math.ceil(totalItems / ps)}`;

      if (paginatedItems.length === 0) {
        tableBody.innerHTML = selectedRencanaKerjaID
          ? '<tr><td colspan="9" class="text-center text-muted py-4">Belum ada data</td></tr>'
          : '<tr><td colspan="9" class="text-center text-muted py-4">Pilih rencana kerja terlebih dahulu</td></tr>';
        renderPaginationControls(ps);
        return;
      }

      paginatedItems.sort((a, b) => a.id - b.id).forEach((item) => tableBody.appendChild(rowTemplate(item)));
      renderPaginationControls(ps);
    } catch (error) {
      tableBody.innerHTML = '<tr><td colspan="9" class="text-center text-danger py-4">Gagal memuat data</td></tr>';
      setStatus(error.message, true);
    }
  }

  function renderPaginationControls(ps) {
    let paginationDiv = document.getElementById("pagination-controls");
    if (!paginationDiv) {
      paginationDiv = document.createElement("div");
      paginationDiv.id = "pagination-controls";
      paginationDiv.className = "d-flex flex-column flex-md-row justify-content-between align-items-md-center mt-3 gap-3";
      
      // Inject after table-responsive, not inside table
      const tableContainer = tableBody.closest('.table-responsive');
      if (tableContainer) {
        tableContainer.insertAdjacentElement('afterend', paginationDiv);
      } else {
        tableBody.parentElement.appendChild(paginationDiv);
      }
    }
    paginationDiv.innerHTML = "";
    
    // Hide the existing text meta text if it exists since we show it here
    if (metaText) metaText.style.display = "none";
    
    const totalPages = Math.ceil(totalItems / ps);
    if (totalPages <= 1) return;

    // Kiri: Pemilihan jumlah data per halaman
    const pageSizeDiv = document.createElement("div");
    pageSizeDiv.className = "d-flex align-items-center gap-2";
    const pageSizeLabel = document.createElement("label");
    pageSizeLabel.textContent = "Data per halaman:";
    pageSizeLabel.className = "form-label mb-0 text-muted small text-nowrap";
    pageSizeLabel.setAttribute("for", "pageSizeSelect");
    const pageSizeSelect = document.createElement("select");
    pageSizeSelect.id = "pageSizeSelect";
    pageSizeSelect.className = "form-select form-select-sm w-auto cursor-pointer";
    [5, 10, 20, 50].forEach((size) => {
      const opt = document.createElement("option");
      opt.value = size;
      opt.textContent = size;
      if (size === ps) opt.selected = true;
      pageSizeSelect.appendChild(opt);
    });
    pageSizeSelect.addEventListener("change", (e) => {
      const val = Number(e.target.value);
      if (val > 0) { window.pageSize = val; page = 1; loadList(); }
    });
    pageSizeDiv.appendChild(pageSizeLabel);
    pageSizeDiv.appendChild(pageSizeSelect);

    // Tengah: Tombol Navigasi Halaman
    const navDiv = document.createElement("nav");
    navDiv.setAttribute("aria-label", "Navigasi halaman");
    const ul = document.createElement("ul");
    ul.className = "pagination pagination-sm mb-0 justify-content-center";
    navDiv.appendChild(ul);
    
    const createPageItem = (content, disabled, active, onClick) => {
      const li = document.createElement("li");
      li.className = `page-item ${disabled ? "disabled" : ""} ${active ? "active" : ""}`;
      const btn = document.createElement("button");
      btn.className = "page-link";
      btn.innerHTML = content;
      if (disabled) btn.setAttribute("tabindex", "-1");
      if (!disabled && onClick) btn.addEventListener("click", onClick);
      li.appendChild(btn);
      return li;
    };

    ul.appendChild(createPageItem("&laquo;", page <= 1, false, () => { if (page > 1) { page--; loadList(); } }));

    for (let i = 1; i <= totalPages; i++) {
      if (totalPages > 7 && i !== 1 && i !== totalPages && Math.abs(i - page) > 2) {
        if (i === page - 3 || i === page + 3) {
          ul.appendChild(createPageItem("...", true, false, null));
        }
        continue;
      }
      ul.appendChild(createPageItem(i, false, i === page, () => { page = i; loadList(); }));
    }

    ul.appendChild(createPageItem("&raquo;", page >= totalPages, false, () => { if (page < totalPages) { page++; loadList(); } }));

    // Kanan: Teks Info Data Total
    const infoDiv = document.createElement("div");
    infoDiv.className = "text-md-end text-center text-muted small";
    infoDiv.innerHTML = `Total <strong>${totalItems}</strong> data<span class="d-none d-md-inline"> | </span><br class="d-md-none">Halaman <strong>${page}</strong> dari <strong>${totalPages}</strong>`;

    paginationDiv.appendChild(pageSizeDiv);
    paginationDiv.appendChild(navDiv);
    paginationDiv.appendChild(infoDiv);
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!canMutateByRole(currentUserRole)) {
      setStatus("Anda tidak memiliki hak akses untuk menyimpan perubahan", true);
      return;
    }

    const payload = {
      rencana_kerja_id: Number(rencanaKerjaInput.value),
      kode: kodeInput.value.trim(),
      nama: namaInput.value.trim(),
      satuan: satuanInput.value.trim(),
      target_tahunan: Number(targetTahunanInput.value || 0),
      harga_satuan: Number(hargaSatuanInput.value || 0),
      anggaran_tahunan: Number(anggaranTahunanInput.value || 0),
    };

    if (!payload.rencana_kerja_id || !payload.kode || !payload.nama) {
      setStatus("Rencana Kerja, kode, dan nama wajib diisi", true);
      return;
    }

    const id = idInput.value.trim();
    const editingID = id ? Number(id) : 0;
    if (isDuplicateKode(payload.kode, editingID)) {
      const suggestedKode = suggestNextKode();
      kodeInput.value = suggestedKode;
      setStatus(`Kode indikator_rencana_kerja sudah digunakan. Saran kode: ${suggestedKode}`, true);
      kodeInput.focus();
      return;
    }

    const isEdit = Boolean(idInput.value);
    const url = isEdit ? `${endpoint}/${idInput.value}` : endpoint;
    const method = isEdit ? "PUT" : "POST";

    try {
      await fetchJSON(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      setStatus(isEdit ? "Data berhasil diperbarui" : "Data berhasil ditambahkan");
      resetForm();
      await loadList();
    } catch (error) {
      setStatus(error.message, true);
    }
  });

  btnReset.addEventListener("click", () => {
    resetForm();
    setStatus("Form direset");
  });

  rencanaKerjaInput.addEventListener("change", () => { page = 1; loadList(); });
  queryInput.addEventListener("keydown", (event) => { if (event.key === "Enter") { event.preventDefault(); page = 1; loadList(); } });
  btnSearch.addEventListener("click", () => { page = 1; loadList(); });

  async function loadCurrentUser() {
    try {
      const token = window.__AUTH__ ? window.__AUTH__.getAccessToken() : (localStorage.getItem("AUTH_TOKEN") || localStorage.getItem("authToken") || "");
      if (!token) return;
      const res = await fetch("/api/v1/auth/me", { headers: { Authorization: `Bearer ${token}` } });
      if (!res.ok) return;
      const body = await res.json();
      currentUserRole = normalizeRole(body?.data?.role ?? "");
    } catch (_) {}
  }

  (async () => {
    targetTahunanInput.addEventListener("input", updateAnggaranTahunan);
    hargaSatuanInput.addEventListener("input", updateAnggaranTahunan);
    targetTahunanInput.addEventListener("change", updateAnggaranTahunan);
    hargaSatuanInput.addEventListener("change", updateAnggaranTahunan);

    function updateAnggaranTahunan() {
      const target = Number(targetTahunanInput.value) || 0;
      const harga = Number(hargaSatuanInput.value) || 0;
      anggaranTahunanInput.value = (target * harga).toFixed(2);
    }

    anggaranTahunanInput.readOnly = true;
    await loadCurrentUser();
    applyRoleAccess();
    await loadRencanaKerjaOptions();
    await loadList();
    setStatus("Data siap dikelola");
  })();
})();
