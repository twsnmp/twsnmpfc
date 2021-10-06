import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

let chart
const showUsersChart = (div, users, filter) => {
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
      orient: 'vertical',
      top: 50,
      right: 10,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上'],
    },
    grid: {
      top: '3%',
      left: '7%',
      right: '10%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: '人数',
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
        stack: '人数',
        data: [],
      },
      {
        name: '33-41',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '42-50',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '51-66',
        type: 'bar',
        stack: '人数',
        data: [],
      },
      {
        name: '67以上',
        type: 'bar',
        stack: '人数',
        data: [],
      },
    ],
  }
  if (!users) {
    return
  }
  const data = {}
  users.forEach((u) => {
    if (!filterUser(u, filter)) {
      return
    }
    const id = u.ServerName + '(' + u.Server + ')'
    if (!data[id]) {
      data[id] = [0, 0, 0, 0, 0, 0]
    }
    data[id][0]++
    const si = getScoreIndex(u.Score)
    data[id][si]++
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

const showUserGraph = (div, users, type, filter) => {
  const nodeMap = new Map()
  const edgeMap = new Map()
  users.forEach((u) => {
    if (!filterUser(u, filter)) {
      return
    }
    Object.keys(u.ClientMap).forEach((k) => {
      const src = u.UserID + '@' + k
      const ek = src + '>' + u.Server
      const e = edgeMap.get(ek)
      const rate = u.ClientMap[k].Total
        ? (100.0 * u.ClientMap[k].Ok) / u.ClientMap[k].Total
        : 0
      if (!e) {
        edgeMap.set(ek, {
          source: src,
          target: u.Server,
          value: rate,
          total: u.ClientMap[k].Total,
        })
      } else {
        e.total += u.ClientMap[k].Total
      }
      const c = rate > 90.0 ? 1 : rate > 50.0 ? 2 : 3
      const n = nodeMap.get(src)
      if (!n) {
        nodeMap.set(src, {
          name: src,
          value: u.Total,
          draggable: true,
          category: c,
        })
      } else {
        if (n.category < c) {
          n.category = c
        }
        n.value += u.Total
      }
    })
    const n = nodeMap.get(u.Server)
    if (!n) {
      nodeMap.set(u.Server, {
        id: u.Server,
        name: u.ServerName + '(' + u.Server + ')',
        value: 0,
        draggable: true,
        category: 0,
      })
    }
  })
  const nodes = Array.from(nodeMap.values())
  const edges = Array.from(edgeMap.values())
  const nvs = []
  const evs = []
  nodes.forEach((e) => {
    nvs.push(e.value)
  })
  edges.forEach((e) => {
    evs.push(e.total)
  })
  const n95 = ecStat.statistics.quantile(nvs, 0.95)
  const n50 = ecStat.statistics.quantile(nvs, 0.5)
  const e95 = ecStat.statistics.quantile(evs, 0.95)
  const categories = [
    { name: 'サーバー' },
    { name: '90%以上' },
    { name: '50%以上' },
    { name: '50%未満' },
  ]
  let mul = 1.0
  if (type === 'gl') {
    mul = 1.5
  }
  nodes.forEach((e) => {
    e.label = { show: e.category === 0 || e.category === 3 }
    e.symbolSize = e.value > n95 || e.category === 0 ? 6 : e.value > n50 ? 4 : 2
    e.symbolSize *= mul
  })
  edges.forEach((e) => {
    e.lineStyle = {
      width: e.total > e95 ? 2 : 1,
      color: e.value > 90.0 ? '#1f78b4' : e.value > 50.0 ? '#ee0' : '#e31a1c',
    }
  })
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
    grid: {
      left: '7%',
      right: '4%',
      bottom: '3%',
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
    tooltip: {},
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
    color: ['#eee', '#1f78b4', '#ee0', '#e31a1c'],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [],
  }
  if (type === 'circular') {
    options.series = [
      {
        name: 'ログイン状況',
        type: 'graph',
        layout: 'circular',
        circular: {
          rotateLabel: true,
        },
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
          color: '#ccc',
        },
        lineStyle: {
          color: 'source',
          curveness: 0.3,
        },
      },
    ]
  } else if (type === 'gl') {
    options.series = [
      {
        name: 'ログイン状況',
        type: 'graphGL',
        nodes,
        edges,
        modularity: {
          resolution: 2,
          sort: true,
        },
        lineStyle: {
          color: 'source',
          opacity: 0.5,
        },
        itemStyle: {
          opacity: 1,
        },
        focusNodeAdjacency: false,
        focusNodeAdjacencyOn: 'click',
        emphasis: {
          label: {
            show: false,
          },
          lineStyle: {
            opacity: 0.5,
            width: 4,
          },
        },
        forceAtlas2: {
          steps: 5,
          stopThreshold: 20,
          jitterTolerence: 10,
          edgeWeight: [0.2, 1],
          gravity: 5,
          edgeWeightInfluence: 0,
        },
        categories,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
          color: '#ccc',
        },
      },
    ]
  } else {
    options.series = [
      {
        name: 'ログイン状況',
        type: 'graph',
        layout: 'force',
        data: nodes,
        links: edges,
        categories,
        roam: true,
        label: {
          position: 'right',
          formatter: '{b}',
          fontSize: 8,
          fontStyle: 'normal',
          color: '#ccc',
        },
        lineStyle: {
          color: 'source',
          curveness: 0,
        },
      },
    ]
  }
  chart.setOption(options)
  chart.resize()
}

const showUser3DChart = (div, users, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const mapz = new Map()
  let maxSym = 0
  users.forEach((u) => {
    if (!filterUser(u, filter)) {
      return
    }
    const sv = u.ServerName + '(' + u.Server + ')'
    Object.keys(u.ClientMap).forEach((k) => {
      const rate = u.ClientMap[k].Total
        ? (100.0 * u.ClientMap[k].Ok) / u.ClientMap[k].Total
        : 0
      const c = rate > 90.0 ? 2 : rate > 50.0 ? 1 : 0
      data.push([u.UserID, k, sv, c, rate, u.ClientMap[k].Total])
      mapy.set(k, true)
      if (maxSym < u.ClientMap[k].Total) {
        maxSym = u.ClientMap[k].Total
      }
    })
    mapx.set(u.UserID, true)
    mapz.set(sv, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  const catz = Array.from(mapz.keys())
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
    visualMap: [
      {
        show: false,
        dimension: 3,
        min: 0,
        max: 2,
        inRange: {
          color: ['#e31a1c', '#dfdf22', '#1f78b4'],
        },
      },
      {
        show: false,
        dimension: 5,
        min: 0,
        max: maxSym / 2,
        inRange: {
          symbolSize: [8, 16],
        },
      },
    ],
    xAxis3D: {
      type: 'category',
      name: 'ユーザー',
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
      name: 'クライアント',
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
      type: 'category',
      name: 'サーバー',
      data: catz,
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
        name: 'ログイン状況',
        type: 'scatter3D',
        symbolSize: 5,
        dimensions: [
          'ユーザー',
          'クライアント',
          'サーバー',
          '成功レベル',
          '成功率',
          '回数',
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const filterUser = (u, filter) => {
  if (filter.user && !u.UserID.includes(filter.user)) {
    return false
  }
  if (filter.server && !u.ServerName.includes(filter.server)) {
    return false
  }
  return true
}

export default (context, inject) => {
  inject('showUsersChart', showUsersChart)
  inject('filterUser', filterUser)
  inject('showUserGraph', showUserGraph)
  inject('showUser3DChart', showUser3DChart)
}
