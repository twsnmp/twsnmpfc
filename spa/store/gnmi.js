export const state = () => ({
  conf: {
    path: '',
    value: '',
    search: '',
    history: '',
    itemsPerPage: 15,
  },
})

export const mutations = {
  setConf(state, c) {
    state.conf = c
  },
}
