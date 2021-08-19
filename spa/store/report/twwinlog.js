export const state = () => ({
  winEventID: {
    level: '',
    computer: '',
    channel: '',
    provider: '',
    eventID: '',
    sortBy: 'Count',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  winLogon: {
    target: '',
    computer: '',
    ip: '',
    sortBy: 'Count',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  winAccount: {
    subject: '',
    target: '',
    computer: '',
    sortBy: 'Count',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
})

export const mutations = {
  setWinEventID(state, c) {
    state.winEventID = c
  },
  setWinLogon(state, c) {
    state.winLogon = c
  },
  setWinAccount(state, c) {
    state.winAccount = c
  },
}
