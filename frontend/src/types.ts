interface Window {
  __DATA__: {
    refresh: number;
    columns: string[];
    portWrap: number;
  };
  setFontSize: (size: string) => void;
  doReset: () => void;
  backupConfig: (ip: string, el: HTMLElement) => void;
  manualRefreshMac: (ip: string) => void;
  openGraph: (ip: string, port: string, name: string, label: string, cumTX: number, cumRX: number) => void;
  openTransceiver: (ip: string, port: string, name: string) => void;
  closeTransceiver: () => void;
  sortMac: (ip: string, col: string) => void;
  toggleIgmpTable: (ip: string) => void;
  toggleMacTable: (ip: string) => void;
  toggleSpeedUnit: (e: Event) => void;
  pausePolling: () => void;
  saveNote: (input: HTMLInputElement) => void;
  saveHost: (input: HTMLInputElement, mac: string) => void;
  focusMacSearch: () => void;
  blurMacSearch: () => void;
  saveMacHeight: (ip: string, el: HTMLElement) => void;
  handleResizeStart: (e: MouseEvent, colId: string) => void;
  handleDragStart: (e: DragEvent, colId: string) => void;
  handleDragEnter: (e: DragEvent, colId: string) => void;
  handleDragOver: (e: DragEvent) => void;
  handleDragLeave: (e: DragEvent) => void;
  handleDragEnd: (e: DragEvent) => void;
  handleDrop: (e: DragEvent, colId: string) => void;
  filterMacTable: (ip: string) => void;
  saveMacScroll: (ip: string, el: HTMLElement) => void;
  bulkRename: () => void;
  switchConsole: (ip: string) => void;
}

interface SnoopingStatus {
  enabled: boolean;
  ports: Record<string, string>;
}

interface IGMPEntry {
  ip: string;
  ports: string;
  vlan: string;
}

interface IGMPStatus {
  enabled: boolean;
  entries: IGMPEntry[];
}

interface JumboFrameStatus {
  enabled: boolean;
  size: string;
}

interface PortState {
  port: string;
  status: string;
  link: string;
  speed: string;
  duplex: string;
  tx_bytes: number;
  rx_bytes: number;
  tx_packets: number;
  rx_packets: number;
  cum_tx: number;
  cum_rx: number;
  speed_tx_bps: number;
  speed_rx_bps: number;
  note?: string;
  is_sfp?: boolean;
  sfp_vendor?: string;
  sfp_model?: string;
}

interface MACEntry {
  mac: string;
  type: string;
  port: string;
  vlan: string;
  vendor?: string;
  host?: string;
}

interface SwitchData {
  name: string;
  ip: string;
  model: string;
  mac: string;
  uptime: string;
  firmware: string;
  hostname: string;
  ports: PortState[];
  mac_table: MACEntry[];
  mac_timestamp: number;
  status: string;
  error?: string;
  timestamp: number;
  dhcp_snooping: SnoopingStatus;
  igmp: IGMPStatus;
  jumbo_frame: JumboFrameStatus;
}

interface TopologyNode {
  id: string;
  name: string;
  type: string;
  ip?: string;
  mac?: string;
  model?: string;
  status: string;
}

interface TopologyLink {
  source: string;
  target: string;
  source_port?: string;
  target_port?: string;
  speed?: string;
  type: string;
}

interface TopologyData {
  nodes: TopologyNode[];
  links: TopologyLink[];
}

interface HistoryData {
  timestamps: number[];
  tx: number[];
  rx: number[];
}

interface MacTableState {
  expanded: boolean;
  filters: { mac: string; host: string; vendor: string; type: string; port: string; vlan: string };
  sortBy: string;
  sortAsc: boolean;
  scrollTop: number;
  height: number;
}

interface ColumnDef {
  label: string;
  thStyle: string;
  tdClass: string | ((p: PortState) => string);
  tdTitleFn?: (p: PortState) => string;
  render: (p: PortState, sw: SwitchData) => string;
}
