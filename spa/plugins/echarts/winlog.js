import * as echarts from 'echarts'
import 'echarts-gl'
import * as ecStat from 'echarts-stat'

let chart

const showWinEventID3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((e) => {
    if (!filterWinEventID(e, filter)) {
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
    xAxis3D: getAxisOption('category', 'コンピュータ', catx),
    yAxis3D: getAxisOption('category', 'イベントID', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows EventID',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions: [
          'コンピュータ',
          'イベントID',
          '回数',
          'レベル',
          'チャネル',
        ],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const filterWinEventID = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.provider && !e.Provider.includes(filter.provider)) {
    return false
  }
  if (filter.eventID && !e.EventID.toString(10).includes(filter.eventID)) {
    return false
  }
  if (filter.channel && !e.Channel.includes(filter.channel)) {
    return false
  }
  if (filter.level && e.Level !== filter.level) {
    return false
  }
  return true
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
    const from = e.IP ? e.IP : 'ローカル'
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
    xAxis3D: getAxisOption('category', '接続元', catx),
    yAxis3D: getAxisOption('category', 'ログイン先', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows Logon',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions: [
          '接続元',
          'ログオン先',
          '回数',
          '色',
          '信用スコア',
          'コンピュータ',
        ],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinLogonGraph = (div, list, filter, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: '接続元' }, { name: 'ログオン先' }]
  const option = getGraphChartOption(div, categories, layout)
  if (!list) {
    return false
  }
  const nodes = {}
  list.forEach((e) => {
    if (!filterWinLogon(e, filter)) {
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

const filterWinLogon = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.target && !e.Target.includes(filter.target)) {
    return false
  }
  if (filter.ip && !e.IP.includes(filter.ip)) {
    return false
  }
  if (filter.service && !e.Service.includes(filter.service)) {
    return false
  }
  if (filter.ticketType && !e.TicketType.includes(filter.ticketType)) {
    return false
  }
  return true
}

const showWinKerberosScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  list.forEach((e) => {
    if (!filterWinKerberos(e, filter)) {
      return
    }
    const from = e.IP ? e.IP : 'ローカル'
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
    xAxis3D: getAxisOption('category', '操作元', catx),
    yAxis3D: getAxisOption('category', '対象', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows Kerberos',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions: [
          '操作元',
          '対象',
          '回数',
          '色',
          '信用スコア',
          'コンピュータ',
          'チケット種別',
          'サービス',
        ],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const filterWinKerberos = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.target && !e.Target.includes(filter.target)) {
    return false
  }
  if (filter.ip && !e.IP.includes(filter.ip)) {
    return false
  }
  if (filter.service && !e.Service.includes(filter.service)) {
    return false
  }
  if (filter.ticketType && !e.TicketType.includes(filter.ticketType)) {
    return false
  }
  return true
}

const showWinPrivilegeScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const number = []
  list.forEach((e) => {
    if (!filterWinPrivilege(e, filter)) {
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
    xAxis3D: getAxisOption('category', 'コンピュータ', catx),
    yAxis3D: getAxisOption('category', '操作アカウント', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows特権アクセス',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions: ['コンピュータ', '操作アカウント', '回数'],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const filterWinPrivilege = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.subject && !e.Subject.includes(filter.subject)) {
    return false
  }
  return true
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

const showWinKerberosGraph = (div, list, filter, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: '操作元' }, { name: '対象' }]
  const option = getGraphChartOption(div, categories, layout)
  if (!list) {
    return false
  }
  const nodes = {}
  list.forEach((e) => {
    if (!filterWinKerberos(e, filter)) {
      return
    }
    const f = e.IP ? e.IP : 'ローカル'
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
    if (!filterWinAccount(e, filter)) {
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
    xAxis3D: getAxisOption('category', '操作アカウント', catx),
    yAxis3D: getAxisOption('category', '対象アカウント', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows Account',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions: [
          '操作アカウント',
          '対象アカウント',
          '回数',
          'コンピュータ',
        ],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinAccountGraph = (div, list, filter, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: '操作アカウント' }, { name: '対象アカウント' }]
  const option = getGraphChartOption(div, categories, layout)
  if (!list) {
    return false
  }
  const q = {}
  const number = []
  const nodes = {}
  list.forEach((e) => {
    if (!filterWinAccount(e, filter)) {
      return
    }
    number.push(e.Count)
    const s = e.Subject ? e.Subject : '不明'
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
      value:
        '回数=' + e.Count + ' 編集=' + e.Edit + ' パスワード=' + e.Password,
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

const filterWinAccount = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.target && !e.Target.includes(filter.target)) {
    return false
  }
  if (filter.subject && !e.Subject.includes(filter.subject)) {
    return false
  }
  return true
}

const showWinProcessScatter3DChart = (div, list, mode, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const dimensions = [
    'コンピュータ',
    'プロセス',
    '回数',
    'ステータス',
    '操作アカウント',
    '親プロセス',
    '色',
  ]
  switch (mode) {
    case 'subject':
      dimensions[0] = '関連アクアンと'
      dimensions[3] = 'コンピュータ'
      break
    case 'parent':
      dimensions[0] = '親プロセス'
      dimensions[3] = 'コンピュータ'
      dimensions[4] = '操作アカウント'
      break
  }
  list.forEach((e) => {
    if (!filterWinProcess(e, filter)) {
      return
    }
    const color = e.LastStatus === '0x0' ? 0 : 1
    switch (mode) {
      case 'subject':
        data.push([
          e.LastSubject,
          e.Process,
          e.Count,
          e.LastStatus,
          e.Computer,
          e.LastParent,
          color,
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
          color,
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
          color,
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
    yAxis3D: getAxisOption('category', 'プロセス', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows Process',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions,
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinProcessGraph = (div, list, mode, filter, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [
    { name: 'コンピュータ' },
    { name: '正常終了' },
    { name: '異常終了' },
  ]
  switch (mode) {
    case 'subject':
      categories[0].name = '操作アカウント'
      break
    case 'parent':
      categories[0].name = '親プロセス'
      break
  }
  const option = getGraphChartOption(div, categories, layout)
  if (!list) {
    return false
  }
  const q = {}
  const number = []
  const nodes = {}
  list.forEach((e) => {
    if (!filterWinProcess(e, filter)) {
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
          '回数=' +
          e.Count +
          ' コンピュータ=' +
          e.Computer +
          '親プロセス=' +
          e.LastParent
        break
      case 'parent':
        s = e.LastParent
        value =
          '回数=' +
          e.Count +
          ' コンピュータ=' +
          e.Computer +
          '操作アカウント=' +
          e.LastSubject
        break
      default:
        s = e.Computer
        value =
          '回数=' +
          e.Count +
          ' 操作アカウント=' +
          e.LastSubject +
          '親プロセス=' +
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

const filterWinProcess = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.process && !e.Process.includes(filter.process)) {
    return false
  }
  if (filter.subject && !e.LastSubject.includes(filter.subject)) {
    return false
  }
  return true
}

const showWinTaskScatter3DChart = (div, list, filter) => {
  const data = []
  const mapx = new Map()
  const mapy = new Map()
  const number = []
  list.forEach((e) => {
    if (!filterWinTask(e, filter)) {
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
    xAxis3D: getAxisOption('category', '操作アカウント', catx),
    yAxis3D: getAxisOption('category', 'タスク', caty),
    zAxis3D: getAxisOption('value', '回数', []),
    series: [
      {
        name: 'Windows Task',
        type: 'scatter3D',
        symbolSize: 6,
        dimensions: ['Subject', 'Task', 'Count', 'Computer'],
        data,
      },
    ],
  }
  chart.setOption(getScatter3DChartBaseOption(div))
  chart.setOption(options)
  chart.resize()
}

const showWinTaskGraph = (div, list, filter, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: '操作アカウント' }, { name: 'タスク' }]
  const option = getGraphChartOption(div, categories, layout)
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

const filterWinTask = (e, filter) => {
  if (filter.computer && !e.Computer.includes(filter.computer)) {
    return false
  }
  if (filter.task && !e.TaskName.includes(filter.task)) {
    return false
  }
  if (filter.subject && !e.Subject.includes(filter.subject)) {
    return false
  }
  return true
}

const getGraphChartOption = (div, categories, layout) => {
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
  inject('showWinLogonGraph', showWinLogonGraph)
  inject('showWinAccountGraph', showWinAccountGraph)
  inject('showWinAccountScatter3DChart', showWinAccountScatter3DChart)
  inject('showWinKerberosScatter3DChart', showWinKerberosScatter3DChart)
  inject('showWinKerberosGraph', showWinKerberosGraph)
  inject('showWinPrivilegeScatter3DChart', showWinPrivilegeScatter3DChart)
  inject('showWinProcessGraph', showWinProcessGraph)
  inject('showWinProcessScatter3DChart', showWinProcessScatter3DChart)
  inject('showWinTaskGraph', showWinTaskGraph)
  inject('showWinTaskScatter3DChart', showWinTaskScatter3DChart)
  inject('filterWinEventID', filterWinEventID)
  inject('filterWinLogon', filterWinLogon)
  inject('filterWinAccount', filterWinAccount)
  inject('filterWinKerberos', filterWinKerberos)
  inject('filterWinPrivilege', filterWinPrivilege)
  inject('filterWinProcess', filterWinProcess)
  inject('filterWinTask', filterWinTask)
}
