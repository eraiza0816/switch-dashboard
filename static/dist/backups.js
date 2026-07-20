var s=null,a=null,i=document.createElement("style");i.innerText=`
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;document.head.appendChild(i);document.addEventListener("DOMContentLoaded",c);async function c(){let e=document.getElementById("table-wrapper");try{let t=await fetch("/api/backups");if(!t.ok)throw Error("HTTP error "+t.status);let n=await t.json();if(n.error){r("Error loading backups: "+n.error),l(n.error);return}if(!n||n.length===0){d();return}y(n)}catch(t){r("Connection error: "+t.message),l(t.message)}}function y(e){let t=document.getElementById("table-wrapper"),n=e.map((o)=>{return`
      <tr id="row-${o.filename.replace(/\./g,"_")}">
        <td class="switch-ip">${o.ip}</td>
        <td class="timestamp">${o.datetime}</td>
        <td class="filename">${o.filename}</td>
        <td class="file-size">${o.size_str}</td>
        <td style="white-space: nowrap;">
          <a class="btn-action btn-download" href="/api/backups/${o.filename}/download" title="Download backup file">
            Download
          </a>
          <button class="btn-action btn-delete" onclick="confirmDelete('${o.filename}', this)" title="Delete backup file">
            Delete
          </button>
        </td>
      </tr>
    `}).join("");t.innerHTML=`
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
          ${n}
        </tbody>
      </table>
    </div>
  `}function d(){let e=document.getElementById("table-wrapper");e.innerHTML=`
    <div class="empty-state">
      <div class="empty-icon" style="font-size:24px;color:#30363d;">-</div>
      <h3>No backups archived</h3>
      <p>Configure your switches in the settings and press the "Backup" action next to any switch IP in the main dashboard to store automated binary backups here.</p>
      <a href="/" style="margin-top: 8px;">&larr; Go to Dashboard</a>
    </div>
  `}function l(e){let t=document.getElementById("table-wrapper");t.innerHTML=`
    <div class="empty-state" style="color: #ff7b72;">
      <div class="empty-icon">-</div>
      <h3>System Error</h3>
      <p>${e}</p>
      <button onclick="loadBackups()" class="btn-action btn-download" style="margin-top: 14px; padding: 10px 20px;">
        Retry Connection
      </button>
    </div>
  `}function w(e,t){s=e,a=t.closest("tr"),document.getElementById("modal-filename-display").textContent=e,document.getElementById("delete-modal").classList.add("open")}function m(){document.getElementById("delete-modal").classList.remove("open"),s=null,a=null}async function f(){if(!s)return;let e=document.getElementById("modal-confirm-btn");e.disabled=!0,e.textContent="Deleting...";try{let t=await fetch(`/api/backups/${encodeURIComponent(s)}`,{method:"DELETE"}),n=await t.json();if(t.ok&&n.status==="ok"){if(g("Backup deleted successfully"),a)a.classList.add("fade-out"),setTimeout(()=>{if(a.remove(),document.querySelectorAll("tbody tr").length===0)d()},400)}else r("Delete failed: "+(n.error||"Unknown error"))}catch(t){r("Error connecting to server: "+t.message)}finally{e.disabled=!1,e.textContent="Delete Permanently",m()}}function g(e){u(e,"success")}function r(e){u(e,"error")}function u(e,t,n=""){let o=document.getElementById("toast-box"),p=document.getElementById("toast-icon"),h=document.getElementById("toast-message");o.className=`toast ${t}`,p.textContent=n,h.textContent=e,o.classList.add("show"),setTimeout(()=>{o.classList.remove("show")},4000)}window.loadBackups=c;window.confirmDelete=w;window.closeDeleteModal=m;window.executeDelete=f;
