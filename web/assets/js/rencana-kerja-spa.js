(() => {
  const auth = window.__AUTH__;
  if (!auth) {
    console.error("Auth client not found");
    return;
  }

  const fetchJSON = auth.fetchJSON;
  const getAccessToken = auth.getAccessToken;
  
  const accessToken = getAccessToken();
  if (!accessToken) {
    window.location.href = '/ui/login?next=' + encodeURIComponent(window.location.pathname + window.location.search);
    return;
  }

  // --- Constants & DOM Elements ---
  const rkEndpoint = "/api/v1/rencana_kerja";
  const irkEndpoint = "/api/v1/indikator_rencana_kerja";
  
  // Filters & Top Bar
  const qInput = document.getElementById("q");
  const tahunInput = document.getElementById("tahun");
  const skFilterInput = document.getElementById("sk-filter");
  const btnRefresh = document.getElementById("btn-refresh");
  const btnResetFilter = document.getElementById("btn-reset-filter");
  const tahunAktifLabel = document.getElementById("tahun-aktif-label");
  const pageStatus = document.getElementById("page-status");

  // Master RK Table & Meta
  const rkBody = document.getElementById("rk-body");
  const rkMeta = document.getElementById("rk-meta");

  // Master RK Form
  const rkKodeInput = document.getElementById("rk-kode");
  const rkNamaInput = document.getElementById("rk-nama");
  const rkTahunInput = document.getElementById("rk-tahun");
  const rkStatusInput = document.getElementById("rk-status");
  const rkSubKegiatanInput = document.getElementById("rk-sub-kegiatan");
  const rkPaguN1Input = document.getElementById("rk-pagu-n1");
  const rkPaguNInput = document.getElementById("rk-pagu-n");
  const rkIndikatorSKInput = document.getElementById("rk-indikator-sk");
  const rkUnitInput = document.getElementById("rk-unit-pengusul");
  const rkNewBtn = document.getElementById("rk-new");
  const rkSaveBtn = document.getElementById("rk-save");
  const rkDeleteBtn = document.getElementById("rk-delete");

  // Detail IRK Table & Meta
  const irkBody = document.getElementById("indikator-body");
  const irkMeta = document.getElementById("indikator-meta");

  // Detail IRK Form
  const irkKodeInput = document.getElementById("irk-kode");
  const irkNamaInput = document.getElementById("irk-nama");
  const irkSatuanInput = document.getElementById("irk-satuan");
  const irkTargetInput = document.getElementById("irk-target");
  const irkAnggaranInput = document.getElementById("irk-anggaran");
  const irkNewBtn = document.getElementById("irk-new");
  const irkSaveBtn = document.getElementById("irk-save");
  const irkDeleteBtn = document.getElementById("irk-delete");

  // Accumulation
  const rkAkumulasiPaguN1 = document.getElementById("rk-akumulasi-pagu-n1");
  const rkAkumulasiPaguN = document.getElementById("rk-akumulasi-pagu-n");
  const rkAkStatusDraft = document.getElementById("rk-ak-status-draft");
  const rkAkStatusDiajukan = document.getElementById("rk-ak-status-diajukan");
  const rkAkStatusDisetujui = document.getElementById("rk-ak-status-disetujui");
  const rkAkStatusDitolak = document.getElementById("rk-ak-status-ditolak");
  const rkAkumulasiKodeBody = document.getElementById("rk-akumulasi-kode-body");
  const rkAkumulasiKodeMeta = document.getElementById("rk-akumulasi-kode-meta");

  const rkPagePrev = document.getElementById("rk-page-prev");
  const rkPageNext = document.getElementById("rk-page-next");
  const rkPageText = document.getElementById("rk-page-text");

  // --- State Variables ---
  let currentUserRole = "";
  let currentUserId = 0;
  let selectedRkId = 0;
  let selectedIndikatorId = 0;
  let currentSearchQuery = "";
  let currentTahun = "";
  let currentSubKegiatanFilter = "";
  let currentPage = 1;
  const PAGE_SIZE = 10;

  let currentRkItems = []; 
  let currentIndikatorItems = []; 
  let subKegiatanItems = [];
  let indikatorSubKegiatanItems = [];
  const paguBySubKegiatanID = new Map();

  const DEFAULT_YEAR_KEY = "DEFAULT_YEAR";
  const MUTATION_ROLES = new Set(["ADMIN", "PERENCANA", "OPERATOR"]);

  // --- Normalization Helpers ---
  function normalizeSubKegiatan(raw) {
    if (!raw) return { id: 0, kode: "", nama: "" };
    return {
      id: Number(raw.id ?? raw.ID ?? 0),
      kode: raw.kode ?? raw.Kode ?? "",
      nama: raw.nama ?? raw.Nama ?? "",
    };
  }

  function isDuplicateKode(items, kode, editingId) {
    const needle = String(kode || "").trim().toLowerCase();
    if (!needle) return false;
    return items.some(item => {
      if (item.id === editingId) return false;
      return String(item.kode || "").trim().toLowerCase() === needle;
    });
  }

  function suggestNextKode(items, prefix) {
    let max = 0;
    items.forEach(item => {
      const parts = String(item.kode || "").split("-");
      const num = parseInt(parts[parts.length - 1]);
      if (!isNaN(num) && num > max) max = num;
    });
    return `${prefix}-${String(max + 1).padStart(3, '0')}`;
  }

  // --- Helpers ---
  function setStatus(message, isError = false) {
    if (!pageStatus) return;
    pageStatus.textContent = message;
    pageStatus.className = isError ? "text-danger" : "text-muted";
  }

  function formatMoney(val) {
    return new Intl.NumberFormat("id-ID", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(Number(val || 0));
  }

  function toNumber(value) {
    const n = Number(value || 0);
    return Number.isFinite(n) ? n : 0;
  }

  function normalizeRole(raw) {
    return String(raw || "").trim().toUpperCase();
  }

  function canMutateByRole(role) {
    return MUTATION_ROLES.has(normalizeRole(role));
  }

  function applyRoleAccess() {
    const canMutate = canMutateByRole(currentUserRole);
    const btns = [rkSaveBtn, rkDeleteBtn, rkNewBtn, irkSaveBtn, irkDeleteBtn, irkNewBtn];
    btns.forEach(btn => { if (btn) btn.disabled = !canMutate; });
    if (!canMutate) {
      document.querySelectorAll('input, select, textarea').forEach(el => el.disabled = true);
      setStatus("Mode baca saja (RBAC active)", false);
    }
  }

  async function loadCurrentUser() {
    try {
      const me = await fetchJSON("/api/v1/auth/me");
      currentUserRole = me?.role || "";
      currentUserId = Number(me?.user_id ?? me?.id ?? me?.ID ?? 0);
      applyRoleAccess();
    } catch (err) {
      console.error("Failed to load user info:", err);
    }
  }

  function defaultTahun() {
    const stored = String(localStorage.getItem(DEFAULT_YEAR_KEY) || "").trim();
    const year = Number(stored);
    if (Number.isInteger(year) && year >= 2000 && year <= 2100) return String(year);
    return String(new Date().getFullYear());
  }

  function normalizeTahunInput(raw) {
    const year = Number(String(raw || "").trim());
    if (!Number.isInteger(year) || year < 2000 || year > 2100) return null;
    return String(year);
  }

  function normalizeRK(raw) {
    return {
      id: Number(raw.id ?? raw.ID ?? 0),
      indikatorSubKegiatanId: Number(raw.indikator_sub_kegiatan_id ?? raw.IndikatorSubKegiatanID ?? 0),
      unitPengusulId: Number(raw.unit_pengusul_id ?? raw.UnitPengusulID ?? 0),
      kode: raw.kode ?? raw.Kode ?? "",
      nama: raw.nama ?? raw.Nama ?? "",
      tahun: Number(raw.tahun ?? raw.Tahun ?? 0),
      status: raw.status ?? raw.Status ?? "DRAFT",
    };
  }


  function normalizeIndikator(raw) {
    return {
      id: Number(raw.id ?? raw.ID ?? 0),
      rencanaKerjaId: Number(raw.rencana_kerja_id ?? raw.RencanaKerjaID ?? 0),
      kode: raw.kode ?? raw.Kode ?? "",
      nama: raw.nama ?? raw.Nama ?? "",
      satuan: raw.satuan ?? raw.Satuan ?? "",
      targetTahunan: Number(raw.target_tahunan ?? raw.TargetTahunan ?? 0),
      anggaranTahunan: Number(raw.anggaran_tahunan ?? raw.AnggaranTahunan ?? 0),
    };
  }

  function fillSelect(selectEl, items, toLabel, placeholder) {
    if (!selectEl) return;
    selectEl.innerHTML = `<option value="">${placeholder}</option>`;
    items.forEach((item) => {
      const option = document.createElement("option");
      option.value = String(item.id ?? item.ID ?? 0);
      option.textContent = toLabel(item);
      selectEl.appendChild(option);
    });
  }

  function indikatorToSubKegiatanID(id) {
    const found = indikatorSubKegiatanItems.find(x => x.id === Number(id));
    return found ? found.subKegiatanId : 0;
  }

  function paguBySubKegiatan(id) {
    return paguBySubKegiatanID.get(Number(id)) || { paguN1: 0, paguN: 0 };
  }

  // --- Master RK Logic ---
  function rowTemplateRK(item) {
    const tr = document.createElement("tr");
    if (selectedRkId === item.id) tr.className = "table-active";
    tr.style.cursor = "pointer";

    const skId = indikatorToSubKegiatanID(item.indikatorSubKegiatanId);
    const pagu = paguBySubKegiatan(skId);

    tr.innerHTML = `
      <td style="width:1%">${item.id}</td>
      <td class="text-nowrap">${item.kode}</td>
      <td>${item.nama}</td>
      <td>${item.tahun}</td>
      <td class="text-end">${formatMoney(pagu.paguN1)}</td>
      <td class="text-end">${formatMoney(pagu.paguN)}</td>
      <td><span class="badge ${item.status === 'DISETUJUI' ? 'bg-success' : 'bg-warning'}">${item.status}</span></td>
    `;

    tr.addEventListener("click", () => {
      selectedRkId = item.id;
      selectedIndikatorId = 0;
      setRkFormFromSelected();
      resetFormIRK();
      renderMasterList();
      loadDetail();
    });
    return tr;
  }

  function renderMasterList() {
    rkBody.innerHTML = "";
    if (currentRkItems.length === 0) {
      rkBody.innerHTML = '<tr><td colspan="7" class="text-center py-3 text-muted">Tidak ada data</td></tr>';
      rkPageText.textContent = "Halaman 1/1";
    } else {
      const start = (currentPage - 1) * PAGE_SIZE;
      const end = start + PAGE_SIZE;
      const paged = currentRkItems.slice(start, end);
      paged.forEach(item => rkBody.appendChild(rowTemplateRK(item)));
      
      const totalPages = Math.ceil(currentRkItems.length / PAGE_SIZE) || 1;
      rkPageText.textContent = `Halaman ${currentPage}/${totalPages}`;
      rkPageNext.disabled = currentPage >= totalPages;
      rkPagePrev.disabled = currentPage <= 1;
    }
    rkMeta.textContent = `Total ${currentRkItems.length} data`;
  }

  async function loadList() {
    rkBody.innerHTML = '<tr><td colspan="7" class="text-center py-3">Memuat...</td></tr>';
    try {
      const params = new URLSearchParams({ all: "true" });
      if (currentSearchQuery) params.append("q", currentSearchQuery);
      if (currentTahun) params.append("tahun", currentTahun);
      if (currentSubKegiatanFilter) params.append("sub_kegiatan_id", currentSubKegiatanFilter);

      const data = await fetchJSON(`${rkEndpoint}?${params.toString()}`);
      currentRkItems = (Array.isArray(data?.items) ? data.items : []).map(normalizeRK);
      renderMasterList();
      updateAccumulation(currentRkItems);
    } catch (err) {
      rkBody.innerHTML = `<tr><td colspan="7" class="text-center py-3 text-danger">Gagal: ${err.message}</td></tr>`;
    }
  }

  function updateAccumulation(items) {
    const stats = items.reduce((acc, item) => {
      const skId = indikatorToSubKegiatanID(item.indikatorSubKegiatanId);
      const p = paguBySubKegiatan(skId);
      acc.n1 += p.paguN1;
      acc.n += p.paguN;
      acc[item.status] = (acc[item.status] || 0) + 1;
      return acc;
    }, { n1: 0, n: 0, DRAFT: 0, DIAJUKAN: 0, DISETUJUI: 0, DITOLAK: 0 });

    if (rkAkStatusDraft) rkAkStatusDraft.textContent = stats.DRAFT;
    if (rkAkStatusDiajukan) rkAkStatusDiajukan.textContent = stats.DIAJUKAN;
    if (rkAkStatusDisetujui) rkAkStatusDisetujui.textContent = stats.DISETUJUI;
    if (rkAkStatusDitolak) rkAkStatusDitolak.textContent = stats.DITOLAK;

    updateAccumulationPerKode(items);
  }

  function updateAccumulationPerKode(items) {
    if (!rkAkumulasiKodeBody) return;
    rkAkumulasiKodeBody.innerHTML = "";
    
    // Group by Kode
    const groups = new Map();
    items.forEach(item => {
      const g = groups.get(item.kode) || { count: 0, budget: 0 };
      g.count++;
      // Note: Backend doesn't provide total budget per RK in master list 
      // yet, so we show budget as 0 or fetch it if possible.
      // For now, let's keep it simple.
      groups.set(item.kode, g);
    });

    if (groups.size === 0) {
      rkAkumulasiKodeBody.innerHTML = '<tr><td colspan="3" class="text-center text-muted">Belum ada data</td></tr>';
    } else {
      groups.forEach((data, kode) => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
          <td>${kode}</td>
          <td class="text-end">${data.count}</td>
          <td class="text-end">${formatMoney(0)}</td>
        `;
        rkAkumulasiKodeBody.appendChild(tr);
      });
    }
    if (rkAkumulasiKodeMeta) rkAkumulasiKodeMeta.textContent = `Menampilkan ${groups.size} kode.`;
  }

  function resetFormRK() {
    selectedRkId = 0;
    rkKodeInput.value = "";
    rkNamaInput.value = "";
    rkTahunInput.value = currentTahun || defaultTahun();
    rkStatusInput.value = "DRAFT";
    rkSubKegiatanInput.value = "";
    rkUnitInput.value = "";
    rkPaguN1Input.value = "0,00";
    rkPaguNInput.value = "0,00";
    rkDeleteBtn.disabled = true;
    rkKodeInput.value = suggestNextKode(currentRkItems, "RK");
    renderMasterList();
  }

  function setRkFormFromSelected() {
    const item = currentRkItems.find(x => x.id === selectedRkId);
    if (!item) return resetFormRK();
    rkKodeInput.value = item.kode;
    rkNamaInput.value = item.nama;
    rkTahunInput.value = item.tahun;
    rkStatusInput.value = item.status;
    rkUnitInput.value = item.unitPengusulId;
    rkSubKegiatanInput.value = indikatorToSubKegiatanID(item.indikatorSubKegiatanId);
    applyIndikatorSubKegiatanFilter();
    rkIndikatorSKInput.value = item.indikatorSubKegiatanId;
    
    const p = paguBySubKegiatan(rkSubKegiatanInput.value);
    rkPaguN1Input.value = formatMoney(p.paguN1);
    rkPaguNInput.value = formatMoney(p.paguN);

    rkDeleteBtn.disabled = false;
  }

  async function handleSaveRK() {
    const isEdit = Boolean(selectedRkId);
    const payload = {
      indikator_sub_kegiatan_id: toNumber(rkIndikatorSKInput.value),
      unit_pengusul_id: toNumber(rkUnitInput.value),
      kode: rkKodeInput.value.trim(),
      nama: rkNamaInput.value.trim(),
      tahun: toNumber(rkTahunInput.value),
      status: rkStatusInput.value,
      catatan: "",
    };
    if (!isEdit) {
      if (!currentUserId) {
        setStatus("Sesi user tidak valid, silahkan login ulang", true);
        return;
      }
      payload.dibuat_oleh = currentUserId;
    }
    if (!payload.indikator_sub_kegiatan_id || !payload.kode || !payload.nama) {
      setStatus("Mohon lengkapi form master", true);
      return;
    }

    if (isDuplicateKode(currentRkItems, payload.kode, selectedRkId)) {
        setStatus(`Kode Master sudah digunakan: ${payload.kode}`, true);
        rkKodeInput.focus();
        return;
    }

    try {
      const url = isEdit ? `${rkEndpoint}/${selectedRkId}` : rkEndpoint;
      await fetchJSON(url, {
        method: isEdit ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setStatus(isEdit ? "Master diperbarui" : "Master dibuat");
      await loadList();
    } catch (err) {
      setStatus(err.message, true);
    }
  }

  // --- Detail IRK Logic ---
  function rowTemplateIRK(item) {
    const tr = document.createElement("tr");
    if (selectedIndikatorId === item.id) tr.className = "table-active";
    tr.style.cursor = "pointer";
    tr.innerHTML = `
      <td style="width:1%">${item.id}</td>
      <td>${item.kode}</td>
      <td>${item.nama}</td>
      <td>${item.satuan}</td>
      <td class="text-end">${item.targetTahunan}</td>
      <td class="text-end">${formatMoney(item.anggaranTahunan)}</td>
    `;
    tr.addEventListener("click", () => {
      selectedIndikatorId = item.id;
      setIndikatorFormFromSelected();
      renderDetailList();
    });
    return tr;
  }

  function renderDetailList() {
    irkBody.innerHTML = "";
    if (currentIndikatorItems.length === 0) {
      irkBody.innerHTML = '<tr><td colspan="6" class="text-center py-3 text-muted">Belum ada indikator</td></tr>';
    } else {
      currentIndikatorItems.forEach(item => irkBody.appendChild(rowTemplateIRK(item)));
    }
    irkMeta.textContent = `Total ${currentIndikatorItems.length} indikator`;
  }

  async function loadDetail() {
    if (!selectedRkId) {
      currentIndikatorItems = [];
      renderDetailList();
      return;
    }
    irkBody.innerHTML = '<tr><td colspan="6" class="text-center py-3">Memuat...</td></tr>';
    try {
      const data = await fetchJSON(`${irkEndpoint}?rencana_kerja_id=${selectedRkId}`);
      currentIndikatorItems = (Array.isArray(data?.items) ? data.items : []).map(normalizeIndikator);
      renderDetailList();
    } catch (err) {
      irkBody.innerHTML = `<tr><td colspan="6" class="text-center py-3 text-danger">Gagal: ${err.message}</td></tr>`;
    }
  }

  function resetFormIRK() {
    selectedIndikatorId = 0;
    irkKodeInput.value = "";
    irkNamaInput.value = "";
    irkSatuanInput.value = "";
    irkTargetInput.value = "0";
    irkAnggaranInput.value = "0";
    irkDeleteBtn.disabled = true;
    irkKodeInput.value = suggestNextKode(currentIndikatorItems, "IRK");
    renderDetailList();
  }

  function setIndikatorFormFromSelected() {
    const item = currentIndikatorItems.find(x => x.id === selectedIndikatorId);
    if (!item) return resetFormIRK();
    irkKodeInput.value = item.kode;
    irkNamaInput.value = item.nama;
    irkSatuanInput.value = item.satuan;
    irkTargetInput.value = item.targetTahunan;
    irkAnggaranInput.value = item.anggaranTahunan;
    irkDeleteBtn.disabled = false;
  }

  async function handleSaveIRK() {
    if (!selectedRkId) {
      setStatus("Pilih master RK terlebih dahulu", true);
      return;
    }
    const payload = {
      rencana_kerja_id: Number(selectedRkId),
      kode: irkKodeInput.value.trim(),
      nama: irkNamaInput.value.trim(),
      satuan: irkSatuanInput.value.trim(),
      target_tahunan: toNumber(irkTargetInput.value),
      anggaran_tahunan: toNumber(irkAnggaranInput.value),
    };
    if (!payload.kode || !payload.nama) {
      setStatus("Mohon lengkapi form detail", true);
      return;
    }

    if (isDuplicateKode(currentIndikatorItems, payload.kode, selectedIndikatorId)) {
        setStatus(`Kode Indikator sudah digunakan: ${payload.kode}`, true);
        irkKodeInput.focus();
        return;
    }

    try {
      const isEdit = Boolean(selectedIndikatorId);
      const url = isEdit ? `${irkEndpoint}/${selectedIndikatorId}` : irkEndpoint;
      await fetchJSON(url, {
        method: isEdit ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setStatus(isEdit ? "Indikator diperbarui" : "Indikator dibuat");
      await loadDetail();
    } catch (err) {
      setStatus(err.message, true);
    }
  }

  // --- External Reference Data ---
  async function loadReferenceData(tahun) {
    try {
      const [units, sks, isks, pagus] = await Promise.all([
        fetchJSON("/api/v1/unit_pengusul?all=true"),
        fetchJSON("/api/v1/sub_kegiatan?all=true"),
        fetchJSON("/api/v1/indikator_sub_kegiatan?all=true"),
        fetchJSON(`/api/v1/pagu_sub_kegiatan?all=true&tahun=${tahun}`)
      ]);

      subKegiatanItems = (sks.items || []).map(normalizeSubKegiatan);
      indikatorSubKegiatanItems = (isks.items || []).map(item => ({
        id: Number(item.id ?? item.ID),
        subKegiatanId: Number(item.sub_kegiatan_id ?? item.SubKegiatanID),
        kode: item.kode ?? item.Kode,
        nama: item.nama ?? item.Nama
      }));

      paguBySubKegiatanID.clear();
      (pagus.items || []).forEach(p => {
        paguBySubKegiatanID.set(Number(p.sub_kegiatan_id ?? p.SubKegiatanID), {
          paguN1: Number(p.pagu_tahun_sebelumnya ?? p.PaguTahunSebelumnya ?? 0),
          paguN: Number(p.pagu_tahun_ini ?? p.PaguTahunIni ?? 0)
        });
      });

      fillSelect(rkUnitInput, units.items || [], u => {
        const k = u.kode ?? u.Kode ?? "";
        const n = u.nama ?? u.Nama ?? "";
        return `${k} - ${n}`.trim();
      }, "Pilih Unit");
      fillSelect(rkSubKegiatanInput, subKegiatanItems, s => `${s.kode} - ${s.nama}`, "Pilih Sub Kegiatan");
      fillSelect(skFilterInput, subKegiatanItems, s => `${s.kode} - ${s.nama}`, "Semua Sub Kegiatan");
    } catch (err) {
      console.error("Reference data failed:", err);
    }
  }

  function applyIndikatorSubKegiatanFilter() {
    const skId = Number(rkSubKegiatanInput.value || 0);
    const filtered = skId 
      ? indikatorSubKegiatanItems.filter(x => x.subKegiatanId === skId)
      : indikatorSubKegiatanItems;
    fillSelect(rkIndikatorSKInput, filtered, x => `${x.kode} - ${x.nama}`, "Pilih Indikator SK");
  }

  // --- App Initialization ---
  (async function main() {
    await loadCurrentUser();
    
    currentTahun = normalizeTahunInput(new URLSearchParams(window.location.search).get("tahun")) || defaultTahun();
    tahunInput.value = currentTahun;
    tahunAktifLabel.textContent = currentTahun;

    await loadReferenceData(currentTahun);
    await loadList();

    // Event Bindings
    btnRefresh.addEventListener("click", () => {
      currentTahun = normalizeTahunInput(tahunInput.value) || currentTahun;
      currentSearchQuery = qInput.value.trim();
      currentSubKegiatanFilter = skFilterInput.value;
      loadList();
    });

    btnResetFilter.addEventListener("click", () => {
      qInput.value = "";
      currentSearchQuery = "";
      skFilterInput.value = "";
      currentSubKegiatanFilter = "";
      loadList();
    });

    rkNewBtn.addEventListener("click", resetFormRK);
    rkSaveBtn.addEventListener("click", handleSaveRK);
    rkDeleteBtn.addEventListener("click", async () => {
        if (!selectedRkId || !confirm("Hapus Rencana Kerja ini?")) return;
        try {
            await fetchJSON(`${rkEndpoint}/${selectedRkId}`, { method: "DELETE" });
            resetFormRK();
            currentPage = 1;
            loadList();
            setStatus("Data dihapus");
        } catch (err) { setStatus(err.message, true); }
    });

    rkPagePrev.addEventListener("click", () => {
        if (currentPage > 1) {
            currentPage--;
            renderMasterList();
        }
    });

    rkPageNext.addEventListener("click", () => {
        const totalPages = Math.ceil(currentRkItems.length / PAGE_SIZE) || 1;
        if (currentPage < totalPages) {
            currentPage++;
            renderMasterList();
        }
    });

    rkSubKegiatanInput.addEventListener("change", () => {
        applyIndikatorSubKegiatanFilter();
        const p = paguBySubKegiatan(rkSubKegiatanInput.value);
        rkPaguN1Input.value = formatMoney(p.paguN1);
        rkPaguNInput.value = formatMoney(p.paguN);
    });

    irkNewBtn.addEventListener("click", resetFormIRK);
    irkSaveBtn.addEventListener("click", handleSaveIRK);
    irkDeleteBtn.addEventListener("click", async () => {
        if (!selectedIndikatorId || !confirm("Hapus Indikator ini?")) return;
        try {
            await fetchJSON(`${irkEndpoint}/${selectedIndikatorId}`, { method: "DELETE" });
            resetFormIRK();
            await loadDetail();
            setStatus("Indikator dihapus");
        } catch (err) { setStatus(err.message, true); }
    });

    if (window.__AUTH__ && typeof window.__AUTH__.initInformasiSwitcher === "function") {
      window.__AUTH__.initInformasiSwitcher("/rencana-kerja-spa");
    }

    setStatus("Siap.");
  })();
})();
