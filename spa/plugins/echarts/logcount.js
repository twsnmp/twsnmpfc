import * as echarts from 'echarts'

let chart

const makeLogCountChart = (div) => {
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
          return echarts.format.formatTime('MM/dd hh:mm', date)
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

const showLogCountChart = (logs) => {
  const data = []
  let count = 0
  let ctm
  logs.forEach((e) => {
    if (!ctm) {
      ctm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60))
      count++
      return
    }
    const newCtm = Math.floor(e.Time / (1000 * 1000 * 1000 * 60))
    if (ctm !== newCtm) {
      let t = new Date(ctm * 60 * 1000)
      data.push({
        name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
        value: [t, count],
      })
      for (; ctm < newCtm; ctm++) {
        t = new Date(ctm * 60 * 1000)
        data.push({
          name: echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t),
          value: [t, 0],
        })
      }
      count = 0
    }
    count++
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

export default (context, inject) => {
  inject('makeLogCountChart', makeLogCountChart)
  inject('showLogCountChart', showLogCountChart)
}
