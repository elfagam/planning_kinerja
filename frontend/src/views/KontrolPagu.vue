<script setup>
import { onMounted, ref, computed } from 'vue'
import { usePaguStore } from '../stores/paguStore'

const store = usePaguStore()
const searchInput = ref('')

const years = computed(() => {
  const currentYear = new Date().getFullYear()
  const list = []
  for (let i = currentYear - 2; i <= currentYear + 2; i++) {
    list.push(i)
  }
  return list
})

const handleSearch = () => {
  store.setSearch(searchInput.value)
  store.fetchPagu()
}

const handleRefresh = () => {
  store.fetchPagu()
}

const formatCurrency = (value) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0
  }).format(value || 0)
}

const getStatusBadgeClass = (diff) => {
  if (diff < 0) return 'bg-danger'
  if (diff === 0) return 'bg-success'
  return 'bg-warning text-dark'
}

const getStatusText = (diff) => {
  if (diff < 0) return 'Over Budget'
  if (diff === 0) return 'Matching'
  return 'Under Pagu'
}

onMounted(() => {
  store.fetchPagu()
})
</script>

<template>
  <div class="p-4 p-md-5">
    <header class="mb-4 p-4 card shadow-sm border-0 bg-primary text-white">
      <div class="d-flex justify-content-between align-items-center">
        <div>
          <h1 class="h3 mb-1">Kontrol Pagu & Rencana Kerja</h1>
          <p class="mb-0 opacity-75">Perbandingan Pagu Anggaran dengan Total Rencana Kerja per Sub Kegiatan</p>
        </div>
      </div>
    </header>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-body">
        <div class="row g-3 align-items-end">
          <div class="col-md-3">
            <label class="form-label small text-muted text-uppercase fw-bold">Tahun Anggaran</label>
            <select v-model="store.tahun" @change="store.fetchPagu()" class="form-select border-0 bg-light">
              <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
            </select>
          </div>
          <div class="col-md-4">
            <label class="form-label small text-muted text-uppercase fw-bold">Cari Sub Kegiatan</label>
            <div class="input-group">
              <input v-model="searchInput" @keyup.enter="handleSearch" type="text" class="form-control border-0 bg-light" placeholder="Kode atau nama sub kegiatan...">
              <button @click="handleSearch" class="btn btn-primary px-4">Cari</button>
            </div>
          </div>
          <div class="col-md-5 text-end">
            <button @click="handleRefresh" class="btn btn-outline-secondary btn-sm">
              Refresh Data
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="row g-4 mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100 bg-primary text-white">
          <div class="card-body p-4">
            <h6 class="text-uppercase small opacity-75">Total Pagu (N)</h6>
            <h3 class="mb-0">{{ formatCurrency(store.totalPagu) }}</h3>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100 bg-info text-white">
          <div class="card-body p-4">
            <h6 class="text-uppercase small opacity-75">Total Rencana Kerja</h6>
            <h3 class="mb-0">{{ formatCurrency(store.totalRK) }}</h3>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm h-100 text-white" :class="store.deficitCount > 0 ? 'bg-danger' : 'bg-success'">
          <div class="card-body p-4">
            <h6 class="text-uppercase small opacity-75">Sub Kegiatan Over Budget</h6>
            <h3 class="mb-0">{{ store.deficitCount }}</h3>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm overflow-hidden">
      <div class="card-body p-0">
        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead class="bg-light">
              <tr>
                <th class="ps-4 py-3 text-uppercase small text-muted">Sub Kegiatan</th>
                <th class="text-end py-3 text-uppercase small text-muted">Pagu N-1</th>
                <th class="text-end py-3 text-uppercase small text-muted">Pagu N</th>
                <th class="text-end py-3 text-uppercase small text-muted">Rencana Kerja (RK)</th>
                <th class="text-end py-3 text-uppercase small text-muted">Selisih (Pagu N - RK)</th>
                <th class="text-center py-3 text-uppercase small text-muted">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="store.loading">
                <td colspan="6" class="text-center py-5">
                  <div class="spinner-border text-primary" role="status"></div>
                </td>
              </tr>
              <tr v-else-if="store.items.length === 0">
                <td colspan="6" class="text-center py-5 text-muted">Data tidak ditemukan untuk tahun {{ store.tahun }}</td>
              </tr>
              <tr v-for="item in store.items" :key="item.kode">
                <td class="ps-4">
                  <div class="fw-bold text-dark">{{ item.kode }}</div>
                  <div class="small text-muted">{{ item.nama }}</div>
                </td>
                <td class="text-end text-nowrap">{{ formatCurrency(item.pagu_tahun_sebelumnya) }}</td>
                <td class="text-end text-nowrap fw-semibold text-primary">{{ formatCurrency(item.pagu_tahun_ini) }}</td>
                <td class="text-end text-nowrap">{{ formatCurrency(item.total_rencana_kerja) }}</td>
                <td class="text-end text-nowrap" :class="item.selisih < 0 ? 'text-danger' : 'text-success'">
                  {{ formatCurrency(item.selisih) }}
                </td>
                <td class="text-center">
                  <span class="badge" :class="getStatusBadgeClass(item.selisih)" style="border-radius: 50rem; padding: 0.4rem 0.8rem; font-size: 0.75rem;">
                    {{ getStatusText(item.selisih) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="card-footer bg-white border-0 py-3 ps-4">
        <small class="text-muted">Menampilkan {{ store.items.length }} sub kegiatan untuk tahun {{ store.tahun }}</small>
      </div>
    </div>
  </div>
</template>

<style scoped>
.header-gradient {
  background: linear-gradient(135deg, #0d6efd 0%, #0a58ca 100%);
}
.bg-soft-pattern {
  background-color: #f8f9fa;
  background-image: radial-gradient(#dee2e6 0.5px, transparent 0.5px);
  background-size: 20px 20px;
}
</style>
