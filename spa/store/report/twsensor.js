export const state = () => ({
  wifiAP: {
    host: '',
    bsid: '',
    ssid: '',
    sortBy: 'LastRSSI',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  blueDevice: {
    host: '',
    address: '',
    vendor: '',
    name: '',
    sortBy: 'LastRSSI',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
  envMonitor: {
    host: '',
    address: '',
    name: '',
    sortBy: 'DataCount',
    sortDesc: true,
    page: 1,
    itemsPerPage: 15,
  },
})

export const mutations = {
  setWifiAP(state, c) {
    state.wifiAP = c
  },
  setBlueDevice(state, c) {
    state.blueDevice = c
  },
  setEnvMonitor(state, c) {
    state.envMonitor = c
  },
}
