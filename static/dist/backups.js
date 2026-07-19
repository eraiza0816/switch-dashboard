var H=document.createElement("style");H.innerText=`
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
`;document.head.appendChild(H);document.addEventListener("DOMContentLoaded",O);async function O(){let x=document.getElementById("table-wrapper");try{let q=await fetch("/api/backups");if(!q.ok)throw Error("HTTP error "+q.status);let z=await q.json();if(z.error){G("Error loading backups: "+z.error),F(z.error);return}if(!z||z.length===0){Q();return}P(z)}catch(q){G("Connection error: "+q.message),F(q.message)}}function P(x){let q=document.getElementById("table-wrapper"),z=x.map((v)=>{return`
      <tr id="row-${v.filename.replace(/\./g,"_")}">
        <td class="switch-ip">${v.ip}</td>
        <td class="timestamp">${v.datetime}</td>
        <td class="filename">${v.filename}</td>
        <td class="file-size">${v.size_str}</td>
        <td style="white-space: nowrap;">
          <a class="btn-action btn-download" href="/api/backups/${v.filename}/download" title="Download backup file">
            Download
          </a>
          <button class="btn-action btn-delete" onclick="confirmDelete('${v.filename}', this)" title="Delete backup file">
            Delete
          </button>
        </td>
      </tr>
    `}).join("");q.innerHTML=`
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
          ${z}
        </tbody>
      </table>
    </div>
  `}function Q(){let x=document.getElementById("table-wrapper");x.innerHTML=`
    <div class="empty-state">
      <div class="empty-icon" style="font-size:24px;color:#30363d;">-</div>
      <h3>No backups archived</h3>
      <p>Configure your switches in the settings and press the "Backup" action next to any switch IP in the main dashboard to store automated binary backups here.</p>
      <a href="/" style="margin-top: 8px;">&larr; Go to Dashboard</a>
    </div>
  `}function F(x){let q=document.getElementById("table-wrapper");q.innerHTML=`
    <div class="empty-state" style="color: #ff7b72;">
      <div class="empty-icon">⚠️</div>
      <h3>System Error</h3>
      <p>${x}</p>
      <button onclick="loadBackups()" class="btn-action btn-download" style="margin-top: 14px; padding: 10px 20px;">
        Retry Connection
      </button>
    </div>
  `}function G(x){R(x,"error")}function R(x,q,z){let v=document.getElementById("toast-box"),J=document.getElementById("toast-icon"),L=document.getElementById("toast-message");v.className=`toast ${q}`,J.textContent=z,L.textContent=x,v.classList.add("show"),setTimeout(()=>{v.classList.remove("show")},4000)}
