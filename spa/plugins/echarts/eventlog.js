import * as echarts from 'echarts'

let chart

const showLogHeatmap = (div, logs) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const hours = [
    '0時',
    '1時',
    '2時',
    '3時',
    '4時',
    '5時',
    '6時',
    '7時',
    '8時',
    '9時',
    '10時',
    '11時',
    '12時',
    '13時',
    '14時',
    '15時',
    '16時',
    '17時',
    '18時',
    '19時',
    '20時',
    '21時',
    '22時',
    '23時',
  ]
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
    grid: {
      left: '10%',
      right: '5%',
      top: 30,
      buttom: 0,
    },
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
      trigger: 'item',
      formatter(params) {
        return (
          params.name +
          ' ' +
          params.data[1] +
          '時 : ' +
          params.data[2].toFixed(1)
        )
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    xAxis: {
      type: 'category',
      name: '日付',
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
      data: [],
    },
    yAxis: {
      type: 'category',
      name: '時間帯',
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
      data: hours,
    },
    visualMap: {
      min: Infinity,
      max: -Infinity,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    series: [
      {
        name: 'Score',
        type: 'heatmap',
        data: [],
        emphasis: {
          itemStyle: {
            borderColor: '#ccc',
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  }
  if (logs) {
    let nD = 0
    let nH = 0
    let x = -1
    let sum = 0
    logs.sort((a, b) => a.Time - b.Time)
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000))
      if (nD === 0) {
        nH = t.getHours()
        nD = t.getDate()
        option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}'))
        x++
        sum++
        return
      }
      if (t.getHours() !== nH) {
        if (nD !== t.getDate()) {
          option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}'))
          nD = t.getDate()
          x++
        }
        option.series[0].data.push([x, t.getHours(), sum])
        if (option.visualMap.min > sum) {
          option.visualMap.min = sum
        }
        if (option.visualMap.max < sum) {
          option.visualMap.max = sum
        }
        sum = 0
        nH = t.getHours()
      }
      sum++
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showEventLogStateChart = (div, logs) => {
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
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#33a02c', '#1f78b4', '#bbb'],
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
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      data: ['重度', '軽度', '注意', '正常', '復帰', 'その他'],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: '状態別',
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [
          { name: '重度', value: 0 },
          { name: '軽度', value: 0 },
          { name: '注意', value: 0 },
          { name: '正常', value: 0 },
          { name: '復帰', value: 0 },
          { name: 'その他', value: 0 },
        ],
      },
    ],
  }
  if (logs) {
    logs.forEach((l) => {
      switch (l.Level) {
        case 'high':
          option.series[0].data[0].value++
          break
        case 'low':
          option.series[0].data[1].value++
          break
        case 'warn':
          option.series[0].data[2].value++
          break
        case 'normal':
          option.series[0].data[3].value++
          break
        case 'repair':
          option.series[0].data[4].value++
          break
        default:
          option.series[0].data[5].value++
      }
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showEventLogTimeChart = (div, type, logs) => {
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
        saveAsImage: { name: 'twsnmp_event_log' + type },
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
      right: '5%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: '日時',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
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
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: type === '' ? '稼働率%' : '使用率%',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
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
    series: [
      {
        color: '#1f78b4',
        type: 'line',
        name: type === '' ? '稼働率' : '使用率',
        showSymbol: false,
        data: [],
      },
    ],
  }
  if (logs) {
    logs.forEach((l) => {
      if (l.Type !== type) {
        return
      }
      const t = new Date(l.Time / (1000 * 1000))
      const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
      const m = l.Event.match(/[0-9.]+%/)
      if (!m || m.length < 1) {
        return
      }
      const val = m[0].replace('%', '') * 1.0
      option.series[0].data.push({
        name: ts,
        value: [t, val],
      })
    })
  }
  chart.setOption(option)
  chart.resize()
}

const getEventLogNodeList = (logs) => {
  const m = new Map()
  logs.forEach((l) => {
    if (!l.NodeID) {
      return
    }
    let e = m.get(l.NodeID)
    if (!e) {
      m.set(l.NodeID, {
        Name: l.NodeName,
        total: 0,
        high: 0,
        low: 0,
        warn: 0,
        normal: 0,
        repair: 0,
        other: 0,
      })
      e = m.get(l.NodeID)
      if (!e) {
        return
      }
    }
    e.total++
    switch (l.Level) {
      case 'high':
        e.high++
        break
      case 'low':
        e.low++
        break
      case 'warn':
        e.warn++
        break
      case 'normal':
        e.normal++
        break
      case 'repair':
        e.repair++
        break
      default:
        e.other++
    }
  })
  const r = Array.from(m.values())
  return r
}

const showEventLogNodeChart = (div, logs) => {
  const list = getEventLogNodeList(logs)
  const high = []
  const low = []
  const warn = []
  const normal = []
  const repair = []
  const other = []
  const category = []
  list.sort((a, b) => b.total - a.total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    high.push(list[i].high)
    low.push(list[i].low)
    warn.push(list[i].warn)
    normal.push(list[i].normal)
    repair.push(list[i].repair)
    other.push(list[i].other)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption({
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
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#33a02c', '#1f78b4', '#bbb'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['重度', '軽度', '注意', '正常', '復帰', 'その他'],
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_eventlog_state' },
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '20%',
      right: '10%',
      top: '10%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: '件数',
    },
    yAxis: {
      type: 'category',
      data: category,
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
        name: '重度',
        type: 'bar',
        stack: '件数',
        data: high,
      },
      {
        name: '軽度',
        type: 'bar',
        stack: '件数',
        data: low,
      },
      {
        name: '注意',
        type: 'bar',
        stack: '件数',
        data: warn,
      },
      {
        name: '正常',
        type: 'bar',
        stack: '件数',
        data: normal,
      },
      {
        name: '復帰',
        type: 'bar',
        stack: '件数',
        data: repair,
      },
      {
        name: 'その他',
        type: 'bar',
        stack: '件数',
        data: other,
      },
    ],
  })
  chart.resize()
}

export default (context, inject) => {
  inject('showEventLogStateChart', showEventLogStateChart)
  inject('showEventLogNodeChart', showEventLogNodeChart)
  inject('showEventLogTimeChart', showEventLogTimeChart)
  inject('showLogHeatmap', showLogHeatmap)
}
