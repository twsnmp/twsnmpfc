import * as echarts from 'echarts'

let chart

const showEnv3DChart = (div, type, list, filter) => {
  const data = []
  const cat = []
  list.forEach((i) => {
    if (!i.EnvData || i.EnvData.length < 1) {
      return
    }
    if (!filterEnvMon(i, filter)) {
      return
    }
    const id = i.Host + ':' + i.Address
    i.EnvData.forEach((e) => {
      data.push([
        id,
        e.Time / (1000 * 1000),
        e[type],
        getEnvLevel(e[type], type),
        i.Name,
      ])
    })
    cat.push(id)
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
      dimension: 3,
      min: 0,
      max: 4,
      inRange: {
        color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host:Address',
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
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
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
      name: getEnvName(type),
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
        name: 'Env Data',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: ['Host:Address', 'Time', type, 'Level', 'Name'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showEnv2DChart = (div, type, list, filter) => {
  const cat = []
  const series = []
  list.forEach((i) => {
    if (!i.EnvData || i.EnvData.length < 1) {
      return
    }
    if (
      i.Name !== 'Rbt' &&
      type !== 'Temp' &&
      type !== 'Humidity' &&
      type !== 'Battery'
    ) {
      return
    }
    if (!filterEnvMon(i, filter)) {
      return
    }
    const data = []
    i.EnvData.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
      data.push({ name, value: [t, e[type]] })
    })
    const id = i.Host + ':' + i.Address
    series.push({
      name: id,
      type: 'line',
      large: true,
      symbol: 'none',
      data,
    })
    cat.push(id)
  })
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
    dataZoom: [{}, {}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: cat,
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
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
      name: getEnvUnit(type),
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
    series,
  }
  chart.setOption(option)
  chart.resize()
}

const getEnvLevel = (v, type) => {
  switch (type) {
    case 'Temp':
      if (v > 35.0) {
        return 0
      } else if (v > 30.0) {
        return 1
      } else if (v > 25.0) {
        return 2
      } else if (v < 10.0) {
        return 0
      }
      return 3
    case 'Humidity':
      if (v > 60.0) {
        return 0
      } else if (v < 40.0) {
        return 2
      }
      return 3
    case 'Illuminance':
      if (v > 300) {
        return 0
      } else if (v > 200) {
        return 1
      } else if (v > 100) {
        return 2
      } else if (v >= 1) {
        return 3
      }
      return 4
    case 'BarometricPressure':
      if (v > 1021) {
        return 0
      } else if (v > 1016) {
        return 1
      } else if (v > 1010) {
        return 3
      } else if (v > 1000) {
        return 2
      }
      return 4
    case 'Sound':
      if (v > 80) {
        return 0
      } else if (v > 70) {
        return 1
      } else if (v > 60) {
        return 2
      } else if (v > 30) {
        return 3
      }
      return 4
    case 'ETVOC':
      if (v > 3000) {
        return 0
      } else if (v > 1000) {
        return 1
      } else if (v > 300) {
        return 2
      } else if (v > 0) {
        return 3
      }
      return 4
    case 'ECo2':
      if (v > 3000) {
        return 0
      } else if (v > 2000) {
        return 1
      } else if (v > 1000) {
        return 2
      } else if (v > 0) {
        return 3
      }
      return 4
    case 'Battery':
      if (v > 50) {
        return 3
      } else if (v > 30) {
        return 2
      } else if (v > 10) {
        return 1
      }
      return 0
    default:
      return 0
  }
}

const filterEnvMon = (i, filter) => {
  if (filter.address && !i.Address.includes(filter.address)) {
    return false
  }
  if (filter.name && !i.Name.includes(filter.name)) {
    return false
  }
  if (filter.host && !i.Host.includes(filter.host)) {
    return false
  }
  return true
}

const envTypeList = [
  { value: 'Temp', text: '気温(℃)', unit: '℃' },
  { value: 'Humidity', text: '湿度(%)', unit: '%' },
  { value: 'Illuminance', text: '照度(lx)', unit: 'lx' },
  { value: 'BarometricPressure', text: '気圧(hpa)', unit: 'hpa' },
  { value: 'Sound', text: '騒音(dB)' },
  { value: 'ETVOC', text: '総揮発性有機化合物(ppb)', unit: 'ppb' },
  { value: 'ECo2', text: '二酸化炭素濃度(ppm)', unit: 'ppm' },
  { value: 'Battery', text: '電池残量(%)', unit: '%' },
]

const getEnvName = (type) => {
  for (let i = 0; i < envTypeList.length; i++) {
    if (type === envTypeList[i].value) {
      return envTypeList[i].text
    }
  }
  return 'Invalid type=' + type
}

const getEnvUnit = (type) => {
  for (let i = 0; i < envTypeList.length; i++) {
    if (type === envTypeList[i].value) {
      return envTypeList[i].unit
    }
  }
  return 'Invalid type=' + type
}

const showPower3DChart = (div, list, filter) => {
  const data = []
  const cat = []
  let min = 1500
  let max = 0
  list.forEach((i) => {
    if (!i.Data || i.Data.length < 1) {
      return
    }
    if (!filterEnvMon(i, filter)) {
      return
    }
    const id = i.Name ? i.Name : i.Host + ':' + i.Address
    i.Data.forEach((e) => {
      data.push([id, e.Time / (1000 * 1000), e.Load, i.Name])
      min = Math.min(min, e.Load)
      max = Math.max(max, e.Load)
    })
    cat.push(id)
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
      dimension: 2,
      min,
      max,
      inRange: {
        color: ['#777', '#1f78b4', '#dfdf22', '#fb9a99', '#e31a1c'],
      },
    },
    xAxis3D: {
      type: 'category',
      name: 'Host:Address',
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
          return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
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
      name: '負荷(W)',
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
        name: 'Load',
        type: 'scatter3D',
        symbolSize: 4,
        dimensions: ['Host:Address', 'Time', 'Load', 'Level', 'Name'],
        data,
      },
    ],
  }
  chart.setOption(options)
  chart.resize()
}

const showPower2DChart = (div, list, filter) => {
  const cat = []
  const series = []
  list.forEach((i) => {
    if (!i.Data || i.Data.length < 1) {
      return
    }
    if (!filterEnvMon(i, filter)) {
      return
    }
    const data = []
    i.Data.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      const name = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
      data.push({ name, value: [t, e.Load] })
    })
    const id = i.Name ? i.Name : i.Host + ':' + i.Address
    series.push({
      name: id,
      type: 'line',
      large: true,
      symbol: 'none',
      data,
    })
    cat.push(id)
  })
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
    dataZoom: [{}, {}],
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow',
      },
    },
    grid: {
      left: '10%',
      right: '5%',
      top: 60,
      buttom: 0,
    },
    legend: {
      data: cat,
      textStyle: {
        color: '#ccc',
        fontSize: 10,
      },
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
      name: 'W',
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
    series,
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showEnv3DChart', showEnv3DChart)
  inject('showEnv2DChart', showEnv2DChart)
  inject('showPower3DChart', showPower3DChart)
  inject('showPower2DChart', showPower2DChart)
  inject('getEnvName', getEnvName)
  inject('envTypeList', envTypeList)
  inject('filterEnvMon', filterEnvMon)
}
