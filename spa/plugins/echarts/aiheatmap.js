import * as echarts from 'echarts'

const showAIHeatMap = (div, scores) => {
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
        return params.name + ' ' + params.data[1] + '時 : ' + params.data[2]
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
      max: 100,
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
      option.xAxis.data.push(echarts.format.formatTime('yyyy/MM/dd', t))
      nD = t.getDate()
      x++
    }
    option.series[0].data.push([x, t.getHours(), e[1]])
  })
  chart.setOption(option)
  chart.resize()
  chart.on('dblclick', function (params) {
    // const d = params.name + ' ' + params.data[1] + ':00:00'
  })
}

export default (context, inject) => {
  inject('showAIHeatMap', showAIHeatMap)
}
