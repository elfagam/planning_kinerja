(() => {
  const dokumenEndpoint = document.body.getAttribute("data-api-endpoint") || "/api/v1/dokumen_pdf";
  const informasiLatestEndpoint = "/api/v1/performance/informasi/latest";
  const activeInfoRoute = "/dokumen_pdf";

  const form = document.getElementById("dokumen-form");
  const tahunInput = document.getElementById("dokumen-tahun");
  const namaInput = document.getElementById("dokumen-nama");
  const fileInput = document.getElementById("dokumen-file");
  const statusText = document.getElementById("dokumen-status");
  const tableBody = document.getElementById("dokumen-table-body");
  const metaText = document.getElementById("dokumen-meta-text");
  const infoSwitcherText = document.getElementById("info-switcher-text");
  const submitBtn = form.querySelector('button[type="submit"]');
  const resetBtn = document.getElementById("dokumen-reset");
  const fileLabel = document.querySelector('label[for="dokumen-file"]');

  let userRole = "";
  let switcherTimerID = null;

  function getAccessToken() {
    return window.__AUTH__ ? window.__AUTH__.getAccessToken() : (localStorage.getItem("AUTH_TOKEN") || localStorage.getItem("authToken") || "");
  }

  function authHeader() {
    const token = getAccessToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  async function fetchUserProfile() {
    try {
      const res = await fetch("/api/v1/auth/me", {
        headers: authHeader(),
      });
      if (!res.ok) throw new Error("Gagal mengambil profil user");
      const data = await res.json();
      userRole = data?.data?.role || "";
    } catch (err) {
      console.error("Error fetchUserProfile:", err);
      userRole = "";
    }
  }

  async function fetchDokumenPDFs() {
    try {
      const response = await fetch(dokumenEndpoint, {
        headers: authHeader(),
      });
      if (!response.ok) throw new Error("Gagal mengambil data");
      const data = await response.json();
      return data;
    } catch (error) {
      console.error("Error fetchDokumenPDFs:", error);
      alert("Error: " + error.message);
      return [];
    }
  }

  function renderTable(dokumenPDFs) {
    const data = Array.isArray(dokumenPDFs)
      ? dokumenPDFs
      : dokumenPDFs.items || [];
    
    tableBody.innerHTML = "";
    
    if (data.length === 0) {
      tableBody.innerHTML = `<tr><td colspan="5" class="text-center text-muted">Tidak ada dokumen</td></tr>`;
      metaText.textContent = "Total 0 dokumen";
      return;
    }

    data.forEach((dokumen) => {
      const row = document.createElement("tr");
      let aksiCell = "";
      // Hanya ADMIN, OPERATOR, PERENCANA yang bisa hapus/edit
      if (["ADMIN", "OPERATOR", "PERENCANA"].includes(userRole)) {
        aksiCell = `
          <div class="btn-group">
            <button class="btn btn-sm btn-outline-primary" data-edit="${dokumen.id}">Edit</button>
            <button class="btn btn-sm btn-outline-danger" data-delete="${dokumen.id}">Hapus</button>
          </div>
        `;
      } else {
        aksiCell = `<span class="text-muted">Tidak diizinkan</span>`;
      }
      
      const filePathLink = dokumen.file_path || dokumen.FilePath || "";
      const namaDoc = dokumen.nama || dokumen.Nama || "Tidak ada nama";
      const tahunDoc = dokumen.tahun || dokumen.Tahun || "-";

      row.innerHTML = `
        <td>${dokumen.id}</td>
        <td>${tahunDoc}</td>
        <td><span class="fw-semibold text-dark">${namaDoc}</span></td>
        <td>
          ${filePathLink ? `<a href="${filePathLink}" target="_blank" class="text-primary fw-semibold">Lihat PDF</a>` : `<span class="text-muted">Tidak ada file</span>`}
        </td>
        <td>${aksiCell}</td>
      `;
      tableBody.appendChild(row);
    });
    metaText.textContent = `Total ${data.length} dokumen`;
  }

  tableBody.addEventListener("click", async function (e) {
    // Delete Handle
    const deleteBtn = e.target.closest("button[data-delete]");
    if (deleteBtn) {
      const id = deleteBtn.getAttribute("data-delete");
      if (!id) return;
      if (!confirm("Yakin hapus dokumen ini?")) return;

      deleteBtn.disabled = true;
      statusText.textContent = "Menghapus...";
      try {
        const response = await fetch(`${dokumenEndpoint}/${id}`, {
          method: "DELETE",
          headers: authHeader(),
        });
        if (!response.ok) throw new Error("Gagal menghapus");
        
        statusText.textContent = "Berhasil dihapus";
        await loadDokumenPDFs();
      } catch (error) {
        alert("Error: " + error.message);
        statusText.textContent = "Gagal menghapus";
      } finally {
        if (document.body.contains(deleteBtn)) {
          deleteBtn.disabled = false;
        }
      }
      return;
    }

    // Edit Handle
    const editBtn = e.target.closest("button[data-edit]");
    if (editBtn) {
      const id = editBtn.getAttribute("data-edit");
      if (!id) return;
      
      const row = editBtn.closest("tr");
      const tahunVal = row.cells[1].innerText;
      const namaVal = row.cells[2].innerText;

      document.getElementById("dokumen-id").value = id;
      tahunInput.value = tahunVal;
      namaInput.value = namaVal;
      fileInput.required = false;
      fileLabel.innerHTML = 'File PDF <span class="text-muted">(Kosongkan jika tidak ingin ubah file)</span>';
      submitBtn.textContent = "Update Dokumen";
      submitBtn.className = "btn btn-info btn-sm";
      statusText.textContent = "Mode Edit: " + namaVal;
      window.scrollTo({ top: 0, behavior: 'smooth' });
    }
  });

  resetBtn.addEventListener("click", () => {
    form.reset();
    document.getElementById("dokumen-id").value = "";
    fileInput.required = true;
    fileLabel.textContent = "File PDF";
    submitBtn.textContent = "Upload";
    submitBtn.className = "btn btn-primary btn-sm";
    statusText.textContent = "Siap";
  });

  async function loadDokumenPDFs() {
    statusText.textContent = "Memuat...";
    const data = await fetchDokumenPDFs();
    renderTable(data);
    statusText.textContent = "Selesai";
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const tahun = tahunInput.value;
    const nama = namaInput.value;
    const file = fileInput.files[0];
    const id = document.getElementById("dokumen-id").value;

    if (!tahun || !nama || (!file && !id)) {
      alert("Tahun, nama, dan file PDF harus diisi!");
      return;
    }

    if (file && file.size > 5 * 1024 * 1024) {
      alert("Ukuran file PDF maksimal 5MB.");
      return;
    }

    const formData = new FormData();
    formData.append("tahun", tahun);
    formData.append("nama", nama);
    if (file) formData.append("file", file);

    try {
      statusText.textContent = "Mengunggah...";
      const method = id ? "PUT" : "POST";
      const url = id ? `${dokumenEndpoint}/${id}` : dokumenEndpoint;

      const response = await fetch(url, {
        method,
        headers: authHeader(),
        body: formData,
      });

      if (!response.ok) throw new Error("Gagal mengunggah/update");

      resetBtn.click();
      await loadDokumenPDFs();
      statusText.textContent = id ? "Berhasil diperbarui" : "Berhasil diunggah";
    } catch (error) {
      alert("Error: " + error.message);
      statusText.textContent = "Gagal";
    }
  });

  (async () => {
    const token = getAccessToken();
    if (!token) {
      if (window.__AUTH__) {
        window.__AUTH__.verifySession();
      } else {
        window.location.href = "/ui/login";
      }
      return;
    }

    await fetchUserProfile();
    await loadDokumenPDFs();
    window.__AUTH__.initInformasiSwitcher("/dokumen_pdf");
  })();
})();
