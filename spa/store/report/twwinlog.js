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
  winKerberos: {
    target: '',
    computer: '',
    ip: '',
    service: '',
    ticketType: '',
    subject: '',
    sortBy: 'Count',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  winPrivilege: {
    subject: '',
    computer: '',
    sortBy: 'Count',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  winProcess: {
    process: '',
    computer: '',
    subject: '',
    sortBy: 'Count',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  winTask: {
    task: '',
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
  setWinKerberos(state, c) {
    state.winKerberos = c
  },
  setWinPrivilege(state, c) {
    state.winPrivilege = c
  },
  setWinProcess(state, c) {
    state.winProcess = c
  },
  setWinTask(state, c) {
    state.winTask = c
  },
}
