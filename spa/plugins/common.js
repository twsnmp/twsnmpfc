import * as echarts from 'echarts'

const stateList = [
  { text: '重度', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'high' },
  { text: '軽度', color: '#fb9a99', icon: 'mdi-alert-circle', value: 'low' },
  { text: '注意', color: '#dfdf22', icon: 'mdi-alert', value: 'warn' },
  { text: '正常', color: '#33a02c', icon: 'mdi-check-circle', value: 'normal' },
  { text: 'Up', color: '#33a02c', icon: 'mdi-check-circle', value: 'up' },
  { text: '復帰', color: '#1f78b4', icon: 'mdi-autorenew', value: 'repair' },
  { text: '情報', color: '#1f78b4', icon: 'mdi-information', value: 'info' },
  { text: '新規', color: '#1f78b4', icon: 'mdi-information', value: 'New' },
  { text: '変化', color: '#e31a1c', icon: 'mdi-autorenew', value: 'Change' },
  { text: 'エラー', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'error' },
  { text: 'Down', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'down' },
  { text: '停止', color: '#777', icon: 'mdi-stop', value: 'off' },
  { text: 'Debug', color: '#777', icon: 'mdi-bug', value: 'debug' },
]
const stateMap = {}

stateList.forEach((e) => {
  stateMap[e.value] = e
})

const getStateColor = (state) => {
  return stateMap[state] ? stateMap[state].color : 'gray'
}

const getStateName = (state) => {
  return stateMap[state] ? stateMap[state].text : '不明'
}

const getStateIconName = (state) => {
  return stateMap[state] ? stateMap[state].icon : 'mdi-comment-question-outline'
}

const levelList = [
  { text: '重度', value: 'high' },
  { text: '軽度', value: 'low' },
  { text: '注意', value: 'warn' },
  { text: '情報', value: 'info' },
  { text: '停止', value: 'off' },
]

const filterEventLevelList = [
  { text: '指定しない', value: '' },
  { text: '重度以上', value: 'high' },
  { text: '軽度以上', value: 'low' },
  { text: '注意以上', value: 'warn' },
]

const filterEventTypeList = [
  { text: '指定しない', value: '' },
  { text: 'システム', value: 'system' },
  { text: 'ポーリング', value: 'polling' },
  { text: 'AI分析', value: 'ai' },
]

const typeList = [
  { text: 'PING', value: 'ping' },
  { text: 'SNMP', value: 'snmp' },
  { text: 'TCP', value: 'tcp' },
  { text: 'HTTP', value: 'http' },
  { text: 'TLS', value: 'tls' },
  { text: 'DNS', value: 'dns' },
  { text: 'NTP', value: 'ntp' },
  { text: 'SYSLOG', value: 'syslog' },
  { text: 'SNMP TRAP', value: 'trap' },
  { text: 'NetFlow', value: 'netflow' },
  { text: 'IPFIX', value: 'ipfix' },
  { text: 'ARP Log', value: 'arp' },
  { text: 'Command', value: 'cmd' },
  { text: 'SSH', value: 'ssh' },
  { text: 'Report', value: 'report' },
  { text: 'TWSNMP', value: 'twsnmp' },
  { text: 'VMware', value: 'vmware' },
]

const logModeList = [
  { text: '記録しない', value: 0 },
  { text: '常に記録', value: 1 },
  { text: '状態変化時のみ記録', value: 2 },
  { text: 'AI分析', value: 3 },
]

const addrModeList = [
  { text: 'IP固定', value: '' },
  { text: 'MAC固定', value: 'mac' },
  { text: 'ホスト名固定', value: 'host' },
]

const snmpModeList = [
  { text: 'SNMPv2c', value: '' },
  { text: 'SNMPv3認証', value: 'v3auth' },
  { text: 'SNMPv3認証暗号化(AES128)', value: 'v3authpriv' },
  { text: 'SNMPv3認証暗号化(SHA256/AES256)', value: 'v3authprivex' },
]

const aiThList = [
  { text: '0.01%以下', value: 88 },
  { text: '0.1%以下', value: 81 },
  { text: '1%以下', value: 74 },
]

/* eslint prettier/prettier: 0 */
const iconList = [
  {
    text: 'デスクトップ',
    icon: 'mdi-desktop-mac',
    value: 'desktop',
    code: 0xF01C4,
  },
  {
    text: '古いデスクトップ',
    icon: 'mdi-desktop-classic',
    value: 'desktop-classic',
    code: 0xF07C0,
  },
  { text: 'ラップトップ', icon: 'mdi-laptop', value: 'laptop' ,code: 0xF0322},
  { text: 'タブレット', icon: 'mdi-tablet-ipad', value: 'tablet' ,code:0xF04F8},
  { text: 'サーバー', icon: 'mdi-server', value: 'server' ,code: 0xF048B},
  { text: 'ネットワーク機器', icon: 'mdi-ip-network', value: 'hdd' ,code: 0xF0A60},
  { text: 'IP機器', icon: 'mdi-ip-network', value: 'ip' ,code: 0xF0A60},
  { text: 'ネットワーク', icon: 'mdi-lan', value: 'network' ,code: 0xF0317},
  { text: '無線LAN', icon: 'mdi-wifi', value: 'wifi' ,code: 0xF05A9},
  { text: 'クラウド', icon: 'mdi-cloud', value: 'cloud' ,code: 0xF015F },
  { text: 'プリンター', icon: 'mdi-printer', value: 'printer' ,code: 0xF042A},
  { text: 'モバイル', icon: 'mdi-cellphone', value: 'cellphone' ,code: 0xF011C},
  { text: 'ルーター1', icon: 'mdi-router', value: 'router' ,code: 0xF11E2},
  { text: 'WEB', icon: 'mdi-web', value: 'web' ,code: 0xF059F},
  { text: 'データベース', icon: 'mdi-database', value: 'db' ,code: 0xF01BC},
  { text: '無線AP', icon: 'mdi-router-wireless', value: 'mdi-router-wireless' ,code: 0xF0469},
  { text: 'ルーター2', icon: 'mdi-router-network', value: 'mdi-router-network' ,code: 0xF1087},
  { text: 'セキュリティー', icon: 'mdi-security', value: 'mdi-security' ,code: 0xF0483},
  { text: 'タワーPC', icon: 'mdi-desktop-tower', value: 'mdi-desktop-tower' ,code: 0xF01C5},
  { text: 'Windows', icon: 'mdi-microsoft-windows', value: 'mdi-microsoft-windows' ,code: 0xF05B3},
  { text: 'Linux', icon: 'mdi-linux', value: 'mdi-linux' ,code: 0xF033D},
  { text: 'Raspberry PI', icon: 'mdi-raspberry-pi', value: 'mdi-raspberry-pi' ,code: 0xF043F},
  { text: 'メールサーバー', icon: 'mdi-mailbox', value: 'mdi-mailbox' ,code: 0xF06EE},
  { text: 'NTPサーバー', icon: 'mdi-clock', value: 'mdi-clock' ,code: 0xF0954},
  { text: 'Android', icon: 'mdi-android', value: 'mdi-android' ,code: 0xF0032},
  { text: 'Azure', icon: 'mdi-microsoft-azure', value: 'mdi-microsoft-azure' ,code: 0xF0805},
  { text: 'Amazon', icon: 'mdi-amazon', value: 'mdi-amazon' ,code: 0xF002D},
  { text: 'Apple', icon: 'mdi-apple', value: 'mdi-apple' ,code: 0xF0035},
  { text: 'Google', icon: 'mdi-google', value: 'mdi-google' ,code: 0xF02AD},
  { text: 'CDプレーヤー', icon: 'mdi-disc-player', value: 'mdi-disc-player' ,code: 0xF0960},
  { text: 'TWSNMP連携マップ', icon: 'mdi-layers-search', value: 'mdi-layers-search' ,code: 0xF1206},
]

const iconMap = {}

iconList.forEach((e) => {
  iconMap[e.value] = e.icon
})

const getIconName = (icon) => {
  return iconMap[icon] ? iconMap[icon] : 'mdi-comment-question-outline'
}

const setIcon = (e) => {
  for( let i = 0; i < iconList; i++) {
    if(iconList[i].value === e.Icon) {
      iconList[i].text = e.Text
      iconList[i].Code = e.Code
      return
    }
  }
  // 追加
  iconList.push({
    text: e.Text,
    icon: e.Icon,
    value: e.Icon,
    code: e.Code,
  })
  iconMap[e.Icon] = e.Icon
}

const delIcon = (icon) => {
  for( let i = 0; i < iconList; i++) {
    if(iconList[i].value === icon) {
      iconList.splice(i+1,1)
      delete(iconMap[icon])
      return
    }
  }
}

const timeFormat = (date, format) => {
  if (!format) {
      format = '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
  }
  return echarts.time.format(date,format)
}

const getScoreColor = (s) => {
  if (s > 66) {
    return getStateColor('repair')
  } else if (s >= 50) {
    return getStateColor('info')
  } else if (s > 42) {
    return getStateColor('warn')
  } else if (s > 33) {
    return getStateColor('low')
  }
  return getStateColor('high')
}

const getScoreIconName = (s) => {
  if (s > 66) {
    return 'mdi-emoticon-excited-outline'
  } else if (s >= 50) {
    return 'mdi-emoticon-outline'
  } else if (s > 42) {
    return 'mdi-emoticon-sad-outline'
  } else if (s > 33) {
    return 'mdi-emoticon-sick-outline'
  }
  return 'mdi-emoticon-dead-outline'
}

// Service Name Map
const serviceNameArray = [
  ["http/tcp", "WEB"],
  ["https/tcp", "WEB"],
  ["ldap/tcp", "LDAP"],
  ["ldaps/tcp", "LDAP"],
  ["domain/tcp", "DNS"],
  ["domain/udp", "DNS"],
  ["snmp/udp", "SNMP"],
  ["ntp/udp", "NTP"],
  ["smtp/tcp", "MAIL"],
  ["submission/tcp", "MAIL"],
  ["pop3/tcp", "MAIL"],
  ["pop3s/tcp", "MAIL"],
  ["imap/tcp", "MAIL"],
  ["imap3/tcp", "MAIL"],
  ["imaps/tcp", "MAIL"],
  ["ssh/tcp", "SSH"],
  ["telnet/tcp", "TELNET"],
  ["ftp/tcp", "FTP"],
  ["bootps/udp", "DHCP"],
  ["syslog/udp", "SYSLOG"],
  ["nfsd/tcp", "NFS"],
  ["microsoft-ds/tcp", "CIFS"],
  ["ms-wbt-server/tcp", "RDP"],
  ["rfb/tcp", "VNC"],
  ["netbios-ns/udp", "NETBIOS"],
  ["netbios-dgm/udp", "NETBIOS"],
  ["kerberos/tcp", "AD"],
  ["icmp", "ICMP"],
  ["igmp", "IGMP"],
  ["wudo/tcp", "WUDO"],
  ["redis/tcp", "REDIS"],
  ["radius/udp", "RADIUS"],
  ["apple-apn/tcp", "APPLE"],
  ["android/tcp", "ANDROID"],
]

const serviceNameMap = new Map(serviceNameArray);

function getServiceName(s) {
  const ret = serviceNameMap.get(s)
  if (ret) {
    return ret
  }
  if (s.indexOf("/icmp") > 0) {
    return "ICMP"
  }
  return 'Other'
}

function getServiceInfo(services) {
  const sns = new Map();
  for (let i = 0; i < services.length; i++) {
    const n = getServiceName(services[i]);
    if (n) {
      sns.set(n, true);
    }
  }
  const ks = Array.from(sns.keys())
  while (ks.length > 10) {
    ks.pop()
  }
  return ks.join() + "(" + services.length + ")"
}

function getLocInfo(l) {
  const loc = l.split(',')
  const r = {
    LatLong: '',
    LocInfo: '',
    Country: '',
  }
  if (loc.length < 3) {
    return r
  }
  if (loc[0] === 'LOCAL') {
    r.LocInfo = 'ローカル'
    return r
  }
  r.Country = loc[0]
  if (loc.length > 3 && loc[3]) {
    r.LocInfo = loc[3] + '/' + loc[0]
  } else {
    r.LocInfo = loc[0]
  }
  r.LatLong = loc[1] + ',' + loc[2]
  return r
}

const protocolFilterList = [
  { text: '指定しない', value: '' },
  { text: 'icmp', value: '1' },
  { text: 'igmp', value: '2' },
  { text: 'tcp', value: '6' },
  { text: 'udp', value: '17' },
]
const tcpFlagFilterList = [
  { text: '指定しない', value: '' },
  { text: 'SYNのみ', value: `\\.S\\.{3}` },
  { text: 'RSTあり', value: 'R' },
  { text: 'FINなし', value: '[^F]' },
  { text: '一般的(SYN/ACK/FIN)', value: 'FS\\.P*A+' },
]

const cmpIP = (a,b) => {
  if (!a.includes(".") || !b.includes(".") ){
    return a < b  ? -1 : a > b ? 1 : 0 
  }
  const pa = a.split('.').map(function(s) {
    return parseInt(s); 
  });
  const pb = b.split('.').map(function(s) {
    return parseInt(s); 
  });
  for(let i =0;i < pa.length;i++){
    if (i >= pb.length){
      return -1;
    }
    if (pa[i] === pb[i]){
      continue;
    }
    if (pa[i] < pb[i]){
      return -1;
    }
    return 1;
  }
  return 0;
}

const getLogModeName = (m) => {
  switch(m){
    case 0:
      return 'off'
    case 1:
      return 'all'
    case 2:
      return 'diff'
    case 3:
      return 'ai'
  }
  return ''
} 

const getRSSIColor = (rssi) => {
  if (rssi >= 0) {
    return getStateColor('debug')
  } else if (rssi >= -70) {
    return getStateColor('info')
  } else if (rssi >= -80) {
    return getStateColor('warn')
  }
  return getStateColor('high')
}

const getRSSIIconName = (rssi) => {
  if (rssi >= 0) {
    return 'mdi-wifi-strength-alert-outline'
  } else if (rssi >= -67) {
    return 'mdi-wifi-strength-4'
  } else if (rssi >= -70) {
    return 'mdi-wifi-strength-3'
  } else if (rssi >= -80) {
    return 'mdi-wifi-strength-2'
  }
  return 'mdi-wifi-strength-1'
}


export default (context, inject) => {
  inject('getIconName', getIconName)
  inject('getStateName', getStateName)
  inject('getScoreColor', getScoreColor)
  inject('getScoreIconName', getScoreIconName)
  inject('getRSSIColor', getRSSIColor)
  inject('getRSSIIconName', getRSSIIconName)
  inject('getStateColor', getStateColor)
  inject('getStateIconName', getStateIconName)
  inject('logModeList', logModeList)
  inject('typeList', typeList)
  inject('levelList', levelList)
  inject('iconList', iconList)
  inject('stateList', stateList)
  inject('addrModeList', addrModeList)
  inject('snmpModeList', snmpModeList)
  inject('aiThList', aiThList)
  inject('filterEventLevelList', filterEventLevelList)
  inject('filterEventTypeList', filterEventTypeList)
  inject('timeFormat', timeFormat)
  inject('getServiceInfo', getServiceInfo)
  inject('getLocInfo', getLocInfo)
  inject('protocolFilterList', protocolFilterList)
  inject('tcpFlagFilterList', tcpFlagFilterList)
  inject('cmpIP', cmpIP)
  inject('getLogModeName', getLogModeName)
  inject('setIcon', setIcon)
  inject('delIcon', delIcon)
}
