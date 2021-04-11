import * as echarts from 'echarts'

const showAIHeatMap = (div, scores, cb) => {
  const chart = echarts.init(document.getElementById(div))
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
      min: 40,
      max: 80,
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
  if (!scores) {
    chart.setOption(option)
    chart.resize()
    return
  }
  let nD = 0
  let x = -1
  scores.forEach((e) => {
    const t = new Date(e[0] * 1000)
    if (nD !== t.getDate()) {
      option.xAxis.data.push(echarts.time.format(t, '{yyyy}/{MM}/{dd}'))
      nD = t.getDate()
      x++
    }
    option.series[0].data.push([x, t.getHours(), e[1]])
  })
  chart.setOption(option)
  chart.resize()
  chart.on('dblclick', function (params) {
    if (cb) {
      const ut =
        Date.parse(params.name + ' ' + params.data[1] + ':00:00') / 1000
      // eslint-disable-next-line standard/no-callback-literal
      cb({ UnixTime: ut })
    }
  })
}

const showAIPieChart = (div, scores) => {
  const chart = echarts.init(document.getElementById(div))
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
    color: ['#1f78b4', '#dfdf22', '#e31a1c'],
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      orient: 'vertical',
      left: 'right',
      data: ['正常', '注意', '異常'],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: '異常スコア',
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [
          { name: '正常', value: 0 },
          { name: '注意', value: 0 },
          { name: '異常', value: 0 },
        ],
      },
    ],
  }
  if (scores) {
    scores.forEach((e) => {
      if (e[1] > 66.0) {
        option.series[0].data[2].value++
      } else if (e[1] > 50.0) {
        option.series[0].data[1].value++
      } else {
        option.series[0].data[0].value++
      }
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showAITimeChart = (div, scores, cb) => {
  const chart = echarts.init(document.getElementById(div))
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
    tooltip: {
      trigger: 'axis',
      formatter(params) {
        const p = params[0]
        return p.name + ' : ' + p.value[1]
      },
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
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
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
      name: '異常スコア',
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
        showSymbol: false,
        hoverAnimation: false,
        data: [],
      },
    ],
  }
  if (scores) {
    scores.forEach((e) => {
      const t = new Date(e[0] * 1000)
      const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
      option.series[0].data.push({
        name: ts,
        value: [t, e[1]],
      })
    })
  }
  chart.setOption(option)
  chart.resize()
  chart.on('dblclick', function (params) {
    if (cb) {
      const ut = Date.parse(params.name) / 1000
      // eslint-disable-next-line standard/no-callback-literal
      cb({ UnixTime: ut })
    }
  })
}

export default (context, inject) => {
  inject('showAIHeatMap', showAIHeatMap)
  inject('showAIPieChart', showAIPieChart)
  inject('showAITimeChart', showAITimeChart)
}
