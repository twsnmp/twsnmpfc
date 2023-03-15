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

const showTLSFlowsChart = (div, tls, filter, layout) => {
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
    if (!filterTLSFlow(f, filter)) {
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
      value:
        f.Service + ':' + f.Version + ':' + f.Cipher + ':' + f.Score.toFixed(2),
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

const filterTLSFlow = (f, filter) => {
  if (filter.client && !f.ClientName.includes(filter.client)) {
    return false
  }
  if (filter.ccountry && !f.CCountry.includes(filter.ccountry)) {
    return false
  }
  if (filter.server && !f.ServerName.includes(filter.server)) {
    return false
  }
  if (filter.scountry && !f.SCountry.includes(filter.scountry)) {
    return false
  }
  if (filter.service && !f.Service.includes(filter.service)) {
    return false
  }
  if (filter.version && !f.Version.includes(filter.version)) {
    return false
  }
  if (filter.cipher && !f.Cipher.includes(filter.cipher)) {
    return false
  }
  return true
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

const showTLSFlows3DChart = (div, tls, filter) => {
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
    if (!filterTLSFlow(f, filter)) {
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

const showTLSVersionPieChart = (div, tls, filter) => {
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
    if (!filterTLSFlow(t, filter)) {
      return
    }
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

const showTLSCipherChart = (div, tls, filter) => {
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
    if (!filterTLSFlow(t, filter)) {
      return
    }
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

const showDNSChart = (div, dns) => {
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
    color: ['#e31a1c', '#1f78b4'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['変化', '固定'],
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
    series: [
      {
        name: '変化',
        type: 'bar',
        stack: '回数',
        data: [],
      },
      {
        name: '固定',
        type: 'bar',
        stack: '回数',
        data: [],
      },
    ],
  }
  if (!dns) {
    return
  }
  const kn =
    div === 'nameChart' ? 'Name' : div === 'typeChart' ? 'Type' : 'Server'

  const m = new Map()
  dns.forEach((d) => {
    const k = d[kn]
    if (!m.has(k)) {
      m.set(k, [0, 0])
    }
    m.get(k)[0] += d.Change
    m.get(k)[1] += d.Count - d.Change
  })
  const keys = Array.from(m.keys())
  keys.sort(function (a, b) {
    return m.get(b)[0] - m.get(a)[0]
  })
  let i = keys.length - 1
  if (i > 49) {
    i = 49
  }
  for (; i >= 0; i--) {
    option.yAxis.data.push(getDNSTypeName(keys[i]))
    option.series[0].data.push(m.get(keys[i])[0])
    option.series[1].data.push(m.get(keys[i])[1])
  }
  chart.setOption(option)
  chart.resize()
}

const dnsTypeList = [
  { text: '', value: '' },
  { text: 'IPv4', value: 'A' },
  { text: 'IPv6', value: 'AAAA' },
  { text: 'ホスト名', value: 'PTR' },
  { text: 'DNSサーバー', value: 'NS' },
  { text: 'メールサーバー', value: 'MX' },
  { text: 'テキスト', value: 'TXT' },
  { text: 'エイリアス', value: 'CNAME' },
  { text: 'ゾーン', value: 'SOA' },
  { text: 'ホスト情報', value: 'HINFO' },
  { text: '不明', value: 'Unknown' },
]

const dnsTypeMap = new Map()

const getDNSTypeName = (t) => {
  if (dnsTypeMap.has(t)) {
    return dnsTypeMap.get(t)
  }
  let n = t
  for (let i = 0; i < dnsTypeList.length; i++) {
    if (dnsTypeList[i].value === t) {
      n = dnsTypeList[i].text
      break
    }
  }
  dnsTypeMap.set(t, n)
  return n
}

const showRADIUSFlowsChart = (div, radius, layout) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [{ name: 'サーバー' }, { name: 'クライアント' }]

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
    color: ['#eee', '#1f78b4'],
    tooltip: {
      trigger: 'item',
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
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [
      {
        type: 'graph',
        layout: layout || 'force',
        categories,
        symbolSize: 6,
        edgeSymbol: ['circle', 'arrow'],
        edgeSymbolSize: [2, 8],
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
  if (!radius) {
    return false
  }
  const nodes = {}
  radius.forEach((f) => {
    if (option.series[0].links.length > 2000) {
      return
    }
    const c = `${f.ClientName}(${f.Client})`
    const s = `${f.ServerName}(${f.Server})`
    if (!nodes[s]) {
      nodes[s] = {
        name: s,
        category: 0,
        draggable: true,
        symbolSize: 8,
        value: f.Count,
        label: {
          show: false,
        },
      }
    } else {
      nodes[s].value += f.Count
    }
    if (!nodes[c]) {
      nodes[c] = {
        name: c,
        category: 1,
        draggable: true,
        value: f.Count,
        label: {
          show: false,
        },
      }
    } else {
      nodes[c].value += f.Count
    }
    option.series[0].links.push({
      source: c,
      target: s,
      value: f.Accept + ':' + f.Reject + ':' + f.Score.toFixed(2),
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
}

const showRADIUSBarChart = (div, type, radius) => {
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
        saveAsImage: { name: 'twsnmp_' + type + div },
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    color: ['#e31a1c', '#1f78b4'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['Reject', 'Accept'],
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
    series: [
      {
        name: 'Reject',
        type: 'bar',
        stack: '回数',
        data: [],
      },
      {
        name: 'Accept',
        type: 'bar',
        stack: '回数',
        data: [],
      },
    ],
  }
  if (!radius) {
    return
  }

  const m = new Map()
  radius.forEach((r) => {
    let k
    switch (type) {
      case 'Client':
        k = r.Client
        break
      case 'Server':
        k = r.Server
        break
      case 'ClientToServer':
        k = r.Client + ':' + r.Server
        break
      default:
        return
    }
    if (!m.has(k)) {
      m.set(k, [0, 0])
    }
    m.get(k)[0] += r.Reject
    m.get(k)[1] += r.Accept
  })
  const keys = Array.from(m.keys())
  keys.sort((a, b) => {
    if (m.get(b)[0] === m.get(a)[0]) {
      return m.get(b)[1] - m.get(a)[1]
    }
    return m.get(b)[0] - m.get(a)[0]
  })
  let i = keys.length - 1
  if (i > 49) {
    i = 49
  }
  for (; i >= 0; i--) {
    option.yAxis.data.push(keys[i])
    option.series[0].data.push(m.get(keys[i])[0])
    option.series[1].data.push(m.get(keys[i])[1])
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
  inject('showDNSChart', showDNSChart)
  inject('showRADIUSFlowsChart', showRADIUSFlowsChart)
  inject('showRADIUSBarChart', showRADIUSBarChart)
  inject('filterTLSFlow', filterTLSFlow)
  inject('dnsTypeList', dnsTypeList)
  inject('getDNSTypeName', getDNSTypeName)
}
