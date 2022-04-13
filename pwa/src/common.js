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
  return stateMap[state] ? stateMap[state] : {color:'gray',icon:'mdi-comment-question-outline', name:'不明'}
}

