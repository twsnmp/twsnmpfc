export const state = () => ({
  conf: {
    client: '',
    server: '',
    country: '',
    service: '',
    version: '',
    cipher: '',
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
