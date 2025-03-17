export const state = () => ({
  conf: {
    Status: '',
    ID: '',
    Subject: '',
    sortBy: 'Created',
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
