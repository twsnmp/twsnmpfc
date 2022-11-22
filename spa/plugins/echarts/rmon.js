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

const showRMONHostsChart = (div, type, list) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('hosts_' + type)
  let topN = []
  switch (type) {
    case 'packtes':
      topN = list
        .slice()
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
      topN = list
        .slice()
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

const showRMONMatrixChart = (div, type, list) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('matrix_' + type)
  let topN = []
  switch (type) {
    case 'packtes':
      topN = list
        .slice()
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
      topN = list
        .slice()
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

const showRMONProtocolChart = (div, type, list) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = baseChartOption('protocol_' + type)
  let topN = []
  switch (type) {
    case 'packtes':
      topN = list
        .slice()
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
      topN = list
        .slice()
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

export default (context, inject) => {
  inject('showRMONStatisticsChart', showRMONStatisticsChart)
  inject('showRMONHistoryChart', showRMONHistoryChart)
  inject('showRMONHostsChart', showRMONHostsChart)
  inject('showRMONMatrixChart', showRMONMatrixChart)
  inject('showRMONProtocolChart', showRMONProtocolChart)
}
