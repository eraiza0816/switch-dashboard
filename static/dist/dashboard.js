var p=window.__DATA__&&window.__DATA__.refresh||30,kq=window.__DATA__&&window.__DATA__.columns||["port","status","speed","packets","bytes","info","notes"],Vq=window.__DATA__&&window.__DATA__.portWrap||0,h=null,v=null,i="live",Hq="",Rq="",B="Bps",Xq=!1,m=p,Qq=null;function Cq(){m=p;let q=document.getElementById("live-text");if(q)if(Xq)q.textContent="Paused";else q.textContent=`Live (update in ${m}s)`}function Dq(){if(Qq)clearInterval(Qq);Qq=setInterval(()=>{let q=document.getElementById("live-text"),J=document.getElementById("live-dot");if(!q||!J)return;if(Xq){q.textContent="Paused",J.style.background="#d29922";return}if(m--,m<=0)q.textContent="Updating...",J.style.background="#d29922",m=p,t();else q.textContent=`Live (update in ${m}s)`,J.style.background="#3fb950"},1000)}var Pq=["port","status","speed","packets","bytes","info","notes"],a=[],Gq={},Zq=[];function bq(){let q=Zq&&Zq.length>0?Zq:null;if(!q)q=Pq;a=q.filter((J)=>kq.includes(J)),kq.forEach((J)=>{if(!a.includes(J))a.push(J)})}bq();var jq=Gq||{};function fq(){jq=Gq||{}}fq();var $q={port:{label:"Port",thStyle:"width:55px",tdClass:"port-num",tdTitleFn:(q)=>{let J=[];if(q.vm_name)J.push(q.vm_name);if(q.interface)J.push(q.interface);return J.length>0?J.join(" / "):""},render:(q,J)=>q.port},status:{label:"Status",thStyle:"width:145px",render:(q,J)=>{let K=(q.status||"unknown").toLowerCase(),Z="";if(J.dhcp_snooping&&J.dhcp_snooping.enabled){let X=J.dhcp_snooping.ports||{},$=q.port.split("/")[0],Y=X[q.port]||X[$];if(Y==="Trusted")Z='<span class="trust-badge trusted" title="DHCP Snooping: Trusted Port">Trusted</span>';else if(Y==="Untrusted")Z='<span class="trust-badge untrusted" title="DHCP Snooping: Untrusted Port">Untrusted</span>'}let j=q.duplex||"Auto";if(j.toLowerCase()==="full")j="Full Duplex";else if(j.toLowerCase()==="half")j="Half Duplex";return`
        <div style="display: flex; flex-direction: column; gap: 4px; align-items: flex-start;">
          <div style="display: flex; align-items: center; gap: 6px;">
            <span class="status-badge ${K}" onclick="openGraph('${J.ip}','${q.port}','${J.name}','Port ${q.port}', ${q.cum_tx||q.tx_bytes||0}, ${q.cum_rx||q.rx_bytes||0})"><span class="status-dot ${K}"></span>${K.toUpperCase()}</span>
            ${Z}
          </div>
          <div style="font-size: 10px; color: #8b949e; padding-left: 4px; font-weight: 500;">${j}</div>
        </div>
      `}},speed:{label:"Speed",thStyle:"width:55px",tdClassFn:(q)=>cq(q.speed,q.status),tdTitleFn:(q)=>q.duplex||"",render:(q,J)=>{let K=q.port==="9"||(q.speed||"").toLowerCase().includes("10g")||q.port==="SFP"||q.port==="9/SFP",Z=q.speed||"Auto";if(K)return`<span class="sfp-speed-link" onclick="openTransceiver('${J.ip}','${q.port}','${J.name}')" title="Click to view SFP+ Transceiver Diagnostics">${Z} <svg viewBox="0 0 24 24" style="width: 10px; height: 10px; fill: currentColor; display: inline-block;"><path d="M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6zm0-7C7 2 2.73 5.11 1 9.5 2.73 13.89 7 17 12 17s9.27-3.11 11-7.5C21.27 5.11 17 2 12 2zm0 13a5.5 5.5 0 1 1 0-11 5.5 5.5 0 0 1 0 11z"/></svg></span>`;else return Z}},packets:{label:'Traffic<br><span style="font-weight:400;color:#8b949e">(packets)</span>',thStyle:"width:90px",tdClass:"mono",tdTitle:"TX / RX packets",render:(q,J)=>{let K=q.tx_packets||0,Z=q.rx_packets||0;return`TX ${Wq(K)}<br>RX ${Wq(Z)}`}},bytes:{label:'Traffic<br><span style="font-weight:400;color:#8b949e">(bytes)</span>',thStyle:"width:90px",tdClass:"mono",tdTitle:"TX / RX bytes formatted",render:(q,J)=>{let K=q.cum_tx||q.tx_bytes||0,Z=q.cum_rx||q.rx_bytes||0;return`TX ${c(K)}<br>RX ${c(Z)}`}},info:{getLabel:()=>`<span id="speed-unit-toggle" class="unit-toggle" onclick="toggleSpeedUnit(event)" title="Toggle speed unit">${B==="bps"?"BPS":"B/s"}</span>`,tdClass:"col-info mono",render:(q,J)=>{let K=q.speed_tx_bps||0,Z=q.speed_rx_bps||0;return`TX ${D(K,B)}<br>RX ${D(Z,B)}`}},notes:{label:"Notes",tdClass:"col-note",render:(q,J)=>{return`<input class="note-input" type="text" placeholder="-" value="${q.note||""}" data-key="${J.ip}:${q.port}" onfocus="pausePolling()" onblur="saveNote(this)">`}}};var Iq=null,Tq=null,gq="";var u={};try{let q=localStorage.getItem("macTableStates");if(q)u=JSON.parse(q)}catch(q){console.error("Failed to load macTableStates from localStorage",q)}function _q(){let q={};for(let J in u)if(u.hasOwnProperty(J)){let K=u[J];q[J]={expanded:K.expanded,filters:K.filters||{mac:"",vendor:"",type:"",port:"",vlan:""},sortBy:K.sortBy||"port",sortAsc:K.sortAsc!==void 0?K.sortAsc:!0,scrollTop:K.scrollTop||0,height:K.height||"450px"}}localStorage.setItem("macTableStates",JSON.stringify(q))}function x(q){if(!u[q])u[q]={};let J=u[q];if(J.expanded===void 0)J.expanded=!1;if(!J.filters)J.filters={mac:"",host:"",vendor:"",type:"",port:"",vlan:""};if(J.sortBy===void 0)J.sortBy="port";if(J.sortAsc===void 0)J.sortAsc=!0;if(!J.data)J.data=[];if(J.scrollTop===void 0)J.scrollTop=0;if(J.height===void 0)J.height="450px";return J}function Aq(q){if(!q)return 0;if(!isNaN(q))return parseInt(q);let J=q.match(/\d+/);if(J)return parseInt(J[0]);return q.toString().toLowerCase()}function Eq(q){if(!q)return 0;let J=parseInt(q);return isNaN(J)?q.toString().toLowerCase():J}function Sq(q){let J=x(q);J.expanded=!J.expanded,_q();let K=document.querySelector(`.mac-content-${CSS.escape(q)}`),Z=document.querySelector(`.mac-toggle-arrow-${CSS.escape(q)}`);if(K&&Z)if(J.expanded)K.style.display="block",Z.style.transform="rotate(90deg)",e(q);else K.style.display="none",Z.style.transform="rotate(0deg)"}function hq(q){let J={};try{let K=localStorage.getItem("igmp_states");if(K)J=JSON.parse(K)}catch(K){console.error("Failed to load igmp_states from localStorage",K)}if(!J[q])J[q]={expanded:!1};return J[q]}function vq(q){let J={};try{let j=localStorage.getItem("igmp_states");if(j)J=JSON.parse(j)}catch(j){console.error("Failed to load igmp_states from localStorage",j)}if(!J[q])J[q]={expanded:!1};J[q].expanded=!J[q].expanded,localStorage.setItem("igmp_states",JSON.stringify(J));let K=document.querySelector(`.igmp-content-${CSS.escape(q)}`),Z=document.querySelector(`.igmp-toggle-arrow-${CSS.escape(q)}`);if(K&&Z)if(J[q].expanded)K.style.display="block",Z.style.transform="rotate(90deg)";else K.style.display="none",Z.style.transform="rotate(0deg)"}function uq(q,J){let K=x(q);if(K.sortBy===J)K.sortAsc=!K.sortAsc;else K.sortBy=J,K.sortAsc=!0;_q(),e(q)}function xq(q){let J=x(q),K=document.querySelector(`.mac-spinner-${CSS.escape(q)}`),Z=document.querySelector(`.mac-refresh-btn-${CSS.escape(q)} span:not(.mac-spinner-${CSS.escape(q)})`);if(K)K.style.display="inline-block";if(Z)Z.textContent="Refreshing...";fetch(`/api/switches/${q}/refresh_mac`,{method:"POST"}).then((j)=>j.json()).then((j)=>{if(j.status==="ok"){J.data=j.mac_table||[];let X=document.querySelector(`.mac-time-${CSS.escape(q)}`);if(X)X.textContent="Last Scraped: "+Yq(Date.now()/1000);let $=document.querySelector(`.mac-count-${CSS.escape(q)}`);if($)$.textContent=`${J.data.length} entries`;e(q)}else alert("Failed to refresh MAC table: "+(j.error||"Unknown error"))}).catch((j)=>{alert("Error refreshing MAC table: "+j.message)}).finally(()=>{if(K)K.style.display="none";if(Z)Z.textContent="Refresh MACs"})}function wq(q,J){if(J.dataset.loading==="true")return;J.dataset.loading="true";let K=J.innerHTML;J.innerHTML="⏳ Backing up...",J.style.color="#8b949e",fetch(`/api/switches/${q}/backup`,{method:"POST"}).then((Z)=>Z.json()).then((Z)=>{if(Z.status==="ok")J.innerHTML="✓ Backed up!",J.style.color="#3fb950";else alert("Failed to backup config: "+(Z.error||"Unknown error")),J.innerHTML="❌ Failed",J.style.color="#ff7b72"}).catch((Z)=>{alert("Error during backup: "+Z.message),J.innerHTML="❌ Failed",J.style.color="#ff7b72"}).finally(()=>{setTimeout(()=>{J.innerHTML=K,J.style.color="#58a6ff",delete J.dataset.loading},3000)})}function e(q){let J=x(q),K=document.querySelector(`.mac-tbody-${CSS.escape(q)}`);if(!K)return;["mac","host","vendor","type","port","vlan"].forEach(($)=>{let Y=document.querySelector(`.sort-indicator-${CSS.escape(q)}-${$}`);if(Y)if(J.sortBy===$)Y.textContent=J.sortAsc?" ▴":" ▾",Y.style.color="#58a6ff";else Y.textContent=""});let Z=J.data,j=J.filters||{mac:"",host:"",vendor:"",type:"",port:"",vlan:""};if(j.mac||j.host||j.vendor||j.type||j.port||j.vlan)Z=Z.filter(($)=>{let Y=!j.mac||($.mac||"").toLowerCase().includes(j.mac),L=!j.host||($.host||"").toLowerCase().includes(j.host),N=!j.vendor||($.vendor||"").toLowerCase().includes(j.vendor),G=!j.type||($.type||"").toLowerCase().includes(j.type),z=!j.port||($.port||"").toString().toLowerCase().includes(j.port),O=!j.vlan||($.vlan||"").toString().toLowerCase().includes(j.vlan);return Y&&L&&N&&G&&z&&O});if(Z.sort(($,Y)=>{let L=$[J.sortBy]||"",N=Y[J.sortBy]||"";if(J.sortBy==="port")L=Aq(L),N=Aq(N);else if(J.sortBy==="vlan")L=Eq(L),N=Eq(N);else L=L.toString().toLowerCase(),N=N.toString().toLowerCase();if(L<N)return J.sortAsc?-1:1;if(L>N)return J.sortAsc?1:-1;return 0}),Z.length===0){K.innerHTML='<tr><td colspan="6" style="padding: 12px; text-align: center; color: #8b949e;">No MAC entries found</td></tr>';let $=document.querySelector(`.mac-scroll-container-${CSS.escape(q)}`);if($)$.scrollTop=J.scrollTop||0;return}K.innerHTML=Z.map(($)=>{let Y=$.vendor?$.vendor:'<a href="/config#vendor-editor-card" style="color: #8b949e; text-decoration: underline;" title="Edit custom vendor list">Unknown</a>',L=`
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
        <td style="padding: 4px 12px; color: #b1bac4;">${L}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">${Y}</td>
        <td style="padding: 8px 12px; color: #8b949e;">${$.type}</td>
        <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${$.port}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">VLAN ${$.vlan}</td>
      </tr>
    `}).join("");let X=document.querySelector(`.mac-scroll-container-${CSS.escape(q)}`);if(X)X.scrollTop=J.scrollTop||0}function c(q){if(q===void 0||q===null)return"";if(q>=1000000000000)return(q/1000000000000).toFixed(1)+" TB";if(q>=1e9)return(q/1e9).toFixed(1)+" GB";if(q>=1e6)return(q/1e6).toFixed(1)+" MB";if(q>=1000)return(q/1000).toFixed(1)+" KB";return q+" B"}function D(q,J){if(!q)return J==="Bps"?"0 B/s":"0 bps";let K=J==="Bps"?q/8:q,Z=J==="Bps"?["B/s","KB/s","MB/s","GB/s"]:["bps","Kbps","Mbps","Gbps"],j=[1e9,1e6,1000];if(K>=j[0])return(K/j[0]).toFixed(1)+" "+Z[3];if(K>=j[1])return(K/j[1]).toFixed(1)+" "+Z[2];if(K>=j[2])return(K/j[2]).toFixed(1)+" "+Z[1];return K.toFixed(0)+" "+Z[0]}function mq(q){if(q)q.stopPropagation();B=B==="bps"?"Bps":"bps",t()}function Wq(q){if(!q)return"0";if(q>=1e6)return(q/1e6).toFixed(1)+"M";if(q>=1000)return(q/1000).toFixed(1)+"K";return q}function cq(q,J){if(!J||J.toLowerCase()!=="up")return"speed-down";if(!q)return"";let K=q.toLowerCase();if(K.includes("10g"))return"speed-10g";if(K.includes("2500")||K.includes("2.5g"))return"speed-2500m";if(K.includes("2000")||K.includes("2g"))return"speed-2000m";if(K.includes("1000")||K.includes("1g"))return"speed-1000m";if(K.includes("100"))return"speed-100m";if(K.includes("10"))return"speed-10m";return""}function Yq(q){if(!q)return"";return new Date(q*1000).toLocaleTimeString()}function lq(q){let J=(Q,U)=>{let k=jq[`mac-${Q}`];return`
      <th style="${`position: relative; padding: 8px 12px; cursor: pointer; color: #8b949e; user-select: none; font-weight: 600;${k?` width: ${k}px;`:""}`}" onclick="sortMac('${q.ip}', '${Q}')" data-col-id="mac-${Q}">
        ${U} <span class="sort-indicator-${q.ip}-${Q}"></span>
        <div class="resizer" onmousedown="handleResizeStart(event, 'mac-${Q}')" style="position: absolute; top: 0; right: 0; width: 6px; height: 100%; cursor: col-resize; user-select: none; z-index: 10;"></div>
      </th>
    `},K=(Q)=>{if(!Q)return"#57606a";let U=(Q.status||"").toLowerCase();if(U==="disable"||U==="disabled")return"#353c45";if(U!=="up")return"#57606a";let k=(Q.speed||"").toLowerCase();if(k.includes("10g"))return"#1f6feb";if(k.includes("2.5g")||k.includes("2500")||k.includes("2g")||k.includes("2000"))return"#58a6ff";if(k.includes("1g")||k.includes("1000"))return"#3fb950";if(k.includes("100")||k.includes("10"))return"#f08a24";return"#57606a"},Z=typeof Vq<"u"?parseInt(Vq):0,j=q.ports||[],X=j.map((Q)=>{let U=Q.port==="9"||(Q.speed||"").toLowerCase().includes("10g")||Q.port==="SFP"||Q.port==="9/SFP",k=K(Q),V=[];if(Q.vm_name)V.push(Q.vm_name);if(Q.interface&&Q.interface!==Q.port)V.push(Q.interface);let _=V.length>0?` (${V.join(" / ")})`:"",F=`Port ${Q.port}${_}: ${Q.status.toUpperCase()} ${Q.speed||""} (${Q.duplex||""}) - Click to view graph`,W=U?"M 4,5 h 3 v 3 h 10 v -3 h 3 v 14 h -16 z":"M 4,8 h 5 v -3 h 6 v 3 h 5 v 11 h -16 z";return`
      <div style="display: flex; flex-direction: column; align-items: center; gap: 2px; cursor: pointer;" 
           title="${F}"
           onclick="openGraph('${q.ip}','${Q.port}','${q.name}','Port ${Q.port}', ${Q.cum_tx||Q.tx_bytes||0}, ${Q.cum_rx||Q.rx_bytes||0})">
        <span style="font-size: 10px; color: #8b949e; font-weight: 700; margin-bottom: 2px;">${Q.port}</span>
        <svg viewBox="0 0 24 24" style="width: 18px; height: 18px; fill: ${k}; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.4)); transition: all 0.15s ease-in-out; transform: scale(1);"
             onmouseover="this.style.transform='scale(1.2)';"
             onmouseout="this.style.transform='scale(1)';">
          <path d="${W}" />
        </svg>
      </div>
    `}),$="";if(Z>0&&X.length>Z)for(let Q=0;Q<X.length;Q+=Z){let U=X.slice(Q,Q+Z);$+=`
        <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
          ${U.join("")}
        </div>
      `}else $=`
      <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
        ${X.join("")}
      </div>
    `;let Y=j.length>0?`
    <div style="background: rgba(13, 17, 23, 0.75); border: 1px solid #30363d; border-radius: 20px; padding: 10px 18px; display: inline-flex; flex-direction: column; gap: 10px; align-items: center; justify-content: center; box-shadow: inset 0 2px 4px rgba(0,0,0,0.4), 0 4px 10px rgba(0,0,0,0.3); margin: 4px 0;">
      ${$}
    </div>
  `:"",L=a.filter((Q)=>$q[Q]),N=L.map((Q)=>{let U=$q[Q],k=U.getLabel?U.getLabel():U.label,V=jq[Q],_=V?`width: ${V}px;`:U.thStyle||"";return`<th draggable="true"
                ondragstart="handleDragStart(event, '${Q}')"
                ondragover="handleDragOver(event)"
                ondragenter="handleDragEnter(event, '${Q}')"
                ondragleave="handleDragLeave(event)"
                ondragend="handleDragEnd(event)"
                ondrop="handleDrop(event, '${Q}')"
                data-col-id="${Q}"
                style="cursor: move; user-select: none; ${_}"
                class="${U.thClass||""}">
              ${k}
              <div class="resizer" onmousedown="handleResizeStart(event, '${Q}')"></div>
            </th>`}).join(""),G=(q.ports||[]).map((Q)=>{return`<tr>${L.map((k)=>{let V=$q[k],_=V.render(Q,q),F=V.tdClassFn?V.tdClassFn(Q):V.tdClass||"",W=V.tdTitleFn?V.tdTitleFn(Q):V.tdTitle||"",M=F?`class="${F}"`:"",Jq=W?`title="${W}"`:"";return`<td ${M} ${Jq}>${_}</td>`}).join("")}</tr>`}).join(""),z=q.error?`<div class="error-card">Error: ${q.error}</div>`:"",O=q.uptime?`<span><span class="label">Uptime:</span> <span class="value">${q.uptime}</span></span>`:"",P=q.mac?`<span><span class="label">MAC:</span> <span class="value">${q.mac}</span></span>`:"",A=q.firmware?`<span><span class="label">Firmware:</span> <span class="value">${q.firmware}</span></span>`:"",y="";if(q.dhcp_snooping){let Q=q.dhcp_snooping.enabled;y=`<span><span class="label">DHCP Snooping:</span> <span class="value" style="color: ${Q?"#56d364":"#8b949e"}">${Q?"\uD83D\uDFE2 Active":"⚪ Inactive"}</span></span>`}let b="";if(q.igmp){let Q=q.igmp.enabled;b=`<span><span class="label">IGMP Snooping:</span> <span class="value" style="color: ${Q?"#56d364":"#8b949e"}">${Q?"\uD83D\uDFE2 Active":"⚪ Inactive"}</span></span>`}let f="";if(q.jumbo_frame){let Q=q.jumbo_frame.enabled,U=q.jumbo_frame.size;f=`<span><span class="label">Jumbo Frame:</span> <span class="value" style="color: ${Q?"#56d364":"#8b949e"}">${Q?"\uD83D\uDFE2 Active ("+U+")":"⚪ Inactive"}</span></span>`}let E=x(q.ip),I=E.expanded,R=I?"block":"none",C=I?"rotate(90deg)":"rotate(0deg)",n=q.mac_table?q.mac_table.length:0,l=q.mac_timestamp?Yq(q.mac_timestamp):"Never",T=`
    <div class="mac-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="mac-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleMacTable('${q.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="mac-toggle-arrow-${q.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${C}; font-size: 10px;">&gt;</span>
          <span>\uD83D\uDCC1 MAC Address Table (<span class="mac-count-${q.ip}">${n} entries</span>)</span>
        </h4>
        <span class="mac-time-${q.ip}" style="font-size: 10px; color: #8b949e;">Last Scraped: ${l}</span>
      </div>
      
      <div class="mac-content-${q.ip}" style="display: ${R}; margin-top: 12px;">
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
  `,o=hq(q.ip).expanded,qq=o?"block":"none",d=o?"rotate(90deg)":"rotate(0deg)",g=q.igmp&&q.igmp.entries?q.igmp.entries:[],w=g.length,S='<tr><td colspan="3" style="padding: 12px; text-align: center; color: #8b949e;">No active multicast groups detected</td></tr>';if(g.length>0)S=g.map((Q)=>`
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; color: #58a6ff; font-weight: 600; font-family: 'SF Mono', Monaco, monospace;">${Q.ip}</td>
        <td style="padding: 8px 12px; color: #c9d1d9; font-weight: 600;">${Q.ports}</td>
        <td style="padding: 8px 12px; color: #8b949e;">VLAN ${Q.vlan}</td>
      </tr>
    `).join("");let s=`
    <div class="igmp-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="igmp-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleIgmpTable('${q.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="igmp-toggle-arrow-${q.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${d}; font-size: 10px;">&gt;</span>
          <span>\uD83D\uDCC1 IGMP Multicast Groups (<span class="igmp-count-${q.ip}">${w} entries</span>)</span>
        </h4>
      </div>
      
      <div class="igmp-content-${q.ip}" style="display: ${qq}; margin-top: 12px;">
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
              ${S}
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
      ${O}
      ${A}
      ${P}
      ${y}
      ${b}
      ${f}
    </div>
    <table class="port-table">
      <thead><tr>
        ${N}
      </tr></thead>
      <tbody>${G||`<tr><td colspan="${L.length}" style="padding:20px;text-align:center;color:#b1bac4">No data</td></tr>`}</tbody>
    </table>
    ${z}
    ${T}
    ${s}
    <div class="last-update">Last update: ${Yq(q.timestamp)}</div>
  </div>`}async function t(){if(Xq)return;let q=document.getElementById("live-dot"),J=document.getElementById("live-text");q.style.background="#d29922",J.textContent="Updating...";try{let Z=await(await fetch("/api/switches")).json(),j=document.getElementById("app");if(!Z||Z.length===0){j.innerHTML='<div style="text-align:center;padding:40px;color:#b1bac4">No nodes configured. <a href="/config" style="color:#58a6ff">Configure now</a></div>',j.className="",q.style.background="#f85149",J.textContent="No nodes";return}if(Z.forEach((X)=>{let $=document.querySelector(`.mac-scroll-container-${CSS.escape(X.ip)}`);if($){let Y=x(X.ip);if(Y.scrollTop=$.scrollTop,$.style.height)Y.height=$.style.height}}),_q(),j.className="switches",j.innerHTML=Z.map(lq).join(""),Z.forEach((X)=>{let $=x(X.ip);if($.data=X.mac_table||[],$.expanded)e(X.ip)}),document.getElementById("graph-overlay").classList.contains("open")&&h&&v){Lq(h,v);let X=Z.find(($)=>$.ip===h);if(X){let $=(X.ports||[]).find((Y)=>Y.port===v);if($){let Y=$.cum_tx||$.tx_bytes||0,L=$.cum_rx||$.rx_bytes||0;document.getElementById("graph-total-tx").innerHTML=`${Y.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${c(Y)})</span>`,document.getElementById("graph-total-rx").innerHTML=`${L.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${c(L)})</span>`}}}q.style.background="#3fb950",Cq()}catch(K){document.getElementById("app").innerHTML=`<div style="text-align:center;padding:40px;color:#f85149">Connection error: ${K.message}</div>`,q.style.background="#f85149",J.textContent="Error"}}function dq(q,J,K,Z,j=0,X=0){h=q,v=J,Hq=K,Rq=Z,document.getElementById("graph-title").textContent=Z+" Speed ("+K+")",document.getElementById("graph-overlay").classList.add("open"),document.getElementById("graph-total-tx").innerHTML=`${j.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${c(j)})</span>`,document.getElementById("graph-total-rx").innerHTML=`${X.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${c(X)})</span>`,iq(i)}function iq(q){if(i=q,document.querySelectorAll(".graph-tab-btn").forEach((J)=>J.classList.remove("active")),q==="live")document.getElementById("tab-live").classList.add("active"),document.getElementById("graph-sub").textContent=`Real-time speed at ${p}s polling intervals`;else if(q==="1h")document.getElementById("tab-1h").classList.add("active"),document.getElementById("graph-sub").textContent=`Last 1 hour speed history (at ${p}s polling intervals)`;else if(q==="24h")document.getElementById("tab-24h").classList.add("active"),document.getElementById("graph-sub").textContent="Last 24 hours speed history (15-minute averages)";Lq(h,v)}function Lq(q,J){let K=document.getElementById("graph-svg"),Z=document.getElementById("graph-box"),j=Z.classList.contains("maximized"),X=Math.max(500,Z.clientWidth-(j?64:48)),$=window.innerHeight*0.35;if(j)$=window.innerHeight-250;$=Math.max(240,$),K.setAttribute("viewBox",`0 0 ${X} ${$}`),K.innerHTML="";let Y={top:20,bottom:40,left:70,right:30},L=X-Y.left-Y.right,N=$-Y.top-Y.bottom;fetch(`/api/history?ip=${q}&port=${J}&range=${i}`).then((G)=>G.json()).then((G)=>{if(!G||!G.tx||G.tx.length<2){K.innerHTML=`<text x="${X/2}" y="${$/2}" text-anchor="middle" fill="#b1bac4" font-size="14">No history data available for this range yet</text>`,document.getElementById("graph-peak-tx").textContent="-",document.getElementById("graph-peak-rx").textContent="-",document.getElementById("graph-now-tx").textContent="-",document.getElementById("graph-now-rx").textContent="-";return}let{tx:z,rx:O,timestamps:P}=G,A=z.length,y=0;for(let _=0;_<A;_++){if(z[_]>y)y=z[_];if(O[_]>y)y=O[_]}if(y===0)y=1000;y*=1.15;let b=Math.max(...z),f=Math.max(...O),E=z[A-1],I=O[A-1];document.getElementById("graph-peak-tx").textContent=D(b,B),document.getElementById("graph-peak-rx").textContent=D(f,B),document.getElementById("graph-now-tx").textContent=D(E,B),document.getElementById("graph-now-rx").textContent=D(I,B);function R(_){return Y.left+_/(A-1)*L}function C(_){return Y.top+N-_/y*N}let n=4,l="";for(let _=0;_<=n;_++){let F=y/n*_,W=C(F);l+=`<line x1="${Y.left}" y1="${W}" x2="${X-Y.right}" y2="${W}" stroke="#21262d" stroke-width="1" stroke-dasharray="${_===0?"0":"4 4"}"/>`,l+=`<text x="${Y.left-10}" y="${W+4}" text-anchor="end" fill="#b1bac4" font-size="11" font-family="'SF Mono', Monaco, monospace">${D(F,B)}</text>`}let T=[],Fq=Math.min(5,A),o=Math.max(1,Math.floor((A-1)/(Fq-1)));for(let _=0;_<A;_+=o)T.push(_);if(T[T.length-1]!==A-1)T.push(A-1);let qq=T.map((_)=>{let F=R(_),W=new Date(P[_]*1000),M=W.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit",second:"2-digit"});if(i==="24h")M=W.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})+"<br>"+(W.getMonth()+1)+"/"+W.getDate();return`
          <line x1="${F}" y1="${Y.top}" x2="${F}" y2="${Y.top+N}" stroke="#21262d" stroke-dasharray="4 4" stroke-width="0.5"/>
          <text x="${F}" y="${$-12}" text-anchor="middle" fill="#b1bac4" font-size="10">${M}</text>
        `}).join(""),d="",g="",w=`M ${R(0)} ${C(0)} `,S=`M ${R(0)} ${C(0)} `;for(let _=0;_<A;_++){let F=R(_).toFixed(1),W=C(z[_]).toFixed(1),M=C(O[_]).toFixed(1);if(_===0)d+=`M ${F} ${W}`,g+=`M ${F} ${M}`,w=`M ${F} ${Y.top+N} L ${F} ${W}`,S=`M ${F} ${Y.top+N} L ${F} ${M}`;else d+=` L ${F} ${W}`,g+=` L ${F} ${M}`,w+=` L ${F} ${W}`,S+=` L ${F} ${M}`}w+=` L ${R(A-1).toFixed(1)} ${Y.top+N} Z`,S+=` L ${R(A-1).toFixed(1)} ${Y.top+N} Z`,K.innerHTML=`
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
        ${l}
        ${qq}

        <!-- Filled Areas (Gradients) -->
        <path d="${w}" fill="url(#tx-grad)" />
        <path d="${S}" fill="url(#rx-grad)" />

        <!-- Line curves -->
        <path d="${d}" fill="none" stroke="#58a6ff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="${g}" fill="none" stroke="#3fb950" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        
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
        <line id="hover-line" x1="0" y1="${Y.top}" x2="0" y2="${Y.top+N}" stroke="#b1bac4" stroke-width="1" stroke-dasharray="3 3" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-tx" r="5" fill="#58a6ff" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-rx" r="5" fill="#3fb950" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>

        <!-- Invisible Overlay rect for hover detection -->
        <rect id="hover-overlay-rect" x="${Y.left}" y="${Y.top}" width="${L}" height="${N}" fill="transparent" style="cursor: crosshair;"/>
      `;let s=document.getElementById("hover-overlay-rect"),Q=document.getElementById("hover-line"),U=document.getElementById("hover-dot-tx"),k=document.getElementById("hover-dot-rx"),V=document.getElementById("graph-tooltip");s.addEventListener("mousemove",(_)=>{let F=K.getBoundingClientRect(),Jq=(_.clientX-F.left)/F.width*X-Y.left,H=Math.round(Jq/L*(A-1));if(H<0)H=0;if(H>=A)H=A-1;let r=R(H),zq=C(z[H]),Oq=C(O[H]);Q.setAttribute("x1",r),Q.setAttribute("x2",r),Q.style.display="block",U.setAttribute("cx",r),U.setAttribute("cy",zq),U.style.display="block",k.setAttribute("cx",r),k.setAttribute("cy",Oq),k.style.display="block";let Kq=new Date(P[H]*1000),Nq=Kq.toLocaleTimeString();if(i==="24h")Nq=Kq.toLocaleDateString()+" "+Kq.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"});V.innerHTML=`
          <div class="time">${Nq}</div>
          <div class="tx">TX: ${D(z[H],B)}</div>
          <div class="rx">RX: ${D(O[H],B)}</div>
        `;let Uq=K.parentElement.getBoundingClientRect(),Bq=_.clientX-Uq.left+15,Mq=_.clientY-Uq.top-60;V.style.left=`${Bq}px`,V.style.top=`${Mq}px`,V.style.display="block"}),s.addEventListener("mouseleave",()=>{Q.style.display="none",U.style.display="none",k.style.display="none",V.style.display="none"})}).catch((G)=>{K.innerHTML=`<text x="${X/2}" y="${$/2}" text-anchor="middle" fill="#f85149" font-size="14">Error loading history: ${G.message}</text>`})}function pq(q,J,K){Iq=q,Tq=J,gq=K;let Z=document.getElementById("transceiver-overlay"),j=document.getElementById("transceiver-title"),X=document.getElementById("transceiver-sub"),$=document.getElementById("transceiver-content");j.textContent=`Port ${J} SFP+ Transceiver Status`,X.textContent=`DDMI Diagnostics & Telemetry (${K})`,$.innerHTML=`
    <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 40px; gap: 12px; color: #8b949e;">
      <div style="width: 24px; height: 24px; border: 2px solid transparent; border-top-color: #58a6ff; border-radius: 50%; animation: spin .8s linear infinite; box-sizing: border-box;"></div>
      <span>Fetching transceiver data from switch...</span>
    </div>
  `,Z.classList.add("open"),fetch(`/api/switches/${q}/transceiver`).then((Y)=>{if(!Y.ok)return Y.json().then((L)=>{throw Error(L.error||"Request failed")});return Y.json()}).then((Y)=>{nq(Y,$)}).catch((Y)=>{$.innerHTML=`
        <div class="error-card" style="margin: 20px 0; font-size: 13px;">
          <strong>No SFP module detected or telemetry unavailable</strong><br>
          <span style="opacity: 0.8; font-size: 11px;">${Y.message}</span>
        </div>
      `})}function nq(q,J){function K(E){if(!E)return 0;let I=E.toString().match(/[-+]?[0-9]*\.?[0-9]+/);return I?parseFloat(I[0]):0}let Z=K(q.temperature),j=K(q.voltage),X=K(q.current),$=-99,Y=-99;if(q.tx_power&&!q.tx_power.includes("-inf"))$=K(q.tx_power);if(q.rx_power&&!q.rx_power.includes("-inf"))Y=K(q.rx_power);let L=Math.min(100,Math.max(0,Z/80*100)),N=Math.min(100,Math.max(0,(j-2.8)/1*100)),G=Math.min(100,Math.max(0,X/40*100)),z=$===-99?0:Math.min(100,Math.max(0,($- -20)/25*100)),O=Y===-99?0:Math.min(100,Math.max(0,(Y- -30)/35*100)),P=Z>65||Z<5?"#f85149":Z>55?"#d29922":"#3fb950",A=j>3.5||j<3.1?"#f85149":"#3fb950",y=X>30?"#f85149":"#3fb950",b=$===-99?"#57606a":$<-12?"#d29922":"#3fb950",f=Y===-99?"#57606a":Y<-18?"#d29922":"#3fb950";J.innerHTML=`
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
            <span class="val" style="color: ${P};">${q.temperature||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${L}%; background: ${P};"></div>
          </div>
        </div>
        
        <!-- Voltage -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">⚡ Supply Voltage</span>
            <span class="val" style="color: ${A};">${q.voltage||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${N}%; background: ${A};"></div>
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
            <span class="val" style="color: ${b};">${q.tx_power||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${z}%; background: ${b};"></div>
          </div>
        </div>
        
        <!-- RX Power -->
        <div class="telemetry-row">
          <div class="header-row">
            <span class="label">\uD83D\uDCE5 Optical RX Power</span>
            <span class="val" style="color: ${f};">${q.rx_power||"N/A"}</span>
          </div>
          <div class="telemetry-bar-bg">
            <div class="telemetry-bar-fill" style="width: ${O}%; background: ${f};"></div>
          </div>
        </div>
        
        <div style="margin-top: 18px; padding: 10px; background: rgba(88,166,255,0.05); border: 1px solid rgba(88,166,255,0.15); border-radius: 6px; font-size: 11px; color: #8b949e; line-height: 1.4;">
          \uD83D\uDCA1 Telemetry values are updated directly from the switch's internal SFP micro-controller registers using DDMI.
        </div>
      </div>
    </div>
  `}function oq(){document.getElementById("confirm-overlay").classList.remove("open")}async function sq(){oq(),await fetch("/api/reset",{method:"POST"}),t()}function yq(q){document.body.dataset.fontSize=q,document.querySelectorAll(".fs-btn").forEach((J)=>J.classList.toggle("fs-active",J.dataset.size===q)),fetch("/api/settings",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({font_size:q})})}async function rq(){try{let J=await(await fetch("/api/settings")).json();if(J.font_size)yq(J.font_size)}catch(q){}}rq();t();Dq();window.addEventListener("resize",()=>{if(document.getElementById("graph-overlay").classList.contains("open")&&h&&v)Lq(h,v)});window.setFontSize=yq;window.doReset=sq;window.backupConfig=wq;window.manualRefreshMac=xq;window.openGraph=dq;window.openTransceiver=pq;window.sortMac=uq;window.toggleIgmpTable=vq;window.toggleMacTable=Sq;window.toggleSpeedUnit=mq;
