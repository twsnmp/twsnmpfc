import * as echarts from 'echarts'

let chart

const baseChartOption = (type) => {
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
    legend: {
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_rmon_' + type },
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
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
      name: type.includes('bytes') ? 'Bytes' : 'Packtes',
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
}

const showRMONStatisticsChart = (div, type, list) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('statistics_' + type)
  option.yAxis.data = list.map((x) => x.etherStatsDataSource)
  switch (type) {
    case 'packtes':
      option.series = [
        {
          name: 'パケット数',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts),
        },
        {
          name: 'ブロードキャスト',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsBroadcastPkts),
        },
        {
          name: 'マルチキャスト',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsMulticastPkts),
        },
        {
          name: 'エラー',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsErrors),
        },
      ]
      break
    case 'bytes':
      option.series = [
        {
          name: 'バイト数',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsOctets),
        },
      ]
      break
    case 'size':
      option.series = [
        {
          name: '=64',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts64Octets),
        },
        {
          name: '65-127',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts65to127Octets),
        },
        {
          name: '128-255',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts128to255Octets),
        },
        {
          name: '256-511',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts256to511Octets),
        },
        {
          name: '512-1023',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts512to1023Octets),
        },
        {
          name: '1024-1518',
          type: 'bar',
          stack: 'packets',
          data: list.map((x) => x.etherStatsPkts1024to1518Octets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONHistoryChart = (div, type, list) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('history_' + type)
  const tmp = list
    .slice()
    .sort((a, b) => a.etherHistoryIntervalStart - b.etherHistoryIntervalStart)
    .reverse()
  option.yAxis.data = tmp.map((x) => x.Index)
  switch (type) {
    case 'packtes':
      option.series = [
        {
          name: 'パケット数',
          type: 'bar',
          stack: 'packets',
          data: tmp.map((x) => x.etherHistoryPkts),
        },
        {
          name: 'ブロードキャスト',
          type: 'bar',
          stack: 'packets',
          data: tmp.map((x) => x.etherHistoryBroadcastPkts),
        },
        {
          name: 'マルチキャスト',
          type: 'bar',
          stack: 'packets',
          data: tmp.map((x) => x.etherHistoryMulticastPkts),
        },
        {
          name: 'エラー',
          type: 'bar',
          stack: 'packets',
          data: tmp.map((x) => x.etherHistoryErrors),
        },
        {
          name: 'ドロップ',
          type: 'bar',
          stack: 'packets',
          data: tmp.map((x) => x.etherHistoryDropEvents),
        },
      ]
      break
    case 'bytes':
      option.series = [
        {
          name: 'バイト数',
          type: 'bar',
          stack: 'packets',
          data: tmp.map((x) => x.etherHistoryOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONHostsChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('hosts_' + type)
  let topN = list
    .slice()
    .filter(
      (x) =>
        (!filter.hostTimeAddress ||
          x.hostTimeAddress.includes(filter.hostTimeAddress)) &&
        (!filter.Vendor || x.Vendor.includes(filter.Vendor))
    )
  switch (type) {
    case 'packtes':
      topN = topN
        .sort(
          (a, b) =>
            b.hostTimeOutPkts +
            b.hostTimeInPkts -
            (a.hostTimeOutPkts + a.hostTimeInPkts)
        )
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.hostTimeAddress)
      option.series = [
        {
          name: '送信パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.hostTimeOutPkts),
        },
        {
          name: 'ブロードキャスト',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.hostTimeOutBroadcastPkts),
        },
        {
          name: 'マルチキャスト',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.hostTimeOutMulticastPkts),
        },
        {
          name: 'エラー',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.hostTimeOutErrors),
        },
        {
          name: '受信パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.hostTimeInPkts),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort(
          (a, b) =>
            b.hostTimeInOctets +
            b.hostTimeOutOctets -
            (a.hostTimeInOctets + a.hostTimeOutOctets)
        )
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.hostTimeAddress)
      option.series = [
        {
          name: '受信バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.hostTimeInOctets),
        },
        {
          name: '送信バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.hostTimeOutOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONMatrixChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('matrix_' + type)
  let topN = list
    .slice()
    .filter(
      (x) =>
        (!filter.matrixSDSourceAddress ||
          x.matrixSDSourceAddress.includes(filter.matrixSDSourceAddress)) &&
        (!filter.matrixSDDestAddress ||
          x.matrixSDDestAddress.includes(filter.matrixSDDestAddress))
    )
  switch (type) {
    case 'packtes':
      topN = topN
        .sort((a, b) => b.matrixSDPkts - a.matrixSDPkts)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map(
        (x) => x.matrixSDSourceAddress + '=>' + x.matrixSDDestAddress
      )
      option.series = [
        {
          name: 'パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.matrixSDPkts),
        },
        {
          name: 'エラー',
          type: 'bar',
          stack: 'packets',
          color: 'red',
          data: topN.map((x) => x.matrixSDErrors),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort((a, b) => b.matrixSDOctets - a.matrixSDOctets)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map(
        (x) => x.matrixSDSourceAddress + '=>' + x.matrixSDDestAddress
      )
      option.series = [
        {
          name: 'バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.matrixSDOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONProtocolChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('protocol_' + type)
  let topN = list.slice().filter((x) => !filter || x.Protocol.includes(filter))
  switch (type) {
    case 'packtes':
      topN = topN
        .sort((a, b) => b.protocolDistStatsPkts - a.protocolDistStatsPkts)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.Protocol)
      option.series = [
        {
          name: 'パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.protocolDistStatsPkts),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort((a, b) => b.protocolDistStatsOctets - a.protocolDistStatsOctets)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.Protocol)
      option.series = [
        {
          name: 'バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.protocolDistStatsOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONNlHostsChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('nlHosts_' + type)
  let topN = list
    .slice()
    .filter((x) => !filter || x.nlHostAddress.includes(filter))
  switch (type) {
    case 'packtes':
      topN = topN
        .sort(
          (a, b) =>
            b.nlHostInPkts + b.nlHostOutPkts - a.nlHostInPkts - a.nlHostOutPkts
        )
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.nlHostAddress)
      option.series = [
        {
          name: '受信パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.nlHostInPkts),
        },
        {
          name: '送信パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.nlHostOutPkts),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort(
          (a, b) =>
            b.nlHostInOctets +
            b.nlHostOutOctets -
            a.nlHostInOctets -
            a.nlHostOutOctets
        )
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.nlHostAddress)
      option.series = [
        {
          name: '受信バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.nlHostInOctets),
        },
        {
          name: '送信バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.nlHostOutOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONNlMatrixChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('nlMaatrix_' + type)
  let topN = list
    .slice()
    .filter(
      (x) =>
        (!filter.nlMatrixSDSourceAddress ||
          x.nlMatrixSDSourceAddress.includes(filter.nlMatrixSDSourceAddress)) &&
        (!filter.nlMatrixSDDestAddress ||
          x.nlMatrixSDDestAddress.includes(filter.nlMatrixSDDestAddress))
    )
  switch (type) {
    case 'packtes':
      topN = topN
        .sort((a, b) => b.nlMatrixSDPkts - a.nlMatrixSDPkts)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map(
        (x) => x.nlMatrixSDSourceAddress + '=>' + x.nlMatrixSDDestAddress
      )
      option.series = [
        {
          name: 'パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.nlMatrixSDPkts),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort((a, b) => b.nlMatrixSDOctets - a.nlMatrixSDOctets)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map(
        (x) => x.nlMatrixSDSourceAddress + '=>' + x.nlMatrixSDDestAddress
      )
      option.series = [
        {
          name: 'バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.nlMatrixSDOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONAlHostsChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('alHosts_' + type)
  let topN = list
    .slice()
    .filter(
      (x) =>
        (!filter.alHostAddress ||
          x.alHostAddress.includes(filter.alHostAddress)) &&
        (!filter.Protocol || x.Protocol.includes(filter.Protocol))
    )
  switch (type) {
    case 'packtes':
      topN = topN
        .sort(
          (a, b) =>
            b.alHostInPkts + b.alHostOutPkts - a.alHostInPkts - a.alHostOutPkts
        )
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.alHostAddress + ':' + x.Protocol)
      option.series = [
        {
          name: '受信パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.alHostInPkts),
        },
        {
          name: '送信パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.alHostOutPkts),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort(
          (a, b) =>
            b.alHostInOctets +
            b.alHostOutOctets -
            a.alHostInOctets -
            a.alHostOutOctets
        )
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map((x) => x.alHostAddress + ':' + x.Protocol)
      option.series = [
        {
          name: '受信バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.alHostInOctets),
        },
        {
          name: '送信バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.alHostOutOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const showRMONAlMatrixChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('alMatrix_' + type)
  let topN = list
    .slice()
    .filter(
      (x) =>
        (!filter.alMatrixSDSourceAddress ||
          x.alMatrixSDSourceAddress.includes(filter.alMatrixSDSourceAddress)) &&
        (!filter.alMatrixSDDestAddress ||
          x.alMatrixSDDestAddress.includes(filter.alMatrixSDDestAddress)) &&
        (!filter.Protocol || x.Protocol.includes(filter.Protocol))
    )
  switch (type) {
    case 'packtes':
      topN = topN
        .sort((a, b) => b.alMatrixSDPkts - a.alMatrixSDPkts)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map(
        (x) =>
          x.alMatrixSDSourceAddress +
          '=>' +
          x.alMatrixSDDestAddress +
          ':' +
          x.Protocol
      )
      option.series = [
        {
          name: 'パケット数',
          type: 'bar',
          stack: 'packets',
          data: topN.map((x) => x.alMatrixSDPkts),
        },
      ]
      break
    case 'bytes':
      topN = topN
        .sort((a, b) => b.alMatrixSDOctets - a.alMatrixSDOctets)
        .slice(0, 20)
        .reverse()
      option.yAxis.data = topN.map(
        (x) =>
          x.alMatrixSDSourceAddress +
          '=>' +
          x.alMatrixSDDestAddress +
          ':' +
          x.Protocol
      )
      option.series = [
        {
          name: 'バイト数',
          type: 'bar',
          stack: 'bytes',
          data: topN.map((x) => x.alMatrixSDOctets),
        },
      ]
      break
  }
  chart.setOption(option)
  chart.resize()
}

const isLocalIP = (ip) => {
  return (
    ip.startsWith('192.168') || ip.startsWith('10.') || ip.startsWith('172.16.')
  )
}

const showRMONAddressMapChart = (div, type, list, filter) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const categories = [
    { name: 'MAC' },
    { name: 'ローカルIP' },
    { name: 'グローバルIP' },
    { name: '重複?IP' },
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
        saveAsImage: { name: 'twsnmp_rmon_addressmap' },
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
    color: ['#1f78b4', '#11cc00', '#cccc00', '#e31a1c'],
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    series: [
      {
        type: 'graph',
        layout: type || 'force',
        symbolSize: 8,
        categories,
        roam: true,
        force: {
          repulsion: 50,
          edgeLength: 10,
        },
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
  const nodes = {}
  list
    .slice()
    .filter(
      (x) =>
        (!filter.addressMapNetworkAddress ||
          x.addressMapNetworkAddress.includes(
            filter.addressMapNetworkAddress
          )) &&
        (!filter.addressMapPhysicalAddress ||
          x.addressMapPhysicalAddress.includes(
            filter.addressMapPhysicalAddress
          )) &&
        (!filter.Vendor || x.Vendor.includes(filter.Vendor))
    )
    .forEach((m) => {
      const mac = m.addressMapPhysicalAddress
      const ip = m.addressMapNetworkAddress
      if (!nodes[mac]) {
        nodes[mac] = {
          name: mac,
          category: 'MAC',
          draggable: true,
          value: 1,
          label: {
            show: false,
          },
        }
      } else {
        nodes[mac].value++
        if (nodes[mac].value > 50) {
          if (!isLocalIP(ip)) {
            return
          }
        }
      }
      if (!nodes[ip]) {
        nodes[ip] = {
          name: ip,
          category: isLocalIP(ip) ? 'ローカルIP' : 'グローバルIP',
          draggable: true,
          value: 1,
          label: {
            show: false,
          },
        }
      } else {
        nodes[ip].value++
        nodes[ip].category = '重複IP'
      }
      option.series[0].links.push({
        source: mac,
        target: ip,
        value: ':',
        lineStyle: {
          color: '#ddd',
        },
      })
    })
  for (const k in nodes) {
    option.series[0].data.push(nodes[k])
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showRMONStatisticsChart', showRMONStatisticsChart)
  inject('showRMONHistoryChart', showRMONHistoryChart)
  inject('showRMONHostsChart', showRMONHostsChart)
  inject('showRMONMatrixChart', showRMONMatrixChart)
  inject('showRMONProtocolChart', showRMONProtocolChart)
  inject('showRMONNlHostsChart', showRMONNlHostsChart)
  inject('showRMONNlMatrixChart', showRMONNlMatrixChart)
  inject('showRMONAlHostsChart', showRMONAlHostsChart)
  inject('showRMONAlMatrixChart', showRMONAlMatrixChart)
  inject('showRMONAddressMapChart', showRMONAddressMapChart)
}
