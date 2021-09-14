import * as echarts from 'echarts'
import 'echarts-gl'

let chart

const showWinEventID3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.provider && !e.Provider.includes(filter.provider)) {
      return
    }
    if (filter.eventID && !e.EventID.toString(10).includes(filter.eventID)) {
      return
    }
    if (filter.channel && !e.Channel.includes(filter.channel)) {
      return
    }
    if (filter.level && e.Level !== filter.level) {
      return
    }
    const id = e.Provider + ':' + e.EventID
    data.push([e.Computer, id, e.Count, getWinEventLevel(e.Level), e.Channel])
    mapx.set(e.Computer, true)
    mapy.set(id, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
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
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      dimension: 3,
      min: 0,
      max: 2,
      inRange: {
        color: ['#e31a1c', '#dfdf22', '#1f78b4'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Computer',
      data: catx,
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
      type: 'category',
      name: 'Provider+EventID',
      data: caty,
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
    zAxis3D: {
      type: 'value',
      name: 'Count',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
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
        name: 'Windows EventID',
        type: 'scatter3D',
        dimensions: [
          'Computer',
          'Provider+EventID',
          'Count',
          'Level',
          'Chaneel',
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const getWinEventLevel = (l) => {
  if (l === 'error') {
    return 0
  }
  if (l === 'warn') {
    return 1
  }
  return 2
}

export default (context, inject) => {
  inject('showWinEventID3DChart', showWinEventID3DChart)
}
