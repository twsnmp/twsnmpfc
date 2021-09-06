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

const showRSSITime3DChart = (div, wifi, list) => {
  const data = []
  const cat = []
  list.forEach((i) => {
    if (!i.RSSI || i.RSSI.length < 1) {
      return
    }
    i.RSSI.forEach((e) => {
      let l = 0
      if (e.Value >= 0) {
        l = 4
      } else if (e.Value >= -67) {
        l = 3
      } else if (e.Value >= -70) {
        l = 2
      } else if (e.Value >= -80) {
        l = 1
      }
      data.push([i.ID, e.Time, e.Value, l, wifi ? i.SSID : i.Name])
    })
    cat.push(i.ID)
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

const showRSSILoc3DChart = (div, wifi, list) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((i) => {
    if (!i.RSSI || i.RSSI.length < 1) {
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

export default (context, inject) => {
  inject('showSensorStatsChart', showSensorStatsChart)
  inject('showSensorCpuMemChart', showSensorCpuMemChart)
  inject('showSensorNetChart', showSensorNetChart)
  inject('showSensorProcChart', showSensorProcChart)
  inject('showRSSITime3DChart', showRSSITime3DChart)
  inject('showRSSILoc3DChart', showRSSILoc3DChart)
}
