(() => {
  // --- JWT check & redirect to login if not present ---
  const accessToken =
    (window.__AUTH__ &&
      window.__AUTH__.getAccessToken &&
      window.__AUTH__.getAccessToken()) ||
    localStorage.getItem("AUTH_TOKEN") ||
    localStorage.getItem("authToken") ||
    "";
  if (!accessToken) {
    window.location.href = "/ui/login";
    return;
  }
  // --- Unit Pengusul Dropdown ---
  const unitPengusulInput = document.getElementById("unitPengusulId");
  let currentUserUnitPengusulId = null;

  async function loadUnitPengusulOptions() {
    if (!unitPengusulInput) return;
    unitPengusulInput.innerHTML =
      '<option value="">Pilih Unit Pengusul</option>';
    try {
      const isOperator = normalizeRole(currentUserRole) === "OPERATOR";
      let url = "/api/v1/unit_pengusul";
      
      // Strict API filtering only for OPERATOR
      if (isOperator && currentUserUnitPengusulId) {
        url += `?user_unit_id=${encodeURIComponent(currentUserUnitPengusulId)}`;
      }
      
      const data = await fetchJSON(url);
      const items = Array.isArray(data?.items) ? data.items : [];
      
      // Filter logically in UI for OPERATOR if API didn't handle it
      let filteredItems = items;
      if (isOperator && currentUserUnitPengusulId) {
        filteredItems = items.filter(
          (item) =>
            String(item.id ?? item.ID) === String(currentUserUnitPengusulId),
        );
      }

      filteredItems.forEach((item) => {
        const id = item.id ?? item.ID;
        const nama = item.nama ?? item.Nama;
        if (id !== undefined && id !== null) {
          const opt = document.createElement("option");
          opt.value = id;
          opt.textContent = nama;
          unitPengusulInput.appendChild(opt);
        }
      });

      // GLOBAL AUTO-FILL: Default to user's unit for ANY role that has one
      if (currentUserUnitPengusulId) {
        unitPengusulInput.value = String(currentUserUnitPengusulId);
        // console.log("[unit_pengusul] auto-selected:", currentUserUnitPengusulId);
      }
    } catch (err) {
      console.error("[unit_pengusul] load error:", err);
    }
  }

  async function refreshSessionInfo() {
    try {
      // Use fetchJSON for consistency (auto-unwraps .data)
      const userData = await fetchJSON("/api/v1/auth/me");
      currentUserRole = normalizeRole(userData?.role || "");
      currentUserID = userData?.user_id ?? userData?.userID ?? 0;
      currentUserUnitPengusulId =
        userData?.unit_pengusul_id ?? userData?.unitPengusulID ?? null;
      // console.log("[session] loaded:", { currentUserRole, currentUserUnitPengusulId });
    } catch (err) {
      console.error("[session] error:", err);
    }
  }
  const endpoint = document.body.dataset.apiEndpoint;
  const rencanaKerjaEndpoint = "/api/v1/rencana_kerja";

  const form = document.getElementById("crud-form");
  const idInput = document.getElementById("entity-id");
  const rencanaKerjaInput = document.getElementById("rencanaKerjaId");
  const rencanaKerjaFilterInput = document.getElementById("rencanaKerjaFilter");
  const kodeInput = document.getElementById("kode");
  const tbStandarHargaIdInput = document.getElementById("tbStandarHargaId");
  const btnBrowseStandarHarga = document.getElementById(
    "btnBrowseStandarHarga",
  );
  const standarHargaLabel = document.getElementById("standarHargaLabel");
  const namaInput = document.getElementById("nama");
  const satuanInput = document.getElementById("satuan");
  const targetTahunanInput = document.getElementById("targetTahunan");
  const hargaSatuanInput = document.getElementById("hargaSatuan");
  const anggaranTahunanInput = document.getElementById("anggaranTahunan");
  const dibuatOlehInput = document.getElementById("dibuatOleh");
  const statusText = document.getElementById("status-text");
  const queryInput = document.getElementById("query");
  const btnSearch = document.getElementById("btn-search");
  const btnReset = document.getElementById("btn-reset");
  const btnGenerateKode = document.getElementById("btn-generate-kode");
  const tableBody = document.getElementById("table-body");
  const metaText = document.getElementById("meta-text");

  const rencanaKerjaMap = new Map();
  let rencanaKerjaItemsCache = [];
  let currentListItems = [];
  let currentUserID = 0;
  let page = 1;
  const pageSize = window.pageSize || 5;
  let totalItems = 0;

  const MUTATION_ROLES = new Set(["ADMIN", "PERENCANA", "OPERATOR"]);
  let currentUserRole = "";

  const formatCurrency = (val) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(Number(val) || 0);

  function normalizeRole(raw) {
    return String(raw || "")
      .trim()
      .toUpperCase();
  }
  function canMutateByRole(role) {
    return MUTATION_ROLES.has(normalizeRole(role));
  }

  // --- Export CSV Logic ---
  const btnExportCSV = document.getElementById("btn-export-csv");
  if (btnExportCSV) {
    btnExportCSV.addEventListener("click", async function () {
      const rencanaKerjaId = document.getElementById("rencanaKerjaId")?.value;
      const unitPengusulId = document.getElementById("unitPengusulId")?.value;
      if (!rencanaKerjaId || !unitPengusulId) {
        alert("Pilih Rencana Kerja dan Unit Pengusul terlebih dahulu.");
        return;
      }
      const url = `/api/v1/renja/export/indikator-csv?rencana_kerja_id=${encodeURIComponent(rencanaKerjaId)}&unit_pengusul_id=${encodeURIComponent(unitPengusulId)}`;
      try {
        const token =
          (window.__AUTH__ &&
            window.__AUTH__.getAccessToken &&
            window.__AUTH__.getAccessToken()) ||
          localStorage.getItem("AUTH_TOKEN") ||
          localStorage.getItem("authToken") ||
          "";
        const resp = await fetch(url, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!resp.ok) {
          const errText = await resp.text();
          alert("Gagal mengunduh CSV: " + errText);
          return;
        }
        const blob = await resp.blob();
        const downloadUrl = window.URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = downloadUrl;
        a.download = "indikator_kinerja.csv";
        document.body.appendChild(a);
        a.click();
        setTimeout(() => {
          window.URL.revokeObjectURL(downloadUrl);
          a.remove();
        }, 100);
      } catch (err) {
        alert("Terjadi kesalahan saat mengunduh CSV");
      }
    });
  }

  function applyRoleAccess() {
    if (canMutateByRole(currentUserRole)) return;
    form
      .querySelectorAll("input, textarea, select, button[type='submit']")
      .forEach((el) => {
        el.disabled = true;
      });
    const info = document.createElement("p");
    info.className = "text-warning small mt-2 mb-0";
    info.textContent =
      "Mode baca saja — role Anda tidak memiliki hak untuk mengubah data.";
    form.appendChild(info);
  }

  function authHeader() {
    const token = window.__AUTH__
      ? window.__AUTH__.getAccessToken()
      : localStorage.getItem("AUTH_TOKEN") ||
        localStorage.getItem("authToken") ||
        "";
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
      tbStandarHargaId: Number(
        raw.tb_standar_harga_id ?? raw.TbStandarHargaID ?? 0,
      ),
      standarHarga: raw.standar_harga ?? raw.StandarHarga ?? null,
      kode: raw.kode ?? raw.Kode ?? "",
      nama: raw.nama ?? raw.Nama ?? "",
      satuan: raw.satuan ?? raw.Satuan ?? "",
      targetTahunan: Number(raw.target_tahunan ?? raw.TargetTahunan ?? 0),
      hargaSatuan: Number(raw.harga_satuan ?? raw.HargaSatuan ?? 0),
      anggaranTahunan: Number(raw.anggaran_tahunan ?? raw.AnggaranTahunan ?? 0),
      dibuatOleh: Number(raw.dibuat_oleh ?? raw.DibuatOleh ?? 0),
    };
  }


  function isDuplicateKode(kode, editingID) {
    const normalizedKode = String(kode || "")
      .trim()
      .toLowerCase();
    if (!normalizedKode) return false;
    return currentListItems.some((item) => {
      const itemID = Number(item.id || 0);
      if (editingID > 0 && itemID === editingID) return false;
      return (
        String(item.kode || "")
          .trim()
          .toLowerCase() === normalizedKode
      );
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

  function generateRandomAlphanumeric(length = 6) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = 'IRK-';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  if (btnGenerateKode) {
    btnGenerateKode.addEventListener("click", () => {
      const rand = generateRandomAlphanumeric(6);
      kodeInput.value = rand;
      setStatus(`Kode random dihasilkan: ${rand}`);
      kodeInput.focus();
    });
  }

  function resetForm() {
    idInput.value = "";
    rencanaKerjaInput.value = "";
    kodeInput.value = suggestNextKode();
    tbStandarHargaIdInput.value = "";
    standarHargaLabel.style.display = "none";
    standarHargaLabel.textContent = "";
    namaInput.value = "";
    satuanInput.value = "";
    hargaSatuanInput.value = "0";
    targetTahunanInput.value = "0";
    anggaranTahunanInput.value = "0";
    if (dibuatOlehInput) {
      dibuatOlehInput.value = currentUserID > 0 ? String(currentUserID) : "";
    }
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
    tdHarga.className = "text-end";
    tdHarga.textContent =
      item.hargaSatuan !== undefined ? formatCurrency(item.hargaSatuan) : "-";
    tr.appendChild(tdHarga);

    const tdAnggaran = document.createElement("td");
    tdAnggaran.className = "text-end";
    tdAnggaran.textContent = formatCurrency(item.anggaranTahunan);
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
        tbStandarHargaIdInput.value =
          item.tbStandarHargaId > 0 ? String(item.tbStandarHargaId) : "";
        if (item.standarHarga && item.standarHarga.uraian_barang) {
          standarHargaLabel.style.display = "block";
          standarHargaLabel.textContent =
            item.standarHarga.uraian_barang +
            (item.standarHarga.spesifikasi
              ? ` (${item.standarHarga.spesifikasi})`
              : "");
        } else {
          standarHargaLabel.style.display = "none";
          standarHargaLabel.textContent = "";
        }
        kodeInput.value = item.kode;
        namaInput.value = item.nama;
        satuanInput.value = item.satuan;
        targetTahunanInput.value = String(item.targetTahunan);
        hargaSatuanInput.value = String(
          item.hargaSatuan !== undefined ? item.hargaSatuan : "0",
        );
        anggaranTahunanInput.value = String(item.anggaranTahunan);
        if (dibuatOlehInput) {
          dibuatOlehInput.value = String(item.dibuatOleh || "");
        }
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

  function filterRencanaKerjaOptions(filterText = "") {
    const selectedValue = rencanaKerjaInput.value;
    const normalizedFilter = String(filterText || "")
      .trim()
      .toLowerCase();

    let visibleItems = rencanaKerjaItemsCache;
    if (normalizedFilter) {
      visibleItems = rencanaKerjaItemsCache.filter((item) => {
        const label = `${item.kode} ${item.nama}`.toLowerCase();
        return label.includes(normalizedFilter);
      });
    }

    rencanaKerjaInput.innerHTML =
      '<option value="">Pilih rencana kerja</option>';
    visibleItems.forEach((item) => {
      const option = document.createElement("option");
      option.value = String(item.id);
      option.textContent = `${item.kode} - ${item.nama}`;
      rencanaKerjaInput.appendChild(option);
    });

    if (
      selectedValue &&
      visibleItems.some((item) => String(item.id) === selectedValue)
    ) {
      rencanaKerjaInput.value = selectedValue;
    } else if (!normalizedFilter) {
      rencanaKerjaInput.value = selectedValue;
    }
  }

  async function loadRencanaKerjaOptions() {
    try {
      // Filter by unit_pengusul_id if selected and valid
      let url = rencanaKerjaEndpoint;
      const unitPengusulIdVal = unitPengusulInput && unitPengusulInput.value;
      if (
        unitPengusulIdVal &&
        unitPengusulIdVal !== "undefined" &&
        unitPengusulIdVal !== "null"
      ) {
        url += `?unit_pengusul_id=${encodeURIComponent(unitPengusulIdVal)}`;
      }
      const data = await fetchJSON(url);
      const items = (Array.isArray(data?.items) ? data.items : [])
        .map(normalizeRencanaKerja)
        .sort((a, b) => {
          const kodeCompare = String(a.kode || "").localeCompare(
            String(b.kode || ""),
            "id",
            { sensitivity: "base" },
          );
          if (kodeCompare !== 0) return kodeCompare;
          return String(a.nama || "").localeCompare(
            String(b.nama || ""),
            "id",
            { sensitivity: "base" },
          );
        });
      rencanaKerjaItemsCache = items;
      rencanaKerjaMap.clear();
      items.forEach((item) => {
        rencanaKerjaMap.set(item.id, `${item.kode} - ${item.nama}`);
      });
      filterRencanaKerjaOptions(
        rencanaKerjaFilterInput ? rencanaKerjaFilterInput.value : "",
      );
    } catch (error) {
      setStatus(
        `Pilihan rencana kerja tidak ditemukan: ${error.message}`,
        true,
      );
    }
  }

  async function loadList() {
    tableBody.innerHTML =
      '<tr><td colspan="9" class="text-center text-muted py-4">Memuat...</td></tr>';
    try {
      const params = new URLSearchParams();
      const q = queryInput.value.trim();
      const selectedRencanaKerjaID =
        rencanaKerjaInput && rencanaKerjaInput.value;
      const selectedUnitPengusulID =
        unitPengusulInput && unitPengusulInput.value;
      if (q) params.set("q", q);
      if (
        selectedUnitPengusulID &&
        selectedUnitPengusulID !== "undefined" &&
        selectedUnitPengusulID !== "null"
      ) {
        params.set("unit_pengusul_id", selectedUnitPengusulID);
      }
      if (
        selectedRencanaKerjaID &&
        selectedRencanaKerjaID !== "undefined" &&
        selectedRencanaKerjaID !== "null"
      ) {
        params.set("rencana_kerja_id", selectedRencanaKerjaID);
      }

      const url = params.size ? `${endpoint}?${params.toString()}` : endpoint;
      const data = await fetchJSON(url);
      const allItems = (Array.isArray(data?.items) ? data.items : []).map(
        normalizeItem,
      );
      currentListItems = allItems;
      const items = allItems;

      const qLower = queryInput.value.trim().toLowerCase();
      const itemsFiltered = items.filter((item) => {
        const rkLabel = rencanaKerjaLabel(item.rencanaKerjaId).toLowerCase();
        return (
          !qLower ||
          item.kode.toLowerCase().includes(qLower) ||
          item.nama.toLowerCase().includes(qLower) ||
          rkLabel.includes(qLower)
        );
      });

      // Log filter and filtered data array
      // console.log(
      //   "[filter] unit_pengusul_id:",
      //   selectedUnitPengusulID,
      //   "rencana_kerja_id:",
      //   selectedRencanaKerjaID,
      //   "| data:",
      //   itemsFiltered,
      // );

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

      paginatedItems
        .sort((a, b) => a.id - b.id)
        .forEach((item) => tableBody.appendChild(rowTemplate(item)));
      renderPaginationControls(ps);
    } catch (error) {
      tableBody.innerHTML =
        '<tr><td colspan="9" class="text-center text-danger py-4">Gagal memuat data</td></tr>';
      setStatus(error.message, true);
    }
  }

  function renderPaginationControls(ps) {
    let paginationDiv = document.getElementById("pagination-controls");
    if (!paginationDiv) {
      paginationDiv = document.createElement("div");
      paginationDiv.id = "pagination-controls";
      paginationDiv.className =
        "d-flex flex-column flex-md-row justify-content-between align-items-md-center mt-3 gap-3";

      // Inject after table-responsive, not inside table
      const tableContainer = tableBody.closest(".table-responsive");
      if (tableContainer) {
        tableContainer.insertAdjacentElement("afterend", paginationDiv);
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
    pageSizeSelect.className =
      "form-select form-select-sm w-auto cursor-pointer";
    [5, 10, 20, 50].forEach((size) => {
      const opt = document.createElement("option");
      opt.value = size;
      opt.textContent = size;
      if (size === ps) opt.selected = true;
      pageSizeSelect.appendChild(opt);
    });
    pageSizeSelect.addEventListener("change", (e) => {
      const val = Number(e.target.value);
      if (val > 0) {
        window.pageSize = val;
        page = 1;
        loadList();
      }
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

    ul.appendChild(
      createPageItem("&laquo;", page <= 1, false, () => {
        if (page > 1) {
          page--;
          loadList();
        }
      }),
    );

    for (let i = 1; i <= totalPages; i++) {
      if (
        totalPages > 7 &&
        i !== 1 &&
        i !== totalPages &&
        Math.abs(i - page) > 2
      ) {
        if (i === page - 3 || i === page + 3) {
          ul.appendChild(createPageItem("...", true, false, null));
        }
        continue;
      }
      ul.appendChild(
        createPageItem(i, false, i === page, () => {
          page = i;
          loadList();
        }),
      );
    }

    ul.appendChild(
      createPageItem("&raquo;", page >= totalPages, false, () => {
        if (page < totalPages) {
          page++;
          loadList();
        }
      }),
    );

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
      setStatus(
        "Anda tidak memiliki hak akses untuk menyimpan perubahan",
        true,
      );
      return;
    }

    const tbShId = Number(tbStandarHargaIdInput.value) || 0;
    const payload = {
      rencana_kerja_id: Number(rencanaKerjaInput.value),
      kode: kodeInput.value.trim(),
      nama: namaInput.value.trim(),
      satuan: satuanInput.value.trim(),
      target_tahunan: Number(targetTahunanInput.value || 0),
      harga_satuan: Number(hargaSatuanInput.value || 0),
      anggaran_tahunan: Number(anggaranTahunanInput.value || 0),
      dibuat_oleh: currentUserID || Number(dibuatOlehInput.value || 0),
    };
    if (tbShId > 0) payload.tb_standar_harga_id = tbShId;

    if (!payload.rencana_kerja_id || !payload.kode || !payload.nama) {
      setStatus("Rencana Kerja, kode, dan nama wajib diisi", true);
      return;
    }

    const id = idInput.value.trim();
    const editingID = id ? Number(id) : 0;
    if (isDuplicateKode(payload.kode, editingID)) {
      const suggestedKode = suggestNextKode();
      kodeInput.value = suggestedKode;
      setStatus(
        `Kode rincian rencana kerja sudah digunakan, saran kode: ${suggestedKode}`,
        true,
      );
      kodeInput.focus();
      return;
    }

    const isEdit = Boolean(idInput.value);
    const url = isEdit ? `${endpoint}/${idInput.value}` : endpoint;
    const method = isEdit ? "PUT" : "POST";

    try {
      // console.log("Payload kirim:", payload);
      await fetchJSON(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      setStatus(
        isEdit ? "Data berhasil diperbarui" : "Data berhasil ditambahkan",
      );
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

  rencanaKerjaInput.addEventListener("change", () => {
    page = 1;
    loadList();
  });
  queryInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter") {
      event.preventDefault();
      page = 1;
      loadList();
    }
  });
  btnSearch.addEventListener("click", () => {
    page = 1;
    loadList();
  });

  if (rencanaKerjaFilterInput) {
    rencanaKerjaFilterInput.addEventListener("input", () => {
      filterRencanaKerjaOptions(rencanaKerjaFilterInput.value);
    });
  }

  let modalStandarHargaInstance = null;
  const tbodyStandarHarga = document.getElementById("tbodyStandarHarga");
  const queryStandarHarga = document.getElementById("queryStandarHarga");
  const metaStandarHarga = document.getElementById("metaStandarHarga");
  const paginationStandarHarga = document.getElementById(
    "paginationStandarHarga",
  );
  let shPage = 1;
  const shPageSize = 5;

  async function loadStandarHargaList() {
    tbodyStandarHarga.innerHTML =
      '<tr><td colspan="5" class="text-center text-muted">Memuat...</td></tr>';
    try {
      const q = queryStandarHarga.value.trim();
      const params = new URLSearchParams({ page: shPage, limit: shPageSize });
      if (q) params.set("q", q);
      const data = await fetchJSON(
        `/api/v1/standar_harga?${params.toString()}`,
      );

      tbodyStandarHarga.innerHTML = "";
      if (!data.items || data.items.length === 0) {
        tbodyStandarHarga.innerHTML =
          '<tr><td colspan="5" class="text-center text-muted">Tidak ada data standar harga</td></tr>';
        metaStandarHarga.textContent = "";
        paginationStandarHarga.innerHTML = "";
        return;
      }

      data.items.forEach((sh) => {
        const tr = document.createElement("tr");
        const tdUraian = document.createElement("td");
        tdUraian.textContent = sh.uraian_barang || "-";
        const tdSpek = document.createElement("td");
        tdSpek.textContent = sh.spesifikasi || "-";
        const tdSatuan = document.createElement("td");
        tdSatuan.textContent = sh.satuan || "-";
        const tdHarga = document.createElement("td");
        tdHarga.className = "text-end text-nowrap";
        const hargaVal = Number(sh.harga_satuan || 0);
        tdHarga.textContent = formatCurrency(hargaVal);

        const tdAksi = document.createElement("td");
        const btnPilih = document.createElement("button");
        btnPilih.className = "btn btn-sm btn-primary";
        btnPilih.textContent = "Pilih";
        btnPilih.onclick = () => {
          tbStandarHargaIdInput.value = sh.id;
          const suggestedName =
            `${sh.uraian_barang || ""} ${sh.spesifikasi ? "(" + sh.spesifikasi + ")" : ""}`.trim();
          standarHargaLabel.textContent = suggestedName;
          standarHargaLabel.style.display = "block";
          namaInput.value = suggestedName;
          satuanInput.value = sh.satuan || "";
          hargaSatuanInput.value = hargaVal;
          hargaSatuanInput.dispatchEvent(new Event("input"));
          modalStandarHargaInstance.hide();
        };
        tdAksi.appendChild(btnPilih);
        tr.append(tdUraian, tdSpek, tdSatuan, tdHarga, tdAksi);
        tbodyStandarHarga.appendChild(tr);
      });

      metaStandarHarga.textContent = `Total ${data.total} data`;

      paginationStandarHarga.innerHTML = "";
      if (data.total_pages > 1) {
        const btnPrev = document.createElement("button");
        btnPrev.className = "btn btn-sm btn-outline-secondary me-1";
        btnPrev.textContent = "Mundur";
        btnPrev.disabled = shPage <= 1;
        btnPrev.onclick = () => {
          shPage--;
          loadStandarHargaList();
        };

        const btnNext = document.createElement("button");
        btnNext.className = "btn btn-sm btn-outline-secondary";
        btnNext.textContent = "Maju";
        btnNext.disabled = shPage >= data.total_pages;
        btnNext.onclick = () => {
          shPage++;
          loadStandarHargaList();
        };

        paginationStandarHarga.append(btnPrev, btnNext);
      }
    } catch (err) {
      tbodyStandarHarga.innerHTML = `<tr><td colspan="5" class="text-center text-danger">Gagal memuat: ${err.message}</td></tr>`;
    }
  }

  if (btnBrowseStandarHarga) {
    btnBrowseStandarHarga.addEventListener("click", () => {
      const modalEl = document.getElementById("modalStandarHarga");
      if (!modalStandarHargaInstance) {
        modalStandarHargaInstance = new bootstrap.Modal(modalEl);
        queryStandarHarga.addEventListener("keydown", (e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            shPage = 1;
            loadStandarHargaList();
          }
        });
        // Set aria-hidden to false on show, true on hide
        modalEl.addEventListener("show.bs.modal", () => {
          modalEl.setAttribute("aria-hidden", "false");
        });
        modalEl.addEventListener("hide.bs.modal", () => {
          modalEl.setAttribute("aria-hidden", "true");
          // Kembalikan fokus ke tombol pembuka modal
          btnBrowseStandarHarga.focus();
        });
      }
      shPage = 1;
      loadStandarHargaList();
      modalStandarHargaInstance.show();
    });
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
 
    // Wait for session info before everything else
    await refreshSessionInfo();
    
    // Load dropdown options
    await loadUnitPengusulOptions();
    applyRoleAccess();
    
    // Load dependent rencana kerja and the final list
    await loadRencanaKerjaOptions();
    await loadList();
    
    setStatus("Data siap dikelola");
    
    if (unitPengusulInput) {
      unitPengusulInput.addEventListener("change", async () => {
        await loadRencanaKerjaOptions();
        await loadList();
      });
    }
  })();
})();
