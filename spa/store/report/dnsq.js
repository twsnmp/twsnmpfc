export const state = () => ({
  conf: {
    host: '',
    server: '',
    type: '',
    name: '',
    client: '',
    sortBy: 'Count',
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
