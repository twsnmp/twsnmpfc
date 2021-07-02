import * as brain from 'brain.js/src/index'
import * as echarts from 'echarts'
import 'echarts-gl'
import * as ecStat from 'echarts-stat'

let chart

const makeErrorChart = (div) => {
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
      name: 'Error',
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
        name: 'error',
        type: 'line',
        color: '#1f78b4',
        showSymbol: false,
        large: true,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const autoEncoder = (divError, divModel, req, callback) => {
  makeErrorChart(divError)
  if (!req || !req.Data || req.Data.length < 5) {
    if (callback) {
      // eslint-disable-next-line node/no-callback-literal
      callback(false)
    }
  }
  const l1 = Math.floor(Math.max(2, req.Data[0].length / 2))
  const l2 = Math.floor(Math.max(1, req.Data[0].length / 4))
  const net = new brain.NeuralNetwork({
    hiddenLayers: [l1, l2, l1],
  })
  const trainingData = []
  req.Data.forEach((l) => {
    trainingData.push({
      input: l,
      output: l,
    })
  })
  const errors = []
  net
    .trainAsync(trainingData, {
      iterations: 20000,
      timeout: 1000 * 20,
      errorThresh: 0.01,
      callbackPeriod: 5,
      callback: (p) => {
        const t = new Date()
        errors.push({
          name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
          value: [t, p.error],
        })
        chart.setOption({
          series: [
            {
              data: errors,
            },
          ],
        })
      },
    })
    .then((res) => {
      const done = setAIScore(net, req)
      if (callback) {
        callback(done)
      }
    })
    .catch((e) => {
      if (callback) {
        // eslint-disable-next-line node/no-callback-literal
        callback(false)
      }
    })
  const model = document.getElementById(divModel)
  if (model) {
    model.innerHTML = brain.utilities.toSVG(net)
  }
}

const setAIScore = (net, req) => {
  const diffs = []
  req.Data.forEach((d) => {
    const o = net.run(d)
    let diff = 0.0
    for (let i = 0; i < d.length; i++) {
      diff += Math.abs(o[i] - d[i])
    }
    diffs.push(diff)
  })
  const mean = ecStat.statistics.mean(diffs)
  const sd = ecStat.statistics.deviation(diffs)
  if (!sd) {
    return false
  }
  req.AIScores = []
  for (let i = 0; i < req.TimeStamp.length; i++) {
    req.AIScores.push([
      req.TimeStamp[i],
      Math.round((10 * (diffs[i] - mean)) / sd + 50),
    ])
  }
  return true
}

const errors = []

const showSyslogAIAssistChart = (divError) => {
  makeErrorChart(divError)
  chart.setOption({
    series: [
      {
        data: errors,
      },
    ],
  })
  chart.resize()
}

const syslogAIAssist = (logs) => {
  errors.length = 0
  const trainingData = []
  const keys = new Map()
  logs.forEach((l) => {
    if (l.AIClass && l.AIClass !== '') {
      trainingData.push({
        input: l.Tag + ' ' + l.Message,
        output: l.AIClass,
      })
      keys.set(l.AIClass, true)
    }
    l.AIResult = ''
  })
  const net = new brain.recurrent.LSTM()
  return new Promise((resolve) => {
    setTimeout(() => {
      net.train(trainingData, {
        iterations: 1000,
        errorThresh: 0.011,
        callbackPeriod: 20,
        callback: (p) => {
          const t = new Date()
          errors.push({
            name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
            value: [t, p.error],
          })
        },
      })
      logs.forEach((l) => {
        const air = net.run(l.Tag + ' ' + l.Message)
        if (keys.get(air)) {
          l.AIResult = air
        }
      })
      resolve('done')
    }, 500)
  })
}

export default (context, inject) => {
  inject('autoEncoder', autoEncoder)
  inject('syslogAIAssist', syslogAIAssist)
  inject('showSyslogAIAssistChart', showSyslogAIAssistChart)
}
