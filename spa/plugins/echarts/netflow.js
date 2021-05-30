import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

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
    },
    yAxis: {
      name: '回数',
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        hoverAnimation: false,
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
  ['636/tcp', 'LDAPS'],
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
  ['943/tcp', 'MAIL'],
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

export default (context, inject) => {
  inject('showNetFlowHistogram', showNetFlowHistogram)
  inject('showNetFlowCluster', showNetFlowCluster)
  inject('showNetFlowTraffic', showNetFlowTraffic)
  inject('showNetFlowTop', showNetFlowTop)
  inject('getNetFlowSenderList', getNetFlowSenderList)
  inject('getNetFlowServiceList', getNetFlowServiceList)
  inject('getNetFlowIPFlowList', getNetFlowIPFlowList)
}
