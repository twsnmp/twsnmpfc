import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

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

const showBlueScanGraph = (div, list, filter) => {
  const nodeMap = new Map()
  const edgeMap = new Map()
  list.forEach((l) => {
    if (!l.Name || !filterBluetoothDev(l, filter)) {
      return
    }
    const ek = l.Host + ':' + l.Name + '->' + l.Address
    const e = edgeMap.get(ek)
    if (!e) {
      edgeMap.set(ek, {
        source: l.Host + ':' + l.Name,
        target: l.Address,
        value: l.Count,
      })
    } else {
      e.value += l.Count
    }
    const src = l.Host + ':' + l.Name
    let n = nodeMap.get(src)
    if (!n) {
      nodeMap.set(src, {
        name: src,
        value: l.Count,
        draggable: true,
        category: 0,
        label: {
          show: true,
        },
      })
    } else {
      n.value += l.Count
    }
    const dst = l.Address
    n = nodeMap.get(dst)
    if (!n) {
      nodeMap.set(dst, {
        name: dst,
        value: l.Count,
        draggable: true,
        category: l.AddressType.includes('Public')
          ? 1
          : l.AddressType.includes('Random')
          ? 2
          : 3,
        label: {
          show: false,
        },
      })
    }
  })
  const nodes = Array.from(nodeMap.values())
  const edges = Array.from(edgeMap.values())
  const nvs = []
  const evs = []
  nodes.forEach((e) => {
    nvs.push(e.value)
  })
  edges.forEach((e) => {
    evs.push(e.value)
  })
  const n95 = ecStat.statistics.quantile(nvs, 0.95)
  const n50 = ecStat.statistics.quantile(nvs, 0.5)
  const e95 = ecStat.statistics.quantile(evs, 0.95)
  const categories = [
    { name: 'Host:Name' },
    { name: 'LE Public' },
    { name: 'LE Random' },
    { name: 'Other' },
  ]
  nodes.forEach((e) => {
    e.symbolSize = e.value > n95 ? 5 : e.value > n50 ? 4 : 2
  })
  edges.forEach((e) => {
    e.lineStyle = {
      width: e.value > e95 ? 2 : 1,
    }
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
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {
      trigger: 'item',
      textStyle: {
        fontSize: 8,
      },
      position: 'bottom',
    },
    legend: [
      {
        orient: 'vertical',
        top: 50,
        right: 20,
        textStyle: {
          fontSize: 10,
          color: '#ccc',
        },
        data: categories.map(function (a) {
          return a.name
        }),
      },
    ],
    color: ['#1f78b4', '#e31a1c', '#dfdf22', '#fb9a99'],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [],
  }
  options.series = [
    {
      name: 'Host:Name->Address',
      type: 'graph',
      layout: 'circular',
      circular: {
        rotateLabel: true,
      },
      data: nodes,
      links: edges,
      categories,
      roam: true,
      label: {
        show: true,
        position: 'right',
        formatter: '{b}',
        fontSize: 8,
        fontStyle: 'normal',
        color: '#ccc',
      },
      lineStyle: {
        color: 'source',
        curveness: 0.3,
      },
    },
  ]
  chart.setOption(options)
  chart.resize()
}

const showBlueVendorChart = (div, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    color: ['#e31a1c', '#dfdf22', '#1f78b4'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['LE Public', 'LE Random', 'Other'],
    },
    grid: {
      top: '10%',
      left: '5%',
      right: '10%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: 'Count',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis: {
      type: 'category',
      axisLine: {
        show: false,
      },
      axisTick: {
        show: false,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      data: [],
    },
    series: [
      {
        name: 'LE Public',
        type: 'bar',
        stack: 'Count',
        data: [],
      },
      {
        name: 'LE Random',
        type: 'bar',
        stack: 'Count',
        data: [],
      },
      {
        name: 'Other',
        type: 'bar',
        stack: 'Count',
        data: [],
      },
    ],
  }
  if (!list) {
    return
  }
  const data = {}
  list.forEach((l) => {
    if (l.Vendor === 'Unknown' || !filterBluetoothDev(l, filter)) {
      return
    }
    if (!data[l.Vendor]) {
      data[l.Vendor] = [0, 0, 0, 0]
    }
    data[l.Vendor][0]++
    const si = l.AddressType.includes('Public')
      ? 1
      : l.AddressType.includes('Random')
      ? 2
      : 3
    data[l.Vendor][si]++
  })
  const keys = Object.keys(data)
  keys.sort(function (a, b) {
    return data[b][0] - data[a][0]
  })
  let i = keys.length - 1
  if (i > 49) {
    i = 49
  }
  for (; i >= 0; i--) {
    option.yAxis.data.push(keys[i])
    for (let j = 0; j < 3; j++) {
      option.series[j].data.push(data[keys[i]][j + 1])
    }
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showRSSITime3DChart', showRSSITime3DChart)
  inject('showRSSILoc3DChart', showRSSILoc3DChart)
  inject('filterBluetoothDev', filterBluetoothDev)
  inject('filterWifiAP', filterWifiAP)
  inject('showBlueScanGraph', showBlueScanGraph)
  inject('showBlueVendorChart', showBlueVendorChart)
}
