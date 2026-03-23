(() => {
  let userRole = "";
  async function fetchUserProfile() {
    try {
      const res = await fetch("/api/v1/auth/me", {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      if (!res.ok) throw new Error("Gagal mengambil profil user");
      const data = await res.json();
      userRole = data?.data?.role || "";
    } catch (err) {
      userRole = "";
    }
  }

  const form = document.getElementById("dokumen-form");
  const tahunInput = document.getElementById("dokumen-tahun");
  const namaInput = document.getElementById("dokumen-nama");
  const fileInput = document.getElementById("dokumen-file");
  const statusText = document.getElementById("dokumen-status");
  const tableBody = document.getElementById("dokumen-table-body");
  const metaText = document.getElementById("dokumen-meta-text");
  
  const accessToken =
    localStorage.getItem("AUTH_TOKEN") ||
    localStorage.getItem("authToken") ||
    "";
    
  const dokumenEndpoint =
    document.body.getAttribute("data-api-endpoint") ||
    "/api/v1/dokumen_pdf";

  async function fetchDokumenPDFs() {
    try {
      const response = await fetch(dokumenEndpoint, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      if (!response.ok) throw new Error("Gagal mengambil data");
      const data = await response.json();
      return data;
    } catch (error) {
      console.error(error);
      alert("Error: " + error.message);
      return [];
    }
  }

  function renderTable(dokumenPDFs) {
    const data = Array.isArray(dokumenPDFs)
      ? dokumenPDFs
      : dokumenPDFs.items || [];
    tableBody.innerHTML = "";
    data.forEach((dokumen) => {
      const row = document.createElement("tr");
      let aksiCell = "";
      // Hanya ADMIN, OPERATOR, PERENCANA yang bisa hapus
      if (["ADMIN", "OPERATOR", "PERENCANA"].includes(userRole)) {
        aksiCell = `<button class="btn btn-sm btn-danger" data-delete="${dokumen.id}">Hapus</button>`;
      } else {
        aksiCell = `<span class="text-muted">Tidak diizinkan</span>`;
      }
      row.innerHTML = `
        <td>${dokumen.id}</td>
        <td>${dokumen.tahun}</td>
        <td>${dokumen.nama}</td>
        <td><a href="${dokumen.file_url || dokumen.file_path}" target="_blank" class="text-primary fw-semibold">Lihat PDF</a></td>
        <td>${aksiCell}</td>
      `;
      tableBody.appendChild(row);
    });
    metaText.textContent = `Total ${data.length} dokumen`;
  }

  // Global Event Delegation (mendukung SSR load maupun AJAX re-render)
  tableBody.addEventListener("click", async function (e) {
    const btn = e.target.closest("button[data-delete]");
    if (!btn) return;
    
    const id = btn.getAttribute("data-delete");
    const row = btn.closest("tr");
    const nama = row ? row.children[2].textContent.trim() : "";
    const filePDF = row && row.children[3].querySelector("a") ? "Lihat PDF" : "Tidak ada file";
    
    if (!id) {
      alert("ID dokumen tidak ditemukan.");
      return;
    }
    
    if (!confirm(`Yakin hapus dokumen ini?\nNama: ${nama}\nFile: ${filePDF}`)) return;
    
    btn.disabled = true;
    statusText.textContent = "Menghapus...";
    try {
      const response = await fetch(`${dokumenEndpoint}/${id}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      if (!response.ok) {
        throw new Error("Gagal menghapus");
      }
      await loadDokumenPDFs();
      statusText.textContent = `Dokumen '${nama}' berhasil dihapus.`;
    } catch (error) {
      alert("Error: " + error.message);
      statusText.textContent = "Gagal menghapus dokumen";
    } finally {
      if (document.body.contains(btn)) {
        btn.disabled = false;
      }
    }
  });

  async function loadDokumenPDFs() {
    statusText.textContent = "Memuat...";
    await fetchUserProfile();
    const dokumenPDFs = await fetchDokumenPDFs();
    renderTable(dokumenPDFs);
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
    
    // Batas maksimal file 5MB
    if (file && file.size > 5 * 1024 * 1024) {
      alert("Ukuran file PDF maksimal 5MB. Silakan pilih file yang lebih kecil.");
      statusText.textContent = "Ukuran file terlalu besar";
      return;
    }
    
    const formData = new FormData();
    formData.append("tahun", tahun);
    formData.append("nama", nama);
    if (file) formData.append("file", file);
    
    try {
      statusText.textContent = id ? "Mengupdate..." : "Mengunggah...";
      const url = id ? `${dokumenEndpoint}/${id}` : dokumenEndpoint;
      const method = id ? "PUT" : "POST";
      const response = await fetch(url, {
        method,
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
        body: formData,
      });
      
      if (!response.ok) throw new Error(id ? "Gagal mengupdate" : "Gagal mengunggah");
      
      await loadDokumenPDFs();
      form.reset();
      document.getElementById("dokumen-id").value = "";
      statusText.textContent = "Selesai";
    } catch (error) {
      alert("Error: " + error.message);
      statusText.textContent = "Gagal";
    }
  });

  // Pre-fetch user profile on load so the SSR table at least runs consistently
  fetchUserProfile();
})();
