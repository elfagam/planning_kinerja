const { createApp, ref, onMounted, computed, watch } = Vue;

const app = createApp({
    setup() {
        const view = ref('list'); // 'list' or 'thread'
        const questions = ref([]);
        const totalQuestions = ref(0);
        const page = ref(1);
        const totalPages = computed(() => Math.ceil(totalQuestions.value / 10));
        const loading = ref(false);
        const filters = ref({
            search: '',
            status: 'all'
        });

        // Thread View State
        const currentThread = ref(null);
        const loadingThread = ref(false);
        const replyForm = ref({ content: '', best_practice: '' });
        const submittingReply = ref(false);

        // Sidebar FAQ State
        const faqs = ref([]);
        const faqLoading = ref(false);

        // Ask Modal State
        const askForm = ref({ title: '', content: '' });
        const submittingQuestion = ref({ loading: false });
        let askModal = null;

        // Auth State
        const currentUser = ref(null);

        const api = axios.create({
            baseURL: '/api/v1',
        });

        // Add request interceptor to attach correct AUTH_TOKEN
        api.interceptors.request.use(config => {
            const token = localStorage.getItem('AUTH_TOKEN') || localStorage.getItem('authToken');
            if (token) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        });

        const fetchQuestions = async () => {
            loading.value = true;
            try {
                const res = await api.get('/qna/questions', {
                    params: {
                        search: filters.value.search,
                        status: filters.value.status,
                        page: page.value
                    }
                });
                questions.value = res.data.data.items || [];
                totalQuestions.value = res.data.data.total;
            } catch (err) {
                console.error("Failed to fetch questions:", err);
            } finally {
                loading.value = false;
            }
        };

        const fetchFaqs = async () => {
            faqLoading.value = true;
            try {
                const res = await api.get('/qna/faq');
                faqs.value = res.data.data || [];
            } catch (err) {
                console.error("Failed to fetch FAQs:", err);
            } finally {
                faqLoading.value = false;
            }
        };

        let searchTimeout = null;
        const debounceSearch = () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                page.value = 1;
                fetchQuestions();
            }, 500);
        };

        const showThread = async (id) => {
            view.value = 'thread';
            loadingThread.value = true;
            currentThread.value = null;
            try {
                const res = await api.get(`/qna/questions/${id}`);
                currentThread.value = res.data.data;
            } catch (err) {
                alert("Gagal memuat detail pertanyaan.");
                view.value = 'list';
            } finally {
                loadingThread.value = false;
            }
        };

        const postAnswer = async () => {
            if (!replyForm.value.content) return;
            submittingReply.value = true;
            try {
                await api.post(`/qna/questions/${currentThread.value.id}/answers`, {
                    content: replyForm.value.content,
                    best_practice: replyForm.value.best_practice
                });
                replyForm.value.content = '';
                replyForm.value.best_practice = '';
                // Refresh thread
                await showThread(currentThread.value.id);
            } catch (err) {
                alert("Gagal mengirim jawaban.");
            } finally {
                submittingReply.value = false;
            }
        };

        const resolveQuestion = async (id) => {
            if (!confirm("Tandai pertanyaan ini sebagai Selesai?")) return;
            try {
                await api.post(`/qna/questions/${id}/resolve`);
                if (currentThread.value) currentThread.value.status = 'resolved';
                fetchQuestions();
            } catch (err) {
                alert("Hanya pembuat pertanyaan yang dapat menandai 'Selesai'.");
            }
        };

        const deleteQuestion = async (id) => {
            if (!confirm("Hapus pertanyaan ini? Tindakan ini tidak dapat dibatalkan.")) return;
            try {
                await api.delete(`/qna/questions/${id}`);
                alert("Pertanyaan berhasil dihapus.");
                view.value = 'list';
                fetchQuestions();
            } catch (err) {
                alert(err.response?.data?.error || "Gagal menghapus pertanyaan.");
            }
        };

        const canDelete = computed(() => {
            const role = currentUser.value?.role || '';
            return role === 'ADMIN' || role === 'PIMPINAN';
        });

        const showAskModal = () => {
            if (!askModal) {
                askModal = new bootstrap.Modal(document.getElementById('askModal'));
            }
            askModal.show();
        };

        const submitQuestion = async () => {
            if (!askForm.value.title || !askForm.value.content) {
                alert("Mohon isi judul dan detail pertanyaan.");
                return;
            }
            submittingQuestion.value = true;
            try {
                await api.post('/qna/questions', askForm.value);
                askForm.value = { title: '', content: '' };
                askModal.hide();
                fetchQuestions();
            } catch (err) {
                alert("Gagal mempublikasikan pertanyaan.");
            } finally {
                submittingQuestion.value = false;
            }
        };

        const formatDate = (dateStr) => {
            if (!dateStr) return '';
            const date = new Date(dateStr);
            return date.toLocaleDateString('id-ID', {
                day: 'numeric',
                month: 'short',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        };

        onMounted(() => {
            // Get user profile for RBAC
            api.get('/auth/me').then(res => {
                currentUser.value = res.data.data;
            });
            fetchQuestions();
            fetchFaqs();
        });

        return {
            view, questions, loading, filters, totalQuestions, page, totalPages,
            currentThread, loadingThread, replyForm, submittingReply,
            faqs, faqLoading,
            askForm, submittingQuestion, currentUser, canDelete,
            fetchQuestions, fetchFaqs, debounceSearch, showThread, postAnswer, resolveQuestion, deleteQuestion,
            showAskModal, submitQuestion, formatDate
        };
    }
});

app.mount('#app');
