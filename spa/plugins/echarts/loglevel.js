import * as echarts from 'echarts'

let chart

const makeLogLevelChart = (div) => {
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
        dataZoom: {},
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    dataZoom: [{}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '5%',
      right: '5%',
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
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
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
      name: '件数',
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
        name: '重度',
        type: 'bar',
        color: '#e31a1c',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '軽度',
        type: 'bar',
        color: '#fb9a99',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '注意',
        type: 'bar',
        color: '#dfdf22',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'その他',
        type: 'bar',
        color: '#1f78b4',
        stack: 'count',
        large: true,
        data: [],
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['重度', '軽度', '注意', 'その他'],
    },
  }
  chart.setOption(option)
  chart.resize()
}

const addChartData = (data, count, ctm, newCtm) => {
  let t = new Date(ctm * 60 * 1000)
  for (const k in count) {
    data[k].push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, count[k]],
    })
  }
  ctm++
  for (; ctm < newCtm; ctm++) {
    t = new Date(ctm * 60 * 1000)
    for (const k in count) {
      data[k].push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, 0],
      })
    }
  }
  return ctm
}

const showLogLevelChart = (div, logs, zoomCallback) => {
  if (chart) {
    chart.dispose()
  }
  makeLogLevelChart(div)
  const data = {
    high: [],
    low: [],
    warn: [],
    other: [],
  }
  const count = {
    high: 0,
    low: 0,
    warn: 0,
    other: 0,
  }
  let ctm
  let st = Infinity
  let lt = 0
  logs.forEach((e) => {
    const lvl = data[e.Level] ? e.Level : 'other'
    const newCtm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60))
    if (!ctm) {
      ctm = newCtm
    }
    if (ctm !== newCtm) {
      ctm = addChartData(data, count, ctm, newCtm)
      for (const k in count) {
        count[k] = 0
      }
    }
    count[lvl]++
    if (st > e.Time) {
      st = e.Time
    }
    if (lt < e.Time) {
      lt = e.Time
    }
  })
  addChartData(data, count, ctm, ctm + 1)
  chart.setOption({
    series: [
      {
        data: data.high,
      },
      {
        data: data.low,
      },
      {
        data: data.warn,
      },
      {
        data: data.other,
      },
    ],
  })
  chart.resize()
  if (zoomCallback) {
    chart.on('datazoom', (e) => {
      if (e.batch && e.batch.length === 2 && e.batch[0].startValue) {
        zoomCallback(
          e.batch[0].startValue * 1000 * 1000,
          e.batch[0].endValue * 1000 * 1000
        )
      } else if (e.start !== undefined && e.end !== undefined) {
        zoomCallback(
          st + (lt - st) * (e.start / 100),
          st + (lt - st) * (e.end / 100)
        )
      }
    })
  }
}

const resizeLogLevelChart = () => {
  if (chart) {
    chart.resize()
  }
}

export default (context, inject) => {
  inject('showLogLevelChart', showLogLevelChart)
  inject('resizeLogLevelChart', resizeLogLevelChart)
}
