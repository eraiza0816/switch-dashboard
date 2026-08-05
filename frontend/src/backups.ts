import './i18n';
import { showToast } from './dashboard-utils';

let fileToDelete = null;
let rowToDelete = null;

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
      showToast("Error loading backups: " + data.error, "error");
      renderErrorState(data.error);
      return;
    }
    
    if (!data || data.length === 0) {
      renderEmptyState();
      return;
    }
    
    renderTable(data);
  } catch (err) {
    showToast("Connection error: " + err.message, "error");
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
            Download
          </a>
          <button class="btn-action btn-delete" onclick="confirmDelete('${b.filename}', this)" title="Delete backup file">
            Delete
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
      <div class="empty-icon" style="font-size:24px;color:#30363d;">-</div>
      <h3>No backups archived</h3>
      <p>Configure your switches in the settings and press the "Backup" action next to any switch IP in the main dashboard to store automated binary backups here.</p>
      <a href="/" style="margin-top: 8px;">&larr; Go to Dashboard</a>
    </div>
  `;
}

function renderErrorState(msg) {
  const wrapper = document.getElementById("table-wrapper");
  wrapper.innerHTML = `
    <div class="empty-state" style="color: #ff7b72;">
      <div class="empty-icon">-</div>
      <h3>System Error</h3>
      <p>${msg}</p>
      <button onclick="loadBackups()" class="btn-action btn-download" style="margin-top: 14px; padding: 10px 20px;">
        Retry Connection
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
  
  const confirmBtn = document.getElementById("modal-confirm-btn") as HTMLButtonElement;
  confirmBtn.disabled = true;
  confirmBtn.textContent = "Deleting...";
  
  try {
    const res = await fetch(`/api/backups/${encodeURIComponent(fileToDelete)}`, {
      method: "DELETE"
    });
    const data = await res.json();
    
    if (res.ok && data.status === "ok") {
      showToast("Backup deleted successfully", "success");
      
      if (rowToDelete) {
        rowToDelete.classList.add("fade-out");
        setTimeout(() => {
          rowToDelete.remove();
          const rows = document.querySelectorAll("tbody tr");
          if (rows.length === 0) {
            renderEmptyState();
          }
        }, 400);
      }
    } else {
      showToast("Delete failed: " + (data.error || "Unknown error"), "error");
    }
  } catch (err) {
    showToast("Error connecting to server: " + err.message, "error");
  } finally {
    confirmBtn.disabled = false;
    confirmBtn.textContent = "Delete Permanently";
    closeDeleteModal();
  }
}


(window as any).loadBackups = loadBackups;
(window as any).confirmDelete = confirmDelete;
(window as any).closeDeleteModal = closeDeleteModal;
(window as any).executeDelete = executeDelete;
export {};
