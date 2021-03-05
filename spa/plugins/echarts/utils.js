export const getScoreIndex = (s) => {
  if (s > 66) {
    return 5
  } else if (s > 50) {
    return 4
  } else if (s > 42) {
    return 3
  } else if (s > 33) {
    return 2
  } else if (s <= 0) {
    return 6
  }
  return 1
}
