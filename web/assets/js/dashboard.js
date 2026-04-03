      const fmtNum = (n) =>
        new Intl.NumberFormat("id-ID").format(Number(n || 0));
      const fmtPct = (n) => `${Number(n || 0).toFixed(2)}%`;
      const fmtRupiah = (n) =>
        new Intl.NumberFormat("id-ID", {
          style: "currency",
          currency: "IDR",
          maximumFractionDigits: 0,
        }).format(Number(n || 0));

      const yearFilter = document.getElementById("yearFilter");
      const refreshBtn = document.getElementById("refreshBtn");
      const msg = document.getElementById("dashboardMsg");
      const targetSearch = document.getElementById("targetSearch");
      const targetLimit = document.getElementById("targetLimit");
      const targetReset = document.getElementById("targetReset");
      const targetPrev = document.getElementById("targetPrev");
      const targetNext = document.getElementById("targetNext");
      const targetMeta = document.getElementById("targetMeta");
      const retrySummary = document.getElementById("retrySummary");
      const retryStats = document.getElementById("retryStats");
      const retryChart = document.getElementById("retryChart");
      const retryYearly = document.getElementById("retryYearly");
      const retryRanking = document.getElementById("retryRanking");
      const retryTargetTable = document.getElementById("retryTargetTable");
      const DEFAULT_YEAR_KEY = "DEFAULT_YEAR";
      const DEFAULT_YEAR_USER_PREFIX = "DEFAULT_YEAR_USER_";

      let targetChart;
      let statusChart;
      let targetPage = 1;
      let targetTotalPages = 1;
      const indikatorLookup = new Map();
      const indikatorToRencanaKerjaLookup = new Map();
      const rencanaKerjaLookup = new Map();
      let indikatorLookupLoaded = false;
      let rencanaKerjaLookupLoaded = false;
      let currentChartMode = "triwulan";
      let currentMetricMode = "kinerja"; // 'kinerja' or 'keuangan'
      let currentUserProfile = null;

      function normalizeYearValue(raw) {
        const year = Number(String(raw || "").trim());
        if (Number.isInteger(year) && year >= 2000 && year <= 2100) {
          return String(year);
        }
        return "";
      }

      function getCurrentUserID() {
        return Number(
          currentUserProfile?.user_id ?? currentUserProfile?.userID ?? 0,
        );
      }

      function userYearStorageKey(userID) {
        return `${DEFAULT_YEAR_USER_PREFIX}${userID}`;
      }


      function getStoredDefaultYear() {
        return normalizeYearValue(localStorage.getItem(DEFAULT_YEAR_KEY) || "");
      }

      function getStoredDefaultYearForCurrentUser() {
        const userID = getCurrentUserID();
        if (!userID) {
          return getStoredDefaultYear();
        }

        const userYear = normalizeYearValue(
          localStorage.getItem(userYearStorageKey(userID)) || "",
        );
        if (userYear) {
          return userYear;
        }

        const fallbackYear = getStoredDefaultYear();
        if (fallbackYear) {
          localStorage.setItem(userYearStorageKey(userID), fallbackYear);
          return fallbackYear;
        }
        return "";
      }

      function getYearFromURL() {
        const params = new URLSearchParams(window.location.search);
        const raw = String(params.get("tahun") || "").trim();
        const year = Number(raw);
        if (Number.isInteger(year) && year >= 2000 && year <= 2100) {
          return String(year);
        }
        return "";
      }

      function getEffectiveYearFilter() {
        return (
          normalizeYearValue(yearFilter.value) ||
          getStoredDefaultYearForCurrentUser()
        );
      }

      function writeYearToURL() {
        const params = new URLSearchParams(window.location.search);
        const tahun = String(getEffectiveYearFilter() || "").trim();

        if (tahun) {
          params.set("tahun", tahun);
        } else {
          params.delete("tahun");
        }

        const query = params.toString();
        const nextURL = query
          ? `${window.location.pathname}?${query}`
          : window.location.pathname;

        window.history.replaceState(null, "", nextURL);
      }

      function applyInitialYearFilter() {
        const preferredYear =
          getStoredDefaultYearForCurrentUser() || getYearFromURL();
        if (!preferredYear || yearFilter.value) {
          return;
        }

        if (
          !Array.from(yearFilter.options).some(
            (opt) => opt.value === preferredYear,
          )
        ) {
          const option = document.createElement("option");
          option.value = preferredYear;
          option.textContent = preferredYear;
          yearFilter.appendChild(option);
        }

        yearFilter.value = preferredYear;
      }


      async function loadCurrentUserProfile() {
        const me = await fetchJSON("/api/v1/auth/me");
        console.log("Current User Role:", me?.role);
        console.log("Is Admin?", me?.role === 'ADMIN');
        currentUserProfile = me || null;

        const userLabel =
          me?.full_name ||
          me?.nama_lengkap ||
          me?.name ||
          me?.email ||
          "Pengguna";
        const userDesktop = document.getElementById("sidebarUserDesktop");
        const userMobile = document.getElementById("sidebarUserMobile");
        if (userDesktop) userDesktop.textContent = userLabel;
        if (userMobile) userMobile.textContent = userLabel;


        const roleLabel = me?.role ? me.role : "";
        if (roleLabel) {
          const el1 = document.getElementById("sidebarRoleDesktop");
          const el2 = document.getElementById("sidebarRoleMobile");
          if (el1) el1.textContent = roleLabel;
          if (el2) el2.textContent = roleLabel;
        }

        // Display unit_pengusul after role in sidebar
        const unitLabel = me?.unit_pengusul_nama || me?.unit_pengusul_name || me?.unit_pengusul || "";
        const unitEl = document.getElementById("sidebarUnitPengusulDesktop");
        if (unitEl) {
          unitEl.textContent = unitLabel ? `Unit: ${unitLabel}` : "";
        }

        const userID = getCurrentUserID();
        if (!userID) {
          return;
        }

        const profileYear = normalizeYearValue(
          me?.default_year ?? me?.tahun_default ?? me?.defaultYear ?? "",
        );
        if (profileYear) {
          localStorage.setItem(userYearStorageKey(userID), profileYear);
        }
      }

      function setSummaryCards(summary) {
        document.getElementById("kpiProgram").textContent = fmtNum(
          summary.total_program,
        );
        document.getElementById("kpiKegiatan").textContent = fmtNum(
          summary.total_kegiatan,
        );
        document.getElementById("kpiAnggaran").textContent = fmtRupiah(
          summary.total_anggaran,
        );
        document.getElementById("kpiRealisasiAnggaran").textContent = fmtRupiah(
          summary.total_realisasi_anggaran,
        );
        document.getElementById("kpiPersenAnggaran").textContent =
          `${fmtPct(summary.persentase_realisasi_anggaran)} dari total anggaran`;
      }

      function drawStatusChart(stats) {
        const onTrack = Number(stats.total_status_on_track || 0);
        const warning = Number(stats.total_status_warning || 0);
        const offTrack = Number(stats.total_status_off_track || 0);
        const unknown = Number(stats.total_status_unknown || 0);
        const ctx = document.getElementById("statusDistributionChart");
        if (statusChart) statusChart.destroy();

        const labels = ["OnTrack", "Warning", "OffTrack", "UnKnown"];
        const values = [onTrack, warning, offTrack, unknown];
        const bgColors = ["#1f8f67", "#cb8a1b", "#c9434f", "#5b6970"];
        const borderColors = ["#ffffff", "#ffffff", "#ffffff", "#ffffff"];

        statusChart = new Chart(ctx, {
          type: "doughnut",
          data: {
            labels,
            datasets: [
              {
                data: values,
                backgroundColor: bgColors,
                borderColor: borderColors,
                borderWidth: 2,
                hoverOffset: 6,
              },
            ],
          },
          options: {
            responsive: true,
            plugins: { legend: { position: "bottom" } },
            cutout: "62%",
          },
        });
      }

      function setStats(stats) {
        const distribution = Array.isArray(stats.status_distribution)
          ? stats.status_distribution
          : [];

        const derivedTotals = {
          total_status_on_track: 0,
          total_status_warning: 0,
          total_status_off_track: 0,
          total_status_unknown: 0,
        };

        distribution.forEach((item) => {
          const normalized = normalizeStatusKey(item?.status);
          if (normalized === "ON_TRACK") {
            derivedTotals.total_status_on_track += Number(item?.total || 0);
          } else if (normalized === "WARNING") {
            derivedTotals.total_status_warning += Number(item?.total || 0);
          } else if (normalized === "OFF_TRACK") {
            derivedTotals.total_status_off_track += Number(item?.total || 0);
          } else {
            derivedTotals.total_status_unknown += Number(item?.total || 0);
          }
        });

        const onTrack = Number(
          stats.total_status_on_track ?? derivedTotals.total_status_on_track,
        );
        const warning = Number(
          stats.total_status_warning ?? derivedTotals.total_status_warning,
        );
        const offTrack = Number(
          stats.total_status_off_track ?? derivedTotals.total_status_off_track,
        );
        const unknown = Number(
          stats.total_status_unknown ?? derivedTotals.total_status_unknown,
        );
        const totalData = Number(
          stats.total_data || onTrack + warning + offTrack + unknown || 0,
        );

        const pct = (value, fallbackField) => {
          if (stats[fallbackField] != null) {
            return fmtPct(Number(stats[fallbackField] || 0));
          }
          if (!totalData) {
            return "0.00%";
          }
          return fmtPct((Number(value || 0) / totalData) * 100);
        };

        document.getElementById("statOnTrack").textContent = fmtNum(onTrack);
        document.getElementById("statWarning").textContent = fmtNum(warning);
        document.getElementById("statOffTrack").textContent = fmtNum(offTrack);
        document.getElementById("statUnknown").textContent = fmtNum(unknown);
        document.getElementById("statOnTrackPct").textContent = pct(
          onTrack,
          "persentase_status_on_track",
        );
        document.getElementById("statWarningPct").textContent = pct(
          warning,
          "persentase_status_warning",
        );
        document.getElementById("statOffTrackPct").textContent = pct(
          offTrack,
          "persentase_status_off_track",
        );
        document.getElementById("statUnknownPct").textContent = pct(
          unknown,
          "persentase_status_unknown",
        );
        document.getElementById("statRataCapaian").textContent = fmtPct(
          stats.rata_rata_capaian_persen,
        );
        drawStatusChart(stats);
      }

      function drawTargetChart(chartData) {
        const ctx = document.getElementById("targetRealisasiChart");
        if (targetChart) targetChart.destroy();

        const isFinance = currentMetricMode === "keuangan";
        const targets = isFinance ? chartData.series?.target_anggaran : chartData.series?.target;
        const reals = isFinance ? chartData.series?.realisasi_anggaran : chartData.series?.realisasi;

        targetChart = new Chart(ctx, {
          type: "line",
          data: {
            labels: chartData.categories || [],
            datasets: [
              {
                label: isFinance ? "Target Anggaran" : "Target",
                data: targets || [],
                borderColor: isFinance ? "#1f6fa8" : "#1f6fa8",
                backgroundColor: isFinance ? "rgba(31,111,168,.16)" : "rgba(31,111,168,.16)",
                fill: true,
                tension: 0.3,
              },
              {
                label: isFinance ? "Realisasi Anggaran" : "Realisasi",
                data: reals || [],
                borderColor: isFinance ? "#e67e22" : "#0a6b65",
                backgroundColor: isFinance ? "rgba(230,126,34,.16)" : "rgba(10,107,101,.16)",
                fill: true,
                tension: 0.3,
              },
            ],
          },
          options: {
            responsive: true,
            plugins: { 
              legend: { position: "bottom" },
              tooltip: {
                callbacks: {
                  label: function(context) {
                    let label = context.dataset.label || '';
                    if (label) label += ': ';
                    if (context.parsed.y !== null) {
                      label += isFinance ? fmtRupiah(context.parsed.y) : fmtNum(context.parsed.y);
                    }
                    return label;
                  }
                }
              }
            },
            scales: { 
              y: { 
                ticks: { 
                  callback: (v) => isFinance ? fmtRupiah(v) : fmtNum(v) 
                } 
              } 
            },
          },
        });
      }

      function setYearlyTable(yearly) {
        const body = document.getElementById("yearlyBody");
        body.innerHTML = "";

        const items = Array.isArray(yearly.items) ? yearly.items : [];
        if (!items.length) {
          body.innerHTML =
            '<tr><td colspan="4" class="text-center text-muted py-4">Tidak ada data ringkasan tahunan untuk filter tahun saat ini.</td></tr>';
          return;
        }

        items.forEach((item) => {
          body.insertAdjacentHTML(
            "beforeend",
            `<tr>
              <td>${item.tahun}</td>
              <td class="text-end">
                <div class="fw-semibold">${fmtNum(item.total_target_nilai)}</div>
                <div class="small text-muted">${fmtRupiah(item.total_target_anggaran)}</div>
              </td>
              <td class="text-end">
                <div class="fw-semibold">${fmtNum(item.total_realisasi_nilai)}</div>
                <div class="small text-muted">${fmtRupiah(item.total_realisasi_anggaran)}</div>
              </td>
              <td class="text-end">
                <div class="fw-semibold">${fmtPct(item.persentase_realisasi_target)}</div>
                <div class="small text-muted">${fmtPct(item.persentase_realisasi_anggaran)}</div>
              </td>
            </tr>`,
          );
        });

        const meta = document.getElementById("yearlyMeta");
        if (meta) {
          meta.textContent =
            yearly.data_source === "realisasi_rencana_kerja"
              ? "Target diproksi dari target tahunan indikator RK · sumber: realisasi rencana kerja"
              : "";
        }
      }

      function updateYearlyActiveYearLabel() {
        const el = document.getElementById("yearlyActiveYear");
        if (!el) return;
        const tahun = getEffectiveYearFilter();
        el.textContent = tahun
          ? `Tahun aktif: ${tahun}`
          : "Tahun aktif: semua tahun";
      }

      function setRankingTable(ranking) {
        const body = document.getElementById("rankingBody");
        const meta = document.getElementById("rankingMeta");
        body.innerHTML = "";

        const items = Array.isArray(ranking.items) ? ranking.items : [];
        if (!items.length) {
          body.innerHTML =
            '<tr><td colspan="3" class="text-center text-muted py-4">Tidak ada data ranking program untuk filter tahun saat ini.</td></tr>';
          if (meta) {
            meta.textContent = "Menampilkan 0 data ranking program.";
          }
          return;
        }

        items.forEach((item) => {
          body.insertAdjacentHTML(
            "beforeend",
            `<tr>
              <td class="fw-semibold">${item.rank}</td>
              <td>
                <div class="fw-semibold">${item.program_nama}</div>
                <div class="small text-muted">${item.program_kode}</div>
              </td>
              <td class="text-end fw-semibold">${fmtPct(item.persentase_realisasi_target)}</td>
            </tr>`,
          );
        });

        if (meta) {
          const total = Number(ranking.total || items.length || 0);
          const srcNote =
            ranking.data_source === "realisasi_rencana_kerja"
              ? " · sumber: realisasi rencana kerja"
              : "";
          meta.textContent = `Menampilkan ${items.length} dari ${total} program${srcNote}.`;
        }
      }

      function statusClass(status) {
        const normalized = normalizeStatusKey(status);
        if (normalized === "ON_TRACK") return "status-on-track";
        if (normalized === "WARNING") return "status-warning";
        if (normalized === "OFF_TRACK") return "status-off-track";
        return "status-unknown";
      }

      function statusLabel(status) {
        const normalized = normalizeStatusKey(status);
        if (normalized === "ON_TRACK") return "OnTrack";
        if (normalized === "WARNING") return "Warning";
        if (normalized === "OFF_TRACK") return "OffTrack";
        return "UnKnown";
      }

      function normalizeStatusKey(status) {
        const raw = String(status || "").trim();
        if (!raw) {
          return "UNKNOWN";
        }

        const compact = raw
          .toUpperCase()
          .replace(/[\s-]+/g, "_")
          .replace(/^ONTRACK$/, "ON_TRACK")
          .replace(/^OFFTRACK$/, "OFF_TRACK");

        if (compact === "ON_TRACK") return "ON_TRACK";
        if (compact === "WARNING") return "WARNING";
        if (compact === "OFF_TRACK") return "OFF_TRACK";
        return "UNKNOWN";
      }

      function indikatorLabel(indikatorID) {
        return (
          indikatorLookup.get(Number(indikatorID)) ||
          `Indikator #${indikatorID}`
        );
      }

      function rencanaKerjaLabel(rkID) {
        return rencanaKerjaLookup.get(Number(rkID)) || `Rencana Kerja #${rkID}`;
      }

      async function loadRencanaKerjaLookup(force = false) {
        if (rencanaKerjaLookupLoaded && !force) {
          return;
        }

        const data = await fetchJSON("/api/v1/rencana_kerja?all=true");
        const items = Array.isArray(data?.items) ? data.items : [];

        rencanaKerjaLookup.clear();
        items.forEach((item) => {
          const id = Number(item.id ?? item.ID ?? 0);
          if (!id) return;
          const kode = item.kode ?? item.Kode ?? "";
          const nama = item.nama ?? item.Nama ?? "";
          rencanaKerjaLookup.set(id, `${kode} - ${nama}`);
        });

        rencanaKerjaLookupLoaded = true;
      }

      async function loadIndikatorLookup(force = false) {
        if (indikatorLookupLoaded && !force) {
          return;
        }

        const data = await fetchJSON(
          "/api/v1/indikator_rencana_kerja?all=true",
        );
        const items = Array.isArray(data?.items) ? data.items : [];

        indikatorLookup.clear();
        indikatorToRencanaKerjaLookup.clear();
        items.forEach((item) => {
          const id = Number(item.id ?? item.ID ?? 0);
          if (!id) return;
          const kode = item.kode ?? item.Kode ?? "";
          const nama = item.nama ?? item.Nama ?? "";
          indikatorLookup.set(id, `${kode} - ${nama}`);

          const rencanaKerjaID = Number(
            item.rencana_kerja_id ?? item.RencanaKerjaID ?? 0,
          );
          if (rencanaKerjaID) {
            indikatorToRencanaKerjaLookup.set(id, rencanaKerjaID);
          }
        });

        indikatorLookupLoaded = true;
      }

      function renderTargetTablePage(rows, page, totalPages, total) {
        const body = document.getElementById("targetBody");
        body.innerHTML = "";

        if (!rows.length) {
          body.innerHTML =
            '<tr><td colspan="7" class="text-center text-muted py-4">Tidak ada data target/realisasi untuk nama indikator tersebut.</td></tr>';
          targetMeta.textContent = "Page 1 | Total 0";
          targetPrev.disabled = true;
          targetNext.disabled = true;
          return;
        }

        rows.forEach((item) => {
          body.insertAdjacentHTML(
            "beforeend",
            `<tr>
              <td>${item.ID}</td>
              <td>
                <div class="small text-muted">ID #${item.RencanaKerjaID}</div>
                <div class="fw-semibold">${rencanaKerjaLabel(item.RencanaKerjaID)}</div>
              </td>
              <td>${item.Tahun}-T${item.Triwulan}</td>
              <td class="text-end">${fmtNum(item.TargetNilai)}</td>
              <td class="text-end">${fmtNum(item.RealisasiNilai)}</td>
              <td class="text-end fw-semibold">${fmtPct(item.CapaianPersen)}</td>
              <td><span class="status-pill ${statusClass(item.Status)}">${statusLabel(item.Status)}</span></td>
            </tr>`,
          );
        });

        targetMeta.textContent = `Page ${page} / ${totalPages} | Total ${total}`;
        targetPrev.disabled = page <= 1;
        targetNext.disabled = page >= totalPages;
      }

      function setTargetTable(payload) {
        const rows = payload.items || [];
        targetPage = Number(payload.page || 1);
        targetTotalPages = Number(payload.total_pages || 1);
        renderTargetTablePage(
          rows,
          targetPage,
          targetTotalPages,
          Number(payload.total || 0),
        );
      }

      async function loadTargetTable() {
        const tahun = getEffectiveYearFilter();
        const params = new URLSearchParams();
        params.set("page", String(targetPage));
        params.set("limit", targetLimit.value || "10");
        const q = (targetSearch.value || "").trim();
        if (q) params.set("q", q);

        if (tahun) params.set("tahun", tahun);

        try {
          await Promise.all([loadIndikatorLookup(), loadRencanaKerjaLookup()]);
          const payload = await fetchJSON(
            `/api/v1/performance/target-realisasi?${params.toString()}`,
          );
          setTargetTable(payload);
        } catch (err) {
          const body = document.getElementById("targetBody");
          body.innerHTML =
            '<tr><td colspan="7" class="text-center text-danger py-4">Gagal memuat tabel target/realisasi.</td></tr>';
          targetMeta.textContent = "Page 1 | Total 0";
          targetPrev.disabled = true;
          targetNext.disabled = true;
          throw err;
        }
      }

      function updateYearOptions(yearly) {
        const years = (yearly.items || []).map((x) => x.tahun);
        const selectedYear = yearFilter.value;
        const urlYear = getYearFromURL();
        const defaultYear = getStoredDefaultYearForCurrentUser();

        const optionYears = defaultYear ? [Number(defaultYear)] : years;

        if (defaultYear) {
          yearFilter.innerHTML = optionYears
            .map((y) => `<option value="${y}">${y}</option>`)
            .join("");
          yearFilter.disabled = true;
          yearFilter.value = defaultYear;
          writeYearToURL();
          updateYearlyActiveYearLabel();
          return;
        }

        yearFilter.innerHTML = `<option value="">Semua Tahun</option>${optionYears.map((y) => `<option value="${y}">${y}</option>`).join("")}`;
        yearFilter.disabled = false;

        if (selectedYear && years.includes(Number(selectedYear))) {
          yearFilter.value = selectedYear;
          writeYearToURL();
          updateYearlyActiveYearLabel();
          return;
        }
        if (urlYear && optionYears.includes(Number(urlYear))) {
          yearFilter.value = urlYear;
          writeYearToURL();
          updateYearlyActiveYearLabel();
          return;
        }
        if (defaultYear && optionYears.includes(Number(defaultYear))) {
          yearFilter.value = defaultYear;
          writeYearToURL();
          updateYearlyActiveYearLabel();
          return;
        }

        if (selectedYear && optionYears.includes(Number(selectedYear))) {
          yearFilter.value = selectedYear;
          writeYearToURL();
          updateYearlyActiveYearLabel();
          return;
        }

        yearFilter.value = "";
        writeYearToURL();
        updateYearlyActiveYearLabel();
      }

      const viewTriwulan = document.getElementById("viewTriwulan");
      const viewBulan = document.getElementById("viewBulan");
      const modeKinerja = document.getElementById("modeKinerja");
      const modeKeuangan = document.getElementById("modeKeuangan");
      const chartSubTitle = document.getElementById("chartSubTitle");

      if (viewTriwulan) {
        viewTriwulan.addEventListener("change", () => {
          if (viewTriwulan.checked) {
            currentChartMode = "triwulan";
            if (chartSubTitle) chartSubTitle.textContent = "Agregat per triwulan";
            loadChartWidget();
          }
        });
      }
      if (viewBulan) {
        viewBulan.addEventListener("change", () => {
          if (viewBulan.checked) {
            currentChartMode = "bulan";
            if (chartSubTitle) chartSubTitle.textContent = "Agregat per bulan";
            loadChartWidget();
          }
        });
      }
      if (modeKinerja) {
        modeKinerja.addEventListener("change", () => {
          if (modeKinerja.checked) {
            currentMetricMode = "kinerja";
            loadChartWidget();
          }
        });
      }
      if (modeKeuangan) {
        modeKeuangan.addEventListener("change", () => {
          if (modeKeuangan.checked) {
            currentMetricMode = "keuangan";
            loadChartWidget();
          }
        });
      }

      async function loadSummaryWidget() {
        const summary = await fetchJSON(
          "/api/v1/performance/dashboard-summary",
        );
        setSummaryCards(summary);
      }

      async function loadStatsWidget() {
        const tahun = getEffectiveYearFilter();
        const q = tahun ? `?tahun=${encodeURIComponent(tahun)}` : "";
        const stats = await fetchJSON(
          `/api/v1/performance/status-distribution${q}`,
        );
        setStats(stats);
      }

      async function loadYearlyWidget() {
        const tahun = getEffectiveYearFilter();
        const params = new URLSearchParams();
        if (tahun) {
          params.set("tahun_start", tahun);
          params.set("tahun_end", tahun);
        }

        const q = params.toString();
        const yearly = await fetchJSON(
          `/api/v1/performance/yearly-summary${q ? `?${q}` : ""}`,
        );
        setYearlyTable(yearly);
        updateYearOptions(yearly);
      }

      async function loadChartWidget() {
        const tahun = getEffectiveYearFilter();
        const params = new URLSearchParams();
        if (tahun) params.set("tahun", tahun);
        params.set("group_by", currentChartMode);

        const chartData = await fetchJSON(
          `/api/v1/performance/chart-target-vs-realisasi?${params.toString()}`,
        );
        drawTargetChart(chartData);
      }

      async function loadRankingWidget() {
        const tahun = getEffectiveYearFilter();
        const rankingParams = new URLSearchParams();
        rankingParams.set("limit", "5");
        if (tahun) {
          rankingParams.set("tahun", tahun);
        }

        let ranking = await fetchJSON(
          `/api/v1/performance/program-ranking?${rankingParams.toString()}`,
        );

        const hasRankingData =
          Array.isArray(ranking?.items) && ranking.items.length > 0;
        if (!hasRankingData && tahun) {
          const fallbackParams = new URLSearchParams();
          fallbackParams.set("limit", "5");
          ranking = await fetchJSON(
            `/api/v1/performance/program-ranking?${fallbackParams.toString()}`,
          );
          const meta = document.getElementById("rankingMeta");
          if (meta) {
            meta.textContent =
              "Data tahun aktif kosong, menampilkan ranking semua tahun.";
          }
        }

        setRankingTable(ranking);
      }

      async function loadQnaWidget() {
        const loadingEl = document.getElementById("qnaWidgetLoading");
        const contentEl = document.getElementById("qnaWidgetContent");
        const emptyEl = document.getElementById("qnaWidgetEmpty");

        loadingEl.classList.remove("d-none");
        contentEl.classList.add("d-none");
        emptyEl.classList.add("d-none");

        try {
          const res = await fetchJSON("/api/v1/qna/faq");
          const questions = res || [];
          
          if (questions.length === 0) {
            emptyEl.classList.remove("d-none");
          } else {
            contentEl.innerHTML = "";
            // Show only top 3
            questions.slice(0, 3).forEach(q => {
              const card = `
                <div class="col-md-4">
                  <div class="card h-100 border-0 shadow-sm" style="border-left: 3px solid #0d6efd !important;">
                    <div class="card-body p-3">
                      <h6 class="card-title text-primary text-truncate mb-2">${q.title}</h6>
                      <p class="card-text small text-muted text-truncate-2 mb-3" style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; height: 2.5rem;">
                        ${q.content}
                      </p>
                      <div class="d-flex justify-content-between align-items-center mt-auto">
                        <small class="text-muted" style="font-size: 0.7rem;">👁️ ${q.view_count} views</small>
                        <a href="/ui/qna?id=${q.id}" class="btn btn-sm btn-link p-0 text-decoration-none" style="font-size: 0.75rem;">Baca &rarr;</a>
                      </div>
                    </div>
                  </div>
                </div>
              `;
              contentEl.insertAdjacentHTML('beforeend', card);
            });
            contentEl.classList.remove("d-none");
          }
        } catch (err) {
          console.error("Failed to load Q&A widget:", err);
          emptyEl.classList.remove("d-none");
        } finally {
          loadingEl.classList.add("d-none");
        }
      }

      async function retryWidget(label, loader) {
        msg.textContent = `Memuat ulang ${label}...`;
        try {
          await loader();
          msg.textContent = `${label} berhasil diperbarui.`;
        } catch (err) {
          msg.textContent = `Gagal memuat ${label}: ${err.message}`;
        }
      }

      async function loadDashboard() {
        const auth = window.__AUTH__;
        if (auth && auth.getSessionState().status === "expired") {
          auth.verifySession();
          return;
        }

        msg.textContent = "Memuat data dashboard...";

        const [summaryRes, statsRes, yearlyRes, chartRes, rankingRes] =
          await Promise.allSettled([
            loadSummaryWidget(),
            loadStatsWidget(),
            loadYearlyWidget(),
            loadChartWidget(),
            loadRankingWidget(),
            loadQnaWidget(),
          ]);

        const failedWidgets = [];

        if (summaryRes.status !== "fulfilled") {
          failedWidgets.push("ringkasan KPI");
        }

        if (statsRes.status !== "fulfilled") {
          failedWidgets.push("statistik status");
        }

        if (yearlyRes.status !== "fulfilled") {
          failedWidgets.push("tabel tahunan");
        }

        if (chartRes.status !== "fulfilled") {
          failedWidgets.push("grafik target vs realisasi");
        }

        if (rankingRes.status !== "fulfilled") {
          failedWidgets.push("ranking program");
        }

        try {
          await loadTargetTable();
        } catch (_) {
          failedWidgets.push("tabel target/realisasi");
        }

        if (failedWidgets.length === 0) {
          msg.textContent = "Data dashboard diperbarui.";
        } else {
          msg.textContent = `Sebagian widget gagal dimuat: ${failedWidgets.join(", ")}.`;
        }
      }

      refreshBtn.addEventListener("click", loadDashboard);
      yearFilter.addEventListener("change", () => {
        targetPage = 1;
        writeYearToURL();
        updateYearlyActiveYearLabel();
        loadDashboard();
      });
      targetSearch.addEventListener("input", () => {
        targetPage = 1;
        loadTargetTable().catch(() => {});
      });
      targetLimit.addEventListener("change", () => {
        targetPage = 1;
        loadTargetTable().catch(() => {});
      });
      targetReset.addEventListener("click", () => {
        targetSearch.value = "";
        targetLimit.value = "10";
        targetPage = 1;
        loadTargetTable().catch(() => {});
      });
      targetPrev.addEventListener("click", () => {
        if (targetPage <= 1) return;
        targetPage -= 1;
        loadTargetTable().catch(() => {});
      });
      targetNext.addEventListener("click", () => {
        if (targetPage >= targetTotalPages) return;
        targetPage += 1;
        loadTargetTable().catch(() => {});
      });

      retrySummary.addEventListener("click", () =>
        retryWidget("ringkasan KPI", loadSummaryWidget),
      );
      retryStats.addEventListener("click", () =>
        retryWidget("statistik status", loadStatsWidget),
      );
      retryChart.addEventListener("click", () =>
        retryWidget("grafik target vs realisasi", loadChartWidget),
      );
      retryYearly.addEventListener("click", () =>
        retryWidget("ringkasan tahunan", loadYearlyWidget),
      );
      retryRanking.addEventListener("click", () =>
        retryWidget("ranking program", loadRankingWidget),
      );
      retryTargetTable.addEventListener("click", () =>
        retryWidget("tabel target/realisasi", loadTargetTable),
      );

      async function initDashboard() {
        const auth = window.__AUTH__;
        if (auth && auth.getSessionState().status === "expired") {
          await auth.verifySession();
          if (auth.getSessionState().status === "expired") {
            return;
          }
        }

        try {
          await loadCurrentUserProfile();
        } catch (err) {
          msg.textContent = `Profil user tidak dapat dimuat: ${err.message}`;
        }

        applyInitialYearFilter();
        writeYearToURL();
        updateYearlyActiveYearLabel();
        loadDashboard();
      }

      initDashboard();

      function setSessionStatusText() {
        const el = document.getElementById("sessionStatus");
        if (!el || !window.__AUTH__ || !window.__AUTH__.getSessionState) {
          return;
        }

        el.classList.remove(
          "session-active",
          "session-warning",
          "session-expired",
        );

        const state = window.__AUTH__.getSessionState();
        if (state.status === "active") {
          el.classList.add("session-active");
          el.textContent = `Status sesi: aktif (${state.secondsRemaining}s)`;
          return;
        }
        if (state.status === "refresh_soon") {
          el.classList.add("session-warning");
          el.textContent = `Status sesi: akan refresh (${state.secondsRemaining}s)`;
          return;
        }

        el.classList.add("session-expired");
        el.textContent = "Status sesi: expired";
      }

      window.addEventListener("load", () => {
        setSessionStatusText();
        window.setInterval(setSessionStatusText, 5000);
      });
      if (window.__AUTH__ && typeof window.__AUTH__.initInformasiSwitcher === "function") {
        window.__AUTH__.initInformasiSwitcher("/dashboard");
      } else {
        console.warn("window.__AUTH__ is undefined or initInformasiSwitcher is missing on Dashboard load");
      }
