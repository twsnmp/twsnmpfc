import * as echarts from 'echarts'

const showEtherTypeChart = (div, etherType) => {
  const chart = echarts.init(document.getElementById(div))
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
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [],
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
      name: '回数',
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
    series: [],
  }
  if (!etherType) {
    return
  }
  const hostMap = new Map()
  const nameMap = new Map()

  etherType.forEach((e) => {
    if (!hostMap.has(e.Host)) {
      hostMap.set(e.Host, new Map())
    }
    hostMap.get(e.Host).set(e.Name, e.Count)
    nameMap.set(e.Name, true)
  })
  option.yAxis.data = Array.from(hostMap.keys())
  option.legend.data = Array.from(nameMap.keys())
  for (let l = 0; l < option.legend.data.length; l++) {
    const data = []
    for (let y = 0; y < option.yAxis.data.length; y++) {
      const c = hostMap.get(option.yAxis.data[y]).has(option.legend.data[l])
        ? hostMap.get(option.yAxis.data[y]).get(option.legend.data[l])
        : 0
      data.push(c)
    }
    option.series.push({
      name: option.legend.data[l],
      type: 'bar',
      stack: '回数',
      data,
    })
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showEtherTypeChart', showEtherTypeChart)
}
