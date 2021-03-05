import * as echarts from 'echarts'

const stateList = [
  { text: '重度', color: '#e31a1c', icon: 'mdi-alert-circle', value: 'high' },
  { text: '軽度', color: '#fb9a99', icon: 'mdi-alert-circle', value: 'low' },
  { text: '注意', color: '#dfdf22', icon: 'mdi-alert', value: 'warn' },
  { text: '正常', color: '#33a02c', icon: 'mdi-check-circle', value: 'normal' },
  { text: '復帰', color: '#1f78b4', icon: 'mdi-autorenew', value: 'repair' },
  { text: '情報', color: '#1f78b4', icon: 'mdi-information', value: 'info' },
  { text: '新規', color: '#1f78b4', icon: 'mdi-information', value: 'New' },
  { text: '変化', color: '#e31a1c', icon: 'mdi-autorenew', value: 'Change' },
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
  { text: 'HTTPS', value: 'https' },
  { text: 'TLS', value: 'tls' },
  { text: 'DNS', value: 'dns' },
  { text: 'NTP', value: 'ntp' },
  { text: 'SYSLOG', value: 'syslog' },
  { text: 'SYSLOG PRI', value: 'syslogpri' },
  { text: 'SYSLOG Device', value: 'syslogdevice' },
  { text: 'SYSLOG User', value: 'sysloguser' },
  { text: 'SYSLOG Flow', value: 'syslogflow' },
  { text: 'SNMP TRAP', value: 'trap' },
  { text: 'NetFlow', value: 'netflow' },
  { text: 'IPFIX', value: 'ipfix' },
  { text: 'Command', value: 'cmd' },
  { text: 'SSH', value: 'ssh' },
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
  { text: 'SNMPv3認証暗号化', value: 'v3authpriv' },
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
  { text: 'ルーター', icon: 'mdi-router', value: 'router' ,code: 0xF11E2},
  { text: 'WEB', icon: 'mdi-web', value: 'web' ,code: 0xF059F},
  { text: 'Database', icon: 'mdi-database', value: 'db' ,code: 0xF01BC},
]

const iconMap = {}

iconList.forEach((e) => {
  iconMap[e.value] = e.icon
})

const getIconName = (icon) => {
  return iconMap[icon] ? iconMap[icon] : 'mdi-comment-question-outline'
}

const timeFormat = (date, format) => {
  if (!format) {
      format = 'yyyy/MM/dd hh:mm:ss'
  }
  return echarts.format.formatTime(format, date)
}

const getScoreColor = (s) => {
  if (s > 66) {
    return getStateColor('repair')
  } else if (s > 50) {
    return getStateColor('info')
  } else if (s > 42) {
    return getStateColor('warn')
  } else if (s > 33) {
    return getStateColor('low')
  } else if (s <= 0) {
    return getStateColor('unknown')
  }
  return getStateColor('high')
}

const getScoreIconName = (s) => {
  if (s > 66) {
    return 'mdi-emoticon-excited-outline'
  } else if (s > 50) {
    return 'mdi-emoticon-outline'
  } else if (s > 42) {
    return 'mdi-emoticon-sad-outline'
  } else if (s > 33) {
    return 'mdi-emoticon-sick-outline'
  } else if (s <= 0) {
    return 'mdi-help-circle-outline'
  }
  return 'mdi-emoticon-dead-outline'
}

// Service Name Map
const serviceNameArray = [
  ["submission/tcp", "SMTP"],
  ["http/tcp", "HTTP"],
  ["https/tcp", "HTTPS"],
  ["ldap/tcp", "LDAP"],
  ["ldaps/tcp", "LDAPS"],
  ["domain/tcp", "DNS"],
  ["domain/udp", "DNS"],
  ["snmp/udp", "SNMP"],
  ["ntp/udp", "NTP"],
  ["smtp/tcp", "SMTP"],
  ["pop3/tcp", "POP3"],
  ["pop3s/tcp", "POP3S"],
  ["imap/tcp", "IMAP"],
  ["imaps/tcp", "IMAPS"],
  ["ssh/tcp", "SSH"],
  ["telnet/tcp", "TELNET"],
  ["ftp/tcp", "FTP"],
  ["bootps/udp", "DHCP"],
  ["syslog/udp", "SYSLOG"],
  ["microsoft-ds/tcp", "CIFS"],
  ["rfb/tcp", "RFB"],
  ["netbios-ns/udp", "NETBIOS"],
  ["netbios-dgm/udp", "NETBIOS"],
  ["icmp", "ICMP"],
  ["igmp", "IGMP"]
]

const serviceNameMap = new Map(serviceNameArray);

function getServiceName(s) {
  const ret = serviceNameMap.get(s);
  return ret || 'Other'
}

function getServiceNames(services) {
  const sns = new Map();
  for (let i = 0; i < services.length; i++) {
    const n = getServiceName(services[i]);
    sns.set(n, true);
  }
  return Array.from(sns.keys()).join();
}

export default (context, inject) => {
  inject('getIconName', getIconName)
  inject('getStateName', getStateName)
  inject('getScoreColor', getScoreColor)
  inject('getScoreIconName', getScoreIconName)
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
  inject('getServiceNames', getServiceNames)
}
