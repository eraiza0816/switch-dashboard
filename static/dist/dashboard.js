var Zq=window.__DATA__.refresh||30,_q=window.__DATA__.columns||["port","status","speed","packets","bytes","info","notes"],Fq=window.__DATA__.portWrap||0,v=null,u=null,t="live";var H="Bps",$q=!1,g=Zq,qq=null;function Oq(){g=Zq;let q=document.getElementById("live-text");if(q)if($q)q.textContent="Paused";else q.textContent=`Live (update in ${g}s)`}function Bq(){if(qq)clearInterval(qq);qq=setInterval(()=>{let q=document.getElementById("live-text"),J=document.getElementById("live-dot");if(!q||!J)return;if($q){q.textContent="Paused",J.style.background="#d29922";return}if(g--,g<=0)q.textContent="Updating...",J.style.background="#d29922",g=Zq,Aq();else q.textContent=`Live (update in ${g}s)`,J.style.background="#3fb950"},1000)}var Mq=["port","status","speed","packets","bytes","info","notes"],n=[],Vq={},Jq=[];function Hq(){let q=Jq&&Jq.length>0?Jq:null;if(!q)q=Mq;n=q.filter((J)=>_q.includes(J)),_q.forEach((J)=>{if(!n.includes(J))n.push(J)})}Hq();var Qq=Vq||{};function Rq(){Qq=Vq||{}}Rq();var Kq={port:{label:"Port",thStyle:"width:55px",tdClass:"port-num",tdTitleFn:(q)=>{let J=[];if(q.vm_name)J.push(q.vm_name);if(q.interface)J.push(q.interface);return J.length>0?J.join(" / "):""},render:(q,J)=>q.port},status:{label:"Status",thStyle:"width:145px",render:(q,J)=>{let Q=(q.status||"unknown").toLowerCase(),X="";if(J.dhcp_snooping&&J.dhcp_snooping.enabled){let _=J.dhcp_snooping.ports||{},Z=q.port.split("/")[0],$=_[q.port]||_[Z];if($==="Trusted")X='<span class="trust-badge trusted" title="DHCP Snooping: Trusted Port">Trusted</span>';else if($==="Untrusted")X='<span class="trust-badge untrusted" title="DHCP Snooping: Untrusted Port">Untrusted</span>'}let j=q.duplex||"Auto";if(j.toLowerCase()==="full")j="Full Duplex";else if(j.toLowerCase()==="half")j="Half Duplex";return`
        <div style="display: flex; flex-direction: column; gap: 4px; align-items: flex-start;">
          <div style="display: flex; align-items: center; gap: 6px;">
            <span class="status-badge ${Q}" onclick="openGraph('${J.ip}','${q.port}','${J.name}','Port ${q.port}', ${q.cum_tx||q.tx_bytes||0}, ${q.cum_rx||q.rx_bytes||0})"><span class="status-dot ${Q}"></span>${Q.toUpperCase()}</span>
            ${X}
          </div>
          <div style="font-size: 10px; color: #8b949e; padding-left: 4px; font-weight: 500;">${j}</div>
        </div>
      `}},speed:{label:"Speed",thStyle:"width:55px",tdClassFn:(q)=>bq(q.speed,q.status),tdTitleFn:(q)=>q.duplex||"",render:(q,J)=>{let Q=q.port==="9"||(q.speed||"").toLowerCase().includes("10g")||q.port==="SFP"||q.port==="9/SFP",X=q.speed||"Auto";if(Q)return`<span class="sfp-speed-link" onclick="openTransceiver('${J.ip}','${q.port}','${J.name}')" title="Click to view SFP+ Transceiver Diagnostics">${X} <svg viewBox="0 0 24 24" style="width: 10px; height: 10px; fill: currentColor; display: inline-block;"><path d="M12 9a3 3 0 1 0 0 6 3 3 0 0 0 0-6zm0-7C7 2 2.73 5.11 1 9.5 2.73 13.89 7 17 12 17s9.27-3.11 11-7.5C21.27 5.11 17 2 12 2zm0 13a5.5 5.5 0 1 1 0-11 5.5 5.5 0 0 1 0 11z"/></svg></span>`;else return X}},packets:{label:'Traffic<br><span style="font-weight:400;color:#8b949e">(packets)</span>',thStyle:"width:90px",tdClass:"mono",tdTitle:"TX / RX packets",render:(q,J)=>{let Q=q.tx_packets||0,X=q.rx_packets||0;return`TX ${kq(Q)}<br>RX ${kq(X)}`}},bytes:{label:'Traffic<br><span style="font-weight:400;color:#8b949e">(bytes)</span>',thStyle:"width:90px",tdClass:"mono",tdTitle:"TX / RX bytes formatted",render:(q,J)=>{let Q=q.cum_tx||q.tx_bytes||0,X=q.cum_rx||q.rx_bytes||0;return`TX ${s(Q)}<br>RX ${s(X)}`}},info:{getLabel:()=>`<span id="speed-unit-toggle" class="unit-toggle" onclick="toggleSpeedUnit(event)" title="Toggle speed unit">${H==="bps"?"BPS":"B/s"}</span>`,tdClass:"col-info mono",render:(q,J)=>{let Q=q.speed_tx_bps||0,X=q.speed_rx_bps||0;return`TX ${D(Q,H)}<br>RX ${D(X,H)}`}},notes:{label:"Notes",tdClass:"col-note",render:(q,J)=>{return`<input class="note-input" type="text" placeholder="-" value="${q.note||""}" data-key="${J.ip}:${q.port}" onfocus="pausePolling()" onblur="saveNote(this)">`}}};var I={};try{let q=localStorage.getItem("macTableStates");if(q)I=JSON.parse(q)}catch(q){console.error("Failed to load macTableStates from localStorage",q)}function Cq(){let q={};for(let J in I)if(I.hasOwnProperty(J)){let Q=I[J];q[J]={expanded:Q.expanded,filters:Q.filters||{mac:"",vendor:"",type:"",port:"",vlan:""},sortBy:Q.sortBy||"port",sortAsc:Q.sortAsc!==void 0?Q.sortAsc:!0,scrollTop:Q.scrollTop||0,height:Q.height||"450px"}}localStorage.setItem("macTableStates",JSON.stringify(q))}function o(q){if(!I[q])I[q]={};let J=I[q];if(J.expanded===void 0)J.expanded=!1;if(!J.filters)J.filters={mac:"",host:"",vendor:"",type:"",port:"",vlan:""};if(J.sortBy===void 0)J.sortBy="port";if(J.sortAsc===void 0)J.sortAsc=!0;if(!J.data)J.data=[];if(J.scrollTop===void 0)J.scrollTop=0;if(J.height===void 0)J.height="450px";return J}function Lq(q){if(!q)return 0;if(!isNaN(q))return parseInt(q);let J=q.match(/\d+/);if(J)return parseInt(J[0]);return q.toString().toLowerCase()}function Uq(q){if(!q)return 0;let J=parseInt(q);return isNaN(J)?q.toString().toLowerCase():J}function Dq(q){let J={};try{let Q=localStorage.getItem("igmp_states");if(Q)J=JSON.parse(Q)}catch(Q){console.error("Failed to load igmp_states from localStorage",Q)}if(!J[q])J[q]={expanded:!1};return J[q]}function Pq(q){let J=o(q),Q=document.querySelector(`.mac-tbody-${CSS.escape(q)}`);if(!Q)return;["mac","host","vendor","type","port","vlan"].forEach((Z)=>{let $=document.querySelector(`.sort-indicator-${CSS.escape(q)}-${Z}`);if($)if(J.sortBy===Z)$.textContent=J.sortAsc?" ▴":" ▾",$.style.color="#58a6ff";else $.textContent=""});let X=J.data,j=J.filters||{mac:"",host:"",vendor:"",type:"",port:"",vlan:""};if(j.mac||j.host||j.vendor||j.type||j.port||j.vlan)X=X.filter((Z)=>{let $=!j.mac||(Z.mac||"").toLowerCase().includes(j.mac),L=!j.host||(Z.host||"").toLowerCase().includes(j.host),N=!j.vendor||(Z.vendor||"").toLowerCase().includes(j.vendor),W=!j.type||(Z.type||"").toLowerCase().includes(j.type),z=!j.port||(Z.port||"").toString().toLowerCase().includes(j.port),O=!j.vlan||(Z.vlan||"").toString().toLowerCase().includes(j.vlan);return $&&L&&N&&W&&z&&O});if(X.sort((Z,$)=>{let L=Z[J.sortBy]||"",N=$[J.sortBy]||"";if(J.sortBy==="port")L=Lq(L),N=Lq(N);else if(J.sortBy==="vlan")L=Uq(L),N=Uq(N);else L=L.toString().toLowerCase(),N=N.toString().toLowerCase();if(L<N)return J.sortAsc?-1:1;if(L>N)return J.sortAsc?1:-1;return 0}),X.length===0){Q.innerHTML='<tr><td colspan="6" style="padding: 12px; text-align: center; color: #8b949e;">No MAC entries found</td></tr>';let Z=document.querySelector(`.mac-scroll-container-${CSS.escape(q)}`);if(Z)Z.scrollTop=J.scrollTop||0;return}Q.innerHTML=X.map((Z)=>{let $=Z.vendor?Z.vendor:'<a href="/config#vendor-editor-card" style="color: #8b949e; text-decoration: underline;" title="Edit custom vendor list">Unknown</a>',L=`
      <input type="text" value="${Z.host||""}" 
             placeholder="Enter nickname..." 
             onfocus="pausePolling()" 
             onblur="saveHost(this, '${Z.mac}')" 
             style="width: 100%; max-width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" 
             onfocus="this.style.borderColor='#58a6ff'" 
             onblur="this.style.borderColor='#30363d'">
    `;return`
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; font-family: 'SF Mono', Monaco, monospace; color: #c9d1d9;">${Z.mac}</td>
        <td style="padding: 4px 12px; color: #b1bac4;">${L}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">${$}</td>
        <td style="padding: 8px 12px; color: #8b949e;">${Z.type}</td>
        <td style="padding: 8px 12px; font-weight: 600; color: #58a6ff;">${Z.port}</td>
        <td style="padding: 8px 12px; color: #b1bac4;">VLAN ${Z.vlan}</td>
      </tr>
    `}).join("");let _=document.querySelector(`.mac-scroll-container-${CSS.escape(q)}`);if(_)_.scrollTop=J.scrollTop||0}function s(q){if(q===void 0||q===null)return"";if(q>=1000000000000)return(q/1000000000000).toFixed(1)+" TB";if(q>=1e9)return(q/1e9).toFixed(1)+" GB";if(q>=1e6)return(q/1e6).toFixed(1)+" MB";if(q>=1000)return(q/1000).toFixed(1)+" KB";return q+" B"}function D(q,J){if(!q)return J==="Bps"?"0 B/s":"0 bps";let Q=J==="Bps"?q/8:q,X=J==="Bps"?["B/s","KB/s","MB/s","GB/s"]:["bps","Kbps","Mbps","Gbps"],j=[1e9,1e6,1000];if(Q>=j[0])return(Q/j[0]).toFixed(1)+" "+X[3];if(Q>=j[1])return(Q/j[1]).toFixed(1)+" "+X[2];if(Q>=j[2])return(Q/j[2]).toFixed(1)+" "+X[1];return Q.toFixed(0)+" "+X[0]}function kq(q){if(!q)return"0";if(q>=1e6)return(q/1e6).toFixed(1)+"M";if(q>=1000)return(q/1000).toFixed(1)+"K";return q}function bq(q,J){if(!J||J.toLowerCase()!=="up")return"speed-down";if(!q)return"";let Q=q.toLowerCase();if(Q.includes("10g"))return"speed-10g";if(Q.includes("2500")||Q.includes("2.5g"))return"speed-2500m";if(Q.includes("2000")||Q.includes("2g"))return"speed-2000m";if(Q.includes("1000")||Q.includes("1g"))return"speed-1000m";if(Q.includes("100"))return"speed-100m";if(Q.includes("10"))return"speed-10m";return""}function Nq(q){if(!q)return"";return new Date(q*1000).toLocaleTimeString()}function fq(q){let J=(K,U)=>{let k=Qq[`mac-${K}`];return`
      <th style="${`position: relative; padding: 8px 12px; cursor: pointer; color: #8b949e; user-select: none; font-weight: 600;${k?` width: ${k}px;`:""}`}" onclick="sortMac('${q.ip}', '${K}')" data-col-id="mac-${K}">
        ${U} <span class="sort-indicator-${q.ip}-${K}"></span>
        <div class="resizer" onmousedown="handleResizeStart(event, 'mac-${K}')" style="position: absolute; top: 0; right: 0; width: 6px; height: 100%; cursor: col-resize; user-select: none; z-index: 10;"></div>
      </th>
    `},Q=(K)=>{if(!K)return"#57606a";let U=(K.status||"").toLowerCase();if(U==="disable"||U==="disabled")return"#353c45";if(U!=="up")return"#57606a";let k=(K.speed||"").toLowerCase();if(k.includes("10g"))return"#1f6feb";if(k.includes("2.5g")||k.includes("2500")||k.includes("2g")||k.includes("2000"))return"#58a6ff";if(k.includes("1g")||k.includes("1000"))return"#3fb950";if(k.includes("100")||k.includes("10"))return"#f08a24";return"#57606a"},X=typeof Fq<"u"?parseInt(Fq):0,j=q.ports||[],_=j.map((K)=>{let U=K.port==="9"||(K.speed||"").toLowerCase().includes("10g")||K.port==="SFP"||K.port==="9/SFP",k=Q(K),V=[];if(K.vm_name)V.push(K.vm_name);if(K.interface&&K.interface!==K.port)V.push(K.interface);let Y=V.length>0?` (${V.join(" / ")})`:"",F=`Port ${K.port}${Y}: ${K.status.toUpperCase()} ${K.speed||""} (${K.duplex||""}) - Click to view graph`,A=U?"M 4,5 h 3 v 3 h 10 v -3 h 3 v 14 h -16 z":"M 4,8 h 5 v -3 h 6 v 3 h 5 v 11 h -16 z";return`
      <div style="display: flex; flex-direction: column; align-items: center; gap: 2px; cursor: pointer;" 
           title="${F}"
           onclick="openGraph('${q.ip}','${K.port}','${q.name}','Port ${K.port}', ${K.cum_tx||K.tx_bytes||0}, ${K.cum_rx||K.rx_bytes||0})">
        <span style="font-size: 10px; color: #8b949e; font-weight: 700; margin-bottom: 2px;">${K.port}</span>
        <svg viewBox="0 0 24 24" style="width: 18px; height: 18px; fill: ${k}; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.4)); transition: all 0.15s ease-in-out; transform: scale(1);"
             onmouseover="this.style.transform='scale(1.2)';"
             onmouseout="this.style.transform='scale(1)';">
          <path d="${A}" />
        </svg>
      </div>
    `}),Z="";if(X>0&&_.length>X)for(let K=0;K<_.length;K+=X){let U=_.slice(K,K+X);Z+=`
        <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
          ${U.join("")}
        </div>
      `}else Z=`
      <div style="display: flex; gap: 10px; align-items: center; justify-content: center;">
        ${_.join("")}
      </div>
    `;let $=j.length>0?`
    <div style="background: rgba(13, 17, 23, 0.75); border: 1px solid #30363d; border-radius: 20px; padding: 10px 18px; display: inline-flex; flex-direction: column; gap: 10px; align-items: center; justify-content: center; box-shadow: inset 0 2px 4px rgba(0,0,0,0.4), 0 4px 10px rgba(0,0,0,0.3); margin: 4px 0;">
      ${Z}
    </div>
  `:"",L=n.filter((K)=>Kq[K]),N=L.map((K)=>{let U=Kq[K],k=U.getLabel?U.getLabel():U.label,V=Qq[K],Y=V?`width: ${V}px;`:U.thStyle||"";return`<th draggable="true"
                ondragstart="handleDragStart(event, '${K}')"
                ondragover="handleDragOver(event)"
                ondragenter="handleDragEnter(event, '${K}')"
                ondragleave="handleDragLeave(event)"
                ondragend="handleDragEnd(event)"
                ondrop="handleDrop(event, '${K}')"
                data-col-id="${K}"
                style="cursor: move; user-select: none; ${Y}"
                class="${U.thClass||""}">
              ${k}
              <div class="resizer" onmousedown="handleResizeStart(event, '${K}')"></div>
            </th>`}).join(""),W=(q.ports||[]).map((K)=>{return`<tr>${L.map((k)=>{let V=Kq[k],Y=V.render(K,q),F=V.tdClassFn?V.tdClassFn(K):V.tdClass||"",A=V.tdTitleFn?V.tdTitleFn(K):V.tdTitle||"",B=F?`class="${F}"`:"",a=A?`title="${A}"`:"";return`<td ${B} ${a}>${Y}</td>`}).join("")}</tr>`}).join(""),z=q.error?`<div class="error-card">Error: ${q.error}</div>`:"",O=q.uptime?`<span><span class="label">Uptime:</span> <span class="value">${q.uptime}</span></span>`:"",x=q.mac?`<span><span class="label">MAC:</span> <span class="value">${q.mac}</span></span>`:"",E=q.firmware?`<span><span class="label">Firmware:</span> <span class="value">${q.firmware}</span></span>`:"",y="";if(q.dhcp_snooping){let K=q.dhcp_snooping.enabled;y=`<span><span class="label">DHCP Snooping:</span> <span class="value" style="color: ${K?"#56d364":"#8b949e"}">${K?"\uD83D\uDFE2 Active":"⚪ Inactive"}</span></span>`}let m="";if(q.igmp){let K=q.igmp.enabled;m=`<span><span class="label">IGMP Snooping:</span> <span class="value" style="color: ${K?"#56d364":"#8b949e"}">${K?"\uD83D\uDFE2 Active":"⚪ Inactive"}</span></span>`}let w="";if(q.jumbo_frame){let K=q.jumbo_frame.enabled,U=q.jumbo_frame.size;w=`<span><span class="label">Jumbo Frame:</span> <span class="value" style="color: ${K?"#56d364":"#8b949e"}">${K?"\uD83D\uDFE2 Active ("+U+")":"⚪ Inactive"}</span></span>`}let G=o(q.ip),c=G.expanded,R=c?"block":"none",C=c?"rotate(90deg)":"rotate(0deg)",l=q.mac_table?q.mac_table.length:0,S=q.mac_timestamp?Nq(q.mac_timestamp):"Never",P=`
    <div class="mac-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="mac-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleMacTable('${q.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="mac-toggle-arrow-${q.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${C}; font-size: 10px;">&gt;</span>
          <span>\uD83D\uDCC1 MAC Address Table (<span class="mac-count-${q.ip}">${l} entries</span>)</span>
        </h4>
        <span class="mac-time-${q.ip}" style="font-size: 10px; color: #8b949e;">Last Scraped: ${S}</span>
      </div>
      
      <div class="mac-content-${q.ip}" style="display: ${R}; margin-top: 12px;">
        <div style="display: flex; justify-content: flex-end; margin-bottom: 10px;">
          <button class="mac-refresh-btn-${q.ip}" onclick="manualRefreshMac('${q.ip}')" style="background: #21262d; border: 1px solid #30363d; border-radius: 6px; padding: 6px 12px; color: #c9d1d9; font-size: 11px; font-weight: 600; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; gap: 6px;" onmouseover="this.style.background='#30363d'; this.style.borderColor='#8b949e';" onmouseout="this.style.background='#21262d'; this.style.borderColor='#30363d';">
            <span class="mac-spinner-${q.ip}" style="display: none; width: 12px; height: 12px; border: 2px solid transparent; border-top-color: #58a6ff; border-radius: 50%; animation: spin .8s linear infinite; box-sizing: border-box;"></span>
            <span>Refresh MACs</span>
          </button>
        </div>
        <div class="mac-scroll-container-${q.ip}" onscroll="saveMacScroll('${q.ip}', this)" onmouseup="saveMacHeight('${q.ip}', this)" ontouchend="saveMacHeight('${q.ip}', this)" style="height: ${G.height||"450px"}; min-height: 150px; max-height: 1200px; resize: vertical; overflow-y: auto; border: 1px solid #30363d; border-radius: 8px; background: #0d1117; box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);">
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
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-mac-${q.ip}" placeholder="Filter MAC..." value="${G.filters?G.filters.mac:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-host-${q.ip}" placeholder="Filter Host..." value="${G.filters?G.filters.host:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-vendor-${q.ip}" placeholder="Filter Vendor..." value="${G.filters?G.filters.vendor:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-type-${q.ip}" placeholder="Filter Type..." value="${G.filters?G.filters.type:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-port-${q.ip}" placeholder="Port..." value="${G.filters?G.filters.port:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
                <th style="padding: 4px 6px;"><input type="text" class="mac-filter-vlan-${q.ip}" placeholder="VLAN..." value="${G.filters?G.filters.vlan:""}" onfocus="focusMacSearch()" onblur="blurMacSearch()" oninput="filterMacTable('${q.ip}')" style="width: 100%; background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 4px 8px; color: #e1e4e8; font-size: 11px; outline: none; transition: border-color 0.2s;" onfocus="this.style.borderColor='#58a6ff'" onblur="this.style.borderColor='#30363d'"></th>
              </tr>
            </thead>
            <tbody class="mac-tbody-${q.ip}">
              <!-- Populate via Javascript -->
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `,d=Dq(q.ip).expanded,r=d?"block":"none",h=d?"rotate(90deg)":"rotate(0deg)",b=q.igmp&&q.igmp.entries?q.igmp.entries:[],T=b.length,f='<tr><td colspan="3" style="padding: 12px; text-align: center; color: #8b949e;">No active multicast groups detected</td></tr>';if(b.length>0)f=b.map((K)=>`
      <tr style="border-bottom: 1px solid #21262d; transition: background 0.15s;" onmouseover="this.style.background='#1c2128'" onmouseout="this.style.background='transparent'">
        <td style="padding: 8px 12px; color: #58a6ff; font-weight: 600; font-family: 'SF Mono', Monaco, monospace;">${K.ip}</td>
        <td style="padding: 8px 12px; color: #c9d1d9; font-weight: 600;">${K.ports}</td>
        <td style="padding: 8px 12px; color: #8b949e;">VLAN ${K.vlan}</td>
      </tr>
    `).join("");let i=`
    <div class="igmp-section" style="border-top: 1px solid #30363d; padding: 12px 20px; background: rgba(22, 27, 34, 0.2);">
      <div class="igmp-header" style="display: flex; justify-content: space-between; align-items: center; cursor: pointer; user-select: none;" onclick="toggleIgmpTable('${q.ip}')">
        <h4 style="font-size: 12px; font-weight: 600; color: #8b949e; display: flex; align-items: center; gap: 6px; margin: 0;">
          <span class="igmp-toggle-arrow-${q.ip}" style="display: inline-block; transition: transform 0.2s; transform: ${h}; font-size: 10px;">&gt;</span>
          <span>\uD83D\uDCC1 IGMP Multicast Groups (<span class="igmp-count-${q.ip}">${T} entries</span>)</span>
        </h4>
      </div>
      
      <div class="igmp-content-${q.ip}" style="display: ${r}; margin-top: 12px;">
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
              ${f}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `;return`<div class="switch-card">
    <div class="switch-header">
      <div style="display: flex; align-items: center; gap: 12px;">
        <img src="/api/switches/${q.ip}/image" alt="Switch" style="height: 24px; width: auto; object-fit: contain; filter: drop-shadow(0 2px 4px rgba(0,0,0,0.5));">
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
      ${$}
      
      <div class="meta">${q.model||""}</div>
    </div>
    <div class="device-bar">
      ${O}
      ${E}
      ${x}
      ${y}
      ${m}
      ${w}
    </div>
    <table class="port-table">
      <thead><tr>
        ${N}
      </tr></thead>
      <tbody>${W||`<tr><td colspan="${L.length}" style="padding:20px;text-align:center;color:#b1bac4">No data</td></tr>`}</tbody>
    </table>
    ${z}
    ${P}
    ${i}
    <div class="last-update">Last update: ${Nq(q.timestamp)}</div>
  </div>`}async function Aq(){if($q)return;let q=document.getElementById("live-dot"),J=document.getElementById("live-text");q.style.background="#d29922",J.textContent="Updating...";try{let X=await(await fetch("/api/switches")).json(),j=document.getElementById("app");if(!X||X.length===0){j.innerHTML='<div style="text-align:center;padding:40px;color:#b1bac4">No nodes configured. <a href="/config" style="color:#58a6ff">Configure now</a></div>',j.className="",q.style.background="#f85149",J.textContent="No nodes";return}if(X.forEach((_)=>{let Z=document.querySelector(`.mac-scroll-container-${CSS.escape(_.ip)}`);if(Z){let $=o(_.ip);if($.scrollTop=Z.scrollTop,Z.style.height)$.height=Z.style.height}}),Cq(),j.className="switches",j.innerHTML=X.map(fq).join(""),X.forEach((_)=>{let Z=o(_.ip);if(Z.data=_.mac_table||[],Z.expanded)Pq(_.ip)}),document.getElementById("graph-overlay").classList.contains("open")&&v&&u){Eq(v,u);let _=X.find((Z)=>Z.ip===v);if(_){let Z=(_.ports||[]).find(($)=>$.port===u);if(Z){let $=Z.cum_tx||Z.tx_bytes||0,L=Z.cum_rx||Z.rx_bytes||0;document.getElementById("graph-total-tx").innerHTML=`${$.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${s($)})</span>`,document.getElementById("graph-total-rx").innerHTML=`${L.toLocaleString()} bytes <span style="font-weight:400;color:#8b949e">(${s(L)})</span>`}}}q.style.background="#3fb950",Oq()}catch(Q){document.getElementById("app").innerHTML=`<div style="text-align:center;padding:40px;color:#f85149">Connection error: ${Q.message}</div>`,q.style.background="#f85149",J.textContent="Error"}}function Eq(q,J){let Q=document.getElementById("graph-svg"),X=document.getElementById("graph-box"),j=X.classList.contains("maximized"),_=Math.max(500,X.clientWidth-(j?64:48)),Z=window.innerHeight*0.35;if(j)Z=window.innerHeight-250;Z=Math.max(240,Z),Q.setAttribute("viewBox",`0 0 ${_} ${Z}`),Q.innerHTML="";let $={top:20,bottom:40,left:70,right:30},L=_-$.left-$.right,N=Z-$.top-$.bottom;fetch(`/api/history?ip=${q}&port=${J}&range=${t}`).then((W)=>W.json()).then((W)=>{if(!W||!W.tx||W.tx.length<2){Q.innerHTML=`<text x="${_/2}" y="${Z/2}" text-anchor="middle" fill="#b1bac4" font-size="14">No history data available for this range yet</text>`,document.getElementById("graph-peak-tx").textContent="-",document.getElementById("graph-peak-rx").textContent="-",document.getElementById("graph-now-tx").textContent="-",document.getElementById("graph-now-rx").textContent="-";return}let{tx:z,rx:O,timestamps:x}=W,E=z.length,y=0;for(let Y=0;Y<E;Y++){if(z[Y]>y)y=z[Y];if(O[Y]>y)y=O[Y]}if(y===0)y=1000;y*=1.15;let m=Math.max(...z),w=Math.max(...O),G=z[E-1],c=O[E-1];document.getElementById("graph-peak-tx").textContent=D(m,H),document.getElementById("graph-peak-rx").textContent=D(w,H),document.getElementById("graph-now-tx").textContent=D(G,H),document.getElementById("graph-now-rx").textContent=D(c,H);function R(Y){return $.left+Y/(E-1)*L}function C(Y){return $.top+N-Y/y*N}let l=4,S="";for(let Y=0;Y<=l;Y++){let F=y/l*Y,A=C(F);S+=`<line x1="${$.left}" y1="${A}" x2="${_-$.right}" y2="${A}" stroke="#21262d" stroke-width="1" stroke-dasharray="${Y===0?"0":"4 4"}"/>`,S+=`<text x="${$.left-10}" y="${A+4}" text-anchor="end" fill="#b1bac4" font-size="11" font-family="'SF Mono', Monaco, monospace">${D(F,H)}</text>`}let P=[],jq=Math.min(5,E),d=Math.max(1,Math.floor((E-1)/(jq-1)));for(let Y=0;Y<E;Y+=d)P.push(Y);if(P[P.length-1]!==E-1)P.push(E-1);let r=P.map((Y)=>{let F=R(Y),A=new Date(x[Y]*1000),B=A.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit",second:"2-digit"});if(t==="24h")B=A.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})+"<br>"+(A.getMonth()+1)+"/"+A.getDate();return`
          <line x1="${F}" y1="${$.top}" x2="${F}" y2="${$.top+N}" stroke="#21262d" stroke-dasharray="4 4" stroke-width="0.5"/>
          <text x="${F}" y="${Z-12}" text-anchor="middle" fill="#b1bac4" font-size="10">${B}</text>
        `}).join(""),h="",b="",T=`M ${R(0)} ${C(0)} `,f=`M ${R(0)} ${C(0)} `;for(let Y=0;Y<E;Y++){let F=R(Y).toFixed(1),A=C(z[Y]).toFixed(1),B=C(O[Y]).toFixed(1);if(Y===0)h+=`M ${F} ${A}`,b+=`M ${F} ${B}`,T=`M ${F} ${$.top+N} L ${F} ${A}`,f=`M ${F} ${$.top+N} L ${F} ${B}`;else h+=` L ${F} ${A}`,b+=` L ${F} ${B}`,T+=` L ${F} ${A}`,f+=` L ${F} ${B}`}T+=` L ${R(E-1).toFixed(1)} ${$.top+N} Z`,f+=` L ${R(E-1).toFixed(1)} ${$.top+N} Z`,Q.innerHTML=`
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
        ${S}
        ${r}

        <!-- Filled Areas (Gradients) -->
        <path d="${T}" fill="url(#tx-grad)" />
        <path d="${f}" fill="url(#rx-grad)" />

        <!-- Line curves -->
        <path d="${h}" fill="none" stroke="#58a6ff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="${b}" fill="none" stroke="#3fb950" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        
        <!-- Legend labels -->
        <g transform="translate(${_-$.right-90}, ${$.top+5})">
          <rect x="0" y="0" width="12" height="12" fill="#58a6ff" rx="3"/>
          <text x="18" y="10" fill="#b1bac4" font-size="11" font-weight="600">TX Speed</text>
        </g>
        <g transform="translate(${_-$.right-90}, ${$.top+22})">
          <rect x="0" y="0" width="12" height="12" fill="#3fb950" rx="3"/>
          <text x="18" y="10" fill="#b1bac4" font-size="11" font-weight="600">RX Speed</text>
        </g>

        <!-- Hover Guideline / Interactive Crosshair components -->
        <line id="hover-line" x1="0" y1="${$.top}" x2="0" y2="${$.top+N}" stroke="#b1bac4" stroke-width="1" stroke-dasharray="3 3" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-tx" r="5" fill="#58a6ff" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>
        <circle id="hover-dot-rx" r="5" fill="#3fb950" stroke="#161b22" stroke-width="2" style="display:none; pointer-events:none;"/>

        <!-- Invisible Overlay rect for hover detection -->
        <rect id="hover-overlay-rect" x="${$.left}" y="${$.top}" width="${L}" height="${N}" fill="transparent" style="cursor: crosshair;"/>
      `;let i=document.getElementById("hover-overlay-rect"),K=document.getElementById("hover-line"),U=document.getElementById("hover-dot-tx"),k=document.getElementById("hover-dot-rx"),V=document.getElementById("graph-tooltip");i.addEventListener("mousemove",(Y)=>{let F=Q.getBoundingClientRect(),a=(Y.clientX-F.left)/F.width*_-$.left,M=Math.round(a/L*(E-1));if(M<0)M=0;if(M>=E)M=E-1;let p=R(M),Gq=C(z[M]),Wq=C(O[M]);K.setAttribute("x1",p),K.setAttribute("x2",p),K.style.display="block",U.setAttribute("cx",p),U.setAttribute("cy",Gq),U.style.display="block",k.setAttribute("cx",p),k.setAttribute("cy",Wq),k.style.display="block";let e=new Date(x[M]*1000),Yq=e.toLocaleTimeString();if(t==="24h")Yq=e.toLocaleDateString()+" "+e.toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"});V.innerHTML=`
          <div class="time">${Yq}</div>
          <div class="tx">TX: ${D(z[M],H)}</div>
          <div class="rx">RX: ${D(O[M],H)}</div>
        `;let Xq=Q.parentElement.getBoundingClientRect(),zq=Y.clientX-Xq.left+15,yq=Y.clientY-Xq.top-60;V.style.left=`${zq}px`,V.style.top=`${yq}px`,V.style.display="block"}),i.addEventListener("mouseleave",()=>{K.style.display="none",U.style.display="none",k.style.display="none",V.style.display="none"})}).catch((W)=>{Q.innerHTML=`<text x="${_/2}" y="${Z/2}" text-anchor="middle" fill="#f85149" font-size="14">Error loading history: ${W.message}</text>`})}function Iq(q){document.body.dataset.fontSize=q,document.querySelectorAll(".fs-btn").forEach((J)=>J.classList.toggle("fs-active",J.dataset.size===q)),fetch("/api/settings",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({font_size:q})})}async function Tq(){try{let J=await(await fetch("/api/settings")).json();if(J.font_size)Iq(J.font_size)}catch(q){}}Tq();Aq();Bq();window.addEventListener("resize",()=>{if(document.getElementById("graph-overlay").classList.contains("open")&&v&&u)Eq(v,u)});
