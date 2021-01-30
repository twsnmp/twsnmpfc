export const state = () => ({
  title: 'TWSNMP FC',
  nodeList: [],
})

export const mutations = {
  setMAP(state, m) {
    state.title = 'TWSNMP FC -' + m.MapConf.MapName
    state.nodeList.length = 0
    Object.keys(m.Nodes).forEach((k) => {
      state.nodeList.push({
        text: m.Nodes[k].Name,
        value: k,
      })
    })
  },
}
