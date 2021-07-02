import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'
import { doFFT } from '~/plugins/echarts/fft.js'

let chart

const makeSyslogHistogram = (div) => {
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

const showSyslogHistogram = (div, logs, type) => {
  makeSyslogHistogram(div)
  if (type === '') {
    type = 'Severity'
  }
  const data = []
  logs.forEach((l) => {
    if (type === 'Severity') {
      data.push(l.Severity)
    } else if (type === 'Facility') {
      data.push(l.Facility)
    } else if (type === 'Priority') {
      data.push(l.Facility * 8 + l.Severity)
    }
  })
  const bins = ecStat.histogram(data)
  chart.setOption({
    xAxis: {
      name: type,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize()
}
const makeSyslogCluster = (div) => {
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
    legend: {
      data: [],
    },
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
      },
    },
    grid: {
      left: '10%',
      right: '10%',
      top: 30,
      buttom: 0,
    },
    xAxis: {
      type: 'value',
    },
    yAxis: {
      type: 'value',
    },
    series: [],
  }
  chart.setOption(option)
  chart.resize()
}

const showSyslogCluster = (div, logs, cluster) => {
  makeSyslogCluster(div)
  const data = []
  logs.forEach((l) => {
    data.push([l.Severity, l.Facility])
  })
  const result = ecStat.clustering.hierarchicalKMeans(data, {
    clusterCount: cluster,
    stepByStep: false,
    outputType: 'multiple',
    outputClusterIndexDimension: cluster,
  })
  const centroids = result.centroids
  const ptsInCluster = result.pointsInCluster
  const series = []
  for (let i = 0; i < centroids.length; i++) {
    series.push({
      name: 'cluster' + (i + 1),
      type: 'scatter',
      data: ptsInCluster[i],
      markPoint: {
        symbolSize: 30,
        label: {
          show: true,
          position: 'top',
          formatter: (params) => {
            return (
              Math.round(params.data.coord[0] * 100) / 100 +
              ' / ' +
              Math.round(params.data.coord[1] * 100) / 100
            )
          },
          color: '#ccc',
          fontSize: 10,
        },
        data: [
          {
            coord: centroids[i],
          },
        ],
      },
    })
  }
  chart.setOption({
    legend: {
      data: [],
    },
    series,
  })
  chart.resize()
}

const showSyslogHost = (div, list) => {
  const high = []
  const low = []
  const warn = []
  const info = []
  const debug = []
  const category = []
  list.sort((a, b) => b.Total - a.Total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    high.push(list[i].High)
    low.push(list[i].Low)
    warn.push(list[i].Warn)
    info.push(list[i].Info)
    debug.push(list[i].Debug)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption({
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
    color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['重度', '軽度', '注意', '情報', 'デバッグ'],
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '20%',
      right: '10%',
      top: '10%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: '件数',
    },
    yAxis: {
      type: 'category',
      data: category,
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
        stack: '件数',
        data: high,
      },
      {
        name: '軽度',
        type: 'bar',
        stack: '件数',
        data: low,
      },
      {
        name: '注意',
        type: 'bar',
        stack: '件数',
        data: warn,
      },
      {
        name: '情報',
        type: 'bar',
        stack: '件数',
        data: info,
      },
      {
        name: 'デバック',
        type: 'bar',
        stack: '件数',
        data: debug,
      },
    ],
  })
  chart.resize()
}

const getSyslogHostList = (logs) => {
  const m = new Map()
  logs.forEach((l) => {
    const e = m.get(l.Host)
    if (!e) {
      m.set(l.Host, {
        Name: l.Host,
        Total: 1,
        High: l.Level === 'high' ? 1 : 0,
        Low: l.Level === 'low' ? 1 : 0,
        Warn: l.Level === 'warn' ? 1 : 0,
        Info: l.Level === 'info' ? 1 : 0,
        Debug: l.Level === 'debug' ? 1 : 0,
      })
    } else {
      e.Total += 1
      e.High += l.Level === 'high' ? 1 : 0
      e.Low += l.Level === 'low' ? 1 : 0
      e.Warn += l.Level === 'warn' ? 1 : 0
      e.Info += l.Level === 'info' ? 1 : 0
      e.Debug += l.Level === 'debug' ? 1 : 0
    }
  })
  const r = Array.from(m.values())
  return r
}

const showSyslogHost3D = (div, logs) => {
  const m = new Map()
  logs.forEach((l) => {
    const level = getSyslogCategory(l.Level)
    const t = new Date(l.Time / (1000 * 1000))
    const e = m.get(l.Host)
    if (!e) {
      m.set(l.Host, {
        Name: l.Host,
        Total: 1,
        Time: [t],
        Priority: [l.Severity + l.Facility * 8],
        Level: [level],
      })
    } else {
      e.Total += 1
      e.Time.push(t)
      e.Priority.push(l.Severity + l.Facility * 8)
      e.Level.push(level)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  l.forEach((e) => {
    for (let i = 0; i < e.Time.length && i < 15000; i++) {
      data.push([e.Name, e.Time[i], e.Priority[i], e.Level[i]])
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
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
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: false,
      min: 0,
      max: 4,
      dimension: 3,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
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
    },
    zAxis3D: {
      type: 'value',
      name: 'Priority',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
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
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'ホスト別ログ件数',
        type: 'scatter3D',
        dimensions: ['Host', 'Time', 'Priority', 'Level'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const getSyslogCategory = (l) => {
  switch (l) {
    case 'high':
      return 0
    case 'low':
      return 1
    case 'warn':
      return 2
    case 'debug':
      return 4
  }
  return 3
}

const showSyslogExtractHistogram = (div, extractDatas, type) => {
  makeSyslogHistogram(div)
  if (type === '') {
    return
  }
  const data = []
  extractDatas.forEach((e) => {
    if (e[type]) {
      data.push(e[type])
    }
  })
  const bins = ecStat.histogram(data)
  chart.setOption({
    xAxis: {
      name: type,
    },
    series: [
      {
        data: bins.data,
      },
    ],
  })
  chart.resize()
}

const showSyslogExtractCluster = (div, extractDatas, numList, cluster) => {
  const nl = []
  numList.forEach((n) => {
    nl.push(n.value)
  })
  makeSyslogCluster(div)
  const data = []
  extractDatas.forEach((e) => {
    const ent = []
    const keys = Object.keys(e)
    for (let i = 0; i < keys.length; i++) {
      if (nl.includes(keys[i])) {
        ent.push(e[keys[i]] * 1)
      }
    }
    data.push(ent)
  })
  const result = ecStat.clustering.hierarchicalKMeans(data, {
    clusterCount: cluster,
    stepByStep: false,
    outputType: 'multiple',
    outputClusterIndexDimension: cluster,
  })
  const centroids = result.centroids
  const ptsInCluster = result.pointsInCluster
  const series = []
  for (let i = 0; i < centroids.length; i++) {
    series.push({
      name: 'cluster' + (i + 1),
      type: 'scatter',
      data: ptsInCluster[i],
      markPoint: {
        symbolSize: 30,
        label: {
          show: true,
          position: 'top',
          formatter: (params) => {
            return (
              Math.round(params.data.coord[0] * 100) / 100 +
              ' / ' +
              Math.round(params.data.coord[1] * 100) / 100
            )
          },
          color: '#ccc',
          fontSize: 10,
        },
        data: [
          {
            coord: centroids[i],
          },
        ],
      },
    })
  }
  chart.setOption({
    legend: {
      data: [],
    },
    series,
  })
  chart.resize()
}

const showSyslogExtractTopList = (div, list) => {
  const total = []
  const category = []
  list.sort((a, b) => b.Total - a.Total)
  for (let i = list.length > 50 ? 49 : list.length - 1; i >= 0; i--) {
    total.push(list[i].Total)
    category.push(list[i].Name)
  }
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  chart.setOption({
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
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '20%',
      right: '10%',
      top: '10%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'value',
      name: '件数',
    },
    yAxis: {
      type: 'category',
      data: category,
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
        name: '件数',
        type: 'bar',
        data: total,
      },
    ],
  })
  chart.resize()
}

const getSyslogExtractTopList = (extractDatas, type) => {
  const m = new Map()
  extractDatas.forEach((ed) => {
    const k = ed[type]
    if (!k) {
      return
    }
    const e = m.get(k)
    if (!e) {
      m.set(k, {
        Name: k,
        Total: 1,
      })
    } else {
      e.Total += 1
    }
  })
  const r = Array.from(m.values())
  return r
}

const showSyslogExtract3D = (div, extractDatas, xType, zType, colorType) => {
  const m = new Map()
  const colors = []
  extractDatas.forEach((ed) => {
    const t = new Date(ed.TimeStr)
    const x = ed[xType]
    const z = ed[zType] * 1
    const c = ed[colorType] * 1
    colors.push(c)
    const e = m.get(x)
    if (!e) {
      m.set(x, {
        Name: x,
        Total: 1,
        Time: [t],
        ZValue: [z],
        Color: [c],
      })
    } else {
      e.Total += 1
      e.Time.push(t)
      e.ZValue.push(z)
      e.Color.push(c)
    }
  })
  const cat = Array.from(m.keys())
  const l = Array.from(m.values())
  const data = []
  l.forEach((e) => {
    for (let i = 0; i < e.Time.length && i < 15000; i++) {
      data.push([e.Name, e.Time[i], e.ZValue[i], e.Color[i]])
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
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
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: true,
      min: ecStat.statistics.min(colors),
      max: ecStat.statistics.max(colors),
      dimension: 3,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: {
      type: 'category',
      name: xType,
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: 'time',
      name: 'Time',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
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
    },
    zAxis3D: {
      type: 'value',
      name: zType,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
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
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: '抽出情報の項目別集計',
        type: 'scatter3D',
        dimensions: [xType, 'Time', zType, colorType],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const getSyslogFFTMap = (logs) => {
  const m = new Map()
  m.set('Total', { Name: 'Total', Count: 0, Data: [] })
  logs.forEach((l) => {
    const e = m.get(l.Host)
    if (!e) {
      m.set(l.Host, { Name: l.Host, Count: 0, Data: [] })
    }
  })
  let cts
  logs.forEach((l) => {
    if (!cts) {
      cts = Math.floor(l.Time / (1000 * 1000 * 1000))
      m.get('Total').Count++
      m.get(l.Host).Count++
      return
    }
    const newCts = Math.floor(l.Time / (1000 * 1000 * 1000))
    if (cts !== newCts) {
      m.forEach((e) => {
        e.Data.push(e.Count)
        e.Count = 0
      })
      cts++
      for (; cts < newCts; cts++) {
        m.forEach((e) => {
          e.Data.push(0)
        })
      }
    }
    m.get('Total').Count++
    m.get(l.Host).Count++
  })
  m.forEach((e) => {
    e.FFT = doFFT(e.Data, 1)
  })
  return m
}

const showSyslogFFT = (div, fftMap, host, type) => {
  if (chart) {
    chart.dispose()
  }
  if (!fftMap || !fftMap.get(host)) {
    return
  }
  const fftData = fftMap.get(host).FFT
  const freq = type === 'hz'
  const fft = []
  if (freq) {
    fftData.forEach((e) => {
      fft.push([e.frequency, e.magnitude])
    })
  } else {
    fftData.forEach((e) => {
      fft.push([e.period, e.magnitude])
    })
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
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
      name: freq ? '周波数(Hz)' : '周期(Sec)',
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
      name: '回数',
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
        name: '回数',
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
  chart.setOption(options)
  chart.resize()
}

const showSyslogFFT3D = (div, fftMap, fftType) => {
  const data = []
  const freq = fftType === 'hz'
  const colors = []
  const cat = []
  fftMap.forEach((e) => {
    if (e.Name === 'Total') {
      return
    }
    cat.push(e.Name)
    if (freq) {
      e.FFT.forEach((f) => {
        if (f.frequency === 0.0) {
          return
        }
        data.push([e.Name, f.frequency, f.magnitude, f.period])
        colors.push(f.magnitude)
      })
    } else {
      e.FFT.forEach((f) => {
        if (f.period === 0.0) {
          return
        }
        data.push([e.Name, f.period, f.magnitude, f.frequency])
        colors.push(f.magnitude)
      })
    }
  })
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const options = {
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
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    animationDurationUpdate: 1500,
    animationEasingUpdate: 'quinticInOut',
    visualMap: {
      show: true,
      min: ecStat.statistics.min(colors),
      max: ecStat.statistics.max(colors),
      dimension: 2,
      inRange: {
        color: [
          '#313695',
          '#4575b4',
          '#74add1',
          '#abd9e9',
          '#e0f3f8',
          '#ffffbf',
          '#fee090',
          '#fdae61',
          '#f46d43',
          '#d73027',
          '#a50026',
        ],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host',
      data: cat,
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    yAxis3D: {
      type: freq ? 'value' : 'log',
      name: freq ? '周波数(Hz)' : '周期(Sec)',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
        margin: 2,
      },
      axisLabel: {
        color: '#eee',
        fontSize: 8,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
    },
    zAxis3D: {
      type: 'value',
      name: '回数',
      nameTextStyle: {
        color: '#eee',
        fontSize: 12,
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
    grid3D: {
      axisLine: {
        lineStyle: { color: '#eee' },
      },
      axisPointer: {
        lineStyle: { color: '#eee' },
      },
      viewControl: {
        projection: 'orthographic',
      },
    },
    series: [
      {
        name: 'Syslog FFT分析',
        type: 'scatter3D',
        dimensions: [
          'Host',
          freq ? '周波数' : '周期',
          '回数',
          freq ? '周期' : '周波数',
        ],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

export default (context, inject) => {
  inject('showSyslogHistogram', showSyslogHistogram)
  inject('showSyslogCluster', showSyslogCluster)
  inject('showSyslogHost', showSyslogHost)
  inject('getSyslogHostList', getSyslogHostList)
  inject('showSyslogHost3D', showSyslogHost3D)
  inject('showSyslogExtractHistogram', showSyslogExtractHistogram)
  inject('showSyslogExtractCluster', showSyslogExtractCluster)
  inject('showSyslogExtractTopList', showSyslogExtractTopList)
  inject('getSyslogExtractTopList', getSyslogExtractTopList)
  inject('showSyslogExtract3D', showSyslogExtract3D)
  inject('showSyslogFFT', showSyslogFFT)
  inject('getSyslogFFTMap', getSyslogFFTMap)
  inject('showSyslogFFT3D', showSyslogFFT3D)
}
