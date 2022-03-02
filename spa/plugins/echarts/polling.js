import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { getChartParams } from '~/plugins/echarts/chartparams.js'

let chart

const makePollingChart = (div) => {
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
      right: '5%',
      top: 30,
      buttom: 0,
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
        color: '#1f78b4',
        type: 'line',
        showSymbol: false,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showPollingChart = (polling, logs, ent, at) => {
  if (ent === '') {
    return
  }
  const data = []
  const dp = getChartParams(polling, ent)
  logs.forEach((l) => {
    const t = new Date(l.Time / (1000 * 1000))
    const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    let numVal = getNumVal(ent, l.Result)
    numVal *= dp.mul
    data.push({
      name: ts,
      value: [new Date(l.Time / (1000 * 1000)), numVal],
    })
  })
  const opt = {
    yAxis: {
      name: dp.axis,
    },
    series: [
      {
        name: dp.axis,
        data,
      },
    ],
  }
  if (at && at.UnixTime) {
    const st = new Date(at.UnixTime * 1000)
    const et = new Date((at.UnixTime + 3600) * 1000)
    opt.series[0].markArea = {
      itemStyle: {
        color: 'rgba(24,173, 172, 0.2)',
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
  chart.setOption(opt)
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

const showPollingHistogram = (polling, logs, ent) => {
  if (ent === '') {
    return
  }
  const data = []
  const dp = getChartParams(polling, ent)
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
  const dp = getChartParams(polling, ent)
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
  const dp = getChartParams(polling, ent)
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
    const dp = getChartParams(polling, ent)
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
  inject('makePollingChart', makePollingChart)
  inject('showPollingChart', showPollingChart)
  inject('makePollingHistogram', makePollingHistogram)
  inject('showPollingHistogram', showPollingHistogram)
  inject('setDataList', setDataList)
  inject('showPollingLogSTL', showPollingLogSTL)
  inject('showPollingLogFFT', showPollingLogFFT)
  inject('showPollingLogForecast', showPollingLogForecast)
}
