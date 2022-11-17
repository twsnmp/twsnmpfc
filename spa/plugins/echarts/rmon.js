import * as echarts from 'echarts'

let chart

const showRMONBarChart = (div, type, xAxis, list, max) => {
  if (chart) {
    chart.dispose()
  }
  const yellow = max ? max * 0.8 : 80
  const red = max ? max * 0.9 : 90
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
        saveAsImage: { name: 'twsnmp_' + type },
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
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
      name: xAxis,
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
        name: type,
        type: 'bar',
        itemStyle: {
          color: (p) => {
            return p.value > red ? '#c00' : p.value > yellow ? '#cc0' : '#0cc'
          },
        },
        data: [],
      },
    ],
  }
  if (!list) {
    return
  }
  list.forEach((e) => {
    option.yAxis.data.push(e.Name)
    option.series[0].data.push(e.Value)
  })
  chart.setOption(option)
  chart.resize()
}

const showRMONSummary = (div, data) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const gaugeData = [
    {
      value: data.CPU.toFixed(1),
      name: 'CPU',
      title: {
        offsetCenter: ['-40%', '80%'],
      },
      detail: {
        offsetCenter: ['-40%', '95%'],
      },
    },
    {
      value: data.Mem.toFixed(1),
      name: 'Mem',
      title: {
        offsetCenter: ['0%', '80%'],
      },
      detail: {
        offsetCenter: ['0%', '95%'],
      },
    },
    {
      value: data.VM.toFixed(1),
      name: 'VM',
      title: {
        offsetCenter: ['40%', '80%'],
      },
      detail: {
        offsetCenter: ['40%', '95%'],
      },
    },
  ]
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
        saveAsImage: { name: 'twsnmp_hr' },
      },
    },
    color: ['#4575b4', '#abd9e9', '#FAC858'],
    series: [
      {
        type: 'gauge',
        anchor: {
          show: true,
          showAbove: true,
          size: 18,
          itemStyle: {
            color: '#FAC858',
          },
        },
        pointer: {
          icon: 'path://M2.9,0.7L2.9,0.7c1.4,0,2.6,1.2,2.6,2.6v115c0,1.4-1.2,2.6-2.6,2.6l0,0c-1.4,0-2.6-1.2-2.6-2.6V3.3C0.3,1.9,1.4,0.7,2.9,0.7z',
          width: 8,
          length: '80%',
          offsetCenter: [0, '8%'],
        },
        progress: {
          show: true,
          overlap: true,
          roundCap: true,
        },
        axisLine: {
          roundCap: true,
        },
        axisLabel: {
          color: '#ccc',
        },
        data: gaugeData,
        title: {
          fontSize: 12,
          color: '#ccc',
        },
        detail: {
          width: 40,
          height: 14,
          fontSize: 12,
          color: '#fff',
          backgroundColor: 'auto',
          borderRadius: 3,
          formatter: '{value}%',
        },
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showRMONBarChart', showRMONBarChart)
  inject('showRMONSummary', showRMONSummary)
}
