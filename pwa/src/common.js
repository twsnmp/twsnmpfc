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

export const getState = (state) => {
  return stateMap[state] ? stateMap[state] : {color:'gray',icon:'mdi-comment-question-outline', text:'不明'}
}

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
const iconCodeMap = {}

iconList.forEach((e) => {
  iconMap[e.value] = e
  iconCodeMap[e.value] = String.fromCodePoint(e.code)
})

export const getIconName = (icon) => {
  return iconMap[icon] ? iconMap[icon].value : 'mdi-comment-question-outline'
}

export const getIconCode = (icon) => {
  return iconCodeMap[icon] ? iconCodeMap[icon] : String.fromCodePoint(0xF0A39)
}

export const setIcon = (e) => {
  for( let i = 0; i < iconList; i++) {
    if(iconList[i].value === e.Icon) {
      iconList[i].text = e.Text
      iconList[i].code = e.Code
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
  iconMap[e.Icon] = e
  iconCodeMap[e.value] = String.fromCodePoint(e.code)
}

