import * as echarts from 'echarts'

let timeLineChart

const showOTelTimeline = (div, data) => {
  if (timeLineChart) {
    timeLineChart.dispose()
  }
  timeLineChart = echarts.init(document.getElementById(div), 'dark')
  const option = {
    title: {
      show: false,
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '25%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    xAxis: {
      type: 'value',
      name: 'mSec',
      splitLine: { show: false },
    },
    yAxis: {
      type: 'category',
      splitLine: { show: false },
      data: [],
    },
    series: [
      {
        name: 'Start',
        type: 'bar',
        stack: 'time',
        itemStyle: {
          borderColor: 'transparent',
          color: 'transparent',
        },
        emphasis: {
          itemStyle: {
            borderColor: 'transparent',
            color: 'transparent',
          },
        },
        data: [],
      },
      {
        name: 'Duration',
        type: 'bar',
        stack: 'time',
        label: {
          show: true,
          position: 'inside',
          formatter: (p) => {
            return p.value.toFixed(3) + ' mSec'
          },
          fontSize: 8,
        },
        data: [],
      },
    ],
  }
  if (data && data.Spans) {
    data.Spans.sort((a, b) => {
      if (a.ParentSpanID === '') {
        if (b.ParentSpanID === '') {
          if (a.Start === b.Start) {
            if (a.Dur > b.Dur) {
              return 1
            }
            return -1
          } else if (a.Start < b.Start) {
            return 1
          }
          return -1
        }
        return 1
      }
      if (a.Start === b.Start) {
        if (a.Dur > b.Dur) {
          return 1
        }
        return -1
      } else if (a.Start < b.Start) {
        return 1
      }
      return -1
    })
    const st = data.Spans[data.Spans.length - 1].Start
    const colors = [
      '#5470c6',
      '#91cc75',
      '#fac858',
      '#ee6666',
      '#73c0de',
      '#3ba272',
      '#fc8452',
      '#9a60b4',
      '#ea7ccc',
    ]
    data.Spans.forEach((s) => {
      option.yAxis.data.push(s.Name)
      option.series[0].data.push((s.Start - st) / (1000.0 * 1000))
      option.series[1].data.push({
        value: (s.End - s.Start) / (1000.0 * 1000),
        itemStyle: {
          color: colors[option.series[1].data.length % colors.length],
        },
      })
    })
  }
  timeLineChart.setOption(option)
  timeLineChart.resize()
  return timeLineChart
}

let traceChart

const showOTelTrace = (div, traces) => {
  if (traceChart) {
    traceChart.dispose()
  }
  traceChart = echarts.init(document.getElementById(div), 'dark')
  const option = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
      },
    },
    dataZoom: [{}],
    tooltip: {
      formatter(params) {
        let ts = params.data[1].toFixed(3) + ' Sec'
        if (params.data[1] < 0.001) {
          ts = (params.data[1] * 1000 * 1000).toFixed(3) + ' uSec'
        } else if (params.data[1] < 1.0) {
          ts = (params.data[1] * 1000).toFixed(3) + ' mSec'
        }
        return (
          'Time     : ' +
          echarts.time.format(params.data[0], '{HH}:{mm}:{ss}.{SSS}', false) +
          '</br>' +
          'TraceID  : ' +
          params.data[3] +
          '</br>' +
          'Duration : ' +
          ts +
          '</br>' +
          'Span     : ' +
          params.data[2]
        )
      },
    },
    grid: {
      left: '5%',
      right: '5%',
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: 'Time',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{HH}:{mm}:{ss}', false)
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
    visualMap: {
      min: 0,
      max: 0.5,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
      dimension: 1,
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
    yAxis: {
      type: 'value',
      name: 'Sec',
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
        name: 'Trace',
        data: [],
        type: 'scatter',
        symbolSize: (data) => {
          if (data[2] > 13) {
            return 16
          }
          return data[2] + 3
        },
      },
    ],
  }
  traces.forEach((t) => {
    const ts = new Date(t.Start / (1000 * 1000))
    if (option.visualMap.max < t.Dur) {
      option.visualMap.max = t.Dur
    }
    option.series[0].data.push([ts, t.Dur, t.NumSpan, t.TraceID])
  })
  traceChart.setOption(option)
  traceChart.resize()
  return traceChart
}

let dagChart

const showOTelDAG = (div, data) => {
  if (dagChart) {
    dagChart.dispose()
  }
  dagChart = echarts.init(document.getElementById(div), 'dark')
  const zoom = data && data.Nodes.length > 0 ? 70 / data.Nodes.length : 1
  const option = {
    title: {
      show: false,
    },
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    tooltip: {
      trigger: 'item',
      textStyle: {
        fontSize: 8,
      },
      position: 'bottom',
    },
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [
      {
        zoom,
        name: 'OTel DAG',
        type: 'graph',
        layout: 'force',
        edgeSymbol: ['none', 'arrow'],
        edgeSymbolSize: [4, 12],
        data: [],
        links: [],
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 10,
          fontStyle: 'normal',
          color: '#ccc',
        },
        lineStyle: {
          color: '#5152e9',
        },
      },
    ],
  }
  const colorMap = [
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
  ]

  if (data && data.Nodes) {
    let nodeMax = 1
    data.Nodes.forEach((n) => {
      option.series[0].data.push({
        name: n.Name,
        draggable: true,
        symbolSize: n.Count,
        label: { show: true },
        itemStyle: {
          color: 'blue',
        },
      })
      if (nodeMax < n.Count) {
        nodeMax = n.Count
      }
    })
    for (let i = 0; i < option.series[0].data.length; i++) {
      const r = option.series[0].data[i].symbolSize / nodeMax
      option.series[0].data[i].symbolSize = 4 + 20 * r
      option.series[0].data[i].itemStyle.color =
        colorMap[Math.floor(r * (colorMap.length - 1))]
    }
    let linkMax = 1
    data.Links.forEach((l) => {
      option.series[0].links.push({
        source: l.Src,
        target: l.Dst,
        lineStyle: {
          width: l.Count,
        },
      })
      if (linkMax < l.Count) {
        linkMax = l.Count
      }
    })
    for (let i = 0; i < option.series[0].links.length; i++) {
      const r = option.series[0].links[i].lineStyle.width / linkMax
      option.series[0].links[i].lineStyle.width = 1 + 8 * r
      option.series[0].links[i].lineStyle.color =
        colorMap[Math.floor(r * (colorMap.length - 1))]
    }
  }
  dagChart.setOption(option)
  dagChart.resize()
  return dagChart
}

let histogramChart

const showOTelHistogram = (div, data) => {
  if (histogramChart) {
    histogramChart.dispose()
  }
  histogramChart = echarts.init(document.getElementById(div), 'dark')
  histogramChart.setOption({
    title: {
      show: false,
    },
    xAxis: {
      type: 'category',
      data: data.ExplicitBounds,
    },
    yAxis: {
      type: 'value',
      name: 'count',
    },
    series: [
      {
        data: data.BucketCounts,
        type: 'bar',
      },
    ],
  })
  histogramChart.resize()
  return histogramChart
}

let timeChart

const showOTelTimeChart = (div, data, key, type) => {
  if (timeChart) {
    timeChart.dispose()
  }
  timeChart = echarts.init(document.getElementById(div), 'dark')
  const option = {
    title: {
      show: false,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        dataZoom: {},
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
      left: '5%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: [],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: 'Time',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value) => {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}', false)
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
    yAxis: {},
    series: [],
  }
  switch (type) {
    case 'Sum':
    case 'Gauge':
      option.series = [
        {
          name: type,
          type: 'bar',
          large: true,
          data: [],
        },
      ]
      option.yAxis = {
        type: 'value',
        name: type,
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
      }
      option.legend.data = [type]
      break
    case 'Histogram':
      option.series = [
        {
          name: 'Count',
          type: 'bar',
          large: true,
          yAxisIndex: 1,
          data: [],
        },
        {
          name: 'Sum',
          type: 'line',
          large: true,
          data: [],
        },
        {
          name: 'Max',
          type: 'line',
          large: true,
          data: [],
        },
        {
          name: 'Min',
          type: 'line',
          large: true,
          data: [],
        },
      ]
      option.yAxis = [
        {
          type: 'value',
          name: 'Sum',
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
      ]
      option.legend.data = ['Count', 'Sum', 'Max', 'Min']
      break
  }
  data.forEach((m) => {
    const t = new Date(m.Time / (1000 * 1000))
    const name = echarts.time.format(
      t,
      '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}',
      false
    )
    if (key !== m.Attributes.join(' ')) {
      return
    }
    switch (type) {
      case 'Sum':
        option.series[0].data.push({
          name,
          value: [t, m.Sum],
        })
        break
      case 'Gauge':
        option.series[0].data.push({
          name,
          value: [t, m.Gauge],
        })
        break
      case 'Histogram':
        option.series[0].data.push({
          name,
          value: [t, m.Count],
        })
        option.series[1].data.push({
          name,
          value: [t, m.Sum],
        })
        option.series[2].data.push({
          name,
          value: [t, m.Max],
        })
        option.series[3].data.push({
          name,
          value: [t, m.Min],
        })
        break
    }
  })
  timeChart.setOption(option)
  timeChart.resize()
  return timeChart
}

export default (context, inject) => {
  inject('showOTelTimeline', showOTelTimeline)
  inject('showOTelTrace', showOTelTrace)
  inject('showOTelHistogram', showOTelHistogram)
  inject('showOTelTimeChart', showOTelTimeChart)
  inject('showOTelDAG', showOTelDAG)
}
