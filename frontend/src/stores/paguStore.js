import { defineStore } from 'pinia'
import axios from 'axios'

export const usePaguStore = defineStore('pagu', {
  state: () => ({
    items: [],
    loading: false,
    error: null,
    tahun: new Date().getFullYear(),
    searchQuery: '',
  }),
  getters: {
    totalPagu: (state) => state.items.reduce((sum, item) => sum + (item.pagu_tahun_ini || 0), 0),
    totalRK: (state) => state.items.reduce((sum, item) => sum + (item.total_rencana_kerja || 0), 0),
    deficitCount: (state) => state.items.filter(item => item.selisih < 0).length,
  },
  actions: {
    async fetchPagu() {
      this.loading = true
      this.error = null
      try {
        const token = localStorage.getItem('AUTH_TOKEN') || localStorage.getItem('authToken')
        // Vite proxy handles the base URL
        const res = await axios.get('/api/v1/pagu_sub_kegiatan_control', {
          params: {
            tahun: this.tahun,
            q: this.searchQuery
          },
          headers: { Authorization: `Bearer ${token}` }
        })
        this.items = res.data.data.items || []
      } catch (err) {
        this.error = "Gagal mengambil data pagu: " + (err.response?.data?.message || err.message)
      } finally {
        this.loading = false
      }
    },
    setTahun(tahun) {
      this.tahun = tahun
      this.fetchPagu()
    },
    setSearch(q) {
      this.searchQuery = q
    }
  }
})
