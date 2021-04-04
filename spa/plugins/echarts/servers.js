import * as echarts from 'echarts'
import WorldData from 'world-map-geojson'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

const showServerMapChart = (div, servers) => {
  const chart = echarts.init(document.getElementById(div))
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
  if (!servers) {
    return
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
    option.series[0].data.push({
      name: a[3] + '/' + a[0],
      value: [a[2] * 1.0, a[1] * 1.0, locMap[k]],
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showCountryChart = (div, list) => {
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
      name: '件数',
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
        stack: '件数',
        data: [],
      },
      {
        name: '33-41',
        type: 'bar',
        stack: '件数',
        data: [],
      },
      {
        name: '42-50',
        type: 'bar',
        stack: '件数',
        data: [],
      },
      {
        name: '51-66',
        type: 'bar',
        stack: '件数',
        data: [],
      },
      {
        name: '67以上',
        type: 'bar',
        stack: '件数',
        data: [],
      },
    ],
  }
  if (!list) {
    return
  }
  const data = {}
  list.forEach((e) => {
    let c = e.Country
    if (!c) {
      if (e.Loc.indexOf('LOCAL') === 0) {
        return
      }
      c = 'Unknown'
    }
    if (!data[c]) {
      data[c] = [0, 0, 0, 0, 0, 0]
    }
    data[c][0]++
    const si = getScoreIndex(e.Score)
    data[c][si]++
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
  inject('showServerMapChart', showServerMapChart)
  inject('showCountryChart', showCountryChart)
}
