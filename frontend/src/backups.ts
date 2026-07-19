// @ts-nocheck
let fileToDelete = null;
let rowToDelete = null;

// CSS Keyframes injection for spin animation
const styleSheet = document.createElement("style");
styleSheet.innerText = `
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;
document.head.appendChild(styleSheet);

document.addEventListener("DOMContentLoaded", loadBackups);

async function loadBackups() {
  const wrapper = document.getElementById("table-wrapper");
  try {
    const res = await fetch("/api/backups");
    if (!res.ok) throw new Error("HTTP error " + res.status);
    const data = await res.json();
    
    if (data.error) {
      showErrorToast("Error loading backups: " + data.error);
      renderErrorState(data.error);
      return;
    }
    
    if (!data || data.length === 0) {
      renderEmptyState();
      return;
    }
    
    renderTable(data);
  } catch (err) {
    showErrorToast("Connection error: " + err.message);
    renderErrorState(err.message);
  }
}

function renderTable(backups) {
  const wrapper = document.getElementById("table-wrapper");
  
  let rows = backups.map(b => {
    return `
      <tr id="row-${b.filename.replace(/\./g, '_')}">
        <td class="switch-ip">${b.ip}</td>
        <td class="timestamp">${b.datetime}</td>
        <td class="filename">${b.filename}</td>
        <td class="file-size">${b.size_str}</td>
        <td style="white-space: nowrap;">
          <a class="btn-action btn-download" href="/api/backups/${b.filename}/download" title="Download backup file">
            <span style="font-size: 13px;">⬇️</span> Download
          </a>
          <button class="btn-action btn-delete" onclick="confirmDelete('${b.filename}', this)" title="Delete backup file">
            <span style="font-size: 13px;">🗑️</span> Delete
          </button>
        </td>
      </tr>
    `;
  }).join('');
  
  wrapper.innerHTML = `
    <div class="table-container">
      <table>
        <thead>
          <tr>
            <th style="width: 15%">Switch IP</th>
            <th style="width: 20%">Date & Time</th>
            <th style="width: 40%">Filename</th>
            <th style="width: 10%">Size</th>
            <th style="width: 15%">Actions</th>
          </tr>
        </thead>
        <tbody>
          ${rows}
        </tbody>
      </table>
    </div>
  `;
}

function renderEmptyState() {
  const wrapper = document.getElementById("table-wrapper");
  wrapper.innerHTML = `
    <div class="empty-state">
      <div class="empty-icon">📂</div>
      <h3>No backups archived</h3>
      <p>Configure your switches in the settings and press the "💾 Backup" action next to any switch IP in the main dashboard to store automated binary backups here.</p>
      <a href="/" style="margin-top: 8px;">&larr; Go to Dashboard</a>
    </div>
  `;
}

function renderErrorState(msg) {
  const wrapper = document.getElementById("table-wrapper");
  wrapper.innerHTML = `
    <div class="empty-state" style="color: #ff7b72;">
      <div class="empty-icon">⚠️</div>
      <h3>System Error</h3>
      <p>${msg}</p>
      <button onclick="loadBackups()" class="btn-action btn-download" style="margin-top: 14px; padding: 10px 20px;">
        🔄 Retry Connection
      </button>
    </div>
  `;
}

function confirmDelete(filename, btnElement) {
  fileToDelete = filename;
  rowToDelete = btnElement.closest("tr");
  
  document.getElementById("modal-filename-display").textContent = filename;
  document.getElementById("delete-modal").classList.add("open");
}

function closeDeleteModal() {
  document.getElementById("delete-modal").classList.remove("open");
  fileToDelete = null;
  rowToDelete = null;
}

async function executeDelete() {
  if (!fileToDelete) return;
  
  const confirmBtn = document.getElementById("modal-confirm-btn");
  confirmBtn.disabled = true;
  confirmBtn.textContent = "Deleting...";
  
  try {
    const res = await fetch(`/api/backups/${encodeURIComponent(fileToDelete)}`, {
      method: "DELETE"
    });
    const data = await res.json();
    
    if (res.ok && data.status === "ok") {
      showSuccessToast("Backup deleted successfully");
      
      // Animate row deletion
      if (rowToDelete) {
        rowToDelete.classList.add("fade-out");
        setTimeout(() => {
          rowToDelete.remove();
          // Check if table is now empty
          const rows = document.querySelectorAll("tbody tr");
          if (rows.length === 0) {
            renderEmptyState();
          }
        }, 400);
      }
    } else {
      showErrorToast("Delete failed: " + (data.error || "Unknown error"));
    }
  } catch (err) {
    showErrorToast("Error connecting to server: " + err.message);
  } finally {
    confirmBtn.disabled = false;
    confirmBtn.textContent = "Delete Permanently";
    closeDeleteModal();
  }
}

// Toast Notifications helper
function showSuccessToast(message) {
  showToast(message, "success", "✓");
}

function showErrorToast(message) {
  showToast(message, "error", "⚠️");
}

function showToast(message, type, icon) {
  const toast = document.getElementById("toast-box");
  const toastIcon = document.getElementById("toast-icon");
  const toastMsg = document.getElementById("toast-message");
  
  toast.className = `toast ${type}`;
  toastIcon.textContent = icon;
  toastMsg.textContent = message;
  
  toast.classList.add("show");
  setTimeout(() => {
    toast.classList.remove("show");
  }, 4000);
}
