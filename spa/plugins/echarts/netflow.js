import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { isPrivateIP, isV4Format } from '~/plugins/echarts/utils.js'
import { doFFT } from '~/plugins/echarts/fft.js'

let chart

const makeNetFlowHistogram = (div) => {
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
    tooltip: {
      trigger: 'axis',
      formatter(params) {
        const p = params[0]
        return p.value[0] + 'の回数:' + p.value[1]
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      scale: true,
      min: 0,
    },
    yAxis: {
      name: '回数',
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showNetFlowHistogram = (div, logs, type) => {
  makeNetFlowHistogram(div)
  if (type === '') {
    type = 'size'
  }
  const data = []
  logs.forEach((l) => {
    if (type === 'size') {
      if (l.Packets > 0) {
        data.push(l.Bytes / l.Packets)
      }
    } else if (type === 'dur') {
      if (l.Duration >= 0.0) {
        data.push(l.Duration)
      }
    } else if (type === 'speed') {
      if (l.Duration > 0) {
        data.push(l.Bytes / l.Duration)
      }
    }
  })
  const bins = ecStat.histogram(data)
  chart.setOption({
    xAxis: {
      name: type,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize()
}
const makeNetFlowCluster = (div) => {
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
    legend: {
      data: [],
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'value',
    },
    yAxis: {
      type: 'value',
    },
    series: [],
  }
  chart.setOption(option)
  chart.resize()
}

const showNetFlowCluster = (div, logs, type, cluster) => {
  makeNetFlowCluster(div)
  if (type === '') {
    type = 'size-bps'
  }
  if (cluster < 2 || cluster > 10) {
    cluster = 2
  }
  const data = []
  logs.forEach((l) => {
    if (l.Packets <= 0 || l.Duration <= 0.0) {
      return
    }
    if (type === 'size-bps') {
      data.push([l.Bytes / l.Packets, l.Bytes / l.Duration])
    } else if (type === 'size-pps') {
      data.push([l.Bytes / l.Packets, l.Packets / l.Duration])
    } else if (type === 'pps-bps') {
      data.push([l.Packets / l.Duration, l.Bytes / l.Duration])
    } else if (type === 'sport-dport') {
      data.push([l.SrcPort, l.DstPort])
    }
  })
  const result = ecStat.clustering.hierarchicalKMeans(data, {
    clusterCount: cluster,
    stepByStep: false,
    outputType: 'multiple',
    outputClusterIndexDimension: cluster,
  })
  const centroids = result.centroids
  const ptsInCluster = result.pointsInCluster
  const series = []
  for (let i = 0; i < centroids.length; i++) {
    series.push({
      name: 'cluster' + (i + 1),
      type: 'scatter',
      data: ptsInCluster[i],
      markPoint: {
        symbolSize: 30,
        label: {
          show: true,
          position: 'top',
          formatter: (params) => {
            return (
              Math.round(params.data.coord[0] * 100) / 100 +
              ' / ' +
              Math.round(params.data.coord[1] * 100) / 100
            )
          },
          color: '#ccc',
          fontSize: 10,
        },
        data: [
          {
            coord: centroids[i],
          },
        ],
      },
    })
  }
  chart.setOption({
    legend: {
      data: [],
    },
    series,
  })
  chart.resize()
}

const makeNetFlowTraffic = (div, type) => {
  let yAxis = ''
  switch (type) {
    case 'bytes':
      yAxis = 'バイト数'
      break
    case 'packets':
      yAxis = 'パケット数'
      break
    case 'bps':
      yAxis = 'バイト/Sec'
      break
    case 'pps':
      yAxis = 'パケット/Sec'
      break
  }
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
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        const p = params[0]
        return p.name + ' : ' + p.value[1]
      },
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 40,
      buttom: 0,
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: yAxis,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        type: 'bar',
        color: '#1f78b4',
        large: true,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showNetFlowTraffic = (div, logs, type) => {
  makeNetFlowTraffic(div, type)
  const data = []
  let bytes = 0
  let packets = 0
  let dur = 0
  let ctm
  logs.forEach((l) => {
    if (!ctm) {
      ctm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60))
      bytes += l.Bytes
      dur += l.Duration
      packets += l.Packets
      return
    }
    const newCtm = Math.floor(l.Time / (1000 * 1000 * 1000 * 60))
    if (ctm !== newCtm) {
      let t = new Date(ctm * 60 * 1000)
      let d = 0
      switch (type) {
        case 'bytes':
          d = bytes
          break
        case 'packets':
          d = packets
          break
        case 'bps':
          if (dur > 0) {
            d = bytes / dur
          }
          break
        case 'pps':
          if (dur > 0) {
            d = packets / dur
          }
          break
      }
      data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, d],
      })
      ctm++
      for (; ctm < newCtm; ctm++) {
        t = new Date(ctm * 60 * 1000)
        data.push({
          name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
          value: [t, 0],
        })
      }
      bytes = 0
      dur = 0
      packets = 0
    }
    bytes += l.Bytes
    dur += l.Duration
    packets += l.Packets
  })
  chart.setOption({
    series: [
      {
        data,
      },
    ],
  })
  chart.resize()
}

const showNetFlowTop = (div, list, type) => {
  const data = []
  const category = []

  let xAxis = ''
  switch (type) {
    case 'bytes':
      xAxis = 'バイト数'
      list.sort((a, b) => b.Bytes - a.Bytes)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].Bytes)
        category.push(list[i].Name)
      }
      break
    case 'packets':
      xAxis = 'パケット数'
      list.sort((a, b) => b.Packets - a.Packets)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].Packets)
        category.push(list[i].Name)
      }
      break
    case 'dur':
      xAxis = '通信期間(Sec)'
      list.sort((a, b) => b.Duration - a.Duration)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].Duration)
        category.push(list[i].Name)
      }
      break
    case 'bps':
      xAxis = 'バイト/Sec'
      list.sort((a, b) => b.bps - a.bps)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].bps)
        category.push(list[i].Name)
      }
      break
    case 'pps':
      xAxis = 'パケット/Sec'
      list.sort((a, b) => b.pps - a.pps)
      for (let i = list.length > 20 ? 19 : list.length - 1; i >= 0; i--) {
        data.push(list[i].pps)
        category.push(list[i].Name)
      }
      break
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption({
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
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '20%',
      right: '10%',
      top: 10,
      buttom: 10,
    },
    xAxis: {
      type: 'value',
      name: xAxis,
      boundaryGap: [0, 0.01],
    },
    yAxis: {
      type: 'category',
      data: category,
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
    },
    series: [
      {
        name: xAxis,
        type: 'bar',
        data,
      },
    ],
  })
  chart.resize()
}

const getNetFlowSenderList = (logs) => {
  const m = new Map()
  logs.forEach((l) => {
    const e = m.get(l.SrcIP)
    if (!e) {
      m.set(l.SrcIP, {
        Name: l.SrcIP,
        Bytes: l.Bytes,
        Packets: l.Packets,
        Duration: l.Duration,
      })
    } else {
      e.Bytes += l.Bytes
      e.Packets += l.Packets
      e.Duration += l.Duration
    }
  })
  const r = Array.from(m.values())
  r.forEach((e) => {
    if (e.Duration > 0) {
      e.bps = (e.Bytes / e.Duration).toFixed(3)
      e.pps = (e.Packets / e.Duration).toFixed(3)
      e.Duration = e.Duration.toFixed(3)
    } else {
      e.bps = 0
      e.pps = 0
    }
  })
  return r
}

const getNetFlowServiceList = (logs) => {
  const m = new Map()
  logs.forEach((l) => {
    let k = getServiceName(l.SrcPort + '/' + l.Protocol)
    if (k === 'Other') {
      k = getServiceName(l.DstPort + '/' + l.Protocol)
    }
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        Bytes: l.Bytes,
        Packets: l.Packets,
        Duration: l.Duration,
      })
    } else {
      e.Bytes += l.Bytes
      e.Packets += l.Packets
      e.Duration += l.Duration
    }
  })
  const r = Array.from(m.values())
  r.forEach((e) => {
    if (e.Duration > 0) {
      e.bps = (e.Bytes / e.Duration).toFixed(3)
      e.pps = (e.Packets / e.Duration).toFixed(3)
      e.Duration = e.Duration.toFixed(3)
    } else {
      e.bps = 0
      e.pps = 0
    }
  })
  return r
}

// Service Name Map
const serviceNameArray = [
  ['80/tcp', 'HTTP'],
  ['443/tcp', 'HTTPS'],
  ['389/tcp', 'LDAP'],
  ['636/tcp', 'LDAP'],
  ['53/tcp', 'DNS'],
  ['53/udp', 'DNS'],
  ['161/udp', 'SNMP'],
  ['162/udp', 'SNMP'],
  ['123/udp', 'NTP'],
  ['25/tcp', 'SMTP'],
  ['587/tcp', 'SMTP'],
  ['110/tcp', 'POP3'],
  ['995/tcp', 'POP3'],
  ['143/tcp', 'IMAP'],
  ['943/tcp', 'IMAP'],
  ['22/tcp', 'SSH'],
  ['21/tcp', 'TELNET'],
  ['23/tcp', 'FTP'],
  ['67/udp', 'DHCP'],
  ['68/udp', 'DHCP'],
  ['514/udp', 'SYSLOG'],
  ['2049/tcp', 'NFS'],
  ['2049/udp', 'NFS'],
  ['445/tcp', 'CIFS'],
  ['3389/tcp', 'RDP'],
  ['5900/tcp', 'VNC'],
  ['137/udp', 'NETBIOS'],
  ['137/tcp', 'NETBIOS'],
  ['138/udp', 'NETBIOS'],
  ['138/tcp', 'NETBIOS'],
  ['139/udp', 'NETBIOS'],
  ['139/tcp', 'NETBIOS'],
  ['88/tcp', 'AD'],
  ['7680/tcp', 'WUDO'],
  ['1812/udp', 'RADIUS'],
  ['5223/tcp', 'APPLE'],
  ['5228/tcp', 'ANDROID'],
]

const serviceNameMap = new Map(serviceNameArray)

const getServiceName = (s) => {
  const ret = serviceNameMap.get(s)
  if (ret) {
    return ret
  }
  if (s.indexOf('/icmp') > 0) {
    return 'ICMP'
  }
  return 'Other'
}

const getNetFlowIPFlowList = (logs) => {
  const m = new Map()
  logs.forEach((l) => {
    let k = l.SrcIP + '<->' + l.DstIP
    let e = m.get(k)
    if (!e) {
      k = l.DstIP + '<->' + l.SrcIP
      e = m.get(k)
    }
    if (!e) {
      m.set(k, {
        Name: k,
        Bytes: l.Bytes,
        Packets: l.Packets,
        Duration: l.Duration,
      })
    } else {
      if (k !== l.SrcIP + '<->' + l.DstIP) {
        // 逆報告もある場合
        e.bidir = true
      }
      e.Bytes += l.Bytes
      e.Packets += l.Packets
      e.Duration += l.Duration
    }
  })
  const r = Array.from(m.values())
  r.forEach((e) => {
    if (e.Duration > 0) {
      if (e.bidir) {
        e.Duration /= 2.0
      }
      e.bps = (e.Bytes / e.Duration).toFixed(3)
      e.pps = (e.Packets / e.Duration).toFixed(3)
      e.Duration = e.Duration.toFixed(3)
    } else {
      e.bps = 0
      e.pps = 0
    }
  })
  return r
}

const showNetFlowGraph = (div, logs, type) => {
  const nodeMap = new Map()
  const edgeMap = new Map()
  logs.forEach((l) => {
    let ek = l.SrcIP + '|' + l.DstIP
    let e = edgeMap.get(ek)
    if (!e) {
      ek = l.DstIP + '|' + l.SrcIP
      e = edgeMap.get(ek)
    }
    if (!e) {
      edgeMap.set(ek, {
        source: l.SrcIP,
        target: l.DstIP,
        value: l.Bytes,
      })
    } else {
      e.value += l.Bytes
    }
    let n = nodeMap.get(l.Src)
    if (!n) {
      nodeMap.set(l.SrcIP, {
        name: l.SrcIP,
        value: l.Bytes,
        draggable: true,
        category: getNodeCategory(l.SrcIP),
      })
    } else {
      n.value += l.Bytes
    }
    n = nodeMap.get(l.DstIP)
    if (!n) {
      nodeMap.set(l.DstIP, {
        name: l.DstIP,
        value: 0,
        draggable: true,
        category: getNodeCategory(l.DstIP),
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
    evs.push(e.value)
  })
  const n95 = ecStat.statistics.quantile(nvs, 0.95)
  const n50 = ecStat.statistics.quantile(nvs, 0.5)
  const e95 = ecStat.statistics.quantile(evs, 0.95)
  const categories = [
    { name: 'IPv4 Private' },
    { name: 'IPv6 Private' },
    { name: 'IPv4 Global' },
    { name: 'IPV6 Global' },
  ]
  let mul = 1.0
  if (type === 'gl') {
    mul = 1.5
  }
  nodes.forEach((e) => {
    e.label = { show: e.value > n95 }
    e.symbolSize = e.value > n95 ? 5 : e.value > n50 ? 4 : 2
    e.symbolSize *= mul
  })
  edges.forEach((e) => {
    e.lineStyle = {
      width: e.value > e95 ? 2 : 1,
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
    color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99'],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [],
  }
  if (type === 'circular') {
    options.series = [
      {
        name: 'IP Flows',
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
        name: 'IP Flows',
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
        name: 'IP Flows',
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

const getNodeCategory = (ip) => {
  if (isPrivateIP(ip)) {
    if (isV4Format(ip)) {
      return 0
    }
    return 1
  }
  if (isV4Format(ip)) {
    return 2
  }
  return 3
}

const showNetFlowService3D = (div, logs, type) => {
  const m = new Map()
  logs.forEach((l) => {
    let k = getServiceName(l.SrcPort + '/' + l.Protocol)
    if (k === 'Other') {
      k = getServiceName(l.DstPort + '/' + l.Protocol)
    }
    const ipt = getNodeCategory(l.SrcIP)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        TotalBytes: l.Bytes,
        Time: [t],
        Bytes: [l.Bytes],
        Packets: [l.Packets],
        Duration: [l.Duration],
        IPType: [ipt],
      })
    } else {
      e.TotalBytes += l.Bytes
      e.Time.push(t)
      e.Bytes.push(l.Bytes)
      e.Packets.push(l.Packets)
      e.Duration.push(l.Duration)
      e.IPType.push(ipt)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  let dim = []
  switch (type) {
    case 'Bytes':
      dim = ['Service', 'Time', 'Bytes', 'Packtes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Bytes[i],
            e.Packets[i],
            e.Duration[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'Packets':
      dim = ['Service', 'Time', 'Packtes', 'Bytes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Packets[i],
            e.Bytes[i],
            e.Duration[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'Duration':
      dim = ['Service', 'Time', 'Duration', 'Bytes', 'Packtes', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Duration[i],
            e.Bytes[i],
            e.Packets[i],
            e.IPType[i],
          ])
        }
      })
      break
  }
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
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 3,
      dimension: 5,
      inRange: {
        color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Service',
      data: cat,
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
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: type,
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
        name: 'サービス別通信量',
        type: 'scatter3D',
        dimensions: dim,
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showNetFlowSender3D = (div, logs, type) => {
  const m = new Map()
  logs.forEach((l) => {
    const ipt = getNodeCategory(l.SrcIP)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(l.SrcIP)
    if (!e) {
      m.set(l.SrcIP, {
        Name: l.SrcIP,
        TotalBytes: l.Bytes,
        Time: [t],
        Bytes: [l.Bytes],
        Packets: [l.Packets],
        Duration: [l.Duration],
        IPType: [ipt],
      })
    } else {
      e.TotalBytes += l.Bytes
      e.Time.push(t)
      e.Bytes.push(l.Bytes)
      e.Packets.push(l.Packets)
      e.Duration.push(l.Duration)
      e.IPType.push(ipt)
    }
  })
  // 上位500件に絞る
  const ks = Array.from(m.keys())
  if (ks.length > 500) {
    ks.sort((a, b) => {
      const ea = m.get(a)
      const eb = m.get(b)
      return eb.TotalBytes - ea.TotalBytes
    })
    for (let i = 500; i < ks.length; i++) {
      m.delete(ks[i])
    }
  }
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  let dim = []
  switch (type) {
    case 'Bytes':
      dim = ['Sender', 'Time', 'Bytes', 'Packtes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Bytes[i],
            e.Packets[i],
            e.Duration[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'Packets':
      dim = ['Sender', 'Time', 'Packtes', 'Bytes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Packets[i],
            e.Bytes[i],
            e.Duration[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'Duration':
      dim = ['Sender', 'Time', 'Duration', 'Bytes', 'Packtes', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Duration[i],
            e.Bytes[i],
            e.Packets[i],
            e.IPType[i],
          ])
        }
      })
      break
  }
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
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 3,
      dimension: 5,
      inRange: {
        color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Sender',
      data: cat,
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
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: type,
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
        name: '送信元別通信量',
        type: 'scatter3D',
        dimensions: dim,
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showNetFlowIPFlow3D = (div, logs, type) => {
  const m = new Map()
  logs.forEach((l) => {
    let k = l.SrcIP + '<->' + l.DstIP
    let e = m.get(k)
    if (!e) {
      k = l.DstIP + '<->' + l.SrcIP
      e = m.get(k)
    }
    const ipt = getNodeCategory(l.SrcIP)
    const t = new Date(l.Time / (1000 * 1000))
    if (!e) {
      m.set(k, {
        Name: k,
        TotalBytes: l.Bytes,
        Time: [t],
        Bytes: [l.Bytes],
        Packets: [l.Packets],
        Duration: [l.Duration],
        IPType: [ipt],
      })
    } else {
      e.TotalBytes += l.Bytes
      e.Time.push(t)
      e.Bytes.push(l.Bytes)
      e.Packets.push(l.Packets)
      e.Duration.push(l.Duration)
      e.IPType.push(ipt)
    }
  })
  // 上位500件に絞る
  const ks = Array.from(m.keys())
  if (ks.length > 500) {
    ks.sort((a, b) => {
      const ea = m.get(a)
      const eb = m.get(b)
      return eb.TotalBytes - ea.TotalBytes
    })
    for (let i = 500; i < ks.length; i++) {
      m.delete(ks[i])
    }
  }
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  let dim = []
  switch (type) {
    case 'Bytes':
      dim = ['IPs', 'Time', 'Bytes', 'Packtes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Bytes[i],
            e.Packets[i],
            e.Duration[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'Packets':
      dim = ['IPs', 'Time', 'Packtes', 'Bytes', 'Duration', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Packets[i],
            e.Bytes[i],
            e.Duration[i],
            e.IPType[i],
          ])
        }
      })
      break
    case 'Duration':
      dim = ['IPs', 'Time', 'Duration', 'Bytes', 'Packtes', 'IPType']
      l.forEach((e) => {
        for (let i = 0; i < e.Time.length && i < 15000; i++) {
          data.push([
            e.Name,
            e.Time[i],
            e.Duration[i],
            e.Bytes[i],
            e.Packets[i],
            e.IPType[i],
          ])
        }
      })
      break
  }
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
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 3,
      dimension: 5,
      inRange: {
        color: ['#1f78b4', '#a6cee3', '#e31a1c', '#fb9a99'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Src <-> Dst',
      data: cat,
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
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
        },
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: type,
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
        name: 'IPペアー別通信量',
        type: 'scatter3D',
        dimensions: dim,
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const getNetFlowFFTMap = (logs) => {
  const m = new Map()
  m.set('Total', { Name: 'Total', Count: 0, Data: [], Total: 0 })
  let packets = 0
  logs.forEach((l) => {
    const e = m.get(l.SrcIP)
    if (!e) {
      m.set(l.SrcIP, { Name: l.SrcIP, Count: 0, Data: [], Total: l.Packets })
    } else {
      e.Total += l.Packets
    }
    packets += l.Packets
  })
  m.get('Total').Total = packets
  let cts
  logs.forEach((l) => {
    if (!cts) {
      cts = Math.floor(l.Time / (1000 * 1000 * 1000))
      m.get('Total').Count++
      m.get(l.SrcIP).Count++
      return
    }
    const newCts = Math.floor(l.Time / (1000 * 1000 * 1000))
    if (cts !== newCts) {
      m.forEach((e) => {
        e.Data.push(e.Count)
        e.Count = 0
      })
      cts++
      for (; cts < newCts; cts++) {
        m.forEach((e) => {
          e.Data.push(0)
        })
      }
    }
    m.get('Total').Count++
    m.get(l.SrcIP).Count++
  })
  const ks = Array.from(m.keys())
  if (ks.length > 50) {
    ks.sort((a, b) => {
      const ea = m.get(a)
      const eb = m.get(b)
      return eb.Total - ea.Total
    })
    for (let i = 0; i < ks.length; i++) {
      const e = m.get(ks[i])
      if (i > 50 || e.Total < 10) {
        m.delete(ks[i])
      }
    }
  }
  m.forEach((e) => {
    e.FFT = doFFT(e.Data, 1)
  })
  return m
}

const showNetFlowFFT = (div, fftMap, src, type) => {
  if (chart) {
    chart.dispose()
  }
  if (!fftMap || !fftMap.get(src)) {
    return
  }
  const fftData = fftMap.get(src).FFT
  const freq = type === 'hz'
  const fft = []
  if (freq) {
    fftData.forEach((e) => {
      fft.push([e.frequency, e.magnitude])
    })
  } else {
    fftData.forEach((e) => {
      fft.push([e.period, e.magnitude])
    })
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
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '10%',
      buttom: '10%',
    },
    dataZoom: [
      {
        type: 'inside',
      },
      {},
    ],
    xAxis: {
      type: 'value',
      name: freq ? '周波数(Hz)' : '周期(Sec)',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      type: 'value',
      name: '回数',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
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
    series: [
      {
        name: '回数',
        type: 'bar',
        color: '#5470c6',
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: fft,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showNetFlowFFT3D = (div, fftMap, fftType) => {
  const data = []
  const freq = fftType === 'hz'
  const colors = []
  const cat = []
  fftMap.forEach((e) => {
    if (e.Name === 'Total') {
      return
    }
    cat.push(e.Name)
    if (freq) {
      e.FFT.forEach((f) => {
        if (f.frequency === 0.0) {
          return
        }
        data.push([e.Name, f.frequency, f.magnitude, f.period])
        colors.push(f.magnitude)
      })
    } else {
      e.FFT.forEach((f) => {
        if (f.period === 0.0) {
          return
        }
        data.push([e.Name, f.period, f.magnitude, f.frequency])
        colors.push(f.magnitude)
      })
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
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: true,
      min: ecStat.statistics.min(colors),
      max: ecStat.statistics.max(colors),
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
    xAxis3D: {
      type: 'category',
      name: 'Src IP',
      data: cat,
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
      type: freq ? 'value' : 'log',
      name: freq ? '周波数(Hz)' : '周期(Sec)',
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
    zAxis3D: {
      type: 'value',
      name: '回数',
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
        name: 'NetFlow FFT分析',
        type: 'scatter3D',
        dimensions: [
          'Host',
          freq ? '周波数' : '周期',
          '回数',
          freq ? '周期' : '周波数',
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

export default (context, inject) => {
  inject('showNetFlowHistogram', showNetFlowHistogram)
  inject('showNetFlowCluster', showNetFlowCluster)
  inject('showNetFlowTraffic', showNetFlowTraffic)
  inject('showNetFlowTop', showNetFlowTop)
  inject('getNetFlowSenderList', getNetFlowSenderList)
  inject('getNetFlowServiceList', getNetFlowServiceList)
  inject('getNetFlowIPFlowList', getNetFlowIPFlowList)
  inject('showNetFlowGraph', showNetFlowGraph)
  inject('showNetFlowService3D', showNetFlowService3D)
  inject('showNetFlowSender3D', showNetFlowSender3D)
  inject('showNetFlowIPFlow3D', showNetFlowIPFlow3D)
  inject('getNetFlowFFTMap', getNetFlowFFTMap)
  inject('showNetFlowFFT', showNetFlowFFT)
  inject('showNetFlowFFT3D', showNetFlowFFT3D)
}
