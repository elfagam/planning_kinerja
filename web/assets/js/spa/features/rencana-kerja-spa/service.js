import { fetchJSON } from "../../core/api.js";

export async function getCurrentUserID() {
  const me = await fetchJSON("/api/v1/auth/me");
  return Number(me?.user_id ?? me?.userID ?? 0);
}

export async function listUnitPengusul() {
  const data = await fetchJSON("/api/v1/unit_pengusul");
  return Array.isArray(data?.items) ? data.items : [];
}

export async function listIndikatorSubKegiatan() {
  const data = await fetchJSON("/api/v1/indikator_sub_kegiatan?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

export async function listSubKegiatan() {
  const data = await fetchJSON("/api/v1/sub_kegiatan?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

export async function listPaguSubKegiatan() {
  const data = await fetchJSON("/api/v1/pagu_sub_kegiatan?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

export async function listRencanaKerja({ q = "", tahun = "" } = {}) {
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

export async function listIndikatorByRencanaKerjaID(rencanaKerjaID) {
  if (!rencanaKerjaID) return [];

  const data = await fetchJSON("/api/v1/indikator_rencana_kerja?all=true");
  const items = Array.isArray(data?.items) ? data.items : [];
  const rkID = Number(rencanaKerjaID);

  return items.filter(
    (item) =>
      Number(item.rencana_kerja_id ?? item.RencanaKerjaID ?? 0) === rkID,
  );
}

export async function listAllIndikatorRencanaKerja() {
  const data = await fetchJSON("/api/v1/indikator_rencana_kerja?all=true");
  return Array.isArray(data?.items) ? data.items : [];
}

export async function saveRencanaKerja(payload, id = "") {
  const isEdit = Boolean(id);
  const url = isEdit ? `/api/v1/rencana_kerja/${id}` : "/api/v1/rencana_kerja";
  const method = isEdit ? "PUT" : "POST";

  return fetchJSON(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export async function deleteRencanaKerja(id) {
  return fetchJSON(`/api/v1/rencana_kerja/${id}`, { method: "DELETE" });
}

export async function saveIndikator(payload, id = "") {
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

export async function deleteIndikator(id) {
  return fetchJSON(`/api/v1/indikator_rencana_kerja/${id}`, {
    method: "DELETE",
  });
}
