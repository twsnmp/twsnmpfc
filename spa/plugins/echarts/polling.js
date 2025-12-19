import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { probit } from 'simple-statistics'
import { getChartParams } from '~/plugins/echarts/chartparams.js'

let chart

const getPollingChartOption = (div) => {
  return {
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
      left: '10%',
      right: '10%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: [''],
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
    },
    xAxis: {
      type: 'time',
      name: '日時',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
        },
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
        nameTextStyle: {
          color: '#ccc',
          fontSize: 10,
          margin: 2,
        },
        axisLabel: {
          color: '#ccc',
          fontSize: 8,
          margin: 2,
        },
        axisLine: {
          lineStyle: {
            color: '#ccc',
          },
        },
      },
    ],
    series: [
      {
        color: '#1f78b4',
        type: 'line',
        showSymbol: false,
        data: [],
      },
    ],
  }
}

const setChartData = (series, t, values) => {
  const data = [t.getTime() * 1000 * 1000]
  const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
  const mean = ecStat.statistics.mean(values)
  series[0].data.push({
    name,
    value: [t, mean],
  })
  data.push(mean)
  const max = ecStat.statistics.max(values)
  series[1].data.push({
    name,
    value: [t, max],
  })
  data.push(max)
  const min = ecStat.statistics.min(values)
  series[2].data.push({
    name,
    value: [t, min],
  })
  data.push(min)
  const median = ecStat.statistics.median(values)
  series[3].data.push({
    name,
    value: [t, median],
  })
  data.push(median)
  const variance = ecStat.statistics.sampleVariance(values)
  series[4].data.push({
    name,
    value: [t, variance],
  })
  data.push(variance)
}

const showPollingChart = (div, polling, logs, ent, at, per1h) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))

  const option = getPollingChartOption(div)

  chart.setOption(option)
  if (ent === '') {
    chart.resize()
    return
  }
  const dp = getChartParams(ent)
  if (per1h) {
    option.series[0].name = '平均値'
    option.series.push({
      name: '最大値',
      type: 'line',
      large: true,
      data: [],
    })
    option.series.push({
      name: '最小値',
      type: 'line',
      large: true,
      data: [],
    })
    option.series.push({
      name: '中央値',
      type: 'line',
      large: true,
      data: [],
    })
    option.series.push({
      name: '分散',
      type: 'line',
      large: true,
      yAxisIndex: 1,
      data: [],
    })
    option.yAxis.push({
      type: 'value',
      name: '分散',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    })
    option.legend.data[0] = '平均値'
    option.legend.data.push('最大値')
    option.legend.data.push('最小値')
    option.legend.data.push('中央値')
    option.legend.data.push('分散')
    let tS = -1
    const values = []
    const dt = 3600 * 1000
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000))
      const tC = Math.floor(t.getTime() / dt)
      if (tS !== tC) {
        if (tS > 0) {
          if (values.length > 0) {
            tS++
            setChartData(option.series, new Date(tS * dt), values)
            values.length = 0
            while (tS < tC) {
              tS++
              setChartData(option.series, new Date(tS * dt), [0, 0, 0, 0])
            }
          }
        }
        tS = tC
      }
      let numVal = getNumVal(ent, l.Result)
      numVal *= dp.mul
      values.push(numVal || 0.0)
    })
    if (values.length > 0) {
      tS++
      setChartData(option.series, new Date(tS * dt), values)
    }
  } else {
    logs.forEach((l) => {
      const t = new Date(l.Time / (1000 * 1000))
      const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
      let numVal = getNumVal(ent, l.Result)
      numVal *= dp.mul
      option.series[0].data.push({
        ts,
        value: [t, numVal || 0.0],
      })
    })
    option.series[0].name = dp.axis
    option.legend.data[0] = dp.axis
  }
  option.yAxis.name = dp.axis
  if (at && at.UnixTime) {
    const st = new Date((at.UnixTime - 3600) * 1000)
    const et = new Date((at.UnixTime + 3600) * 1000)
    option.series[0].markArea = {
      itemStyle: {
        color: 'rgba(92,92, 192, 0.2)',
      },
      data: [
        [
          {
            xAxis: st,
          },
          {
            xAxis: et,
          },
        ],
      ],
    }
  }
  chart.setOption(option)
  chart.resize()
}

const makePollingHistogram = (div) => {
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
      min: 0,
    },
    yAxis: {
      name: '回数',
    },
    series: [
      {
        color: '#1f78b4',
        type: 'bar',
        showSymbol: false,
        barWidth: '99.3%',
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showPollingHistogram = (div, logs, ent) => {
  makePollingHistogram(div)
  if (ent === '') {
    return
  }
  const data = []
  const dp = getChartParams(ent)
  logs.forEach((l) => {
    if (!l.Result.error) {
      let numVal = getNumVal(ent, l.Result)
      numVal *= dp.mul
      data.push(numVal)
    }
  })
  if (data.length < 1) {
    return
  }
  const bins = ecStat.histogram(data)
  chart.setOption({
    xAxis: {
      name: dp.axis,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize()
}

const getChartModeName = (mode) => {
  const r = getChartParams(mode)
  if (r && r.axis) {
    return r.axis
  }
  return mode
}

const setDataList = (r, numValEntList) => {
  Object.keys(r).forEach((k) => {
    const v = r[k]
    if (
      !isNaN(parseFloat(v)) &&
      k !== 'lastTime' &&
      numValEntList.findIndex((e) => e.value === k) === -1
    ) {
      numValEntList.push({
        text: getChartModeName(k),
        value: k,
      })
    }
  })
}

const getNumVal = (key, r) => {
  return r[key] || 0.0
}

const makeSTLChart = (div) => {
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
        type: 'cross',
        label: {
          backgroundColor: '#6a7985',
        },
      },
    },
    legend: {
      data: ['Seasonal', 'Trend', 'Resid'],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: '10%',
      buttom: '10%',
    },
    xAxis: {
      type: 'time',
      name: '日時',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
        formatter(value, index) {
          const date = new Date(value)
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
        },
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
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    series: [
      {
        name: 'Resid',
        type: 'line',
        stack: 'stl',
        color: '#fac858',
        areaStyle: {},
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: [],
      },
      {
        name: 'Trend',
        type: 'line',
        stack: 'stl',
        color: '#91cc75',
        areaStyle: {},
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: [],
      },
      {
        name: 'Seasonal',
        type: 'line',
        color: '#5470c6',
        stack: 'stl',
        areaStyle: {},
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showPollingLogSTL = (div, polling, data, ent, unit) => {
  makeSTLChart(div)
  if (ent === '' || !data || !data.StlMapH || !data.StlMapH[ent]) {
    return
  }
  const timeList = unit !== 'h' ? data.TimePX2 : data.TimeH
  const stlMap = unit !== 'h' ? data.StlMapPX2 : data.StlMapH
  const seasonal = []
  const trend = []
  const resid = []
  const dp = getChartParams(ent)
  for (let i = 0; i < timeList.length; i++) {
    const t = new Date(timeList[i] * 1000)
    const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    seasonal.push({
      name: ts,
      value: [t, (stlMap[ent].Seasonal[i] || 0.0) * dp.mul],
    })
    trend.push({
      name: ts,
      value: [t, (stlMap[ent].Trend[i] || 0.0) * dp.mul],
    })
    resid.push({
      name: ts,
      value: [t, (stlMap[ent].Resid[i] || 0.0) * dp.mul],
    })
  }
  const opt = {
    yAxis: {
      name: dp.axis,
    },
    series: [
      {
        name: 'Resid',
        type: 'line',
        stack: 'stl',
        color: '#fac858',
        areaStyle: {},
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: resid,
      },
      {
        name: 'Trend',
        type: 'line',
        stack: 'stl',
        color: '#91cc75',
        areaStyle: {},
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: trend,
      },
      {
        name: 'Seasonal',
        type: 'line',
        color: '#5470c6',
        stack: 'stl',
        areaStyle: {},
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: seasonal,
      },
    ],
  }
  chart.setOption(opt)
  chart.resize()
}

const makeFFTChart = (div) => {
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
      left: '10%',
      right: '10%',
      top: '10%',
      buttom: '10%',
    },
    xAxis: {
      type: 'value',
      name: 'Hz',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: '8px',
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
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    series: [],
  }
  chart.setOption(option)
  chart.resize()
}

const showPollingLogFFT = (div, polling, data, ent, unit, fftType) => {
  makeFFTChart(div)
  if (ent === '' || !data.FFTH || !data.FFTH[ent]) {
    return
  }
  const fftin = unit !== 'h' ? data.FFTPX2[ent] : data.FFTH[ent]
  if (!fftin) {
    return
  }
  const freq = fftType === 'hz'
  const dp = getChartParams(ent)
  const fft = []
  for (let i = 0; i < fftin.length; i++) {
    if (freq) {
      fft.push([fftin[i][0], fftin[i][1] * dp.mul])
    } else {
      fft.push([
        fftin[i][0] > 0.0 ? 1.0 / fftin[i][0] : 0.0,
        fftin[i][1] * dp.mul,
      ])
    }
  }
  const opt = {
    yAxis: {
      name: dp.axis,
    },
    xAxis: {
      name: freq ? '周波数(Hz)' : '周期(Sec)',
    },
    series: [
      {
        name: dp.axis,
        type: 'bar',
        color: '#5470c6',
        emphasis: {
          focus: 'series',
        },
        showSymbol: false,
        data: fft,
      },
    ],
  }
  chart.setOption(opt)
  chart.resize()
}

let forecastChart

const showPollingLogForecast = (div, polling, logs, ent) => {
  if (forecastChart) {
    forecastChart.dispose()
  }
  forecastChart = echarts.init(document.getElementById(div))
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
    yAxis: [
      {
        type: 'value',
        name: '',
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
        name: '',
        type: 'line',
        large: true,
        symbol: 'none',
        data: [],
      },
    ],
  }
  if (ent !== '') {
    const dp = getChartParams(ent)
    const data = []
    logs.forEach((l) => {
      if (!l.Result.error) {
        let numVal = getNumVal(ent, l.Result)
        numVal *= dp.mul
        data.push([l.Time / (1000 * 1000), numVal])
      }
    })
    const reg = ecStat.regression('linear', data)
    const sd = Math.floor(Date.now() / (24 * 3600 * 1000))
    for (let d = sd; d < sd + 365; d++) {
      const x = d * 24 * 3600 * 1000
      const t = new Date(x)
      option.series[0].data.push({
        name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        value: [t, reg.parameter.intercept + reg.parameter.gradient * x],
      })
    }
    if (dp.vmap) {
      option.visualMap = {
        top: 50,
        right: 10,
        textStyle: {
          color: '#ccc',
          fontSize: 8,
        },
        pieces: dp.vmap,
        outOfRange: {
          color: '#eee',
        },
      }
    }
  }
  forecastChart.setOption(option)
  forecastChart.resize()
}

/**
 * QQプロット用のデータセットを計算する関数
 * @param {Array<number>} data - 標本データ
 * @returns {Array<Array<number>>} [[理論分位点, 標本分位点], ...] の形式
 */
const calculateQQPlotData = (data, mu, sigma) => {
  const n = data.length
  if (n === 0) return []

  // 1. 標本データをソート (Sample Quantiles)
  const sortedData = [...data].sort((a, b) => a - b)

  const qqData = []

  // 2. 理論的分位点 (Theoretical Quantiles) の計算
  // 標準正規分布 (平均0, 標準偏差1) を仮定
  // Cunnaneのプロット位置: p_i = (i - 0.5) / n を使用

  for (let i = 0; i < n; i++) {
    const rank = i + 1 // 順位 (1 から n)

    // 累積確率 p
    const p = (rank - 0.5) / n

    // 平均 mu, 標準偏差 sigma の正規分布の場合
    const theoreticalQuantile = mu + sigma * probit(p)
    const sampleQuantile = sortedData[i]

    // [理論分位点, 標本分位点] のペアを保存
    qqData.push([theoreticalQuantile, sampleQuantile])
  }
  return qqData
}

const showPollingQQPlot = (div, logs, ent) => {
  if (ent === '') {
    return
  }
  const data = []
  const dp = getChartParams(ent)
  logs.forEach((l) => {
    if (!l.Result.error) {
      let numVal = getNumVal(ent, l.Result)
      numVal *= dp.mul
      data.push(numVal)
    }
  })
  const mu = ecStat.statistics.mean(data)
  const sigma = ecStat.statistics.deviation(data)
  const qqPlotData = calculateQQPlotData(data, mu, sigma)
  const scatterData = qqPlotData
  // 基準線の座標 (対角線 y=x)
  const lineMin = mu - 4 * sigma // 平均から下方向に4標準偏差
  const lineMax = mu + 4 * sigma // 平均から上方向に4標準偏差
  const lineData = [
    [lineMin, lineMin],
    [lineMax, lineMax],
  ]
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div), 'dark')
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        if (params.seriesName === 'Sample plot') {
          return `理論分位点(Z): 
            ${params.value[0].toFixed(4)}
            <br/>標本分位点: ${params.value[1].toFixed(4)}`
        }
        return params.seriesName
      },
    },
    graphic: [
      {
        type: 'text',
        left: '10%',
        bottom: '15%',
        style: {
          text: `μ: ${mu.toFixed(2)}\nσ: ${sigma.toFixed(2)}`,
          fill: '#fff',
          font: '14px sans-serif',
        },
      },
    ],
    grid: {
      left: '10%',
      right: '15%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      name: '理論分位点',
      type: 'value',
      scale: true,
      splitLine: {
        show: false,
      },
    },
    yAxis: {
      name: '標本分位点',
      type: 'value',
      scale: true,
      splitLine: {
        show: false,
      },
    },
    series: [
      {
        name: 'Reference line (y=x)',
        type: 'line',
        data: lineData,
        symbol: 'none',
        lineStyle: {
          color: 'red',
          type: 'dashed',
        },
        z: 1,
      },
      {
        name: 'Sample plot',
        type: 'scatter',
        data: scatterData,
        symbolSize: 8,
        itemStyle: {
          color: 'steelblue',
        },
        z: 2,
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showPollingChart', showPollingChart)
  inject('showPollingHistogram', showPollingHistogram)
  inject('setDataList', setDataList)
  inject('showPollingLogSTL', showPollingLogSTL)
  inject('showPollingLogFFT', showPollingLogFFT)
  inject('showPollingLogForecast', showPollingLogForecast)
  inject('showPollingQQPlot', showPollingQQPlot)
}
