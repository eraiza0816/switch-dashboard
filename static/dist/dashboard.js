var t=window.__DATA__&&window.__DATA__.refresh||30,Wq=window.__DATA__&&window.__DATA__.columns||["port","status","speed","packets","bytes","info","notes"],Gq=window.__DATA__&&window.__DATA__.portWrap||0,w=null,m=null,e="live",hq="",xq="",O="Bps",l=!1,p=t,Fq=null;function jq(){p=t;let q=document.getElementById("live-text");if(q)if(l)q.textContent="Paused";else q.textContent=`Live (update in ${p}s)`}function uq(){if(Fq)clearInterval(Fq);Fq=setInterval(()=>{let q=document.getElementById("live-text"),J=document.getElementById("live-dot");if(!q||!J)return;if(l){q.textContent="Paused",J.style.background="#d29922";return}if(p--,p<=0)q.textContent="Updating...",J.style.background="#d29922",p=t,d();else q.textContent=`Live (update in ${p}s)`,J.style.background="#3fb950"},1000)}var vq=["port","status","speed","packets","bytes","info","notes"],b=[],Hq={},Lq=[];function wq(){let q=Lq&&Lq.length>0?Lq:null;if(!q)q=vq;b=q.filter((J)=>Wq.includes(J)),Wq.forEach((J)=>{if(!b.includes(J))b.push(J)})}wq();var qq=Hq||{};async function Mq(){try{await fetch("/api/config/settings",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({column_widths:qq,column_order:b})})}catch(q){console.error("Failed to save column settings to server:",q)}}function mq(q,J){qq[q]=J,Mq()}function cq(){qq=Hq||{}}cq();var I=null,Rq=0,Cq=0,f=null;function lq(q,J){q.stopPropagation(),q.preventDefault(),I=J,f=q.target.closest("th"),Rq=q.clientX,Cq=f.offsetWidth,document.addEventListener("mousemove",Pq),document.addEventListener("mouseup",Dq),f.classList.add("resizing"),document.querySelectorAll(".port-table th, .mac-table th").forEach((K)=>{K.setAttribute("draggable","false")})}function Pq(q){if(!I||!f)return;let J=q.clientX-Rq,K=Math.max(30,Cq+J);if(I.startsWith("mac-"))document.querySelectorAll(`.mac-table th[data-col-id="${I}"]`).forEach((Z)=>{Z.style.width=K+"px"});else document.querySelectorAll(`.port-table th[data-col-id="${I}"]`).forEach((Z)=>{Z.style.width=K+"px"})}function Dq(q){if(I&&f){f.classList.remove("resizing");let J=f.offsetWidth;mq(I,J)}I=null,f=null,document.removeEventListener("mousemove",Pq),document.removeEventListener("mouseup",Dq),document.querySelectorAll(".port-table th, .mac-table th").forEach((J)=>{J.setAttribute("draggable","true")})}var Uq={port:{label:"Port",thStyle:"width:55px",tdClass:"port-num",tdTitleFn:(q)=>{let J=[];if(q.vm_name)J.push(q.vm_name);if(q.interface)J.push(q.interface);return J.length>0?J.join(" / "):""},render:(q,J)=>q.port},status:{label:"Status",thStyle:"width:145px",render:(q,J)=>{let K=(q.status||"unknown").toLowerCase(),Q="";if(J.dhcp_snooping&&J.dhcp_snooping.enabled){let X=J.dhcp_snooping.ports||{},$=q.port.split("/")[0],Y=X[q.port]||X[$];if(Y==="Trusted")Q='<span class="trust-badge trusted" title="DHCP Snooping: Trusted Port">Trusted</span>';else if(Y==="Untrusted")Q='<span class="trust-badge untrusted" title="DHCP Snooping: Untrusted Port">Untrusted</span>'}let Z=q.duplex||"Auto";if(Z.toLowerCase()==="full")Z="Full Duplex";else if(Z.toLowerCase()==="half")Z="Half Duplex";return`
        <div style="display: flex; flex-direction: column; gap: 4px; align-items: flex-start;">
          <div style="display: flex; align-items: center; gap: 6px;">
            <span class="status-badge ${K}" onclick="openGraph('${J.ip}','${q.port}','${J.name}','Port ${q.port}', ${q.cum_tx||q.tx_bytes||0}, ${q.cum_rx||q.rx_bytes||0})"><span class="status-dot ${K}"></span>${K.toUpperCase()}</span>
            ${Q}
          </div>
          <div style="font-size: 10px; color: #8b949e; padding-left: 4px; font-weight: 500;">${Z}</div>
        </div>
      `}},speed:{label:"Speed",thStyle:"width:55px",tdClassFn:(q)=>kJ(q.speed,q.status),tdTitleFn:(q)=>q.duplex||"",render:(q,J)=>{let K=q.port==="9"||(q.speed||"").toLowerCase().includes("10g")||q.port==="SFP"||q.port==="9/SFP",Q=q.speed||"Auto";if(K)return`<span class="sfp-speed-link" onclick="openTransceiver('${J.ip}','${q.port}','${J.name}')" title="Click to view SFP+ Transceiver Diagnostics">${Q} <svg viewBox="0 0 24 24" style="width: 10px; height: 10px; fill: currentColor; display: inline-block;"><path d="M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6zm0-7C7 2 2.73 5.11 1 9.5 2.73 13.89 7 17 12 17s9.27-3.11 11-7.5C21.27 5.11 17 2 12 2zm0 13a5.5 5.5 0 1 1 0-11 5.5 5.5 0 0 1 0 11z"/></svg></span>`;else return Q}},packets:{label:'Traffic<br><span style="font-weight:400;color:#8b949e">(packets)</span>',thStyle:"width:90px",tdClass:"mono",tdTitle:"TX / RX packets",render:(q,J)=>{let K=q.tx_packets||0,Q=q.rx_packets||0;return`TX ${Oq(K)}<br>RX ${Oq(Q)}`}},bytes:{label:'Traffic<br><span style="font-weight:400;color:#8b949e">(bytes)</span>',thStyle:"width:90px",tdClass:"mono",tdTitle:"TX / RX bytes formatted",render:(q,J)=>{let K=q.cum_tx||q.tx_bytes||0,Q=q.cum_rx||q.rx_bytes||0;return`TX ${o(K)}<br>RX ${o(Q)}`}},info:{getLabel:()=>`<span id="speed-unit-toggle" class="unit-toggle" onclick="toggleSpeedUnit(event)" title="Toggle speed unit">${O==="bps"?"BPS":"B/s"}</span>`,tdClass:"col-info mono",render:(q,J)=>{let K=q.speed_tx_bps||0,Q=q.speed_rx_bps||0;return`TX ${D(K,O)}<br>RX ${D(Q,O)}`}},notes:{label:"Notes",tdClass:"col-note",render:(q,J)=>{return`<input class="note-input" type="text" placeholder="-" value="${q.note||""}" data-key="${J.ip}:${q.port}" onfocus="pausePolling()" onblur="saveNote(this)">`}}},n=null;function dq(q,J){n=J,q.dataTransfer.effectAllowed="move",q.dataTransfer.setData("text/plain",J)}function iq(q){q.preventDefault(),q.dataTransfer.dropEffect="move"}function pq(q,J){q.preventDefault();let K=q.target.closest("th");if(K&&n!==J)K.classList.add("drag-over")}function nq(q){let J=q.target.closest("th");if(J)J.classList.remove("drag-over")}function oq(q){document.querySelectorAll(".port-table th").forEach((J)=>{J.classList.remove("drag-over")})}function sq(q,J){q.preventDefault();let K=q.target.closest("th");if(K)K.classList.remove("drag-over");if(n&&n!==J){let Q=b.indexOf(n),Z=b.indexOf(J);if(Q!==-1&&Z!==-1)b.splice(Q,1),b.splice(Z,0,n),Mq(),d()}}var rq=null,aq=null,eq="";var c={};try{let q=localStorage.getItem("macTableStates");if(q)c=JSON.parse(q)}catch(q){console.error("Failed to load macTableStates from localStorage",q)}function s(){let q={};for(let J in c)if(c.hasOwnProperty(J)){let K=c[J];q[J]={expanded:K.expanded,filters:K.filters||{mac:"",vendor:"",type:"",port:"",vlan:""},sortBy:K.sortBy||"port",sortAsc:K.sortAsc!==void 0?K.sortAsc:!0,scrollTop:K.scrollTop||0,height:K.height||"450px"}}localStorage.setItem("macTableStates",JSON.stringify(q))}function R(q){if(!c[q])c[q]={};let J=c[q];if(J.expanded===void 0)J.expanded=!1;if(!J.filters)J.filters={mac:"",host:"",vendor:"",type:"",port:"",vlan:""};if(J.sortBy===void 0)J.sortBy="port";if(J.sortAsc===void 0)J.sortAsc=!0;if(!J.data)J.data=[];if(J.scrollTop===void 0)J.scrollTop=0;if(J.height===void 0)J.height="450px";return J}function yq(q){if(!q)return 0;if(!isNaN(q))return parseInt(q);let J=q.match(/\d+/);if(J)return parseInt(J[0]);return q.toString().toLowerCase()}function zq(q){if(!q)return 0;let J=parseInt(q);return isNaN(J)?q.toString().toLowerCase():J}function tq(){bq()}function qJ(){l=!1,document.getElementById("live-dot").style.background="#3fb950",jq()}var Bq=null;function JJ(q,J){let K=R(q);K.scrollTop=J.scrollTop,clearTimeout(Bq),Bq=setTimeout(()=>{s()},500)}function KJ(q,J){if(J&&J.style.height){let K=R(q);K.height=J.style.height,s()}}function QJ(q){let J=R(q);J.expanded=!J.expanded,s();let K=document.querySelector(`.mac-content-${CSS.escape(q)}`),Q=document.querySelector(`.mac-toggle-arrow-${CSS.escape(q)}`);if(K&&Q)if(J.expanded)K.style.display="block",Q.style.transform="rotate(90deg)",Jq(q);else K.style.display="none",Q.style.transform="rotate(0deg)"}function ZJ(q){let J={};try{let K=localStorage.getItem("igmp_states");if(K)J=JSON.parse(K)}catch(K){console.error("Failed to load igmp_states from localStorage",K)}if(!J[q])J[q]={expanded:!1};return J[q]}function $J(q){let J={};try{let Z=localStorage.getItem("igmp_states");if(Z)J=JSON.parse(Z)}catch(Z){console.error("Failed to load igmp_states from localStorage",Z)}if(!J[q])J[q]={expanded:!1};J[q].expanded=!J[q].expanded,localStorage.setItem("igmp_states",JSON.stringify(J));let K=document.querySelector(`.igmp-content-${CSS.escape(q)}`),Q=document.querySelector(`.igmp-toggle-arrow-${CSS.escape(q)}`);if(K&&Q)if(J[q].expanded)K.style.display="block",Q.style.transform="rotate(90deg)";else K.style.display="none",Q.style.transform="rotate(0deg)"}function jJ(q){let J=R(q),K=document.querySelector(`.mac-filter-mac-${CSS.escape(q)}`),Q=document.querySelector(`.mac-filter-host-${CSS.escape(q)}`),Z=document.querySelector(`.mac-filter-vendor-${CSS.escape(q)}`),X=document.querySelector(`.mac-filter-type-${CSS.escape(q)}`),$=document.querySelector(`.mac-filter-port-${CSS.escape(q)}`),Y=document.querySelector(`.mac-filter-vlan-${CSS.escape(q)}`);if(K)J.filters.mac=K.value.trim().toLowerCase();if(Q)J.filters.host=Q.value.trim().toLowerCase();if(Z)J.filters.vendor=Z.value.trim().toLowerCase();if(X)J.filters.type=X.value.trim().toLowerCase();if($)J.filters.port=$.value.trim().toLowerCase();if(Y)J.filters.vlan=Y.value.trim().toLowerCase();s(),Jq(q)}function YJ(q,J){let K=R(q);if(K.sortBy===J)K.sortAsc=!K.sortAsc;else K.sortBy=J,K.sortAsc=!0;s(),Jq(q)}function XJ(q){let J=R(q),K=document.querySelector(`.mac-spinner-${CSS.escape(q)}`),Q=document.querySelector(`.mac-refresh-btn-${CSS.escape(q)} span:not(.mac-spinner-${CSS.escape(q)})`);if(K)K.style.display="inline-block";if(Q)Q.textContent="Refreshing...";fetch(`/api/switches/${q}/refresh_mac`,{method:"POST"}).then((Z)=>Z.json()).then((Z)=>{if(Z.status==="ok"){J.data=Z.mac_table||[];let X=document.querySelector(`.mac-time-${CSS.escape(q)}`);if(X)X.textContent="Last Scraped: "+kq(Date.now()/1000);let $=document.querySelector(`.mac-count-${CSS.escape(q)}`);if($)$.textContent=`${J.data.length} entries`;Jq(q)}else alert("Failed to refresh MAC table: "+(Z.error||"Unknown error"))}).catch((Z)=>{alert("Error refreshing MAC table: "+Z.message)}).finally(()=>{if(K)K.style.display="none";if(Q)Q.textContent="Refresh MACs"})}function _J(q,J){if(J.dataset.loading==="true")return;J.dataset.loading="true";let K=J.innerHTML;J.innerHTML="⏳ Backing up...",J.style.color="#8b949e",fetch(`/api/switches/${q}/backup`,{method:"POST"}).then((Q)=>Q.json()).then((Q)=>{if(Q.status==="ok")J.innerHTML="✓ Backed up!",J.style.color="#3fb950";else alert("Failed to backup config: "+(Q.error||"Unknown error")),J.innerHTML="❌ Failed",J.style.color="#ff7b72"}).catch((Q)=>{alert("Error during backup: "+Q.message),J.innerHTML="❌ Failed",J.style.color="#ff7b72"}).finally(()=>{setTimeout(()=>{J.innerHTML=K,J.style.color="#58a6ff",delete J.dataset.loading},3000)})}function Jq(q){let J=R(q),K=document.querySelector(`.mac-tbody-${CSS.escape(q)}`);if(!K)return;["mac","host","vendor","type","port","vlan"].forEach(($)=>{let Y=document.querySelector(`.sort-indicator-${CSS.escape(q)}-${$}`);if(Y)if(J.sortBy===$)Y.textContent=J.sortAsc?" ▴":" ▾",Y.style.color="#58a6ff";else Y.textContent=""});let Q=J.data,Z=J.filters||{mac:"",host:"",vendor:"",type:"",port:"",vlan:""};if(Z.mac||Z.host||Z.vendor||Z.type||Z.port||Z.vlan)Q=Q.filter(($)=>{let Y=!Z.mac||($.mac||"").toLowerCase().includes(Z.mac),F=!Z.host||($.host||"").toLowerCase().includes(Z.host),U=!Z.vendor||($.vendor||"").toLowerCase().includes(Z.vendor),G=!Z.type||($.type||"").toLowerCase().includes(Z.type),z=!Z.port||($.port||"").toString().toLowerCase().includes(Z.port),B=!Z.vlan||($.vlan||"").toString().toLowerCase().includes(Z.vlan);return Y&&F&&U&&G&&z&&B});if(Q.sort(($,Y)=>{let F=$[J.sortBy]||"",U=Y[J.sortBy]||"";if(J.sortBy==="port")F=yq(F),U=yq(U);else if(J.sortBy==="vlan")F=zq(F),U=zq(U);else F=F.toString().toLowerCase(),U=U.toString().toLowerCase();if(F<U)return J.sortAsc?-1:1;if(F>U)return J.sortAsc?1:-1;return 0}),Q.length===0){K.innerHTML='<tr><td colspan="6" style="padding: 12px; text-align: center; color: #8b949e;">No MAC entries found</td></tr>';let $=document.querySelector(`.mac-scroll-container-${CSS.escape(q)}`);if($)$.scrollTop=J.scrollTop||0;return}K.innerHTML=Q.map(($)=>{let Y=$.vendor?$.vendor:'<a href="/config#vendor-editor-card" style="color: #8b949e; text-decoration: underline;" title="Edit custom vendor list">Unknown</a>',F=`
      <input type="text" value="${$.host||""}" 
             placeholder="Enter nickname..." 
             onfocus="pausePolling()" 
             onblur="saveHost(this, '${$.mac}')" 
             style="width: 100%; max-width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" 
             onfocus="this.style.borderColor='#58a6ff'" 
             onblur="this.style.borderColor='#30363d'">
    `;return`
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; font-family: 'SF Mono', Monaco, monospace; color: #c9d1d9;">${$.mac}</td>
        <td style="padding: 4px 12px; color: #b1bac4;">${F}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">${Y}</td>
        <td style="padding: 8px 12px; color: #8b949e;">${$.type}</td>
        <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${$.port}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">VLAN ${$.vlan}</td>
      </tr>
    `}).join("");let X=document.querySelector(`.mac-scroll-container-${CSS.escape(q)}`);if(X)X.scrollTop=J.scrollTop||0}async function FJ(q,J){let K=q.value;try{await fetch("/api/clients/update_host",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({mac:J,host:K})})}catch(Q){console.error("Failed to save client host nickname:",Q)}l=!1,document.getElementById("live-dot").style.background="#3fb950",jq(),d()}function bq(){l=!0,document.getElementById("live-dot").style.background="#d29922",document.getElementById("live-text").textContent="Paused"}async function LJ(q){let J=q.dataset.key,K=q.value;await fetch("/api/notes",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({key:J,note:K})}),l=!1,document.getElementById("live-dot").style.background="#3fb950",jq(),d()}function o(q){if(q===void 0||q===null)return"";if(q>=1000000000000)return(q/1000000000000).toFixed(1)+" TB";if(q>=1e9)return(q/1e9).toFixed(1)+" GB";if(q>=1e6)return(q/1e6).toFixed(1)+" MB";if(q>=1000)return(q/1000).toFixed(1)+" KB";return q+" B"}function D(q,J){if(!q)return J==="Bps"?"0 B/s":"0 bps";let K=J==="Bps"?q/8:q,Q=J==="Bps"?["B/s","KB/s","MB/s","GB/s"]:["bps","Kbps","Mbps","Gbps"],Z=[1e9,1e6,1000];if(K>=Z[0])return(K/Z[0]).toFixed(1)+" "+Q[3];if(K>=Z[1])return(K/Z[1]).toFixed(1)+" "+Q[2];if(K>=Z[2])return(K/Z[2]).toFixed(1)+" "+Q[1];return K.toFixed(0)+" "+Q[0]}function UJ(q){if(q)q.stopPropagation();O=O==="bps"?"Bps":"bps",d()}function Oq(q){if(!q)return"0";if(q>=1e6)return(q/1e6).toFixed(1)+"M";if(q>=1000)return(q/1000).toFixed(1)+"K";return q}function kJ(q,J){if(!J||J.toLowerCase()!=="up")return"speed-down";if(!q)return"";let K=q.toLowerCase();if(K.includes("10g"))return"speed-10g";if(K.includes("2500")||K.includes("2.5g"))return"speed-2500m";if(K.includes("2000")||K.includes("2g"))return"speed-2000m";if(K.includes("1000")||K.includes("1g"))return"speed-1000m";if(K.includes("100"))return"speed-100m";if(K.includes("10"))return"speed-10m";return""}function kq(q){if(!q)return"";return new Date(q*1000).toLocaleTimeString()}function NJ(q){let J=(j,k)=>{let N=qq[`mac-${j}`];return`
      <th style="${`position: relative; padding: 8px 12px; cursor: pointer; color: #8b949e; user-select: none; font-weight: 600;${N?` width: ${N}px;`:""}`}" onclick="sortMac('${q.ip}', '${j}')" data-col-id="mac-${j}">
        ${k} <span class="sort-indicator-${q.ip}-${j}"></span>
        <div class="resizer" onmousedown="handleResizeStart(event, 'mac-${j}')" style="position: absolute; top: 0; right: 0; width: 6px; height: 100%; cursor: col-resize; user-select: none; z-index: 10;"></div>
      </th>
    `},K=(j)=>{if(!j)return"#57606a";let k=(j.status||"").toLowerCase();if(k==="disable"||k==="disabled")return"#353c45";if(k!=="up")return"#57606a";let N=(j.speed||"").toLowerCase();if(N.includes("10g"))return"#1f6feb";if(N.includes("2.5g")||N.includes("2500")||N.includes("2g")||N.includes("2000"))return"#58a6ff";if(N.includes("1g")||N.includes("1000"))return"#3fb950";if(N.includes("100")||N.includes("10"))return"#f08a24";return"#57606a"},Q=typeof Gq<"u"?parseInt(Gq):0,Z=q.ports||[],X=Z.map((j)=>{let k=j.port==="9"||(j.speed||"").toLowerCase().includes("10g")||j.port==="SFP"||j.port==="9/SFP",N=K(j),V=[];if(j.vm_name)V.push(j.vm_name);if(j.interface&&j.interface!==j.port)V.push(j.interface);let _=V.length>0?` (${V.join(" / ")})`:"",L=`Port ${j.port}${_}: ${j.status.toUpperCase()} ${j.speed||""} (${j.duplex||""}) - Click to view graph`,W=k?"M 4,5 h 3 v 3 h 10 v -3 h 3 v 14 h -16 z":"M 4,8 h 5 v -3 h 6 v 3 h 5 v 11 h -16 z";return`
      <div style="display: flex; flex-direction: column; align-items: center; gap: 2px; cursor: pointer;" 
           title="${L}"
           onclick="openGraph('${q.ip}','${j.port}','${q.name}','Port ${j.port}', ${j.cum_tx||j.tx_bytes||0}, ${j.cum_rx||j.rx_bytes||0})">
        <span style="font-size: 10px; color: #8b949e; font-weight: 700; margin-bottom: 2px;">${j.port}</span>
        <svg viewBox="0 0 24 24" style="width: 18px; height: 18px; fill: ${N}; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.4)); transition: all 0.15s ease-in-out; transform: scale(1);"
             onmouseover="this.style.transform='scale(1.2)';"
             onmouseout="this.style.transform='scale(1)';">
          <path d="${W}" />
        </svg>
      </div>
    `}),$="";if(Q>0&&X.length>Q)for(let j=0;j<X.length;j+=Q){let k=X.slice(j,j+Q);$+=`
        <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
          ${k.join("")}
        </div>
      `}else $=`
      <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
        ${X.join("")}
      </div>
    `;let Y=Z.length>0?`
    <div style="background: rgba(13, 17, 23, 0.75); border: 1px solid #30363d; border-radius: 20px; padding: 10px 18px; display: inline-flex; flex-direction: column; gap: 10px; align-items: center; justify-content: center; box-shadow: inset 0 2px 4px rgba(0,0,0,0.4), 0 4px 10px rgba(0,0,0,0.3); margin: 4px 0;">
      ${$}
    </div>
  `:"",F=b.filter((j)=>Uq[j]),U=F.map((j)=>{let k=Uq[j],N=k.getLabel?k.getLabel():k.label,V=qq[j],_=V?`width: ${V}px;`:k.thStyle||"";return`<th draggable="true"
                ondragstart="handleDragStart(event, '${j}')"
                ondragover="handleDragOver(event)"
                ondragenter="handleDragEnter(event, '${j}')"
                ondragleave="handleDragLeave(event)"
                ondragend="handleDragEnd(event)"
                ondrop="handleDrop(event, '${j}')"
                data-col-id="${j}"
                style="cursor: move; user-select: none; ${_}"
                class="${k.thClass||""}">
              ${N}
              <div class="resizer" onmousedown="handleResizeStart(event, '${j}')"></div>
            </th>`}).join(""),G=(q.ports||[]).map((j)=>{return`<tr>${F.map((N)=>{let V=Uq[N],_=V.render(j,q),L=V.tdClassFn?V.tdClassFn(j):V.tdClass||"",W=V.tdTitleFn?V.tdTitleFn(j):V.tdTitle||"",H=L?`class="${L}"`:"",Xq=W?`title="${W}"`:"";return`<td ${H} ${Xq}>${_}</td>`}).join("")}</tr>`}).join(""),z=q.error?`<div class="error-card">Error: ${q.error}</div>`:"",B=q.uptime?`<span><span class="label">Uptime:</span> <span class="value">${q.uptime}</span></span>`:"",T=q.mac?`<span><span class="label">MAC:</span> <span class="value">${q.mac}</span></span>`:"",A=q.firmware?`<span><span class="label">Firmware:</span> <span class="value">${q.firmware}</span></span>`:"",y="";if(q.dhcp_snooping){let j=q.dhcp_snooping.enabled;y=`<span><span class="label">DHCP Snooping:</span> <span class="value" style="color: ${j?"#56d364":"#8b949e"}">${j?"\uD83D\uDFE2 Active":"⚪ Inactive"}</span></span>`}let g="";if(q.igmp){let j=q.igmp.enabled;g=`<span><span class="label">IGMP Snooping:</span> <span class="value" style="color: ${j?"#56d364":"#8b949e"}">${j?"\uD83D\uDFE2 Active":"⚪ Inactive"}</span></span>`}let S="";if(q.jumbo_frame){let j=q.jumbo_frame.enabled,k=q.jumbo_frame.size;S=`<span><span class="label">Jumbo Frame:</span> <span class="value" style="color: ${j?"#56d364":"#8b949e"}">${j?"\uD83D\uDFE2 Active ("+k+")":"⚪ Inactive"}</span></span>`}let E=R(q.ip),h=E.expanded,C=h?"block":"none",P=h?"rotate(90deg)":"rotate(0deg)",Kq=q.mac_table?q.mac_table.length:0,r=q.mac_timestamp?kq(q.mac_timestamp):"Never",x=`
    <div class="mac-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="mac-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleMacTable('${q.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="mac-toggle-arrow-${q.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${P}; font-size: 10px;">&gt;</span>
          <span>\uD83D\uDCC1 MAC Address Table (<span class="mac-count-${q.ip}">${Kq} entries</span>)</span>
        </h4>
        <span class="mac-time-${q.ip}" style="font-size: 10px; color: #8b949e;">Last Scraped: ${r}</span>
      </div>
      
      <div class="mac-content-${q.ip}" style="display: ${C}; margin-top: 12px;">
        <div style="display: flex; justify-content: flex-end; margin-bottom: 10px;">
          <button class="mac-refresh-btn-${q.ip}" onclick="manualRefreshMac('${q.ip}')" style="background: #21262d; border: 1px solid #30363d; border-radius: 6px; padding: 6px 12px; color: #c9d1d9; font-size: 11px; font-weight: 600; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;" onmouseover="this.style.background='#30363d'; this.style.borderColor='#8b949e';" onmouseout="this.style.background='#21262d'; this.style.borderColor='#30363d';">
            <span class="mac-spinner-${q.ip}" style="display: none; width: 12px; height: 12px; border: 2px solid transparent; border-top-color: #58a6ff; border-radius: 50%; animation: spin .8s linear infinite; box-sizing: border-box;"></span>
            <span>Refresh MACs</span>
          </button>
        </div>
        <div class="mac-scroll-container-${q.ip}" onscroll="saveMacScroll('${q.ip}', this)" onmouseup="saveMacHeight('${q.ip}', this)" ontouchend="saveMacHeight('${q.ip}', this)" style="height: ${E.height||"450px"}; min-height: 150px; max-height: 1200px; resize: vertical; overflow-y: auto; border: 1px solid #30363d; border-radius: 8px; background: #0d1117; box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);">
          <table class="mac-table" style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left; table-layout: fixed;">
            <thead>
              <tr style="background: #161b22; position: sticky; top: 0; z-index: 2; border-bottom: 1px solid #30363d;">
                ${J("mac","MAC Address")}
                ${J("host","Host")}
                ${J("vendor","Vendor")}
                ${J("type","Type")}
                ${J("port","Port")}
                ${J("vlan","VLAN")}
              </tr>
              <tr style="background: #0d1117; position: sticky; top: 31px; z-index: 2; border-bottom: 1px solid #30363d;">
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-mac-${q.ip}" placeholder="Filter MAC..." value="${E.filters?E.filters.mac:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-host-${q.ip}" placeholder="Filter Host..." value="${E.filters?E.filters.host:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-vendor-${q.ip}" placeholder="Filter Vendor..." value="${E.filters?E.filters.vendor:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-type-${q.ip}" placeholder="Filter Type..." value="${E.filters?E.filters.type:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-port-${q.ip}" placeholder="Port..." value="${E.filters?E.filters.port:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-vlan-${q.ip}" placeholder="VLAN..." value="${E.filters?E.filters.vlan:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
              </tr>
            </thead>
            <tbody class="mac-tbody-${q.ip}">
              <!-- Populate via Javascript -->
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `,Qq=ZJ(q.ip).expanded,Yq=Qq?"block":"none",a=Qq?"rotate(90deg)":"rotate(0deg)",u=q.igmp&&q.igmp.entries?q.igmp.entries:[],i=u.length,v='<tr><td colspan="3" style="padding: 12px; text-align: center; color: #8b949e;">No active multicast groups detected</td></tr>';if(u.length>0)v=u.map((j)=>`
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; color: #58a6ff; font-weight: 600; font-family: 'SF Mono', Monaco, monospace;">${j.ip}</td>
        <td style="padding: 8px 12px; color: #c9d1d9; font-weight: 600;">${j.ports}</td>
        <td style="padding: 8px 12px; color: #8b949e;">VLAN ${j.vlan}</td>
      </tr>
    `).join("");let Zq=`
    <div class="igmp-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="igmp-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleIgmpTable('${q.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="igmp-toggle-arrow-${q.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${a}; font-size: 10px;">&gt;</span>
          <span>\uD83D\uDCC1 IGMP Multicast Groups (<span class="igmp-count-${q.ip}">${i} entries</span>)</span>
        </h4>
      </div>
      
      <div class="igmp-content-${q.ip}" style="display: ${Yq}; margin-top: 12px;">
        <div style="border: 1px solid #30363d; border-radius: 8px; background: #0d1117; overflow: hidden; box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);">
          <table style="width: 100%; border-collapse: collapse; font-size: 11px; text-align: left;">
            <thead>
              <tr style="background: #161b22; border-bottom: 1px solid #30363d;">
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600; width: 45%;">Multicast IP</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600; width: 35%;">Member Ports</th>
                <th style="padding: 8px 12px; color: #8b949e; font-weight: 600; width: 20%;">VLAN</th>
              </tr>
            </thead>
            <tbody>
              ${v}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;return`<div class="switch-card">
    <div class="switch-header">
      <div style="display: flex; align-items: center; gap: 12px;">
        <img src="/static/logo.png" alt="Switch" style="height: 24px; width: auto; object-fit: contain; filter: drop-shadow(0 2px 4px rgba(0,0,0,0.5));">
        <div>
          <h2>${q.name}</h2>
          <span style="font-size:12px;color:#b1bac4;display:flex;align-items:center;gap:8px;">
            <a href="http://${q.ip}/" target="_blank" style="color:#58a6ff;text-decoration:none" title="Open switch web UI">${q.ip}</a>
            <span style="color:#30363d">|</span>
            <a href="#" onclick="backupConfig('${q.ip}', this); return false;" style="color:#58a6ff;text-decoration:none;font-weight:600;display:inline-flex;align-items:center;gap:4px;" title="Backup configuration on server">\uD83D\uDCBE Backup</a>
          </span>
        </div>
      </div>
      
      <!-- Port Graphic -->
      ${Y}
      
      <div class="meta">${q.model||""}</div>
    </div>
    <div class="device-bar">
      ${B}
      ${A}
      ${T}
      ${y}
      ${g}
      ${S}
    </div>
    <table class="port-table">
      <thead><tr>
        ${U}
      </tr></thead>
      <tbody>${G||`<tr><td colspan="${F.length}" style="padding:20px;text-align:center;color:#b1bac4">No data</td></tr>`}</tbody>
    </table>
    ${z}
    ${x}
    ${Zq}
    <div class="last-update">Last update: ${kq(q.timestamp)}</div>
  </div>`}async function d(){if(l)return;let q=document.getElementById("live-dot"),J=document.getElementById("live-text");q.style.background="#d29922",J.textContent="Updating...";try{let Q=await(await fetch("/api/switches")).json(),Z=document.getElementById("app");if(!Q||Q.length===0){Z.innerHTML='<div style="text-align:center;padding:40px;color:#b1bac4">No nodes configured. <a href="/config" style="color:#58a6ff">Configure now</a></div>',Z.className="",q.style.background="#f85149",J.textContent="No nodes";return}if(Q.forEach((X)=>{let $=document.querySelector(`.mac-scroll-container-${CSS.escape(X.ip)}`);if($){let Y=R(X.ip);if(Y.scrollTop=$.scrollTop,$.style.height)Y.height=$.style.height}}),s(),Z.className="switches",Z.innerHTML=Q.map(NJ).join(""),Q.forEach((X)=>{let $=R(X.ip);if($.data=X.mac_table||[],$.expanded)Jq(X.ip)}),document.getElementById("graph-overlay").classList.contains("open")&&w&&m){Nq(w,m);let X=Q.find(($)=>$.ip===w);if(X){let $=(X.ports||[]).find((Y)=>Y.port===m);if($){let Y=$.cum_tx||$.tx_bytes||0,F=$.cum_rx||$.rx_bytes||0;document.getElementById("graph-total-tx").innerHTML=`${Y.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${o(Y)})</span>`,document.getElementById("graph-total-rx").innerHTML=`${F.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${o(F)})</span>`}}}q.style.background="#3fb950",jq()}catch(K){document.getElementById("app").innerHTML=`<div style="text-align:center;padding:40px;color:#f85149">Connection error: ${K.message}</div>`,q.style.background="#f85149",J.textContent="Error"}}function VJ(q,J,K,Q,Z=0,X=0){w=q,m=J,hq=K,xq=Q,document.getElementById("graph-title").textContent=Q+" Speed ("+K+")",document.getElementById("graph-overlay").classList.add("open"),document.getElementById("graph-total-tx").innerHTML=`${Z.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${o(Z)})</span>`,document.getElementById("graph-total-rx").innerHTML=`${X.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${o(X)})</span>`,AJ(e)}function AJ(q){if(e=q,document.querySelectorAll(".graph-tab-btn").forEach((J)=>J.classList.remove("active")),q==="live")document.getElementById("tab-live").classList.add("active"),document.getElementById("graph-sub").textContent=`Real-time speed at ${t}s polling intervals`;else if(q==="1h")document.getElementById("tab-1h").classList.add("active"),document.getElementById("graph-sub").textContent=`Last 1 hour speed history (at ${t}s polling intervals)`;else if(q==="24h")document.getElementById("tab-24h").classList.add("active"),document.getElementById("graph-sub").textContent="Last 24 hours speed history (15-minute averages)";Nq(w,m)}function Nq(q,J){let K=document.getElementById("graph-svg"),Q=document.getElementById("graph-box"),Z=Q.classList.contains("maximized"),X=Math.max(500,Q.clientWidth-(Z?64:48)),$=window.innerHeight*0.35;if(Z)$=window.innerHeight-250;$=Math.max(240,$),K.setAttribute("viewBox",`0 0 ${X} ${$}`),K.innerHTML="";let Y={top:20,bottom:40,left:70,right:30},F=X-Y.left-Y.right,U=$-Y.top-Y.bottom;fetch(`/api/history?ip=${q}&port=${J}&range=${e}`).then((G)=>G.json()).then((G)=>{if(!G||!G.tx||G.tx.length<2){K.innerHTML=`<text x="${X/2}" y="${$/2}" text-anchor="middle" fill="#b1bac4" font-size="14">No history data available for this range yet</text>`,document.getElementById("graph-peak-tx").textContent="-",document.getElementById("graph-peak-rx").textContent="-",document.getElementById("graph-now-tx").textContent="-",document.getElementById("graph-now-rx").textContent="-";return}let{tx:z,rx:B,timestamps:T}=G,A=z.length,y=0;for(let _=0;_<A;_++){if(z[_]>y)y=z[_];if(B[_]>y)y=B[_]}if(y===0)y=1000;y*=1.15;let g=Math.max(...z),S=Math.max(...B),E=z[A-1],h=B[A-1];document.getElementById("graph-peak-tx").textContent=D(g,O),document.getElementById("graph-peak-rx").textContent=D(S,O),document.getElementById("graph-now-tx").textContent=D(E,O),document.getElementById("graph-now-rx").textContent=D(h,O);function C(_){return Y.left+_/(A-1)*F}function P(_){return Y.top+U-_/y*U}let Kq=4,r="";for(let _=0;_<=Kq;_++){let L=y/Kq*_,W=P(L);r+=`<line x1="${Y.left}" y1="${W}" x2="${X-Y.right}" y2="${W}" stroke="#21262d" stroke-width="1" stroke-dasharray="${_===0?"0":"4 4"}"/>`,r+=`<text x="${Y.left-10}" y="${W+4}" text-anchor="end" fill="#b1bac4" font-size="11" font-family="'SF Mono', Monaco, monospace">${D(L,O)}</text>`}let x=[],Vq=Math.min(5,A),Qq=Math.max(1,Math.floor((A-1)/(Vq-1)));for(let _=0;_<A;_+=Qq)x.push(_);if(x[x.length-1]!==A-1)x.push(A-1);let Yq=x.map((_)=>{let L=C(_),W=new Date(T[_]*1000),H=W.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit",second:"2-digit"});if(e==="24h")H=W.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})+"<br>"+(W.getMonth()+1)+"/"+W.getDate();return`
          <line x1="${L}" y1="${Y.top}" x2="${L}" y2="${Y.top+U}" stroke="#21262d" stroke-dasharray="4 4" stroke-width="0.5"/>
          <text x="${L}" y="${$-12}" text-anchor="middle" fill="#b1bac4" font-size="10">${H}</text>
        `}).join(""),a="",u="",i=`M ${C(0)} ${P(0)} `,v=`M ${C(0)} ${P(0)} `;for(let _=0;_<A;_++){let L=C(_).toFixed(1),W=P(z[_]).toFixed(1),H=P(B[_]).toFixed(1);if(_===0)a+=`M ${L} ${W}`,u+=`M ${L} ${H}`,i=`M ${L} ${Y.top+U} L ${L} ${W}`,v=`M ${L} ${Y.top+U} L ${L} ${H}`;else a+=` L ${L} ${W}`,u+=` L ${L} ${H}`,i+=` L ${L} ${W}`,v+=` L ${L} ${H}`}i+=` L ${C(A-1).toFixed(1)} ${Y.top+U} Z`,v+=` L ${C(A-1).toFixed(1)} ${Y.top+U} Z`,K.innerHTML=`
        <defs>
          <linearGradient id="tx-grad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#58a6ff" stop-opacity="0.35"/>
            <stop offset="100%" stop-color="#58a6ff" stop-opacity="0.0"/>
          </linearGradient>
          <linearGradient id="rx-grad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#3fb950" stop-opacity="0.35"/>
            <stop offset="100%" stop-color="#3fb950" stop-opacity="0.0"/>
          </linearGradient>
        </defs>
        
        <!-- Grid and Axes -->
        ${r}
        ${Yq}

        <!-- Filled Areas (Gradients) -->
        <path d="${i}" fill="url(#tx-grad)" />
        <path d="${v}" fill="url(#rx-grad)" />

        <!-- Line curves -->
        <path d="${a}" fill="none" stroke="#58a6ff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="${u}" fill="none" stroke="#3fb950" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        
        <!-- Legend labels -->
        <g transform="translate(${X-Y.right-90}, ${Y.top+5})">
          <rect x="0" y="0" width="12" height="12" fill="#58a6ff" rx="3"/>
          <text x="18" y="10" fill="#b1bac4" font-size="11" font-weight="600">TX Speed</text>
        </g>
        <g transform="translate(${X-Y.right-90}, ${Y.top+22})">
          <rect x="0" y="0" width="12" height="12" fill="#3fb950" rx="3"/>
          <text x="18" y="10" fill="#b1bac4" font-size="11" font-weight="600">RX Speed</text>
        </g>

        <!-- Hover Guideline / Interactive Crosshair components -->
        <line id="hover-line" x1="0" y1="${Y.top}" x2="0" y2="${Y.top+U}" stroke="#b1bac4" stroke-width="1" stroke-dasharray="3 3" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-tx" r="5" fill="#58a6ff" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-rx" r="5" fill="#3fb950" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>

        <!-- Invisible Overlay rect for hover detection -->
        <rect id="hover-overlay-rect" x="${Y.left}" y="${Y.top}" width="${F}" height="${U}" fill="transparent" style="cursor: crosshair;"/>
      `;let Zq=document.getElementById("hover-overlay-rect"),j=document.getElementById("hover-line"),k=document.getElementById("hover-dot-tx"),N=document.getElementById("hover-dot-rx"),V=document.getElementById("graph-tooltip");Zq.addEventListener("mousemove",(_)=>{let L=K.getBoundingClientRect(),Xq=(_.clientX-L.left)/L.width*X-Y.left,M=Math.round(Xq/F*(A-1));if(M<0)M=0;if(M>=A)M=A-1;let $q=C(M),fq=P(z[M]),Tq=P(B[M]);j.setAttribute("x1",$q),j.setAttribute("x2",$q),j.style.display="block",k.setAttribute("cx",$q),k.setAttribute("cy",fq),k.style.display="block",N.setAttribute("cx",$q),N.setAttribute("cy",Tq),N.style.display="block";let _q=new Date(T[M]*1000),Aq=_q.toLocaleTimeString();if(e==="24h")Aq=_q.toLocaleDateString()+" "+_q.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"});V.innerHTML=`
          <div class="time">${Aq}</div>
          <div class="tx">TX: ${D(z[M],O)}</div>
          <div class="rx">RX: ${D(B[M],O)}</div>
        `;let Eq=K.parentElement.getBoundingClientRect(),gq=_.clientX-Eq.left+15,Sq=_.clientY-Eq.top-60;V.style.left=`${gq}px`,V.style.top=`${Sq}px`,V.style.display="block"}),Zq.addEventListener("mouseleave",()=>{j.style.display="none",k.style.display="none",N.style.display="none",V.style.display="none"})}).catch((G)=>{K.innerHTML=`<text x="${X/2}" y="${$/2}" text-anchor="middle" fill="#f85149" font-size="14">Error loading history: ${G.message}</text>`})}function EJ(q,J,K){rq=q,aq=J,eq=K;let Q=document.getElementById("transceiver-overlay"),Z=document.getElementById("transceiver-title"),X=document.getElementById("transceiver-sub"),$=document.getElementById("transceiver-content");Z.textContent=`Port ${J} SFP+ Transceiver Status`,X.textContent=`DDMI Diagnostics & Telemetry (${K})`,$.innerHTML=`
    <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px; gap: 12px; color: #8b949e;">
      <div style="width: 24px; height: 24px; border: 2px solid transparent; border-top-color: #58a6ff; border-radius: 50%; animation: spin .8s linear infinite; box-sizing: border-box;"></div>
      <span>Fetching transceiver data from switch...</span>
    </div>
  `,Q.classList.add("open"),fetch(`/api/switches/${q}/transceiver`).then((Y)=>{if(!Y.ok)return Y.json().then((F)=>{throw Error(F.error||"Request failed")});return Y.json()}).then((Y)=>{WJ(Y,$)}).catch((Y)=>{$.innerHTML=`
        <div class="error-card" style="margin: 20px 0; font-size: 13px;">
          <strong>No SFP module detected or telemetry unavailable</strong><br>
          <span style="opacity: 0.8; font-size: 11px;">${Y.message}</span>
        </div>
      `})}function WJ(q,J){function K(E){if(!E)return 0;let h=E.toString().match(/[-+]?[0-9]*\.?[0-9]+/);return h?parseFloat(h[0]):0}let Q=K(q.temperature),Z=K(q.voltage),X=K(q.current),$=-99,Y=-99;if(q.tx_power&&!q.tx_power.includes("-inf"))$=K(q.tx_power);if(q.rx_power&&!q.rx_power.includes("-inf"))Y=K(q.rx_power);let F=Math.min(100,Math.max(0,Q/80*100)),U=Math.min(100,Math.max(0,(Z-2.8)/1*100)),G=Math.min(100,Math.max(0,X/40*100)),z=$===-99?0:Math.min(100,Math.max(0,($- -20)/25*100)),B=Y===-99?0:Math.min(100,Math.max(0,(Y- -30)/35*100)),T=Q>65||Q<5?"#f85149":Q>55?"#d29922":"#3fb950",A=Z>3.5||Z<3.1?"#f85149":"#3fb950",y=X>30?"#f85149":"#3fb950",g=$===-99?"#57606a":$<-12?"#d29922":"#3fb950",S=Y===-99?"#57606a":Y<-18?"#d29922":"#3fb950";J.innerHTML=`
    <div class="transceiver-grid">
      <!-- SFP Module Details Card -->
      <div class="transceiver-card">
        <h4>
          <svg viewBox="0 0 24 24" style="width: 14px; height: 14px; fill: currentColor; display: inline-block; vertical-align: middle;"><path d="M4,5 h3 v3 h10 v -3 h 3 v 14 h -16 z" /></svg>
          SFP Module Details
        </h4>
        <div class="transceiver-item"><span class="label">OE Present</span><span class="val">${q.oe_present||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Loss of Signal</span><span class="val">${q.loss_of_signal||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Transceiver Type</span><span class="val">${q.transceiver_type||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Connector Type</span><span class="val">${q.connector_type||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Compliance Codes</span><span class="val">${q.eth_compliance_codes||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Wavelength</span><span class="val">${q.wavelength||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Bitrate</span><span class="val">${q.bitrate||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Vendor Name</span><span class="val">${(q.vendor_name||"Unknown").trim()}</span></div>
        <div class="transceiver-item"><span class="label">Vendor OUI</span><span class="val">${q.vendor_oui||"Unknown"}</span></div>
        <div class="transceiver-item"><span class="label">Part Number</span><span class="val">${(q.vendor_pn||"Unknown").trim()}</span></div>
        <div class="transceiver-item"><span class="label">Revision</span><span class="val">${(q.vendor_revision||"Unknown").trim()}</span></div>
        <div class="transceiver-item"><span class="label">Serial Number</span><span class="val">${(q.vendor_sn||"Unknown").trim()}</span></div>
        <div class="transceiver-item"><span class="label">Date Code</span><span class="val">${q.date_code||"Unknown"}</span></div>
      </div>
      
      <!-- DDMI Real-Time Telemetry Card -->
      <div class="transceiver-card">
        <h4>
          <svg viewBox="0 0 24 24" style="width: 14px; height: 14px; fill: currentColor; display: inline-block; vertical-align: middle;"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h-2v-2h4v8zm0-10h-2V5h2v2z"/></svg>
          DDMI Diagnostics
        </h4>
        
        <!-- Temperature -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">\uD83C\uDF21️ Temperature</span>
            <span class="val" style="color: ${T};">${q.temperature||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${F}%; background: ${T};"></div>
          </div>
        </div>
        
        <!-- Voltage -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">⚡ Supply Voltage</span>
            <span class="val" style="color: ${A};">${q.voltage||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${U}%; background: ${A};"></div>
          </div>
        </div>
        
        <!-- Current -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">\uD83D\uDD0B Bias Current</span>
            <span class="val" style="color: ${y};">${q.current||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${G}%; background: ${y};"></div>
          </div>
        </div>
        
        <!-- TX Power -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">\uD83D\uDCE4 Optical TX Power</span>
            <span class="val" style="color: ${g};">${q.tx_power||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${z}%; background: ${g};"></div>
          </div>
        </div>
        
        <!-- RX Power -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">\uD83D\uDCE5 Optical RX Power</span>
            <span class="val" style="color: ${S};">${q.rx_power||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${B}%; background: ${S};"></div>
          </div>
        </div>
        
        <div style="margin-top: 18px; padding: 10px; background: rgba(88,166,255,0.05); border: 1px solid rgba(88,166,255,0.15); border-radius: 6px; font-size: 11px; color: #8b949e; line-height: 1.4;">
          \uD83D\uDCA1 Telemetry values are updated directly from the switch's internal SFP micro-controller registers using DDMI.
        </div>
      </div>
    </div>
  `}function GJ(){document.getElementById("confirm-overlay").classList.remove("open")}async function yJ(){GJ(),await fetch("/api/reset",{method:"POST"}),d()}function Iq(q){document.body.dataset.fontSize=q,document.querySelectorAll(".fs-btn").forEach((J)=>J.classList.toggle("fs-active",J.dataset.size===q)),fetch("/api/settings",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({font_size:q})})}async function zJ(){try{let J=await(await fetch("/api/settings")).json();if(J.font_size)Iq(J.font_size)}catch(q){}}zJ();d();uq();window.addEventListener("resize",()=>{if(document.getElementById("graph-overlay").classList.contains("open")&&w&&m)Nq(w,m)});window.setFontSize=Iq;window.doReset=yJ;window.backupConfig=_J;window.manualRefreshMac=XJ;window.openGraph=VJ;window.openTransceiver=EJ;window.sortMac=YJ;window.toggleIgmpTable=$J;window.toggleMacTable=QJ;window.toggleSpeedUnit=UJ;window.pausePolling=bq;window.saveNote=LJ;window.saveHost=FJ;window.focusMacSearch=tq;window.blurMacSearch=qJ;window.saveMacHeight=KJ;window.handleResizeStart=lq;window.handleDragStart=dq;window.handleDragEnter=pq;window.handleDragOver=iq;window.handleDragLeave=nq;window.handleDragEnd=oq;window.handleDrop=sq;window.filterMacTable=jJ;window.saveMacScroll=JJ;
