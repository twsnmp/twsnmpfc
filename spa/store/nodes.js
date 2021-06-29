export const state = () => ({
  conf: {
    state: '',
    name: '',
    ip: '',
    descr: '',
    sortBy: 'State',
    sortDesc: false,
    page: 1,
    itemsPerPage: 15,
  },
})

export const mutations = {
  setConf(state, c) {
    state.conf = c
  },
}
