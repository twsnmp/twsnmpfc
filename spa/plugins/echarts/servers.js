import * as echarts from 'echarts'
import WorldData from 'world-map-geojson'

let chart

const makeServersChart = (div) => {
  chart = echarts.init(document.getElementById(div))
  echarts.registerMap('world', WorldData)
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
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    geo: {
      map: 'world',
      silent: true,
      emphasis: {
        label: {
          show: false,
          areaColor: '#eee',
        },
      },
      itemStyle: {
        borderWidth: 0.2,
        borderColor: '#404a59',
      },
      roam: true,
    },
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        return params.name + ' : ' + params.value[2]
      },
    },
    series: [
      {
        type: 'scatter',
        coordinateSystem: 'geo',
        label: {
          formatter: '{b}',
          position: 'right',
          color: '#eee',
          show: false,
        },
        emphasis: {
          label: {
            show: true,
          },
        },
        symbolSize: 6,
        itemStyle: {
          color: (params) => {
            const s = params.data.value[2]
            if (s > 66) {
              return '#1f78b4'
            } else if (s > 50) {
              return '#a6cee3'
            } else if (s > 42) {
              return '#dfdf22'
            } else if (s > 33) {
              return '#fb9a99'
            } else if (s <= 0) {
              return '#aaa'
            }
            return '#e31a1c'
          },
        },
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showServersChart = (servers) => {
  if (!servers) {
    return
  }
  const opt = {
    series: [{ data: [] }],
  }
  const locMap = {}
  servers.forEach((s) => {
    if (locMap.length > 10000) {
      return
    }
    const loc = s.Loc
    if (!loc || loc.indexOf('LOCAL') === 0) {
      return
    }
    if (!locMap[loc] || locMap[loc] > s.Score) {
      locMap[loc] = s.Score
    }
  })
  for (const k in locMap) {
    const a = k.split(',')
    if (a.length < 4 || !a[1]) {
      continue
    }
    opt.series[0].data.push({
      name: a[3] + '/' + a[0],
      value: [a[2] * 1.0, a[1] * 1.0, locMap[k]],
    })
  }
  chart.setOption(opt)
  chart.resize()
}

export default (context, inject) => {
  inject('makeServersChart', makeServersChart)
  inject('showServersChart', showServersChart)
}
