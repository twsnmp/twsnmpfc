import * as echarts from 'echarts'
import 'echarts-gl'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

let chart
const showFlowsChart = (div, flows, filter, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [
    { name: 'RU' },
    { name: 'CN' },
    { name: 'US' },
    { name: 'JP' },
    { name: 'LOCAL' },
    { name: 'Other' },
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
    grid: {
      left: '7%',
      right: '5%',
      bottom: '5%',
      containLabel: true,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        return (
          params.name.replace(' > ', '<br/>') +
          '<br/>' +
          params.value.replaceAll(':', '<br/>')
        )
      },
      textStyle: {
        fontSize: 8,
      },
      position: 'bottom',
    },
    legend: [
      {
        orient: 'vertical',
        top: 50,
        right: 20,
        textStyle: {
          fontSize: 10,
          color: '#ccc',
        },
        data: categories.map(function (a) {
          return a.name
        }),
      },
    ],
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#a6cee3', '#1f78b4', '#999'],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [
      {
        type: 'graph',
        layout: layout || 'force',
        symbolSize: 6,
        categories,
        roam: true,
        label: {
          show: false,
        },
        data: [],
        links: [],
        lineStyle: {
          width: 1,
          curveness: 0,
        },
      },
    ],
  }
  if (!flows) {
    return false
  }
  let bOver = false
  const nodes = {}
  flows.forEach((f) => {
    if (option.series[0].links.length > 2000) {
      bOver = true
      return
    }
    if (!filterFlow(f, filter)) {
      return
    }
    const c = `${f.ClientName}(${f.Client})`
    const s = `${f.ServerName}(${f.Server})`
    if (!nodes[s]) {
      nodes[s] = {
        name: s,
        category: getLocCategory(f.ServerLoc),
        draggable: true,
        value: f.ServerLoc,
        label: {
          show: false,
        },
      }
    }
    if (!nodes[c]) {
      nodes[c] = {
        name: c,
        category: getLocCategory(f.ClientLoc),
        draggable: true,
        value: f.ClientLoc,
        label: {
          show: false,
        },
      }
    }
    option.series[0].links.push({
      source: c,
      target: s,
      value: f.ServiceInfo + ':' + f.Score.toFixed(2),
      lineStyle: {
        color: getScoreColor(f.Score),
      },
    })
  })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
  return bOver
}

const filterFlow = (f, filter) => {
  if (filter.client && !f.ClientName.includes(filter.client)) {
    return false
  }
  if (filter.server && !f.ServerName.includes(filter.server)) {
    return false
  }
  if (filter.country && !f.Country.includes(filter.country)) {
    return false
  }
  if (filter.service && !f.ServiceInfo.includes(filter.service)) {
    return false
  }
  return true
}

const getScoreColor = (s) => {
  if (s > 66) {
    return '#1f78b4'
  } else if (s > 50) {
    return '#a6cee3'
  } else if (s > 42) {
    return '#dfdf22'
  } else if (s > 33) {
    return '#fb9a99'
  }
  return '#e31a1c'
}

const getLocCategory = (l) => {
  if (!l) {
    return 4
  }
  const a = l.split(',')
  if (a.length < 2) {
    return 4
  }
  switch (a[0]) {
    case 'LOCAL':
      return 4
    case 'JP':
      return 3
    case 'US':
      return 2
    case 'CN':
      return 1
    case 'RU':
      return 0
  }
  return 5
}

const showFlows3DChart = (div, flows, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
    backgroundColor: '#000',
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上'],
    },
    globe: {
      baseTexture: '/images/world.topo.bathy.200401.jpg',
      heightTexture: '/images/bathymetry_bw_composite_4k.jpg',
      shading: 'lambert',
      light: {
        ambient: {
          intensity: 0.4,
        },
        main: {
          intensity: 0.4,
        },
      },
      viewControl: {
        autoRotate: false,
      },
    },
    series: [
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '32以下',
        lineStyle: {
          width: 2,
          color: '#e31a1c',
          opacity: 0.8,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '33-41',
        lineStyle: {
          width: 2,
          color: '#fb9a99',
          opacity: 0.5,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '42-50',
        lineStyle: {
          width: 1,
          color: '#dfdf22',
          opacity: 0.3,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '51-66',
        lineStyle: {
          width: 1,
          color: '#a6cee3',
          opacity: 0.1,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '67以上',
        lineStyle: {
          width: 1,
          color: '#1f78b4',
          opacity: 0.1,
        },
        data: [],
      },
    ],
  }
  if (!flows) {
    return
  }
  let count = 0
  flows.forEach((f) => {
    if (count > 20000) {
      return
    }
    if (!f.ServerLatLong && !f.ClientLatLong) {
      return
    }
    if (!filterFlow(f, filter)) {
      return
    }
    const s = getLatLong(f.ServerLatLong)
    const c = getLatLong(f.ClientLatLong)
    const si = getScoreIndex(f.Score) - 1
    if (si > 2 && count > 1000) {
      return
    }
    count++
    option.series[si].data.push([c, s])
  })
  chart.setOption(option)
  chart.resize()
}

const getLatLong = (loc) => {
  if (!loc) {
    return [139.548088, 35.856222]
  }
  const a = loc.split(',')
  if (a.length !== 2) {
    return [139.548088, 35.856222]
  }
  return [a[1], a[0]]
}

export default (context, inject) => {
  inject('showFlowsChart', showFlowsChart)
  inject('showFlows3DChart', showFlows3DChart)
  inject('filterFlow', filterFlow)
}
