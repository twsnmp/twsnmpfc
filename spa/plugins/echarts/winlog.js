import * as echarts from 'echarts'
import 'echarts-gl'
import * as ecStat from 'echarts-stat'

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
    visualMap: {
      show: false,
      dimension: 3,
      min: 0,
      max: 2,
      inRange: {
        color: ['#e31a1c', '#dfdf22', '#1f78b4'],
      },
    },
    xAxis3D: getAxisOption('category', 'Computer', catx),
    yAxis3D: getAxisOption('category', 'EventID', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
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
  chart.setOption(getScatter3DChartBaseOption(div))
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

const showWinLogonScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.target && !e.Target.includes(filter.target)) {
      return
    }
    if (filter.ip && !e.IP.includes(filter.ip)) {
      return
    }
    const from = e.IP ? e.IP : 'Local'
    data.push([
      from,
      e.Target,
      e.Count,
      getScoreIndex(e.Score),
      e.Score,
      e.Computer,
    ])
    mapx.set(from, true)
    mapy.set(e.Target, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    visualMap: {
      show: false,
      dimension: 3,
      min: 0,
      max: 4,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#a6cee3', '#1f78b4'],
      },
    },
    xAxis3D: getAxisOption('category', 'From', catx),
    yAxis3D: getAxisOption('category', 'Target', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
    series: [
      {
        name: 'Windows Logon',
        type: 'scatter3D',
        dimensions: ['From', 'Target', 'Count', 'Color', 'Score', 'Computer'],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinLogonForceChart = (div, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: 'From' }, { name: 'Target' }]
  const option = getForceChartOption(div, categories)
  if (!list) {
    return false
  }
  const nodes = {}
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.target && !e.Target.includes(filter.target)) {
      return
    }
    if (filter.ip && !e.IP.includes(filter.ip)) {
      return
    }
    if (filter.service && !e.Service.includes(filter.service)) {
      return
    }
    if (filter.ticketType && !e.TicketType.includes(filter.ticketType)) {
      return
    }
    const f = e.IP ? e.IP : 'Local'
    const t = e.Target
    if (!nodes[f]) {
      nodes[f] = {
        name: f,
        category: 0,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    if (!nodes[t]) {
      nodes[t] = {
        name: t,
        category: 1,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    option.series[0].links.push({
      source: f,
      target: t,
      value: e.Score.toFixed(2),
      lineStyle: {
        color: getScoreColor(e.Score),
      },
    })
  })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
}

const showWinKerberosScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.target && !e.Target.includes(filter.target)) {
      return
    }
    if (filter.ip && !e.IP.includes(filter.ip)) {
      return
    }
    if (filter.service && !e.Service.includes(filter.service)) {
      return
    }
    if (filter.ticketType && !e.TicketType.includes(filter.ticketType)) {
      return
    }
    const from = e.IP ? e.IP : 'Local'
    data.push([
      from,
      e.Target,
      e.Count,
      getScoreIndex(e.Score),
      e.Score,
      e.Computer,
      e.TicketType,
      e.Service,
    ])
    mapx.set(from, true)
    mapy.set(e.Target, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    visualMap: {
      show: false,
      dimension: 3,
      min: 0,
      max: 4,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#a6cee3', '#1f78b4'],
      },
    },
    xAxis3D: getAxisOption('category', 'From', catx),
    yAxis3D: getAxisOption('category', 'Target', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
    series: [
      {
        name: 'Windows Kerberos',
        type: 'scatter3D',
        dimensions: [
          'From',
          'Target',
          'Count',
          'Color',
          'Score',
          'Computer',
          'TicketType',
          'Service',
        ],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinPrivilegeScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const number = []
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.subject && !e.Subject.includes(filter.subject)) {
      return
    }
    number.push(e.Count)
    data.push([e.Computer, e.Subject, e.Count])
    mapx.set(e.Computer, true)
    mapy.set(e.Subject, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    visualMap: {
      min: ecStat.statistics.min(number),
      max: ecStat.statistics.max(number),
      calculable: true,
      realtime: false,
      dimension: 2,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: getAxisOption('category', 'Computer', catx),
    yAxis3D: getAxisOption('category', 'Subject', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
    series: [
      {
        name: 'Windows Privilege',
        type: 'scatter3D',
        dimensions: ['Computer', 'Subject', 'Count'],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const getAxisOption = (type, name, categories) => {
  return {
    type,
    name,
    data: categories,
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
  }
}

const getScatter3DChartBaseOption = (div) => {
  return {
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
    },
    xAxis3D: {},
    yAxis3D: {},
    zAxis3D: {},
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
  }
}

const showWinKerberosForceChart = (div, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: 'From' }, { name: 'Target' }]
  const option = getForceChartOption(div, categories)
  if (!list) {
    return false
  }
  const nodes = {}
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.target && !e.Target.includes(filter.target)) {
      return
    }
    if (filter.ip && !e.IP.includes(filter.ip)) {
      return
    }
    const f = e.IP ? e.IP : 'Local'
    const t = e.Target
    if (!nodes[f]) {
      nodes[f] = {
        name: f,
        category: 0,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    if (!nodes[t]) {
      nodes[t] = {
        name: t,
        category: 1,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    option.series[0].links.push({
      source: f,
      target: t,
      value: e.TicketType + ':' + e.Service + ':' + e.Score.toFixed(2),
      lineStyle: {
        width: 1,
        color: getScoreColor(e.Score),
      },
    })
  })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
}

const showWinAccountScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const number = []
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.target && !e.Target.includes(filter.target)) {
      return
    }
    if (filter.subject && !e.Subject.includes(filter.subject)) {
      return
    }
    number.push(e.Count)
    data.push([e.Subject, e.Target, e.Count, e.Computer])
    mapx.set(e.Subject, true)
    mapy.set(e.Target, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    visualMap: {
      min: ecStat.statistics.min(number),
      max: ecStat.statistics.max(number),
      calculable: true,
      realtime: false,
      dimension: 2,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: getAxisOption('category', 'Subject', catx),
    yAxis3D: getAxisOption('category', 'Target', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
    series: [
      {
        name: 'Windows Account',
        type: 'scatter3D',
        dimensions: ['Subject', 'Target', 'Count', 'Computer'],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinAccountForceChart = (div, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: 'Subject' }, { name: 'Target' }]
  const option = getForceChartOption(div, categories)
  if (!list) {
    return false
  }
  const q = {}
  const number = []
  const nodes = {}
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.target && !e.Target.includes(filter.target)) {
      return
    }
    if (filter.subject && !e.Subject.includes(filter.subject)) {
      return
    }
    number.push(e.Count)
    const s = e.Subject ? e.Subject : 'Unknown'
    const t = e.Target
    if (!nodes[s]) {
      nodes[s] = {
        name: s,
        category: 0,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    if (!nodes[t]) {
      nodes[t] = {
        name: t,
        category: 1,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    option.series[0].links.push({
      source: s,
      target: t,
      value: 'Count=' + e.Count + ' Edit=' + e.Edit + ' Passwd=' + e.Password,
      count: e.Count,
      lineStyle: getLineStyle(e.Count, q),
    })
  })
  q.q1 = ecStat.statistics.quantile(number, 0.25)
  q.q2 = ecStat.statistics.quantile(number, 0.5)
  q.q3 = ecStat.statistics.quantile(number, 0.75)
  option.series[0].links.forEach((l) => {
    l.lineStyle = getLineStyle(l.count, q)
  })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
}

const showWinProcessScatter3DChart = (div, list, mode, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const dimensions = [
    'Computer',
    'Process',
    'Count',
    'Subject',
    'Parent',
    'Status',
  ]
  switch (mode) {
    case 'subject':
      dimensions[0] = 'Subject'
      dimensions[3] = 'Computer'
      break
    case 'parent':
      dimensions[0] = 'Parent'
      dimensions[3] = 'Computer'
      dimensions[4] = 'Subject'
      break
  }
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.process && !e.Process.includes(filter.process)) {
      return
    }
    if (filter.subject && !e.LastSubject.includes(filter.subject)) {
      return
    }
    const status = e.LastStatus === '0x0' ? 0 : 1
    switch (mode) {
      case 'subject':
        data.push([
          e.LastSubject,
          e.Process,
          e.Count,
          e.LastStatus,
          e.Computer,
          e.LastParent,
          status,
        ])
        mapx.set(e.LastSubject, true)
        break
      case 'parent':
        data.push([
          e.LastParent,
          e.Process,
          e.Count,
          e.LastStatus,
          e.Computer,
          e.LastSubject,
          status,
        ])
        mapx.set(e.LastParent, true)
        break
      default:
        data.push([
          e.Computer,
          e.Process,
          e.Count,
          e.LastStatus,
          e.LastSubject,
          e.LastParent,
          status,
        ])
        mapx.set(e.Computer, true)
        break
    }
    mapy.set(e.Process, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    visualMap: {
      show: false,
      min: 0,
      max: 1,
      realtime: false,
      dimension: 6,
      inRange: {
        color: ['#4575b4', '#a50026'],
      },
    },
    xAxis3D: getAxisOption('category', dimensions[0], catx),
    yAxis3D: getAxisOption('category', 'Process', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
    series: [
      {
        name: 'Windows Process',
        type: 'scatter3D',
        dimensions,
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinProcessForceChart = (div, list, mode, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [
    { name: 'Computer' },
    { name: 'Normal Process' },
    { name: 'Failed Process' },
  ]
  switch (mode) {
    case 'subject':
      categories[0].name = 'Subject'
      break
    case 'parent':
      categories[0].name = 'Parent'
      break
  }
  const option = getForceChartOption(div, categories)
  if (!list) {
    return false
  }
  const q = {}
  const number = []
  const nodes = {}
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.process && !e.Process.includes(filter.process)) {
      return
    }
    if (filter.subject && !e.LastSubject.includes(filter.subject)) {
      return
    }
    number.push(e.Count)
    let s
    let value
    const t = e.Process
    switch (mode) {
      case 'subject':
        s = e.LastSubject
        value =
          'Count=' +
          e.Count +
          ' Compuert=' +
          e.Computer +
          'Parent=' +
          e.LastParent
        break
      case 'parent':
        s = e.LastParent
        value =
          'Count=' +
          e.Count +
          ' Compuert=' +
          e.Computer +
          'Subject=' +
          e.LastSubject
        break
      default:
        s = e.Computer
        value =
          'Count=' +
          e.Count +
          ' Subject=' +
          e.LastSubject +
          'Parent=' +
          e.LastParent
    }
    if (!nodes[s]) {
      nodes[s] = {
        name: s,
        category: 0,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    if (!nodes[t]) {
      nodes[t] = {
        name: t,
        category: e.LastStatus === '0x0' ? 1 : 2,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    option.series[0].links.push({
      source: s,
      target: t,
      value,
      count: e.Count,
      lineStyle: getLineStyle(e.Count, q),
    })
  })
  q.q1 = ecStat.statistics.quantile(number, 0.25)
  q.q2 = ecStat.statistics.quantile(number, 0.5)
  q.q3 = ecStat.statistics.quantile(number, 0.75)
  option.series[0].links.forEach((l) => {
    l.lineStyle = getLineStyle(l.count, q)
  })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
}

const showWinTaskScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const number = []
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.task && !e.TaskName.includes(filter.task)) {
      return
    }
    if (filter.subject && !e.Subject.includes(filter.subject)) {
      return
    }
    number.push(e.Count)
    data.push([e.Subject, e.TaskName, e.Count, e.Computer])
    mapx.set(e.Subject, true)
    mapy.set(e.TaskName, true)
  })
  const catx = Array.from(mapx.keys())
  const caty = Array.from(mapy.keys())
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
    visualMap: {
      min: ecStat.statistics.min(number),
      max: ecStat.statistics.max(number),
      calculable: true,
      realtime: false,
      dimension: 2,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: getAxisOption('category', 'Subject', catx),
    yAxis3D: getAxisOption('category', 'Task', caty),
    zAxis3D: getAxisOption('value', 'Count', []),
    series: [
      {
        name: 'Windows Task',
        type: 'scatter3D',
        dimensions: ['Subject', 'Task', 'Count', 'Computer'],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinTaskForceChart = (div, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: 'Subject' }, { name: 'Task' }]
  const option = getForceChartOption(div, categories)
  if (!list) {
    return false
  }
  const q = {}
  const number = []
  const nodes = {}
  list.forEach((e) => {
    if (filter.computer && !e.Computer.includes(filter.computer)) {
      return
    }
    if (filter.task && !e.TaskName.includes(filter.task)) {
      return
    }
    if (filter.subject && !e.Subject.includes(filter.subject)) {
      return
    }
    number.push(e.Count)
    const s = e.Subject ? e.Subject : 'Unknown'
    const t = e.TaskName
    if (!nodes[s]) {
      nodes[s] = {
        name: s,
        category: 0,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    if (!nodes[t]) {
      nodes[t] = {
        name: t,
        category: 1,
        draggable: true,
        label: {
          show: false,
        },
      }
    }
    option.series[0].links.push({
      source: s,
      target: t,
      value: 'Count=' + e.Count + ' Computer=' + e.Computer,
      count: e.Count,
      lineStyle: getLineStyle(e.Count, q),
    })
  })
  q.q1 = ecStat.statistics.quantile(number, 0.25)
  q.q2 = ecStat.statistics.quantile(number, 0.5)
  q.q3 = ecStat.statistics.quantile(number, 0.75)
  option.series[0].links.forEach((l) => {
    l.lineStyle = getLineStyle(l.count, q)
  })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
}

const getForceChartOption = (div, categories) => {
  return {
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
      textStyle: {
        fontSize: 8,
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
    color: ['#eee', '#1f78b4', '#e31a1c'],
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
          curveness: 0,
        },
      },
    ],
  }
}

const getLineStyle = (c, q) => {
  if (c < q.q1) {
    return { color: '#ccc', width: 1 }
  }
  if (c < q.q2) {
    return { color: '#eee', width: 2 }
  }
  if (c < q.q3) {
    return { color: '#dfdf22', width: 3 }
  }
  return { color: '#e31a1c', width: 4 }
}

const getScoreColor = (s) => {
  if (s > 66) {
    return '#1f78b4'
  } else if (s >= 50) {
    return '#a6cee3'
  } else if (s > 42) {
    return '#dfdf22'
  } else if (s > 33) {
    return '#fb9a99'
  }
  return '#e31a1c'
}

const getScoreIndex = (s) => {
  if (s > 66) {
    return 4
  } else if (s >= 50) {
    return 3
  } else if (s > 42) {
    return 2
  } else if (s > 33) {
    return 1
  }
  return 0
}

export default (context, inject) => {
  inject('showWinEventID3DChart', showWinEventID3DChart)
  inject('showWinLogonScatter3DChart', showWinLogonScatter3DChart)
  inject('showWinLogonForceChart', showWinLogonForceChart)
  inject('showWinAccountForceChart', showWinAccountForceChart)
  inject('showWinAccountScatter3DChart', showWinAccountScatter3DChart)
  inject('showWinKerberosScatter3DChart', showWinKerberosScatter3DChart)
  inject('showWinKerberosForceChart', showWinKerberosForceChart)
  inject('showWinPrivilegeScatter3DChart', showWinPrivilegeScatter3DChart)
  inject('showWinProcessForceChart', showWinProcessForceChart)
  inject('showWinProcessScatter3DChart', showWinProcessScatter3DChart)
  inject('showWinTaskForceChart', showWinTaskForceChart)
  inject('showWinTaskScatter3DChart', showWinTaskScatter3DChart)
}
