import * as echarts from 'echarts'

let chart

const showRSSITime3DChart = (div, wifi, list, filter) => {
  const data = []
  const cat = []
  list.forEach((i) => {
    if (!i.RSSI || i.RSSI.length < 1) {
      return
    }
    if (wifi) {
      if (!filterWifiAP(i, filter)) {
        return
      }
    } else if (!filterBluetoothDev(i, filter)) {
      return
    }
    const id = i.Host + ':' + (wifi ? i.BSSID : i.Address)
    i.RSSI.forEach((e) => {
      data.push([
        id,
        e.Time / (1000 * 1000),
        e.Value,
        getRSSILevel(e.Value),
        wifi ? i.SSID : i.Name,
      ])
    })
    cat.push(id)
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: '#4b5769',
      },
      {
        offset: 1,
        color: '#404a59',
      },
    ]),
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      dimension: 3,
      min: 0,
      max: 4,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: wifi ? 'Host:BSSID' : 'Host:Address',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: 'RSSI',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: wifi ? 'Wifi AP' : 'Bluetooth Device',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: [
          wifi ? 'Host:BSSID' : 'Host:Address',
          'Time',
          'RSSI',
          'Level',
          wifi ? 'SSID' : 'Name',
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showRSSILoc3DChart = (div, wifi, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((i) => {
    if (!i.RSSI || i.RSSI.length < 1) {
      return
    }
    if (wifi) {
      if (!filterWifiAP(i, filter)) {
        return
      }
    } else if (!filterBluetoothDev(i, filter)) {
      return
    }
    i.RSSI.forEach((e) => {
      data.push([
        i.Host,
        wifi ? i.BSSID : i.Address,
        e.Value,
        getRSSILevel(e.Value),
        wifi ? i.SSID : i.Name,
      ])
    })
    mapx.set(i.Host, true)
    mapy.set(wifi ? i.BSSID : i.Address, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    title: {
      show: false,
    },
    backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
      {
        offset: 0,
        color: '#4b5769',
      },
      {
        offset: 1,
        color: '#404a59',
      },
    ]),
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      dimension: 3,
      min: 0,
      max: 4,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host',
      data: catx,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'category',
      name: wifi ? 'BSSID' : 'Address',
      data: caty,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: 'RSSI',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: wifi ? 'Wifi AP' : 'Bluetooth Device',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: [
          'Host',
          wifi ? 'BSSID' : 'Address',
          'RSSI',
          'Level',
          wifi ? 'SSID' : 'Name',
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const filterBluetoothDev = (i, filter) => {
  if (filter.address && !i.Address.includes(filter.address)) {
    return false
  }
  if (filter.addressType && !i.AddressType.includes(filter.addressType)) {
    return false
  }
  if (filter.name && !i.Name.includes(filter.name)) {
    return false
  }
  if (filter.vendor && !i.Vendor.includes(filter.vendor)) {
    return false
  }
  if (filter.host && !i.Host.includes(filter.host)) {
    return false
  }
  return true
}

const filterWifiAP = (i, filter) => {
  if (filter.bssid && !i.BSSID.includes(filter.bssid)) {
    return false
  }
  if (filter.ssid && !i.SSID.includes(filter.ssid)) {
    return false
  }
  if (filter.host && !i.Host.includes(filter.host)) {
    return false
  }
  return true
}

const getRSSILevel = (rssi) => {
  if (rssi >= 0) {
    return 4
  } else if (rssi >= -67) {
    return 3
  } else if (rssi >= -70) {
    return 2
  } else if (rssi >= -80) {
    return 1
  }
  return 0
}

export default (context, inject) => {
  inject('showRSSITime3DChart', showRSSITime3DChart)
  inject('showRSSILoc3DChart', showRSSILoc3DChart)
  inject('filterBluetoothDev', filterBluetoothDev)
  inject('filterWifiAP', filterWifiAP)
}
