import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

let chart
const showSysResChart = (div, monitor) => {
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
          params[1].value[1].toFixed(2) +
          '<br>' +
          params[2].seriesName +
          ':' +
          params[2].value[1].toFixed(2)
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
      data: ['CPU', 'Mem', 'Disk'],
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
      {
        name: 'Disk',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.At * 1000)
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, m.CPU],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Mem],
    })
    option.series[2].data.push({
      name,
      value: [t, m.Disk],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showSysNetChart = (div, monitor) => {
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
      data: ['Speed', 'Connection'],
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
        name: 'bps',
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
        name: 'count',
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
        name: 'Speed',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
      {
        name: 'Connection',
        type: 'bar',
        large: true,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
  monitor.forEach((m) => {
    const t = new Date(m.At * 1000)
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, m.Net],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Conn],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showSysProcChart = (div, monitor) => {
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
        const p = params[0]
        return p.name + ' CPU: ' + p.value[0]
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
    const t = new Date(m.At * 1000)
    const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    option.series[0].data.push({
      name,
      value: [t, m.Load],
    })
    option.series[1].data.push({
      name,
      value: [t, m.Proc],
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showDiskUsageForecast = (div, monitor) => {
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
        name: 'ディスク使用量(%)',
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
    visualMap: {
      top: 50,
      right: 10,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      pieces: [
        {
          gt: 0,
          lte: 90,
          color: '#0062f7',
        },
        {
          gt: 90,
          lte: 100,
          color: '#FBFB0F',
        },
        {
          gt: 100,
          color: '#FF2000',
        },
      ],
      outOfRange: {
        color: '#eee',
      },
    },
    series: [
      {
        name: 'ディスク使用量',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
    ],
  }
  if (monitor) {
    const data = []
    monitor.forEach((m) => {
      data.push([m.At * 1000, m.Disk])
    })
    const reg = ecStat.regression('linear', data)
    const sd = Math.floor(Date.now() / (24 * 3600 * 1000))
    for (let d = sd; d < sd + 365; d++) {
      const x = d * 24 * 3600 * 1000
      const t = new Date(x)
      option.series[0].data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, reg.parameter.intercept + reg.parameter.gradient * x],
      })
    }
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showSysResChart', showSysResChart)
  inject('showSysNetChart', showSysNetChart)
  inject('showSysProcChart', showSysProcChart)
  inject('showDiskUsageForecast', showDiskUsageForecast)
}
