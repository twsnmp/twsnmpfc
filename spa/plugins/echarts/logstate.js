import * as echarts from 'echarts'

let chart

const makeLogStateChart = (div) => {
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
        name: '正常',
        type: 'bar',
        color: '#33a02c',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '不明',
        type: 'bar',
        color: 'gray',
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
      data: ['重度', '軽度', '注意', '正常', '不明'],
    },
  }
  chart.setOption(option)
  chart.resize()
}

const showLogStateChart = (logs) => {
  const data = {
    high: [],
    low: [],
    warn: [],
    normal: [],
    unknown: [],
  }
  const count = {
    high: 0,
    low: 0,
    warn: 0,
    normal: 0,
    unknown: 0,
  }
  let cth
  logs.forEach((e) => {
    const lvl = data[e.State] ? e.State : 'normal'
    if (!cth) {
      cth = Math.floor(e.Time / (1000 * 1000 * 1000 * 3600))
      count[lvl]++
      return
    }
    const newCth = Math.floor(e.Time / (1000 * 1000 * 1000 * 3600))
    if (cth !== newCth) {
      let t = new Date(cth * 3600 * 1000)
      for (const k in count) {
        data[k].push({
          name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
          value: [t, count[k]],
        })
      }
      cth++
      for (; cth < newCth; cth++) {
        t = new Date(cth * 3600 * 1000)
        for (const k in count) {
          data[k].push({
            name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
            value: [t, 0],
          })
        }
      }
      for (const k in count) {
        count[k] = 0
      }
    }
    count[lvl]++
  })
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
        data: data.normal,
      },
      {
        data: data.unknown,
      },
    ],
  })
  chart.resize()
}

export default (context, inject) => {
  inject('makeLogStateChart', makeLogStateChart)
  inject('showLogStateChart', showLogStateChart)
}
