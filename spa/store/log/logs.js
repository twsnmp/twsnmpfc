export const state = () => ({
  eventLog: {
    level: '',
    logtype: '',
    node: '',
    event: '',
    sortBy: 'TimeStr',
    sortDesc: false,
    page: 1,
    itemsPerPage: 15,
  },
  syslog: {
    pri: '',
    host: '',
    tag: '',
    msg: '',
    sortBy: 'TimeStr',
    sortDesc: false,
    page: 1,
    itemsPerPage: 15,
  },
  trapLog: {
    src: '',
    traptype: '',
    varbind: '',
    level: '',
    sortBy: 'TimeStr',
    sortDesc: false,
    page: 1,
    itemsPerPage: 15,
  },
  arpLog: {
    state: '',
    ip: '',
    mac: '',
    vendor: '',
    oldmac: '',
    oldvendor: '',
    sortBy: 'TimeStr',
    sortDesc: false,
    page: 1,
    itemsPerPage: 15,
  },
})

export const mutations = {
  setEventLog(state, c) {
    state.eventLog = c
  },
  setSyslog(state, c) {
    state.syslog = c
  },
  setTrapLog(state, c) {
    state.trapLog = c
  },
  setArpLog(state, c) {
    state.arpLog = c
  },
}
