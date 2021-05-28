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

const showNetFlowHistogram = (div, logs, type) => {
  makeNetFlowHistogram(div)
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
const makeNetFlowCluster = (div) => {
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
    legend: {
      data: [],
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'value',
    },
    yAxis: {
      type: 'value',
    },
    series: [],
  }
  chart.setOption(option)
  chart.resize()
}

const showNetFlowCluster = (div, logs, type, cluster) => {
  makeNetFlowCluster(div)
  if (type === '') {
    type = 'size-bps'
  }
  if (cluster < 2 || cluster > 10) {
    cluster = 2
  }
  const data = []
  logs.forEach((l) => {
    if (l.Packets <= 0 || l.Duration <= 0.0) {
      return
    }
    if (type === 'size-bps') {
      data.push([l.Bytes / l.Packets, l.Bytes / l.Duration])
    } else if (type === 'size-pps') {
      data.push([l.Bytes / l.Packets, l.Packets / l.Duration])
    } else if (type === 'pps-bps') {
      data.push([l.Packets / l.Duration, l.Bytes / l.Duration])
    } else if (type === 'sport-dport') {
      data.push([l.SrcPort, l.DstPort])
    }
  })
  const result = ecStat.clustering.hierarchicalKMeans(data, {
    clusterCount: cluster,
    stepByStep: false,
    outputType: 'multiple',
    outputClusterIndexDimension: cluster,
  })
  const centroids = result.centroids
  const ptsInCluster = result.pointsInCluster
  const series = []
  for (let i = 0; i < centroids.length; i++) {
    series.push({
      name: 'cluster' + (i + 1),
      type: 'scatter',
      data: ptsInCluster[i],
      markPoint: {
        symbolSize: 30,
        label: {
          show: true,
          position: 'top',
          formatter: (params) => {
            return (
              Math.round(params.data.coord[0] * 100) / 100 +
              ' / ' +
              Math.round(params.data.coord[1] * 100) / 100
            )
          },
          color: '#ccc',
          fontSize: 10,
        },
        data: [
          {
            coord: centroids[i],
          },
        ],
      },
    })
  }
  chart.setOption({
    legend: {
      data: [],
    },
    series,
  })
  chart.resize()
}

export default (context, inject) => {
  inject('showNetFlowHistogram', showNetFlowHistogram)
  inject('showNetFlowCluster', showNetFlowCluster)
}
