import * as echarts from 'echarts'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

let chart
const showUsersChart = (div, users) => {
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
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#a6cee3', '#1f78b4'],
    legend: {
      orient: 'vertical',
      top: 50,
      right: 10,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上'],
    },
    grid: {
      top: '3%',
      left: '7%',
      right: '10%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: '人数',
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
        name: '32以下',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '33-41',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '42-50',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '51-66',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '67以上',
        type: 'bar',
        stack: '人数',
        data: [],
      },
    ],
  }
  if (!users) {
    return
  }
  const data = {}
  users.forEach((u) => {
    if (!data[u.Server]) {
      data[u.Server] = [0, 0, 0, 0, 0, 0]
    }
    data[u.Server][0]++
    const si = getScoreIndex(u.Score)
    data[u.Server][si]++
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
    for (let j = 0; j < 5; j++) {
      option.series[j].data.push(data[keys[i]][j + 1])
    }
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showUsersChart', showUsersChart)
}
