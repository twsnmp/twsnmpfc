export const state = () => ({
  title: 'TWSNMP FC',
  nodeList: [],
  lastUpdate: 0,
  readOnly: false,
})

export const mutations = {
  setMAP(state, m) {
    state.title = 'TWSNMP FC -' + m.MapConf.MapName
    state.nodeList.length = 0
    state.lastUpdate = m.LastUpdate
    Object.keys(m.Nodes).forEach((k) => {
      state.nodeList.push({
        text: m.Nodes[k].Name,
        value: k,
      })
    })
  },
  setReadOnly(state, v) {
    if (!v || v === 'false') {
      v = false
    } else {
      v = true
    }
    state.readOnly = v
  },
}
