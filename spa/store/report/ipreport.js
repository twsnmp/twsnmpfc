export const state = () => ({
  conf: {
    ip: '',
    name: '',
    country: '',
    mac: '',
    vendor: '',
    excludeFlag: false,
    sortBy: 'Score',
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
