(() => {
  const fetchJSON = window.fetchJSON || window.__AUTH__?.fetchJSON;
  const getAccessToken = window.__AUTH__?.getAccessToken || (() => localStorage.getItem("AUTH_TOKEN") || localStorage.getItem("authToken") || "");
  const redirectToLogin = () => { window.location.href = '/ui/login?next=' + encodeURIComponent(window.location.pathname + window.location.search); };

// --- router.js ---
function readUrlState(defaults) {
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

function writeUrlState(state, { replace = false } = {}) {
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


// --- service.js ---

async function getCurrentUserID() {
  const me = await fetchJSON("/api/v1/auth/me");
  return Number(me?.user_id ?? me?.userID ?? 0);
}

async function listUnitPengusul() {
  const data = await fetchJSON("/api/v1/unit_pengusul");
  return Array.isArray(data?.items) ? data.items : [];
}

async function listIndikatorSubKegiatan() {
  const data = await fetchJSON("/api/v1/indikator_sub_kegiatan?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

async function listSubKegiatan() {
  const data = await fetchJSON("/api/v1/sub_kegiatan?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

async function listPaguSubKegiatan() {
  const data = await fetchJSON("/api/v1/pagu_sub_kegiatan?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

async function listRencanaKerja({ q = "", tahun = "" } = {}) {
  const params = new URLSearchParams();
  params.set("all", "true");
  if (q) params.set("q", q);
  if (tahun) params.set("tahun", tahun);
  const url = params.size
    ? `/api/v1/rencana_kerja?${params.toString()}`
    : "/api/v1/rencana_kerja";

  const data = await fetchJSON(url);
  return Array.isArray(data?.items) ? data.items : [];
}

async function listIndikatorByRencanaKerjaID(rencanaKerjaID) {
  if (!rencanaKerjaID) return [];

  const data = await fetchJSON("/api/v1/indikator_rencana_kerja?all=true");
  const items = Array.isArray(data?.items) ? data.items : [];
  const rkID = Number(rencanaKerjaID);

  return items.filter(
    (item) =>
      Number(item.rencana_kerja_id ?? item.RencanaKerjaID ?? 0) === rkID,
  );
}

async function listAllIndikatorRencanaKerja() {
  const data = await fetchJSON("/api/v1/indikator_rencana_kerja?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

async function saveRencanaKerja(payload, id = "") {
  const isEdit = Boolean(id);
  const url = isEdit ? `/api/v1/rencana_kerja/${id}` : "/api/v1/rencana_kerja";
  const method = isEdit ? "PUT" : "POST";

  return fetchJSON(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function deleteRencanaKerja(id) {
  return fetchJSON(`/api/v1/rencana_kerja/${id}`, { method: "DELETE" });
}

async function saveIndikator(payload, id = "") {
  const isEdit = Boolean(id);
  const url = isEdit
    ? `/api/v1/indikator_rencana_kerja/${id}`
    : "/api/v1/indikator_rencana_kerja";
  const method = isEdit ? "PUT" : "POST";

  return fetchJSON(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

async function deleteIndikator(id) {
  return fetchJSON(`/api/v1/indikator_rencana_kerja/${id}`, {
    method: "DELETE",
  });
}


// --- index.js ---
import {
  deleteIndikator,
  deleteRencanaKerja,
  getCurrentUserID,
  listAllIndikatorRencanaKerja,
  listIndikatorByRencanaKerjaID,
  listIndikatorSubKegiatan,
  listPaguSubKegiatan,
  listRencanaKerja,
  listSubKegiatan,
  listUnitPengusul,
  saveIndikator,
  saveRencanaKerja,
} from "./service.js";

const stateDefaults = {
  q: "",
  tahun: "",
  sk: "",
  page: "1",
  selected: "",
  selectedIndikator: "",
};

let state = readUrlState(stateDefaults);
let currentUserID = 0;
let allRkItems = [];
let currentRkItems = [];
let currentRkPagedItems = [];
let currentIndikatorItems = [];
let allIndikatorItems = [];
let subKegiatanItems = [];
let indikatorSubKegiatanItems = [];
const paguBySubKegiatanID = new Map();
const DEFAULT_PAGE_SIZE = 2;
const DEFAULT_YEAR_KEY = "DEFAULT_YEAR";

const qInput = document.getElementById("q");
const tahunInput = document.getElementById("tahun");
const skFilterInput = document.getElementById("sk-filter");
const btnRefresh = document.getElementById("btn-refresh");
const btnResetFilter = document.getElementById("btn-reset-filter");
const rkBody = document.getElementById("rk-body");
const indikatorBody = document.getElementById("indikator-body");
const rkMeta = document.getElementById("rk-meta");
const tahunAktifLabel = document.getElementById("tahun-aktif-label");
const rkAkumulasiPaguN1 = document.getElementById("rk-akumulasi-pagu-n1");
const rkAkumulasiPaguN = document.getElementById("rk-akumulasi-pagu-n");
const rkAkStatusDraft = document.getElementById("rk-ak-status-draft");
const rkAkStatusDiajukan = document.getElementById("rk-ak-status-diajukan");
const rkAkStatusDisetujui = document.getElementById("rk-ak-status-disetujui");
const rkAkStatusDitolak = document.getElementById("rk-ak-status-ditolak");
const rkAkumulasiKodeBody = document.getElementById("rk-akumulasi-kode-body");
const rkAkumulasiKodeMeta = document.getElementById("rk-akumulasi-kode-meta");
const rkPagePrevBtn = document.getElementById("rk-page-prev");
const rkPageNextBtn = document.getElementById("rk-page-next");
const rkPageText = document.getElementById("rk-page-text");
const indikatorMeta = document.getElementById("indikator-meta");
const pageStatus = document.getElementById("page-status");

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

const irkKodeInput = document.getElementById("irk-kode");
const irkNamaInput = document.getElementById("irk-nama");
const irkSatuanInput = document.getElementById("irk-satuan");
const irkTargetInput = document.getElementById("irk-target");
const irkAnggaranInput = document.getElementById("irk-anggaran");
const irkNewBtn = document.getElementById("irk-new");
const irkSaveBtn = document.getElementById("irk-save");
const irkDeleteBtn = document.getElementById("irk-delete");

function setStatus(message, isError = false) {
  pageStatus.textContent = message;
  pageStatus.className = isError ? "text-danger" : "text-muted";
}

function toNumber(value) {
  const n = Number(value || 0);
  return Number.isFinite(n) ? n : 0;
}

function defaultTahun() {
  const stored = String(localStorage.getItem(DEFAULT_YEAR_KEY) || "").trim();
  const year = Number(stored);
  if (Number.isInteger(year) && year >= 2000 && year <= 2100) {
    return String(year);
  }
  return String(new Date().getFullYear());
}

function normalizeTahunInput(raw) {
  const value = String(raw || "").trim();
  if (!value) return "";
  const year = Number(value);
  if (!Number.isInteger(year) || year < 2000 || year > 2100) {
    return null;
  }
  return String(year);
}

function normalizeRK(raw) {
  return {
    id: Number(raw.id ?? raw.ID ?? 0),
    indikatorSubKegiatanId: Number(
      raw.indikator_sub_kegiatan_id ?? raw.IndikatorSubKegiatanID ?? 0,
    ),
    unitPengusulId: Number(raw.unit_pengusul_id ?? raw.UnitPengusulID ?? 0),
    kode: raw.kode ?? raw.Kode ?? "",
    nama: raw.nama ?? raw.Nama ?? "",
    tahun: Number(raw.tahun ?? raw.Tahun ?? 0),
    status: raw.status ?? raw.Status ?? "DRAFT",
  };
}

function normalizeSubKegiatan(raw) {
  return {
    id: Number(raw.id ?? raw.ID ?? 0),
    kode: raw.kode ?? raw.Kode ?? "",
    nama: raw.nama ?? raw.Nama ?? "",
  };
}

function normalizeIndikatorSubKegiatan(raw) {
  return {
    id: Number(raw.id ?? raw.ID ?? 0),
    subKegiatanId: Number(raw.sub_kegiatan_id ?? raw.SubKegiatanID ?? 0),
    kode: raw.kode ?? raw.Kode ?? "",
    nama: raw.nama ?? raw.Nama ?? "",
  };
}

function normalizePaguSubKegiatan(raw) {
  return {
    subKegiatanId: Number(raw.sub_kegiatan_id ?? raw.SubKegiatanID ?? 0),
    paguN1: Number(raw.pagu_tahun_sebelumnya ?? raw.PaguTahunSebelumnya ?? 0),
    paguN: Number(raw.pagu_tahun_ini ?? raw.PaguTahunIni ?? 0),
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
  selectEl.innerHTML = `<option value="">${placeholder}</option>`;
  items.forEach((item) => {
    const option = document.createElement("option");
    option.value = String(item.id ?? item.ID ?? 0);
    option.textContent = toLabel(item);
    selectEl.appendChild(option);
  });
}

function formatMoney(value) {
  return new Intl.NumberFormat("id-ID", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(Number(value || 0));
}

function paguToneClass(value) {
  return Number(value || 0) > 0 ? "text-success fw-semibold" : "text-muted";
}

function indikatorToSubKegiatanID(indikatorID) {
  const indikatorItem = indikatorSubKegiatanItems.find(
    (item) => item.id === Number(indikatorID),
  );
  return Number(indikatorItem?.subKegiatanId || 0);
}

function paguBySubKegiatan(subKegiatanID) {
  return (
    paguBySubKegiatanID.get(Number(subKegiatanID)) || { paguN1: 0, paguN: 0 }
  );
}

function setPaguFields(subKegiatanID) {
  const pagu = paguBySubKegiatan(subKegiatanID);
  rkPaguN1Input.value = formatMoney(pagu.paguN1);
  rkPaguNInput.value = formatMoney(pagu.paguN);
  rkPaguN1Input.className = `form-control ${paguToneClass(pagu.paguN1)}`;
  rkPaguNInput.className = `form-control ${paguToneClass(pagu.paguN)}`;
}

function indikatorSubKegiatanSearchText(item) {
  const indikator = indikatorSubKegiatanItems.find(
    (x) => x.id === Number(item.indikatorSubKegiatanId),
  );
  if (!indikator) return "";
  return `${indikator.kode || ""} ${indikator.nama || ""}`.trim();
}

function matchesListSearch(item, q) {
  if (!q) return true;
  const needle = String(q || "")
    .trim()
    .toLowerCase();
  if (!needle) return true;
  const nama = String(item.nama || "").toLowerCase();
  const indikatorText = indikatorSubKegiatanSearchText(item).toLowerCase();
  return nama.includes(needle) || indikatorText.includes(needle);
}

function filterIndikatorRencanaKerjaByRencanaKerjaItems(
  indikatorItems,
  rkItems,
) {
  const rkIDs = new Set((rkItems || []).map((item) => Number(item.id)));
  return (indikatorItems || []).filter((item) =>
    rkIDs.has(Number(item.rencanaKerjaId)),
  );
}

function setAkumulasiPagu(items, indikatorItems) {
  const total = items.reduce(
    (acc, item) => {
      const subKegiatanID = indikatorToSubKegiatanID(
        item.indikatorSubKegiatanId,
      );
      const pagu = paguBySubKegiatan(subKegiatanID);
      acc.n1 += Number(pagu.paguN1 || 0);
      acc.n += Number(pagu.paguN || 0);
      return acc;
    },
    { n1: 0, n: 0 },
  );

  const byStatus = {
    DRAFT: 0,
    DIAJUKAN: 0,
    DISETUJUI: 0,
    DITOLAK: 0,
  };
  const rkStatusByID = new Map(
    items.map((item) => [Number(item.id), item.status]),
  );
  (indikatorItems || []).forEach((item) => {
    const status = String(rkStatusByID.get(Number(item.rencanaKerjaId)) || "")
      .trim()
      .toUpperCase();
    if (!Object.prototype.hasOwnProperty.call(byStatus, status)) {
      return;
    }
    byStatus[status] += Number(item.anggaranTahunan || 0);
  });

  rkAkumulasiPaguN1.textContent = formatMoney(total.n1);
  rkAkumulasiPaguN.textContent = formatMoney(total.n);
  rkAkStatusDraft.textContent = formatMoney(byStatus.DRAFT);
  rkAkStatusDiajukan.textContent = formatMoney(byStatus.DIAJUKAN);
  rkAkStatusDisetujui.textContent = formatMoney(byStatus.DISETUJUI);
  rkAkStatusDitolak.textContent = formatMoney(byStatus.DITOLAK);
}

function buildAkumulasiByKode(rkItems, indikatorItems) {
  const rkByID = new Map(
    (rkItems || []).map((item) => [Number(item.id), item]),
  );
  const summary = new Map();

  (indikatorItems || []).forEach((item) => {
    const rk = rkByID.get(Number(item.rencanaKerjaId));
    if (!rk) return;

    const kode = String(rk.kode || "-").trim() || "-";
    if (!summary.has(kode)) {
      summary.set(kode, { kode, jumlahIndikator: 0, totalAnggaranTahunan: 0 });
    }

    const row = summary.get(kode);
    row.jumlahIndikator += 1;
    row.totalAnggaranTahunan += Number(item.anggaranTahunan || 0);
  });

  return Array.from(summary.values()).sort((a, b) =>
    String(a.kode).localeCompare(String(b.kode), "id", {
      sensitivity: "base",
      numeric: true,
    }),
  );
}

function renderAkumulasiByKode(rows, page, totalPages) {
  if (!rkAkumulasiKodeBody || !rkAkumulasiKodeMeta) return;

  if (!rows || rows.length === 0) {
    rkAkumulasiKodeBody.innerHTML =
      '<tr><td colspan="3" class="text-center text-muted py-2">Tidak ada data akumulasi untuk filter saat ini</td></tr>';
    rkAkumulasiKodeMeta.textContent = `Menampilkan 0 kode pada halaman ${page}/${totalPages}.`;
    return;
  }

  rkAkumulasiKodeBody.innerHTML = rows
    .map(
      (row) => `
        <tr>
          <td>${row.kode}</td>
          <td class="text-end">${row.jumlahIndikator}</td>
          <td class="text-end">${formatMoney(row.totalAnggaranTahunan)}</td>
        </tr>
      `,
    )
    .join("");

  rkAkumulasiKodeMeta.textContent = `Menampilkan ${rows.length} kode pada halaman ${page}/${totalPages}.`;
}

function setRkLoading() {
  rkBody.innerHTML =
    '<tr><td colspan="7" class="text-center text-muted py-3">Memuat rencana kerja...</td></tr>';
  rkMeta.textContent = "Memuat...";
  rkPageText.textContent = "Page -/-";
  rkPagePrevBtn.disabled = true;
  rkPageNextBtn.disabled = true;
}

function setIndikatorLoading() {
  indikatorBody.innerHTML =
    '<tr><td colspan="6" class="text-center text-muted py-3">Memuat indikator...</td></tr>';
  indikatorMeta.textContent = "Memuat...";
}

function setAkumulasiLoading() {
  rkAkumulasiPaguN1.textContent = "...";
  rkAkumulasiPaguN.textContent = "...";
  rkAkStatusDraft.textContent = "...";
  rkAkStatusDiajukan.textContent = "...";
  rkAkStatusDisetujui.textContent = "...";
  rkAkStatusDitolak.textContent = "...";
  if (rkAkumulasiKodeBody) {
    rkAkumulasiKodeBody.innerHTML =
      '<tr><td colspan="3" class="text-center text-muted py-2">Menghitung akumulasi...</td></tr>';
  }
  if (rkAkumulasiKodeMeta) {
    rkAkumulasiKodeMeta.textContent = "Menyiapkan data akumulasi...";
  }
}

function updatePaginationUI(totalItems) {
  const totalPages = Math.max(1, Math.ceil(totalItems / DEFAULT_PAGE_SIZE));
  const currentPage = Math.min(
    Math.max(1, Number(state.page || 1)),
    totalPages,
  );
  state.page = String(currentPage);
  writeUrlState(state, { replace: true });

  rkPageText.textContent = `Page ${currentPage}/${totalPages}`;
  rkPagePrevBtn.disabled = currentPage <= 1;
  rkPageNextBtn.disabled = currentPage >= totalPages;

  return { currentPage, totalPages };
}

function isDuplicateKode(items, kode, excludeID = 0) {
  const key = String(kode || "")
    .trim()
    .toLowerCase();
  if (!key) return false;
  return (items || []).some((item) => {
    const currentKode = String(item.kode || "")
      .trim()
      .toLowerCase();
    const id = Number(item.id || 0);
    return currentKode === key && id !== Number(excludeID || 0);
  });
}

function suggestNextKode(items, prefix) {
  const re = new RegExp(`^${prefix}-(\\d+)$`, "i");
  const maxNum = (items || []).reduce((acc, item) => {
    const m = String(item.kode || "")
      .trim()
      .match(re);
    if (!m) return acc;
    const n = Number(m[1]);
    return Number.isFinite(n) ? Math.max(acc, n) : acc;
  }, 0);
  return `${prefix}-${String(maxNum + 1).padStart(3, "0")}`;
}

function resetRkForm() {
  rkKodeInput.value = suggestNextKode(allRkItems, "RK");
  rkNamaInput.value = "";
  rkTahunInput.value = state.tahun || defaultTahun();
  rkStatusInput.value = "DRAFT";
  rkSubKegiatanInput.value = "";
  setPaguFields(0);
  rkIndikatorSKInput.value = "";
  applyIndikatorSubKegiatanFilter();
  rkDeleteBtn.disabled = !state.selected;
}

function resetIndikatorForm() {
  irkKodeInput.value = suggestNextKode(allIndikatorItems, "IRK");
  irkNamaInput.value = "";
  irkSatuanInput.value = "";
  irkTargetInput.value = "0";
  irkAnggaranInput.value = "0";
  irkDeleteBtn.disabled = !state.selectedIndikator;
}

function setRkFormFromSelected() {
  const selected = currentRkItems.find((x) => x.id === Number(state.selected));
  if (!selected) {
    resetRkForm();
    return;
  }

  rkKodeInput.value = selected.kode;
  rkNamaInput.value = selected.nama;
  rkTahunInput.value = String(selected.tahun || "");
  rkStatusInput.value = selected.status || "DRAFT";
  const indikatorSelected = indikatorSubKegiatanItems.find(
    (item) => item.id === selected.indikatorSubKegiatanId,
  );
  rkSubKegiatanInput.value = indikatorSelected
    ? String(indikatorSelected.subKegiatanId)
    : "";
  setPaguFields(rkSubKegiatanInput.value);
  applyIndikatorSubKegiatanFilter();
  rkIndikatorSKInput.value = selected.indikatorSubKegiatanId
    ? String(selected.indikatorSubKegiatanId)
    : "";
  rkUnitInput.value = selected.unitPengusulId
    ? String(selected.unitPengusulId)
    : "";
  rkDeleteBtn.disabled = false;
}

function applyIndikatorSubKegiatanFilter() {
  const subKegiatanID = Number(rkSubKegiatanInput.value || 0);
  const filteredItems = subKegiatanID
    ? indikatorSubKegiatanItems.filter(
        (item) => item.subKegiatanId === subKegiatanID,
      )
    : indikatorSubKegiatanItems;

  fillSelect(
    rkIndikatorSKInput,
    filteredItems,
    (item) => `${item.kode} - ${item.nama}`,
    "Pilih indikator sub kegiatan",
  );
}

function setIndikatorFormFromSelected() {
  const selected = currentIndikatorItems.find(
    (x) => x.id === Number(state.selectedIndikator),
  );
  if (!selected) {
    resetIndikatorForm();
    return;
  }

  irkKodeInput.value = selected.kode;
  irkNamaInput.value = selected.nama;
  irkSatuanInput.value = selected.satuan;
  irkTargetInput.value = String(selected.targetTahunan || 0);
  irkAnggaranInput.value = String(selected.anggaranTahunan || 0);
  irkDeleteBtn.disabled = false;
}

function renderRencanaKerja(items) {
  rkBody.innerHTML = "";

  if (!items.length) {
    rkBody.innerHTML =
      '<tr><td colspan="7" class="text-center text-muted py-3">Tidak ada data</td></tr>';
    rkMeta.textContent = "Total 0";
    return;
  }

  items.forEach((item) => {
    const tr = document.createElement("tr");
    tr.className = Number(state.selected) === item.id ? "table-active" : "";
    tr.style.cursor = "pointer";
    const subKegiatanID = indikatorToSubKegiatanID(item.indikatorSubKegiatanId);
    const pagu = paguBySubKegiatan(subKegiatanID);
    tr.innerHTML = `
      <td>${item.id}</td>
      <td>${item.kode}</td>
      <td>${item.nama}</td>
      <td>${item.tahun || "-"}</td>
      <td class="text-end"><span class="${paguToneClass(pagu.paguN1)}">${formatMoney(pagu.paguN1)}</span></td>
      <td class="text-end"><span class="${paguToneClass(pagu.paguN)}">${formatMoney(pagu.paguN)}</span></td>
      <td>${item.status}</td>
    `;

    tr.addEventListener("click", () => {
      state.selected = String(item.id);
      state.selectedIndikator = "";
      writeUrlState(state);
      setRkFormFromSelected();
      resetIndikatorForm();
      void loadIndikator();
      renderRencanaKerja(currentRkPagedItems);
    });

    rkBody.appendChild(tr);
  });

  rkMeta.textContent = `Total ${currentRkItems.length}`;
}

function renderIndikator(items) {
  indikatorBody.innerHTML = "";

  if (!state.selected) {
    indikatorBody.innerHTML =
      '<tr><td colspan="6" class="text-center text-muted py-3">Pilih rencana kerja dulu</td></tr>';
    indikatorMeta.textContent = "Total 0";
    return;
  }

  if (!items.length) {
    indikatorBody.innerHTML =
      '<tr><td colspan="6" class="text-center text-muted py-3">Belum ada indikator</td></tr>';
    indikatorMeta.textContent = "Total 0";
    return;
  }

  items.forEach((item) => {
    const tr = document.createElement("tr");
    tr.className =
      Number(state.selectedIndikator) === item.id ? "table-active" : "";
    tr.style.cursor = "pointer";
    tr.innerHTML = `
      <td>${item.id}</td>
      <td>${item.kode}</td>
      <td>${item.nama}</td>
      <td>${item.satuan || "-"}</td>
      <td class="text-end">${item.targetTahunan}</td>
      <td class="text-end">${formatMoney(item.anggaranTahunan)}</td>
    `;

    tr.addEventListener("click", () => {
      state.selectedIndikator = String(item.id);
      writeUrlState(state);
      setIndikatorFormFromSelected();
      renderIndikator(items);
    });

    indikatorBody.appendChild(tr);
  });

  indikatorMeta.textContent = `Total ${items.length}`;
}

async function loadReferenceOptions() {
  const [
    userID,
    unitItems,
    subKegiatanRawItems,
    indikatorSkRawItems,
    paguSubKegiatanRawItems,
  ] = await Promise.all([
    getCurrentUserID(),
    listUnitPengusul(),
    listSubKegiatan(),
    listIndikatorSubKegiatan(),
    listPaguSubKegiatan(),
  ]);

  currentUserID = userID;
  subKegiatanItems = subKegiatanRawItems.map(normalizeSubKegiatan);
  indikatorSubKegiatanItems = indikatorSkRawItems.map(
    normalizeIndikatorSubKegiatan,
  );
  paguBySubKegiatanID.clear();
  paguSubKegiatanRawItems.map(normalizePaguSubKegiatan).forEach((item) => {
    if (item.subKegiatanId > 0) {
      paguBySubKegiatanID.set(item.subKegiatanId, item);
    }
  });

  fillSelect(
    rkUnitInput,
    unitItems,
    (item) =>
      `${item.kode ?? item.Kode ?? ""} - ${item.nama ?? item.Nama ?? ""}`,
    "Pilih unit pengusul",
  );
  if (unitItems.length) {
    rkUnitInput.value = String(unitItems[0].id ?? unitItems[0].ID ?? "");
  }

  fillSelect(
    rkSubKegiatanInput,
    subKegiatanItems,
    (item) => `${item.kode} - ${item.nama}`,
    "Pilih sub kegiatan",
  );
  fillSelect(
    skFilterInput,
    subKegiatanItems,
    (item) => `${item.kode} - ${item.nama}`,
    "Semua sub kegiatan",
  );
  if (state.sk) {
    skFilterInput.value = state.sk;
  }
  applyIndikatorSubKegiatanFilter();
}

async function loadAllIndikatorRencanaKerja() {
  const data = await listAllIndikatorRencanaKerja();
  allIndikatorItems = data.map(normalizeIndikator);
}

async function loadRencanaKerja() {
  const data = await listRencanaKerja({ tahun: state.tahun });
  allRkItems = data.map(normalizeRK);

  let filteredItems = allRkItems;
  if (state.sk) {
    const selectedSubKegiatanID = Number(state.sk);
    const indikatorIDs = new Set(
      indikatorSubKegiatanItems
        .filter((item) => item.subKegiatanId === selectedSubKegiatanID)
        .map((item) => item.id),
    );
    filteredItems = filteredItems.filter((item) =>
      indikatorIDs.has(item.indikatorSubKegiatanId),
    );
  }

  currentRkItems = filteredItems.filter((item) =>
    matchesListSearch(item, state.q),
  );

  const { currentPage, totalPages } = updatePaginationUI(currentRkItems.length);
  const start = (currentPage - 1) * DEFAULT_PAGE_SIZE;
  currentRkPagedItems = currentRkItems.slice(start, start + DEFAULT_PAGE_SIZE);

  renderRencanaKerja(currentRkPagedItems);

  const indikatorForFilteredRk = filterIndikatorRencanaKerjaByRencanaKerjaItems(
    allIndikatorItems,
    currentRkItems,
  );
  const indikatorForPagedRk = filterIndikatorRencanaKerjaByRencanaKerjaItems(
    indikatorForFilteredRk,
    currentRkPagedItems,
  );

  setAkumulasiPagu(currentRkItems, indikatorForFilteredRk);
  const akRows = buildAkumulasiByKode(currentRkPagedItems, indikatorForPagedRk);
  renderAkumulasiByKode(akRows, currentPage, totalPages);

  const tahunLabel = state.tahun || defaultTahun();
  tahunAktifLabel.textContent = tahunLabel;
  rkMeta.textContent = `Total ${currentRkItems.length} | Halaman ${currentPage}/${totalPages}`;

  if (
    state.selected &&
    !currentRkItems.some((x) => x.id === Number(state.selected))
  ) {
    state.selected = "";
    state.selectedIndikator = "";
    writeUrlState(state, { replace: true });
  }

  setRkFormFromSelected();
}

async function loadIndikator() {
  const data = await listIndikatorByRencanaKerjaID(state.selected);
  currentIndikatorItems = data.map(normalizeIndikator);
  renderIndikator(currentIndikatorItems);

  if (
    state.selectedIndikator &&
    !currentIndikatorItems.some((x) => x.id === Number(state.selectedIndikator))
  ) {
    state.selectedIndikator = "";
    writeUrlState(state, { replace: true });
  }

  setIndikatorFormFromSelected();
}

async function reloadAll() {
  setStatus("Memuat data...");
  setRkLoading();
  setIndikatorLoading();
  setAkumulasiLoading();
  try {
    await loadAllIndikatorRencanaKerja();
    await loadRencanaKerja();
    await loadIndikator();
    setStatus("Data dimuat.");
  } catch (err) {
    setStatus(`Gagal memuat data: ${err.message}`, true);
  }
}

async function handleSaveRk() {
  if (!rkKodeInput.value.trim() || !rkNamaInput.value.trim()) {
    setStatus("Kode dan nama rencana kerja wajib diisi", true);
    return;
  }
  if (!rkIndikatorSKInput.value) {
    setStatus("Indikator sub kegiatan wajib dipilih", true);
    return;
  }
  if (!rkUnitInput.value) {
    setStatus("Unit pengusul wajib dipilih", true);
    return;
  }
  if (!currentUserID) {
    setStatus("User login tidak valid", true);
    return;
  }

  const payload = {
    indikator_sub_kegiatan_id: toNumber(rkIndikatorSKInput.value),
    unit_pengusul_id: toNumber(rkUnitInput.value),
    kode: rkKodeInput.value.trim(),
    nama: rkNamaInput.value.trim(),
    tahun: toNumber(rkTahunInput.value || new Date().getFullYear()),
    status: rkStatusInput.value || "DRAFT",
    dibuat_oleh: currentUserID,
    catatan: "",
  };

  if (isDuplicateKode(allRkItems, payload.kode, state.selected)) {
    setStatus(`Kode rencana kerja sudah digunakan: ${payload.kode}`, true);
    rkKodeInput.value = suggestNextKode(allRkItems, "RK");
    rkKodeInput.focus();
    rkKodeInput.select();
    return;
  }

  try {
    await saveRencanaKerja(payload, state.selected);
    setStatus(
      state.selected ? "Rencana kerja diperbarui." : "Rencana kerja dibuat.",
    );
    await loadAllIndikatorRencanaKerja();
    await loadRencanaKerja();
    await loadIndikator();
  } catch (err) {
    const message = String(err?.message || "");
    if (message.toLowerCase().includes("kode rencana_kerja sudah digunakan")) {
      setStatus("Kode rencana kerja sudah digunakan, gunakan kode lain.", true);
      rkKodeInput.value = suggestNextKode(allRkItems, "RK");
      rkKodeInput.focus();
      rkKodeInput.select();
      return;
    }
    setStatus(`Gagal simpan rencana kerja: ${message}`, true);
  }
}

async function handleDeleteRk() {
  if (!state.selected) {
    setStatus("Pilih rencana kerja yang akan dihapus", true);
    return;
  }

  if (!confirm(`Hapus rencana kerja ID ${state.selected}?`)) {
    return;
  }

  try {
    await deleteRencanaKerja(state.selected);
    state.selected = "";
    state.selectedIndikator = "";
    writeUrlState(state);
    resetRkForm();
    resetIndikatorForm();
    await reloadAll();
    setStatus("Rencana kerja dihapus.");
  } catch (err) {
    setStatus(`Gagal hapus rencana kerja: ${err.message}`, true);
  }
}

async function handleSaveIndikator() {
  if (!state.selected) {
    setStatus("Pilih rencana kerja dulu", true);
    return;
  }
  if (!irkKodeInput.value.trim() || !irkNamaInput.value.trim()) {
    setStatus("Kode dan nama indikator wajib diisi", true);
    return;
  }

  const payload = {
    rencana_kerja_id: toNumber(state.selected),
    kode: irkKodeInput.value.trim(),
    nama: irkNamaInput.value.trim(),
    satuan: irkSatuanInput.value.trim(),
    target_tahunan: toNumber(irkTargetInput.value),
    anggaran_tahunan: toNumber(irkAnggaranInput.value),
  };

  if (
    isDuplicateKode(allIndikatorItems, payload.kode, state.selectedIndikator)
  ) {
    setStatus(`Kode indikator sudah digunakan: ${payload.kode}`, true);
    irkKodeInput.value = suggestNextKode(allIndikatorItems, "IRK");
    irkKodeInput.focus();
    irkKodeInput.select();
    return;
  }

  try {
    await saveIndikator(payload, state.selectedIndikator);
    setStatus(
      state.selectedIndikator ? "Indikator diperbarui." : "Indikator dibuat.",
    );
    await loadAllIndikatorRencanaKerja();
    await loadRencanaKerja();
    await loadIndikator();
  } catch (err) {
    const message = String(err?.message || "");
    if (
      message
        .toLowerCase()
        .includes("kode indikator_rencana_kerja sudah digunakan")
    ) {
      setStatus(
        "Kode indikator rencana kerja sudah digunakan, gunakan kode lain.",
        true,
      );
      irkKodeInput.value = suggestNextKode(allIndikatorItems, "IRK");
      irkKodeInput.focus();
      irkKodeInput.select();
      return;
    }
    setStatus(`Gagal simpan indikator: ${message}`, true);
  }
}

async function handleDeleteIndikator() {
  if (!state.selectedIndikator) {
    setStatus("Pilih indikator yang akan dihapus", true);
    return;
  }

  if (!confirm(`Hapus indikator ID ${state.selectedIndikator}?`)) {
    return;
  }

  try {
    await deleteIndikator(state.selectedIndikator);
    state.selectedIndikator = "";
    writeUrlState(state);
    resetIndikatorForm();
    await loadAllIndikatorRencanaKerja();
    await loadRencanaKerja();
    await loadIndikator();
    setStatus("Indikator dihapus.");
  } catch (err) {
    setStatus(`Gagal hapus indikator: ${err.message}`, true);
  }
}

function bindEvents() {
  qInput.value = state.q;
  const initialTahun = normalizeTahunInput(state.tahun);
  state.tahun =
    initialTahun === null ? defaultTahun() : initialTahun || defaultTahun();
  tahunInput.value = state.tahun;
  tahunAktifLabel.textContent = state.tahun;
  skFilterInput.value = state.sk;

  let filterDebounceTimer = null;

  function applyTopFilters({ replaceHistory = false } = {}) {
    state.q = qInput.value.trim();
    const normalizedTahun = normalizeTahunInput(tahunInput.value);
    if (normalizedTahun === null) {
      setStatus("Tahun harus angka 2000-2100", true);
      tahunInput.focus();
      return;
    }
    state.tahun = normalizedTahun || defaultTahun();
    tahunInput.value = state.tahun;
    state.sk = skFilterInput.value;
    state.page = "1";
    writeUrlState(state, { replace: replaceHistory });
    void reloadAll();
  }

  function debounceApplyTopFilters() {
    if (filterDebounceTimer) {
      clearTimeout(filterDebounceTimer);
    }
    filterDebounceTimer = setTimeout(() => {
      applyTopFilters({ replaceHistory: true });
    }, 350);
  }

  btnRefresh.addEventListener("click", () => {
    applyTopFilters();
  });

  btnResetFilter.addEventListener("click", () => {
    qInput.value = "";
    state.q = "";
    state.page = "1";
    writeUrlState(state);
    void reloadAll();
  });

  qInput.addEventListener("input", debounceApplyTopFilters);
  qInput.addEventListener("keydown", (event) => {
    if (event.key !== "Enter") return;
    event.preventDefault();
    applyTopFilters();
  });
  tahunInput.addEventListener("input", debounceApplyTopFilters);
  tahunInput.addEventListener("blur", () => {
    const normalizedTahun = normalizeTahunInput(tahunInput.value);
    if (normalizedTahun === null) {
      tahunInput.value = state.tahun || defaultTahun();
      setStatus("Tahun tidak valid, dikembalikan ke tahun aktif.", true);
      return;
    }
    tahunInput.value = normalizedTahun || defaultTahun();
  });
  skFilterInput.addEventListener("change", () => {
    applyTopFilters({ replaceHistory: true });
  });

  rkPagePrevBtn.addEventListener("click", () => {
    const currentPage = Math.max(1, Number(state.page || 1));
    if (currentPage <= 1) return;
    state.page = String(currentPage - 1);
    writeUrlState(state);
    void loadRencanaKerja();
  });

  rkPageNextBtn.addEventListener("click", () => {
    const totalPages = Math.max(
      1,
      Math.ceil(currentRkItems.length / DEFAULT_PAGE_SIZE),
    );
    const currentPage = Math.max(1, Number(state.page || 1));
    if (currentPage >= totalPages) return;
    state.page = String(currentPage + 1);
    writeUrlState(state);
    void loadRencanaKerja();
  });

  rkNewBtn.addEventListener("click", () => {
    state.selected = "";
    state.selectedIndikator = "";
    writeUrlState(state);
    resetRkForm();
    resetIndikatorForm();
    renderRencanaKerja(currentRkPagedItems);
    renderIndikator([]);
    setStatus("Mode tambah rencana kerja.");
  });
  rkSaveBtn.addEventListener("click", () => {
    void handleSaveRk();
  });
  rkDeleteBtn.addEventListener("click", () => {
    void handleDeleteRk();
  });
  rkSubKegiatanInput.addEventListener("change", () => {
    setPaguFields(rkSubKegiatanInput.value);
    applyIndikatorSubKegiatanFilter();
  });

  irkNewBtn.addEventListener("click", () => {
    state.selectedIndikator = "";
    writeUrlState(state);
    resetIndikatorForm();
    renderIndikator(currentIndikatorItems);
    setStatus("Mode tambah indikator.");
  });
  irkSaveBtn.addEventListener("click", () => {
    void handleSaveIndikator();
  });
  irkDeleteBtn.addEventListener("click", () => {
    void handleDeleteIndikator();
  });

  window.addEventListener("popstate", () => {
    state = readUrlState(stateDefaults);
    qInput.value = state.q;
    const normalizedTahun = normalizeTahunInput(state.tahun);
    state.tahun =
      normalizedTahun === null
        ? defaultTahun()
        : normalizedTahun || defaultTahun();
    tahunInput.value = state.tahun;
    skFilterInput.value = state.sk;
    void reloadAll();
  });
}

(async function init() {
  if (!getAccessToken()) {
    redirectToLogin();
    return;
  }

  bindEvents();
  resetRkForm();
  resetIndikatorForm();

  try {
    await loadReferenceOptions();
  } catch (err) {
    setStatus(`Gagal memuat referensi: ${err.message}`, true);
  }

  void reloadAll();
})();

})();
