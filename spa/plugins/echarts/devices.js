import * as echarts from 'echarts'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

let chart
const showVendorChart = (div, list) => {
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
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上'],
    },
    grid: {
      top: '10%',
      left: '5%',
      right: '10%',
      bottom: '10%',
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
    ],
  }
  if (!list) {
    return
  }
  const data = {}
  list.forEach((d) => {
    if (!data[d.Vendor]) {
      data[d.Vendor] = [0, 0, 0, 0, 0, 0]
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
    option.yAxis.data.push(keys[i])
    for (let j = 0; j < 5; j++) {
      option.series[j].data.push(data[keys[i]][j + 1])
    }
  }
  chart.setOption(option)
  chart.resize()
}

const filterDevice = (d, filter) => {
  if (filter.mac && !d.ID.includes(filter.mac)) {
    return false
  }
  if (filter.name && !d.Name.includes(filter.name)) {
    return false
  }
  if (filter.ip && !d.IP.includes(filter.ip)) {
    return false
  }
  if (filter.vendor && !d.Vendor.includes(filter.vendor)) {
    return false
  }
  if (filter.excludeVM && d.Vendor.includes('VMware')) {
    return false
  }
  return true
}

export default (context, inject) => {
  inject('showVendorChart', showVendorChart)
  inject('filterDevice', filterDevice)
}
