import * as echarts from 'echarts'

let chart

const makeDevicesChart = (div) => {
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
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#a6cee3', '#1f78b4', '#999'],
    legend: {
      orient: 'vertical',
      top: 50,
      right: 10,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上', '調査中'],
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
      name: '台数',
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
        stack: '台数',
        data: [],
      },
      {
        name: '33-41',
        type: 'bar',
        stack: '台数',
        data: [],
      },
      {
        name: '42-50',
        type: 'bar',
        stack: '台数',
        data: [],
      },
      {
        name: '51-66',
        type: 'bar',
        stack: '台数',
        data: [],
      },
      {
        name: '67以上',
        type: 'bar',
        stack: '台数',
        data: [],
      },
      {
        name: '調査中',
        type: 'bar',
        stack: '台数',
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const getScoreIndex = (s) => {
  if (s > 66) {
    return 5
  } else if (s > 50) {
    return 4
  } else if (s > 42) {
    return 3
  } else if (s > 33) {
    return 2
  } else if (s <= 0) {
    return 6
  }
  return 1
}

const showDevicesChart = (devices) => {
  if (!devices) {
    return
  }
  const opt = {
    yAxis: {
      data: [],
    },
    series: [
      { data: [] },
      { data: [] },
      { data: [] },
      { data: [] },
      { data: [] },
      { data: [] },
    ],
  }
  const data = {}
  devices.forEach((d) => {
    if (!data[d.Vendor]) {
      data[d.Vendor] = [0, 0, 0, 0, 0, 0, 0]
    }
    data[d.Vendor][0]++
    const si = getScoreIndex(d.Score)
    data[d.Vendor][si]++
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
    opt.yAxis.data.push(keys[i])
    for (let j = 0; j < 6; j++) {
      opt.series[j].data.push(data[keys[i]][j + 1])
    }
  }
  chart.setOption(opt)
  chart.resize()
}

export default (context, inject) => {
  inject('makeDevicesChart', makeDevicesChart)
  inject('showDevicesChart', showDevicesChart)
}
