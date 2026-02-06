export const state = () => ({
  conf: {
    ipv6: '',
    ipv4: '',
    node: '',
    mac: '',
    vendor: '',
    sortBy: 'IPv6',
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
