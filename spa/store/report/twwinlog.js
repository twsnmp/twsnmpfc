export const state = () => ({
  winEventID: {
    level: '',
    computer: '',
    channel: '',
    provider: '',
    eventID: '',
    sortBy: 'Count',
    sortDesc: false,
    page: 1,
    itemsPerPage: 15,
  },
})

export const mutations = {
  setWinEventID(state, c) {
    state.winEventID = c
  },
}
