import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import * as numeral from 'numeral'

import WorldData from 'world-map-geojson'

let chart

const getPingChartOption = () => {
  return {
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
        saveAsImage: { name: 'twsnmp_ping' },
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
      top: 60,
      buttom: 0,
    },
    legend: {
      data: ['応答時間(秒)', '送信TTL', '受信TTL'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
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
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
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
    yAxis: [
      {
        type: 'value',
        name: '応答時間(秒)',
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
      {
        type: 'value',
        name: 'TTL',
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
    ],
    series: [
      {
        name: '応答時間(秒)',
        color: '#1f78b4',
        type: 'line',
        showSymbol: false,
        data: [],
      },
      {
        name: '送信TTL',
        color: '#dfdf22',
        type: 'line',
        showSymbol: false,
        yAxisIndex: 1,
        data: [],
      },
      {
        name: '受信TTL',
        color: '#e31a1c',
        type: 'line',
        showSymbol: false,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
}

const showPing3DChart = (div, results) => {
  if (chart) {
    chart.dispose()
  }
  let maxRtt = 0.0
  const data = []
  results.forEach((r) => {
    if (r.Stat !== 1) {
      return
    }
    const t = new Date(r.TimeStamp * 1000)
    const rtt = r.Time / (1000 * 1000 * 1000)
    if (rtt > maxRtt) {
      maxRtt = rtt
    }
    data.push([r.Size, t, rtt])
  })
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
        saveAsImage: { name: 'twsnmp_ping_3d' },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: maxRtt,
      dimension: 2,
      inRange: {
        color: [
          '#1710c0',
          '#0b9df0',
          '#00fea8',
          '#00ff0d',
          '#f5f811',
          '#f09a09',
          '#fe0300',
        ],
      },
    },
    xAxis3D: {
      type: 'value',
      name: 'サイズ',
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
      name: '日時',
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
      name: '応答時間(秒)',
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
        name: 'PING分析(3D)',
        type: 'scatter3D',
        symbolSize: 10,
        dimensions: ['サイズ', '日時', '応答時間(秒)'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showPingMapChart = (div, results) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  echarts.registerMap('world', WorldData)
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
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    geo: {
      map: 'world',
      silent: true,
      emphasis: {
        label: {
          show: false,
          areaColor: '#eee',
        },
      },
      itemStyle: {
        borderWidth: 0.2,
        borderColor: '#404a59',
      },
      roam: true,
    },
    toolbox: {
      iconStyle: {
        color: '#eee',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    series: [
      {
        type: 'scatter',
        coordinateSystem: 'geo',
        label: {
          formatter: '{b}',
          position: 'right',
          color: '#0ef',
          show: true,
          fontSize: 12,
        },
        emphasis: {
          label: {
            show: true,
          },
        },
        symbolSize: 10,
        itemStyle: {
          color: (params) => {
            const t = params.data.value[2]
            if (t < 0.005) {
              return '#1f78b4'
            } else if (t < 0.05) {
              return '#a6cee3'
            } else if (t < 0.2) {
              return '#dfdf22'
            } else if (t < 0.8) {
              return '#fb9a99'
            }
            return '#e31a1c'
          },
        },
        data: [],
      },
    ],
  }
  if (!results) {
    return
  }
  const locMap = {}
  results.forEach((e) => {
    const loc = e.Loc
    if (!loc || loc.indexOf('LOCAL') === 0) {
      return
    }
    if (!locMap[loc] || locMap[loc].time > e.Time) {
      locMap[loc] = {
        time: e.Time,
        ip: e.RecvSrc,
      }
    }
  })
  for (const k in locMap) {
    const a = k.split(',')
    if (a.length < 4 || !a[1]) {
      continue
    }
    option.series[0].data.push({
      name: locMap[k].ip + ':' + a[3] + '/' + a[0],
      value: [
        a[2] * 1.0,
        a[1] * 1.0,
        (locMap[k].time / (1000 * 1000 * 100)).toFixed(6),
      ],
    })
  }
  chart.setOption(option)
  chart.resize()
  chart.on('dblclick', (p) => {
    const url =
      'https://www.google.com/maps/search/?api=1&zoom=10&query=' +
      p.value[1] +
      ',' +
      p.value[0]
    window.open(url, '_blank')
  })
}

const showPingHistgram = (div, results) => {
  if (chart) {
    chart.dispose()
  }
  const data = []
  results.forEach((r) => {
    if (r.Stat !== 1) {
      return
    }
    data.push(r.Time / (1000 * 1000 * 1000))
  })
  const bins = ecStat.histogram(data)
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
        saveAsImage: { name: 'twsnmp_ping_histgram' },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      formatter(params) {
        const p = params[0]
        return p.value[0] + 'の回数:' + p.value[1]
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      scale: true,
      name: '応答時間(秒)',
      min: 0,
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
    },
    yAxis: {
      name: '回数',
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
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: bins.data,
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showPingLinearChart = (div, results) => {
  if (chart) {
    chart.dispose()
  }
  this.linear = echarts.init(document.getElementById('linear'))
  const data = []
  this.results.forEach((r) => {
    if (r.Stat !== 1) {
      return
    }
    data.push([r.Size, r.Time / (1000 * 1000 * 1000)])
  })
  const reg = ecStat.regression('linear', data)
  const speed =
    numeral(reg.parameter.gradient ? 1.0 / reg.parameter.gradient : 0.0).format(
      '0.00a'
    ) + 'bps'
  const delay = reg.parameter.intercept.toFixed(6) + `sec`
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
        saveAsImage: { name: 'twsnmp_ping_size' },
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
      type: 'value',
      name: 'サイズ',
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
        name: '応答時間(秒)',
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
        name: 'scatter',
        type: 'scatter',
        label: {
          emphasis: {
            show: true,
          },
        },
        data,
      },
      {
        name: 'line',
        type: 'line',
        showSymbol: false,
        data: reg.points,
        markPoint: {
          itemStyle: {
            normal: {
              color: 'transparent',
            },
          },
          label: {
            normal: {
              show: true,
              formatter: `回線速度=${speed} 遅延=${delay}`,
              textStyle: {
                color: '#eee',
                fontSize: 12,
              },
            },
          },
          data: [
            {
              coord: reg.points[reg.points.length - 1],
            },
          ],
        },
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('getPingChartOption', getPingChartOption)
  inject('showPingMapChart', showPingMapChart)
  inject('showPing3DChart', showPing3DChart)
  inject('showPingHistgram', showPingHistgram)
  inject('showPingLinearChart', showPingLinearChart)
}
