import * as echarts from 'echarts'

let chart

const makeFlowsChart = (div) => {
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
      right: '4%',
      bottom: '3%',
      containLabel: true,
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
  chart.setOption(option)
  chart.resize()
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
  } else if (s <= 0) {
    return '#aaa'
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

const showFlowsChart = (flows) => {
  if (!flows) {
    return
  }
  const opt = {
    series: [
      {
        data: [],
        links: [],
      },
    ],
  }
  const nodes = {}
  flows.forEach((f) => {
    if (opt.series[0].links.length > 1000) {
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
      }
    }
    if (!nodes[c]) {
      nodes[c] = {
        name: c,
        category: getLocCategory(f.ClientLoc),
        draggable: true,
        value: f.ClientLoc,
      }
    }
    opt.series[0].links.push({
      source: c,
      target: s,
      value: f.ServiceInfo + ':' + f.Score.toFixed(2),
      lineStyle: {
        color: getScoreColor(f.Score),
      },
    })
  })
  for (const k in nodes) {
    opt.series[0].data.push(nodes[k])
  }
  chart.setOption(opt)
  chart.resize()
}

export default (context, inject) => {
  inject('makeFlowsChart', makeFlowsChart)
  inject('showFlowsChart', showFlowsChart)
}
