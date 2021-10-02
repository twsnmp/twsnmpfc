import * as echarts from 'echarts'

let chart

const showSensorStatsChart = (div, stats) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 40,
      buttom: 0,
    },
    legend: {
      data: ['PS', 'Count'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'PS',
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          margin: 2,
        },
      },
      {
        type: 'value',
        name: 'Count',
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          margin: 2,
        },
      },
    ],
    series: [
      {
        name: 'PS',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Count',
        type: 'bar',
        large: true,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
  stats.forEach((s) => {
    const t = new Date(s.Time / (1000 * 1000))
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, s.PS],
    })
    option.series[1].data.push({
      name,
      value: [t, s.Count],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showSensorCpuMemChart = (div, monitor) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        return (
          params[0].name +
          '<br>' +
          params[0].seriesName +
          ':' +
          params[0].value[1].toFixed(2) +
          '<br>' +
          params[1].seriesName +
          ':' +
          params[1].value[1].toFixed(2)
        )
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: ['CPU', 'Mem'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: '%',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: 'CPU',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Mem',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.Time / (1000 * 1000))
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, m.CPU],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Mem],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showSensorNetChart = (div, monitor) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        return (
          params[0].name +
          '<br>' +
          params[0].seriesName +
          ':' +
          params[0].value[1].toFixed(2) +
          '<br>' +
          params[1].seriesName +
          ':' +
          params[1].value[1].toFixed(2)
        )
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 40,
      buttom: 0,
    },
    legend: {
      data: ['TxSpeed', 'RxSpeed'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'Mbps',
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          margin: 2,
        },
      },
    ],
    series: [
      {
        name: 'TxSpeed',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'RxSpeed',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.Time / (1000 * 1000))
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, m.TxSpeed],
    })
    option.series[1].data.push({
      name,
      value: [t, m.RxSpeed],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showSensorProcChart = (div, monitor) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 40,
      buttom: 0,
    },
    legend: {
      data: ['Load', 'Process'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'Load',
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          margin: 2,
        },
      },
      {
        type: 'value',
        name: 'Count',
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          margin: 2,
        },
      },
    ],
    series: [
      {
        name: 'Load',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Process',
        type: 'bar',
        large: true,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.Time / (1000 * 1000))
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, m.Load],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Process],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showRSSITime3DChart = (div, wifi, list, filter) => {
  const data = []
  const cat = []
  list.forEach((i) => {
    if (!i.RSSI || i.RSSI.length < 1) {
      return
    }
    if (filter.address && !i.Address.includes(filter.address)) {
      return
    }
    if (filter.addressType && !i.AddressType.includes(filter.addressType)) {
      return
    }
    if (filter.name && !i.Name.includes(filter.name)) {
      return
    }
    if (filter.host && !i.Host.includes(filter.host)) {
      return
    }
    if (filter.vendor && !i.Vendor.includes(filter.vendor)) {
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
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
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
      if (filter.bssid && !i.BSSID.includes(filter.bssid)) {
        return
      }
      if (filter.ssid && !i.SSID.includes(filter.ssid)) {
        return
      }
    } else {
      if (filter.address && !i.Address.includes(filter.address)) {
        return
      }
      if (filter.addressType && !i.AddressType.includes(filter.addressType)) {
        return
      }
      if (filter.name && !i.Name.includes(filter.name)) {
        return
      }
      if (filter.vendor && !i.Vendor.includes(filter.vendor)) {
        return
      }
    }
    if (filter.host && !i.Host.includes(filter.host)) {
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

const showEnv3DChart = (div, type, list) => {
  const data = []
  const cat = []
  list.forEach((i) => {
    if (!i.EnvData || i.EnvData.length < 1) {
      return
    }
    const id = i.Host + ':' + i.Address
    i.EnvData.forEach((e) => {
      data.push([
        id,
        e.Time / (1000 * 1000),
        e[type],
        getEnvLevel(e[type], type),
        i.Name,
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
      name: 'Host:Address',
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
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
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
      name: getEnvName(type),
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
        name: 'Env Data',
        type: 'scatter3D',
        dimensions: ['Host:Address', 'Time', type, 'Level', 'Name'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showEnv2DChart = (div, type, list) => {
  const cat = []
  const series = []
  list.forEach((i) => {
    if (!i.EnvData || i.EnvData.length < 1) {
      return
    }
    if (
      i.Name !== 'Rbt' &&
      type !== 'Temp' &&
      type !== 'Humidity' &&
      type !== 'Battery'
    ) {
      return
    }
    const data = []
    i.EnvData.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
      data.push({ name, value: [t, e[type]] })
    })
    const id = i.Host + ':' + i.Address
    series.push({
      name: id,
      type: 'line',
      large: true,
      symbol: 'none',
      data,
    })
    cat.push(id)
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}, {}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: cat,
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: getEnvUnit(type),
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series,
  }
  chart.setOption(option)
  chart.resize()
}

const getEnvLevel = (v, type) => {
  switch (type) {
    case 'Temp':
      if (v > 35.0) {
        return 0
      } else if (v > 30.0) {
        return 1
      } else if (v > 25.0) {
        return 2
      } else if (v < 10.0) {
        return 0
      }
      return 3
    case 'Humidity':
      if (v > 60.0) {
        return 0
      } else if (v < 40.0) {
        return 2
      }
      return 3
    case 'Illuminance':
      if (v > 300) {
        return 0
      } else if (v > 200) {
        return 1
      } else if (v > 100) {
        return 2
      } else if (v >= 1) {
        return 3
      }
      return 4
    case 'BarometricPressure':
      if (v > 1021) {
        return 0
      } else if (v > 1016) {
        return 1
      } else if (v > 1010) {
        return 3
      } else if (v > 1000) {
        return 2
      }
      return 4
    case 'Sound':
      if (v > 80) {
        return 0
      } else if (v > 70) {
        return 1
      } else if (v > 60) {
        return 2
      } else if (v > 30) {
        return 3
      }
      return 4
    case 'ETVOC':
      if (v > 3000) {
        return 0
      } else if (v > 1000) {
        return 1
      } else if (v > 300) {
        return 2
      } else if (v > 0) {
        return 3
      }
      return 4
    case 'ECo2':
      if (v > 3000) {
        return 0
      } else if (v > 2000) {
        return 1
      } else if (v > 1000) {
        return 2
      } else if (v > 0) {
        return 3
      }
      return 4
    case 'Battery':
      if (v > 50) {
        return 3
      } else if (v > 30) {
        return 2
      } else if (v > 10) {
        return 1
      }
      return 0
    default:
      return 0
  }
}

const envTypeList = [
  { value: 'Temp', text: '気温(℃)', unit: '℃' },
  { value: 'Humidity', text: '湿度(%)', unit: '%' },
  { value: 'Illuminance', text: '照度(lx)', unit: 'lx' },
  { value: 'BarometricPressure', text: '気圧(hpa)', unit: 'hpa' },
  { value: 'Sound', text: '騒音(dB)' },
  { value: 'ETVOC', text: '総揮発性有機化合物(ppb)', unit: 'ppb' },
  { value: 'ECo2', text: '二酸化炭素濃度(ppm)', unit: 'ppm' },
  { value: 'Battery', text: '電池残量(%)', unit: '%' },
]

const getEnvName = (type) => {
  for (let i = 0; i < envTypeList.length; i++) {
    if (type === envTypeList[i].value) {
      return envTypeList[i].text
    }
  }
  return 'Invalid type=' + type
}

const getEnvUnit = (type) => {
  for (let i = 0; i < envTypeList.length; i++) {
    if (type === envTypeList[i].value) {
      return envTypeList[i].unit
    }
  }
  return 'Invalid type=' + type
}

export default (context, inject) => {
  inject('showSensorStatsChart', showSensorStatsChart)
  inject('showSensorCpuMemChart', showSensorCpuMemChart)
  inject('showSensorNetChart', showSensorNetChart)
  inject('showSensorProcChart', showSensorProcChart)
  inject('showRSSITime3DChart', showRSSITime3DChart)
  inject('showRSSILoc3DChart', showRSSILoc3DChart)
  inject('showEnv3DChart', showEnv3DChart)
  inject('showEnv2DChart', showEnv2DChart)
  inject('getEnvName', getEnvName)
  inject('envTypeList', envTypeList)
}
