const { createApp, ref, reactive, onMounted, computed, watch } = Vue;

createApp({
    setup() {
        const loading = ref(true);
        const submitting = ref(false);
        const treeData = ref([]);
        const expandedParents = ref([]);
        const subKegiatanItems = ref([]);
        const indikatorSkItems = ref([]);
        const unitItems = ref([]);
        const currentUserId = ref(null);
        
        const filters = reactive({
            search: '',
            tahun: new Date().getFullYear(),
            triwulan: '', 
            sub_kegiatan_id: '',
            sk_search: '' 
        });

        // Form RK (Parent)
        const rkForm = reactive({
            id: null,
            sub_kegiatan_id: '', // Dropdown level 1
            indikator_sub_kegiatan_id: '', // Dropdown level 2
            unit_pengusul_id: '',
            kode: '',
            nama: '',
            tahun: new Date().getFullYear(),
            triwulan: 1, // Default Triwulan 1
            status: 'DRAFT',
            dibuat_oleh: null,
            sk_search: ''
        });

        // Form IRK (Child)
        const irkForm = reactive({
            id: null,
            rencana_kerja_id: null,
            tb_standar_harga_id: null,
            kode: '',
            nama: '',
            satuan: '',
            target_tahunan: 0,
            harga_satuan: 0,
            anggaran_tahunan: 0,
            isManual: false
        });

        // Standar Harga Modal
        const shItems = ref([]);
        const shSearch = ref('');
        const shLoading = ref(false);
        let shModal = null;
        let rkDrawer = null;
        let irkDrawer = null;

        const api = axios.create({
            baseURL: '/api/v1'
        });

        // Auth handling
        const token = localStorage.getItem('AUTH_TOKEN');
        if (token) {
            api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
        }

        const fetchReferenceData = async () => {
            try {
                const [skRes, iskRes, unitRes] = await Promise.all([
                    api.get('/sub_kegiatan?all=true'),
                    api.get('/indikator_sub_kegiatan?all=true'),
                    api.get('/unit_pengusul?all=true')
                ]);
                
                const skItems = (skRes.data.data.items || []).map(normalizeData);
                skItems.sort((a, b) => (a.kode || "").localeCompare(b.kode || "", undefined, { numeric: true }));
                subKegiatanItems.value = skItems;
                
                indikatorSkItems.value = (iskRes.data.data.items || []).map(normalizeData);
                unitItems.value = (unitRes.data.data.items || []).map(normalizeData);
            } catch (err) {
                console.error("Reference data error:", err);
            }
        };

        const normalizeData = (item) => {
            return {
                id: item.id || item.ID,
                kode: item.kode || item.Kode,
                nama: item.nama || item.Nama,
                tahun: item.tahun || item.Tahun,
                triwulan: item.triwulan || item.Triwulan || 1,
                status: item.status || item.Status,
                unit_pengusul_id: item.unit_pengusul_id || item.UnitPengusulID,
                indikator_sub_kegiatan_id: item.indikator_sub_kegiatan_id || item.IndikatorSubKegiatanID,
                sub_kegiatan_id: item.sub_kegiatan_id || item.SubKegiatanID,
                dibuat_oleh: item.dibuat_oleh || item.DibuatOleh,
                // For Indikator
                rencana_kerja_id: item.rencana_kerja_id || item.RencanaKerjaID,
                tb_standar_harga_id: item.tb_standar_harga_id || item.TbStandarHargaID,
                satuan: item.satuan || item.Satuan,
                target_tahunan: item.target_tahunan || item.TargetTahunan,
                harga_satuan: item.harga_satuan || item.HargaSatuan,
                anggaran_tahunan: item.anggaran_tahunan || item.AnggaranTahunan
            };
        };

        const fetchData = async () => {
            loading.value = true;
            try {
                const params = {
                    all: true,
                    q: filters.search,
                    tahun: filters.tahun,
                    unit_pengusul_id: filters.unit_pengusul_id
                };

                // Filter Triwulan di API (jika backend mendukung)
                if (filters.triwulan) {
                    params.triwulan = filters.triwulan;
                }
                
                const [rkRes, irkRes] = await Promise.all([
                    api.get('/rencana_kerja', { params }),
                    api.get('/indikator_rencana_kerja', { params })
                ]);

                let rkList = (rkRes.data.data.items || []).map(normalizeData);
                const irkList = (irkRes.data.data.items || []).map(normalizeData);

                // Local Filter: Jika sub_kegiatan_id dipilih, saring rkList berdasarkan indikator yang terikat
                if (filters.sub_kegiatan_id) {
                    rkList = rkList.filter(rk => {
                        const isk = indikatorSkItems.value.find(i => i.id === rk.indikator_sub_kegiatan_id);
                        return isk && (isk.sub_kegiatan_id || isk.SubKegiatanID) == filters.sub_kegiatan_id;
                    });
                }

                // Local Filter: Filter berdasarkan Triwulan jika dipilih
                if (filters.triwulan) {
                    rkList = rkList.filter(rk => rk.triwulan == filters.triwulan);
                }

                // Build Hierarchy
                treeData.value = rkList.map(rk => {
                    const children = irkList.filter(irk => irk.rencana_kerja_id === rk.id);
                    const total_anggaran = children.reduce((sum, c) => sum + (Number(c.anggaran_tahunan) || 0), 0);
                    return { ...rk, children, total_anggaran };
                });

            } catch (err) {
                console.error("Fetch error:", err);
            } finally {
                loading.value = false;
            }
        };

        const toggleExpand = (id) => {
            const index = expandedParents.value.indexOf(id);
            if (index === -1) expandedParents.value.push(id);
            else expandedParents.value.splice(index, 1);
        };

        // RK Drawer Methods
        const openRkDrawer = (rk = null) => {
            if (rk) {
                // Find parent sub_kegiatan from selected indikator
                const isk = indikatorSkItems.value.find(i => i.id === rk.indikator_sub_kegiatan_id);
                Object.assign(rkForm, {
                    id: rk.id,
                    sub_kegiatan_id: isk ? (isk.sub_kegiatan_id || isk.SubKegiatanID) : '',
                    indikator_sub_kegiatan_id: rk.indikator_sub_kegiatan_id,
                    unit_pengusul_id: rk.unit_pengusul_id,
                    kode: rk.kode,
                    nama: rk.nama,
                    tahun: rk.tahun,
                    triwulan: rk.triwulan || 1,
                    status: rk.status,
                    dibuat_oleh: rk.dibuat_oleh
                });
            } else {
                Object.assign(rkForm, {
                    id: null,
                    sub_kegiatan_id: filters.sub_kegiatan_id,
                    indikator_sub_kegiatan_id: '',
                    unit_pengusul_id: unitItems.value[0]?.id || '',
                    kode: 'RK-' + Math.floor(Math.random() * 1000).toString().padStart(3, '0'),
                    nama: '',
                    tahun: filters.tahun,
                    triwulan: 1,
                    status: 'DRAFT',
                    dibuat_oleh: currentUserId.value
                });
            }
            if (!rkDrawer) rkDrawer = new bootstrap.Offcanvas(document.getElementById('rkDrawer'));
            rkDrawer.show();
        };

        const saveRk = async () => {
            submitting.value = true;
            try {
                const payload = { ...rkForm };
                // Bersihkan field yang tidak ada di database
                delete payload.sub_kegiatan_id;
                delete payload.sk_search;
                
                const method = payload.id ? 'put' : 'post';
                const url = payload.id ? `/rencana_kerja/${payload.id}` : '/rencana_kerja';
                await api[method](url, payload);
                rkDrawer.hide();
                fetchData();
            } catch (err) {
                alert(err.response?.data?.error || "Gagal menyimpan Rencana Kerja");
            } finally {
                submitting.value = false;
            }
        };

        const deleteRk = async (id) => {
            if (!confirm("Hapus Rencana Kerja ini beserta seluruh rinciannya?")) return;
            try {
                await api.delete(`/rencana_kerja/${id}`);
                rkDrawer.hide();
                fetchData();
            } catch (err) {
                alert("Gagal menghapus.");
            }
        };

        // IRK Drawer Methods
        const openIrkDrawer = (irk = null, parentId) => {
            if (irk) {
                Object.assign(irkForm, {
                    id: irk.id,
                    rencana_kerja_id: parentId,
                    tb_standar_harga_id: irk.tb_standar_harga_id,
                    kode: irk.kode,
                    nama: irk.nama,
                    satuan: irk.satuan,
                    target_tahunan: Number(irk.target_tahunan),
                    harga_satuan: Number(irk.harga_satuan),
                    isManual: !irk.tb_standar_harga_id
                });
            } else {
                Object.assign(irkForm, {
                    id: null,
                    rencana_kerja_id: parentId,
                    tb_standar_harga_id: null,
                    kode: 'IRK-' + Math.floor(Math.random() * 1000).toString().padStart(3, '0'),
                    nama: '',
                    satuan: '',
                    target_tahunan: 1,
                    harga_satuan: 0,
                    isManual: false
                });
            }
            if (!irkDrawer) irkDrawer = new bootstrap.Offcanvas(document.getElementById('irkDrawer'));
            irkDrawer.show();
        };

        const calculatedAnggaran = computed(() => {
            return (irkForm.target_tahunan || 0) * (irkForm.harga_satuan || 0);
        });

        const saveIrk = async () => {
            submitting.value = true;
            try {
                const { isManual, ...payload } = irkForm; // Hapus isManual dari payload
                payload.anggaran_tahunan = calculatedAnggaran.value;
                
                if (irkForm.isManual) payload.tb_standar_harga_id = null;
                
                const method = payload.id ? 'put' : 'post';
                const url = payload.id ? `/indikator_rencana_kerja/${payload.id}` : '/indikator_rencana_kerja';
                await api[method](url, payload);
                irkDrawer.hide();
                fetchData();
            } catch (err) {
                alert(err.response?.data?.error || "Gagal menyimpan rincian");
            } finally {
                submitting.value = false;
            }
        };

        const deleteIrk = async (id) => {
            if (!confirm("Hapus rincian ini?")) return;
            try {
                await api.delete(`/indikator_rencana_kerja/${id}`);
                irkDrawer.hide();
                fetchData();
            } catch (err) {
                alert("Gagal menghapus.");
            }
        };

        // Standar Harga Logic
        const openStandarHargaModal = () => {
            const el = document.getElementById('shModal');
            if (!shModal) shModal = new bootstrap.Modal(el);
            // Paksa hapus aria-hidden untuk menghindari konflik aksesibilitas saat drawer dibuka
            el.removeAttribute('aria-hidden');
            shModal.show();
            fetchStandarHarga();
        };

        const fetchStandarHarga = async () => {
            shLoading.value = true;
            try {
                const res = await api.get('/standar_harga', { params: { q: shSearch.value, limit: 20 } });
                shItems.value = res.data.data.items || [];
            } catch (err) {
                console.error("Standar Harga error:", err);
            } finally {
                shLoading.value = false;
            }
        };

        const selectStandarHarga = (sh) => {
            irkForm.tb_standar_harga_id = sh.id;
            
            // Gabungkan Uraian Barang dan Spesifikasi
            let fullNama = sh.uraian_barang || '';
            if (sh.spesifikasi && sh.spesifikasi.trim() !== "") {
                fullNama += " - " + sh.spesifikasi;
            }
            
            irkForm.nama = fullNama;
            irkForm.satuan = sh.satuan;
            irkForm.harga_satuan = sh.harga_satuan;
            irkForm.isManual = false;
            shModal.hide();
        };

        let searchTimeout = null;
        const debounceSearch = () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(fetchData, 500);
        };

        const debounceShSearch = () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(fetchStandarHarga, 500);
        };

        const filteredSkItems = computed(() => {
            const q = filters.sk_search.toLowerCase();
            return subKegiatanItems.value.filter(s => 
                s.kode.toLowerCase().includes(q) || s.nama.toLowerCase().includes(q)
            );
        });

        const filteredSkItemsDrawer = computed(() => {
            const q = rkForm.sk_search.toLowerCase();
            return subKegiatanItems.value.filter(s => 
                s.kode.toLowerCase().includes(q) || s.nama.toLowerCase().includes(q)
            );
        });

        const filteredIskItemsDrawer = computed(() => {
            if (!rkForm.sub_kegiatan_id) return [];
            return indikatorSkItems.value.filter(i => 
                (i.sub_kegiatan_id || i.SubKegiatanID) == rkForm.sub_kegiatan_id
            );
        });

        const formatMoney = (val) => {
            return new Intl.NumberFormat("id-ID", {
                style: "currency",
                currency: "IDR",
                minimumFractionDigits: 0
            }).format(val || 0);
        };

        const getStatusClass = (status) => {
            switch (status) {
                case 'DISETUJUI': return 'bg-success text-white px-2 rounded';
                case 'DIAJUKAN': return 'bg-primary text-white px-2 rounded';
                default: return 'bg-warning text-dark px-2 rounded';
            }
        };

        watch(() => filters.sub_kegiatan_id, () => {
            fetchData();
        });

        watch(() => filters.tahun, () => {
            fetchData();
        });

        watch(() => filters.triwulan, () => {
            fetchData();
        });

        onMounted(async () => {
            // Check auth first
            if (!token) {
                window.location.href = '/ui/login';
                return;
            }
            
            // Get current user info for ownership
            try {
                const meRes = await api.get('/auth/me');
                if (meRes.data.success) {
                    currentUserId.value = meRes.data.data.user_id || meRes.data.data.userID;
                }
            } catch (err) {
                console.error("Auth me error:", err);
            }

            await fetchReferenceData();
            fetchData();
        });

        return {
            loading, submitting, treeData, expandedParents, subKegiatanItems, indikatorSkItems, unitItems, filters,
            rkForm, irkForm, shItems, shSearch, shLoading, calculatedAnggaran,
            filteredSkItems, filteredSkItemsDrawer, filteredIskItemsDrawer,
            fetchData, toggleExpand, openRkDrawer, saveRk, deleteRk,
            openIrkDrawer, saveIrk, deleteIrk, openStandarHargaModal, fetchStandarHarga, selectStandarHarga,
            debounceSearch, debounceShSearch, formatMoney, getStatusClass
        };
    }
}).mount('#app');
