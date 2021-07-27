import * as echarts from 'echarts'
import 'echarts-gl'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

let chart
const showEtherTypeChart = (div, etherType) => {
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

const showTLSFlowsChart = (div, tls, filter) => {
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
        return params.name + ':' + params.value
      },
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
        layout: 'force',
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
  if (!tls) {
    return false
  }
  const nodes = {}
  let over = false
  tls.forEach((f) => {
    if (option.series[0].links.length > 2000) {
      over = true
      return
    }
    if (filter.Service && f.Service !== filter.Service) {
      return
    }
    if (filter.Client && f.Client !== filter.Client) {
      return
    }
    if (filter.Server && f.Server !== filter.Server) {
      return
    }
    if (filter.Version && f.Version !== filter.Version) {
      return
    }
    if (filter.Cipher && f.Cipher !== filter.Cipher) {
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
      value: f.Service + ':' + f.Score.toFixed(2),
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
  return over
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

const showTLSFlows3DChart = (div, tls) => {
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
  if (!tls) {
    return
  }
  let count = 0
  tls.forEach((f) => {
    if (count > 20000) {
      return
    }
    if (!f.ServerLatLong && !f.ClientLatLong) {
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

const showTLSVersionPieChart = (div, tls) => {
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
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      data: [],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: 'TLSバージョン',
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [],
      },
    ],
  }
  const verMap = new Map()
  tls.forEach((t) => {
    if (!verMap.has(t.Version)) {
      verMap.set(t.Version, { count: 0 })
    }
    verMap.get(t.Version).count++
  })
  verMap.forEach((v, k) => {
    if (option.legend.data.length > 20) {
      return
    }
    option.legend.data.push(k)
    option.series[0].data.push({
      name: k,
      value: v.count,
    })
  })
  chart.setOption(option)
  chart.resize()
}

const showTLSCipherChart = (div, tls) => {
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
  if (!tls) {
    return
  }
  const csMap = new Map()
  tls.forEach((t) => {
    if (!csMap.has(t.Cipher)) {
      csMap.set(t.Cipher, [0, 0, 0, 0, 0, 0])
    }
    csMap.get(t.Cipher)[0]++
    const si = getScoreIndex(t.Score)
    csMap.get(t.Cipher)[si]++
  })
  const keys = Array.from(csMap.keys())
  keys.sort(function (a, b) {
    return csMap.get(b)[0] - csMap.get(a)[0]
  })
  let i = keys.length - 1
  if (i > 49) {
    i = 49
  }
  for (; i >= 0; i--) {
    option.yAxis.data.push(keys[i])
    for (let j = 0; j < 5; j++) {
      option.series[j].data.push(csMap.get(keys[i])[j + 1])
    }
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showEtherTypeChart', showEtherTypeChart)
  inject('showTLSFlowsChart', showTLSFlowsChart)
  inject('showTLSFlows3DChart', showTLSFlows3DChart)
  inject('showTLSVersionPieChart', showTLSVersionPieChart)
  inject('showTLSCipherChart', showTLSCipherChart)
}
