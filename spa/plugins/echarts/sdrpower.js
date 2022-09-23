import * as echarts from 'echarts'

let chart

const showSdrPower2DChart = (div, list) => {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_sdrpower' },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: [''],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      name: '周波数(MHz)',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter(value, index) {
          return (value / 1e6).toFixed(3)
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
    yAxis: [
      {
        name: 'DBm',
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
    ],
    series: [],
  }
  const catMap = new Map()
  list.forEach((l) => {
    l.forEach((e) => {
      const ts = echarts.time.format(
        new Date(e.Time * 1000),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const id = e.Host + ':' + ts
      if (catMap.has(id)) {
        catMap.get(id).push([e.Freq, e.Dbm])
      } else {
        catMap.set(id, [[e.Freq, e.Dbm]])
      }
    })
  })
  catMap.forEach((v, k) => {
    options.legend.data.push(k)
    options.series.push({
      type: 'line',
      name: k,
      showSymbol: false,
      data: v,
    })
  })
  chart.setOption(options)
  chart.resize()
}

const showSdrPower3DChart = (div, list) => {
  const data = []
  const mapx = new Map()
  let max = -100
  let min = 100
  list.forEach((l) => {
    l.forEach((e) => {
      const ts = echarts.time.format(
        new Date(e.Time * 1000),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const id = e.Host + ':' + ts
      data.push([id, e.Freq, e.Dbm])
      mapx.set(id, true)
      max = Math.max(max, e.Dbm)
      min = Math.min(min, e.Dbm)
    })
  })
  const catx = Array.from(mapx.keys())
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
      dimension: 2,
      min,
      max,
      inRange: {
        color: ['#1f78b4', '#dfdf22', '#fb9a99', '#e31a1c'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host:Time',
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
      type: 'value',
      name: 'Freq',
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
      name: 'DBm',
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
        name: 'SDR Power',
        type: 'scatter3D',
        symbolSize: 5,
        dimensions: ['Host:Time', 'Freq', 'DBm'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

export default (context, inject) => {
  inject('showSdrPower2DChart', showSdrPower2DChart)
  inject('showSdrPower3DChart', showSdrPower3DChart)
}
