import * as echarts from 'echarts'

const showFlowsChart = (div, flows, filter) => {
  const chart = echarts.init(document.getElementById(div))
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
        return params.name + '<br/>' + params.value
      },
      textStyle: {
        fontSize: 10,
      },
      position: 'right',
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
  if (!flows) {
    return false
  }
  if (filter.ClientName) {
    filter.ClientNameReg = new RegExp(filter.ClientName)
  }
  if (filter.ServerName) {
    filter.ServerNameReg = new RegExp(filter.ServerName)
  }
  let serviceReg
  if (filter.ServiceReg) {
    serviceReg = new RegExp(filter.ServiceReg)
  }
  let bOver = false
  const nodes = {}
  flows.forEach((f) => {
    if (option.series[0].links.length > 2000) {
      bOver = true
      return
    }
    if (filter.Service) {
      if (!f.Services || !f.Services[filter.Service]) {
        return
      }
    }
    if (serviceReg) {
      let hit = false
      const svs = Object.keys(f.Services)
      for (let i = 0; i < svs.length; i++) {
        if (svs[i].match(serviceReg)) {
          hit = true
          break
        }
      }
      if (!hit) return
    }
    if (filter.ClientNameReg) {
      if (!f.ClientName.match(filter.ClientNameReg)) {
        return
      }
    }
    if (filter.ClientIP) {
      if (f.Client !== filter.ClientIP) {
        return
      }
    }
    if (filter.ServerNameReg) {
      if (!f.ServerName.match(filter.ServerNameReg)) {
        return
      }
    }
    if (filter.ServerIP) {
      if (f.Server !== filter.ServerIP) {
        return
      }
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

export default (context, inject) => {
  inject('showFlowsChart', showFlowsChart)
}
