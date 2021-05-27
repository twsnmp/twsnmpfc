import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

let chart

const makeNetFlowHistogram = (div) => {
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
    },
    yAxis: {
      name: '回数',
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        hoverAnimation: false,
        barWidth: '99.3%',
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showNetFlowHistogram = (logs, type) => {
  if (type === '') {
    type = 'size'
  }
  const data = []
  logs.forEach((l) => {
    if (type === 'size') {
      if (l.Packets > 0) {
        data.push(l.Bytes / l.Packets)
      }
    } else if (type === 'dur') {
      if (l.Duration >= 0.0) {
        data.push(l.Duration)
      }
    } else if (type === 'speed') {
      if (l.Duration > 0) {
        data.push(l.Bytes / l.Duration)
      }
    }
  })
  const bins = ecStat.histogram(data)
  chart.setOption({
    xAxis: {
      name: type,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize()
}

export default (context, inject) => {
  inject('makeNetFlowHistogram', makeNetFlowHistogram)
  inject('showNetFlowHistogram', showNetFlowHistogram)
}
