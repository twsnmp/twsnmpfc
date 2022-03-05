export const getScoreIndex = (s) => {
  if (s > 66) {
    return 5
  } else if (s > 50) {
    return 4
  } else if (s > 42) {
    return 3
  } else if (s > 33) {
    return 2
  }
  return 1
}

const ipv4Regex = /^(\d{1,3}\.){3,3}\d{1,3}$/
const ipv6Regex =
  /^(::)?(((\d{1,3}\.){3}(\d{1,3}){1})?([0-9a-f]){0,4}:{0,2}){1,8}(::)?$/i

export const isV4Format = (ip) => {
  return ipv4Regex.test(ip)
}

export const isV6Format = (ip) => {
  return ipv6Regex.test(ip)
}

export const isPrivateIP = (addr) => {
  return (
    /^(::f{4}:)?10\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})$/i.test(addr) ||
    /^(::f{4}:)?192\.168\.([0-9]{1,3})\.([0-9]{1,3})$/i.test(addr) ||
    /^(::f{4}:)?172\.(1[6-9]|2\d|30|31)\.([0-9]{1,3})\.([0-9]{1,3})$/i.test(
      addr
    ) ||
    /^(::f{4}:)?127\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})$/i.test(addr) ||
    /^(::f{4}:)?169\.254\.([0-9]{1,3})\.([0-9]{1,3})$/i.test(addr) ||
    /^f[cd][0-9a-f]{2}:/i.test(addr) ||
    /^fe80:/i.test(addr) ||
    /^::1$/.test(addr) ||
    /^::$/.test(addr)
  )
}

export const setZoomCallback = (chart, cb, st, lt) => {
  if (!cb) {
    return
  }
  chart.on('datazoom', (e) => {
    if (e.batch && e.batch.length === 2) {
      if (e.batch[0].startValue) {
        // Select ZOOM
        // eslint-disable-next-line node/no-callback-literal
        cb(
          e.batch[0].startValue * 1000 * 1000,
          e.batch[0].endValue * 1000 * 1000
        )
      } else if (e.batch[0].end === 100 && e.batch[0].start === 0) {
        // Reset ZOOM
        // eslint-disable-next-line node/no-callback-literal
        cb(false, false)
      }
    } else if (e.start !== undefined && e.end !== undefined) {
      // Scroll ZOOM
      // eslint-disable-next-line node/no-callback-literal
      cb(st + (lt - st) * (e.start / 100), st + (lt - st) * (e.end / 100))
    }
  })
}
