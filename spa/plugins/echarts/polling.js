import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
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

const showPollingHistogram = (div, polling, logs, ent) => {
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

export default (context, inject) => {
  inject('showPollingChart', showPollingChart)
  inject('showPollingHistogram', showPollingHistogram)
  inject('setDataList', setDataList)
  inject('showPollingLogSTL', showPollingLogSTL)
  inject('showPollingLogFFT', showPollingLogFFT)
  inject('showPollingLogForecast', showPollingLogForecast)
}
