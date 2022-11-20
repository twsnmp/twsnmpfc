import * as echarts from 'echarts'

let chart

const showRMONStatisticsChart = (div, type, list) => {
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
        saveAsImage: { name: 'twsnmp_rmon_statistics_' + type },
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
      name: type === 'bytes' ? 'Bytes' : 'Packtes',
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

export default (context, inject) => {
  inject('showRMONStatisticsChart', showRMONStatisticsChart)
}
