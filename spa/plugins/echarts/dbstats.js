import * as echarts from 'echarts'

let chart

const makeDBStatsChart = (div) => {
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
    legend: {
      data: ['データ数/秒', 'DBサイズ'],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter: (value, index) => {
          const date = new Date(value)
          return echarts.time.format(date, '{MM}/{dd} {hh}:{mm}')
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
    yAxis: [
      {
        type: 'value',
        name: 'Write/Sec',
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
      {
        type: 'value',
        name: 'Bytes',
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
    ],
    series: [
      {
        name: 'データ数/秒',
        type: 'bar',
        color: '#1f78b4',
        large: true,
        data: [],
      },
      {
        name: 'DBサイズ',
        type: 'line',
        color: '#e31a1c',
        large: true,
        yAxisIndex: 1,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showDBStatsChart = (logs) => {
  const speed = []
  const size = []
  if (!logs) {
    return
  }
  logs.forEach((e) => {
    const t = new Date(e.Time / (1000 * 1000))
    speed.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {hh}:{mm}:{ss}'),
      value: [t, e.Speed],
    })
    size.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {hh}:{mm}:{ss}'),
      value: [t, e.Size],
    })
  })
  chart.setOption({
    series: [
      {
        data: speed,
      },
      {
        data: size,
      },
    ],
  })
  chart.resize()
}

export default (context, inject) => {
  inject('makeDBStatsChart', makeDBStatsChart)
  inject('showDBStatsChart', showDBStatsChart)
}
