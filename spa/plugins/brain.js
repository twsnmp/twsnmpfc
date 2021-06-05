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
        name: 'cpu',
        type: 'line',
        color: '#1f78b4',
        showSymbol: false,
        hoverAnimation: false,
        large: true,
        data: [],
      },
      {
        name: 'gpu',
        type: 'line',
        color: '#f11',
        showSymbol: false,
        hoverAnimation: false,
        large: true,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const testBrain = (divError, divModel) => {
  makeErrorChart(divError)
  const net = new brain.NeuralNetwork()
  const xorTrainingData = [
    { input: [0, 0], output: [0] },
    { input: [0, 1], output: [1] },
    { input: [1, 0], output: [1] },
    { input: [1, 1], output: [0] },
  ]
  const data = []
  net.trainAsync(xorTrainingData, {
    iterations: 20000,
    errorThresh: 0.005,
    callbackPeriod: 100,
    callback: (p) => {
      const t = new Date()
      data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, p.error],
      })
      chart.setOption({
        series: [
          {
            data,
          },
        ],
      })
    },
  })
  const model = document.getElementById(divModel)
  if (model) {
    model.innerHTML = brain.utilities.toSVG(net)
  }
}

const autoEncoder = (divError, divModel, req, callback) => {
  makeErrorChart(divError)
  if (req.Data.length < 5) {
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

export default (context, inject) => {
  inject('testBrain', testBrain)
  inject('autoEncoder', autoEncoder)
}
