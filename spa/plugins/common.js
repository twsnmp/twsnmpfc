const stateList = [
  { text: '重度', color: 'red', icon: 'mdi-alert-circle', value: 'high' },
  { text: '軽度', color: 'pink', icon: 'mdi-alert-circle', value: 'low' },
  { text: '注意', color: 'yellow', icon: 'mdi-alert', value: 'warn' },
  { text: '正常', color: 'green', icon: 'mdi-check-circle', value: 'normal' },
  { text: '復帰', color: 'blue', icon: 'mdi-autorenew', value: 'repair' },
  { text: '情報', color: 'blue', icon: 'mdi-information', value: 'info' },
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

const iconList = [
  { text: 'デスクトップ', icon: 'mdi-desktop-mac', value: 'desktop' },
  { text: 'ラップトップ', icon: 'mdi-laptop', value: 'laptop' },
  { text: 'タブレット', icon: 'mdi-tablet-ipad', value: 'tablet' },
  { text: 'サーバー', icon: 'mdi-server', value: 'server' },
  { text: 'ネットワーク機器', icon: 'mdi-ip-network', value: 'hdd' },
  { text: 'IP機器', icon: 'mdi-ip-network', value: 'ip' },
  { text: 'ネットワーク', icon: 'mdi-lan', value: 'network' },
  { text: '無線LAN', icon: 'mdi-wifi', value: 'wifi' },
  { text: 'クラウド', icon: 'mdi-cloud', value: 'cloud' },
  { text: 'プリンター', icon: 'mdi-printer', value: 'printer' },
  { text: 'モバイル', icon: 'mdi-cellphone', value: 'cellphone' },
  { text: 'ルーター', icon: 'mdi-router', value: 'router' },
  { text: 'WEB', icon: 'mdi-desktop-mac', value: 'web' },
]

const iconMap = {}

iconList.forEach((e) => {
  iconMap[e.value] = e.icon
})

const getIconName = (icon) => {
  return iconMap[icon] ? iconMap[icon] : 'comment-question-outline'
}

export default (context, inject) => {
  inject('getIconName', getIconName)
  inject('getStateName', getStateName)
  inject('getStateColor', getStateColor)
  inject('getStateIconName', getStateIconName)
  inject('logModeList', logModeList)
  inject('typeList', typeList)
  inject('levelList', levelList)
  inject('iconList', iconList)
  inject('addrModeList', addrModeList)
  inject('snmpModeList', snmpModeList)
  inject('aiThList', aiThList)
  inject('filterEventLevelList', filterEventLevelList)
  inject('filterEventTypeList', filterEventTypeList)
}
