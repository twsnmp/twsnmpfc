import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

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
    tooltip: {
      trigger: 'axis',
      formatter(params) {
        const p = params[0]
        return p.name + ' : ' + p.value[1]
      },
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
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
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
  const dp = getDispParams(polling, ent)
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
        data,
      },
    ],
  }
  if (at && at.UnixTime) {
    const st = new Date(at.UnixTime * 1000)
    const et = new Date((at.UnixTime + 3600) * 1000)
    opt.series[0].markArea = {
      itemStyle: {
        color: 'rgba(245,173, 172, 0.4)',
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
  const dp = getDispParams(polling, ent)
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
  const r = chartDispInfo[mode]
  if (r) {
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

const chartDispInfo = {
  rtt: {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: '応答時間(秒)',
  },
  rtt_cv: {
    mul: 1.0,
    axis: '応答時間変動係数',
  },
  successRate: {
    mul: 100.0,
    axis: '成功率(%)',
  },
  speed: {
    mul: 1.0,
    axis: '回線速度(Mbps)',
  },
  speed_cv: {
    mul: 1.0,
    axis: '回線速度変動係数',
  },
  feels_like: {
    mul: 1.0,
    axis: '体感温度(℃）',
  },
  humidity: {
    mul: 1.0,
    axis: '湿度(%)',
  },
  pressure: {
    mul: 1.0,
    axis: '気圧(hPa)',
  },
  temp: {
    mul: 1.0,
    axis: '温度(℃）',
  },
  temp_max: {
    mul: 1.0,
    axis: '最高温度(℃）',
  },
  temp_min: {
    mul: 1.0,
    axis: '最低温度(℃）',
  },
  wind: {
    mul: 1.0,
    axis: '風速(m/sec)',
  },
  offset: {
    mul: 1.0 / (1000 * 1000 * 1000),
    axis: '時刻差(秒)',
  },
  stratum: {
    mul: 1,
    axis: '階層',
  },
  load1m: {
    mul: 1.0,
    axis: '１分間負荷',
  },
  load5m: {
    mul: 1.0,
    axis: '５分間負荷',
  },
  load15m: {
    mul: 1.0,
    axis: '１５分間負荷',
  },
  up: {
    mul: 1.0,
    axis: '稼働数',
  },
  total: {
    mul: 1.0,
    axis: '総数',
  },
  rate: {
    mul: 1.0,
    axis: '稼働率',
  },
  capacity: {
    mul: 0.000000001,
    axis: '総容量(GB)',
  },
  freeSpace: {
    mul: 0.000000001,
    axis: '空き容量(GB)',
  },
  usage: {
    mul: 1.0,
    axis: '使用率(%)',
  },
  totalCPU: {
    mul: 0.001,
    axis: '総CPUクロック(GHz)',
  },
  usedCPU: {
    mul: 0.001,
    axis: '使用中のCPUクロック(GHz)',
  },
  usageCPU: {
    mul: 1.0,
    axis: 'CPU使用率(%)',
  },
  totalMEM: {
    mul: 0.000001,
    axis: '総メモリー(MB)',
  },
  usedMEM: {
    mul: 0.000001,
    axis: '使用中メモリー(MB)',
  },
  usageMEM: {
    mul: 1.0,
    axis: 'メモリー使用率(%)',
  },
  totalHost: {
    mul: 1.0,
    axis: 'ホスト数',
  },
  fail: {
    mul: 1.0,
    axis: '失敗回数',
  },
  count: {
    mul: 1.0,
    axis: '回数',
  },
}

const getDispParams = (p, ent) => {
  const r = chartDispInfo[ent]
  if (r) {
    return r
  }
  return {
    mul: 1.0,
    axis: '',
  }
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
    toolbox: {
      feature: {
        saveAsImage: {},
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
          return echarts.time.format(date, '{MM}/{dd} {HH}:{mm}')
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
  const dp = getDispParams(polling, ent)
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
    dataZoom: [
      {
        type: 'inside',
      },
      {},
    ],
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
  const dp = getDispParams(polling, ent)
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

export default (context, inject) => {
  inject('makePollingChart', makePollingChart)
  inject('showPollingChart', showPollingChart)
  inject('makePollingHistogram', makePollingHistogram)
  inject('showPollingHistogram', showPollingHistogram)
  inject('setDataList', setDataList)
  inject('showPollingLogSTL', showPollingLogSTL)
  inject('showPollingLogFFT', showPollingLogFFT)
}
