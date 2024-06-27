import * as echarts from 'echarts'

let chart

/*
InOctets
OutOctets
*/

const ifCounterBPSChart = (div, logs) => {
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
        saveAsImage: { name: 'twsnmp_sFlow_Counter_BPS' },
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
    yAxis: {
      type: 'value',
      name: 'バイト/秒',
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
        name: '受信',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '送信',
        type: 'bar',
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
      data: ['受信', '送信'],
    },
  }
  logs.sort((a, b) => {
    return a.Time - b.Time
  })
  for (let i = 1; i < logs.length; i++) {
    const dt = (logs[i].Time - logs[i - 1].Time) / (1000 * 1000 * 1000.0)
    if (dt === 0) {
      continue
    }
    const t = new Date(logs[i].Time / (1000 * 1000))
    const obps = (logs[i].OutOctets - logs[i - 1].OutOctets) / dt
    const ibps = (logs[i].InOctets - logs[i - 1].InOctets) / dt
    option.series[0].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, ibps],
    })
    option.series[1].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, obps],
    })
  }
  chart.setOption(option)
  chart.resize()
}

const ifCounterPPSChart = (div, logs) => {
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
        saveAsImage: { name: 'twsnmp_sFlow_Counter_PPS' },
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
    yAxis: {
      type: 'value',
      name: 'パケット/秒',
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
        name: '受信エラー',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '受信ユニキャスト',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '受信ノンユニキャスト',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '送信エラー',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '送信ユニキャスト',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '送信ノンユニキャスト',
        type: 'bar',
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
      data: [
        '受信エラー',
        '受信ユニキャスト',
        '受信ノンユニキャスト',
        '送信エラー',
        '送信ユニキャスト',
        '送信ノンユニキャスト',
      ],
    },
  }
  logs.sort((a, b) => {
    return a.Time - b.Time
  })
  for (let i = 1; i < logs.length; i++) {
    const dt = (logs[i].Time - logs[i - 1].Time) / (1000 * 1000 * 1000.0)
    if (dt === 0) {
      continue
    }
    const t = new Date(logs[i].Time / (1000 * 1000))
    const ouc = (logs[i].OutUnicastPackets - logs[i - 1].OutUnicastPackets) / dt
    const onuc =
      (logs[i].OutBroadcastPackets -
        logs[i - 1].OutBroadcastPackets +
        logs[i].OutMulticastPackets -
        logs[i - 1].OutMulticastPackets) /
      dt
    const oerr =
      (logs[i].OutDiscards -
        logs[i - 1].OutDiscards +
        logs[i].OutErrors -
        logs[i - 1].OutErrors) /
      dt
    const iuc = (logs[i].InUnicastPackets - logs[i - 1].InUnicastPackets) / dt
    const inuc =
      (logs[i].InBroadcastPackets -
        logs[i - 1].InBroadcastPackets +
        logs[i].InMulticastPackets -
        logs[i - 1].InMulticastPackets) /
      dt
    const ierr =
      (logs[i].InDiscards -
        logs[i - 1].InDiscards +
        logs[i].InErrors -
        logs[i - 1].InErrors +
        logs[i].InUnknownProtocols -
        logs[i - 1].InUnknownProtocols) /
      dt
    option.series[0].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, ierr],
    })
    option.series[1].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, iuc],
    })
    option.series[2].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, inuc],
    })
    option.series[3].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, oerr],
    })
    option.series[4].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, ouc],
    })
    option.series[5].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, onuc],
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showSFlowIFCounter = (div, logs, type) => {
  switch (type) {
    case 'bps':
      ifCounterBPSChart(div, logs)
      break
    case 'pps':
      ifCounterPPSChart(div, logs)
      break
  }
}

const showSFlowCpuCounter = (div, logs) => {
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
        saveAsImage: { name: 'twsnmp_sFlow_Counter_CPU' },
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
        name: 'CPU使用率',
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
      {
        type: 'value',
        name: '負荷',
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
        name: 'CPU Sys',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'CPU User',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'CPU I/O',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'CPU Other',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'Load 1M',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: 'Load 5M',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: 'Load 15M',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [
        'CPU Sys',
        'CPU User',
        'CPU I/O',
        'CPU Other',
        'Load 1M',
        'Load 5M',
        'Load 15M',
      ],
    },
  }
  logs.sort((a, b) => {
    return a.Time - b.Time
  })
  for (let i = 1; i < logs.length; i++) {
    const dt = (logs[i].Time - logs[i - 1].Time) / (1000 * 1000 * 1000.0)
    if (dt === 0) {
      continue
    }
    const t = new Date(logs[i].Time / (1000 * 1000))
    const sys = logs[i].CPUSys - logs[i - 1].CPUSys
    const user = logs[i].CPUUser - logs[i - 1].CPUUser
    const io = logs[i].CPUWio - logs[i - 1].CPUWio
    const idle = logs[i].CPUIdle - logs[i - 1].CPUIdle
    let other = logs[i].CPUIntr - logs[i - 1].CPUIntr
    other += logs[i].CPUNice - logs[i - 1].CPUNice
    other += logs[i].CPUSoftIntr - logs[i - 1].CPUSoftIntr
    other += logs[i].CPUSteal - logs[i - 1].CPUSteal
    const total = sys + user + io + idle + other
    if (!total) {
      continue
    }
    option.series[0].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, (100 * sys) / total],
    })
    option.series[1].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, (100 * user) / total],
    })
    option.series[2].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, (100 * io) / total],
    })
    option.series[3].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, (100 * other) / total],
    })
    option.series[4].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, logs[i].Load1m],
    })
    option.series[5].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, logs[i].Load5m],
    })
    option.series[6].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, logs[i].Load15m],
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showSFlowMemCounter = (div, logs) => {
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
        saveAsImage: { name: 'twsnmp_sFlow_Counter_Mem' },
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
        name: '利用可能容量',
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
      {
        type: 'value',
        name: '使用率',
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
        name: 'Free',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'Buffers',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'Cached',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: 'Real',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: 'Swap',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['Free', 'Buffers', 'Cached', 'Real', 'Swap'],
    },
  }
  logs.sort((a, b) => {
    return a.Time - b.Time
  })
  /*
  {"Total":498462720,"Free":10137600,"Shared":0,"Buffers":31383552,"Cached":204365824,"SwapTotal":0,"SwapFree":0,"PageIn":705849,"PageOut":2831993,"SwapIn":0,"SwapOut":0}
  */
  for (let i = 0; i < logs.length; i++) {
    const t = new Date(logs[i].Time / (1000 * 1000))
    const mem = logs[i].Total
      ? (logs[i].Total - logs[i].Free - logs[i].Buffers - logs[i].Cached) /
        logs[i].Total
      : 0.0
    const swap = logs[i].SwapTotal
      ? (logs[i].SwapTotal - logs[i].SwapFree) / logs[i].SwapTotal
      : 0.0

    option.series[0].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, logs[i].Free],
    })
    option.series[1].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, logs[i].Buffers],
    })
    option.series[2].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, logs[i].Cached],
    })
    option.series[3].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, 100 * mem],
    })
    option.series[4].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, 100 * swap],
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showSFlowDiskCounter = (div, logs) => {
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
        saveAsImage: { name: 'twsnmp_sFlow_Counter_Disk' },
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
        name: '使用率',
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
      {
        type: 'value',
        name: 'Bytes/秒',
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
        name: '使用率',
        type: 'bar',
        large: true,
        data: [],
      },
      {
        name: 'Read',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: 'Write',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['使用率', 'Read', 'Write'],
    },
  }
  logs.sort((a, b) => {
    return a.Time - b.Time
  })
  /*
  {"Total":10196652032,"Free":9010069504,"MaxUsedPercent":1.63e-42,"Reads":41267,"BytesRead":1094646784,"ReadTime":15219,"Writes":1251316,"BytesWritten":5808449536,"WriteTime":2297004}
  */
  for (let i = 1; i < logs.length; i++) {
    const dt = (logs[i].Time - logs[i - 1].Time) / (1000 * 1000 * 1000.0)
    if (dt === 0) {
      continue
    }
    const t = new Date(logs[i].Time / (1000 * 1000))
    const usage = logs[i].Total
      ? (100 * (logs[i].Total - logs[i].Free)) / logs[i].Total
      : 0.0
    const read = (logs[i].BytesRead - logs[i - 1].BytesRead) / dt
    const write = (logs[i].BytesWritten - logs[i - 1].BytesWritten) / dt
    option.series[0].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, usage],
    })
    option.series[1].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, read],
    })
    option.series[2].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, write],
    })
  }
  chart.setOption(option)
  chart.resize()
}

const showSFlowNetCounter = (div, logs) => {
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
        saveAsImage: { name: 'twsnmp_sFlow_Counter_Mem' },
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
        name: '利用可能容量',
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
      {
        type: 'value',
        name: '使用率',
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
        name: '受信BPS',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '送信BPS',
        type: 'bar',
        stack: 'count',
        large: true,
        data: [],
      },
      {
        name: '受信PPS',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: '受信エラー',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: '送信PPS',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
      {
        name: '送信エラー',
        type: 'line',
        symbol: 'none',
        yAxisIndex: 1,
        large: true,
        data: [],
      },
    ],
    legend: {
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: [
        '受信BPS',
        '送信BPS',
        '受信PPS',
        '受信エラー',
        '送信PPS',
        '送信エラー',
      ],
    },
  }
  logs.sort((a, b) => {
    return a.Time - b.Time
  })
  /*
  {"BytesIn":107406561,"PacketsIn":822319,"ErrorsIn":100061,"DropsIn":0,"BytesOut":85470554,"PacketsOut":293398,"ErrorsOut":0,"DropsOut":0}
  */
  for (let i = 1; i < logs.length; i++) {
    const dt = (logs[i].Time - logs[i - 1].Time) / (1000 * 1000 * 1000.0)
    if (dt === 0) {
      continue
    }
    const t = new Date(logs[i].Time / (1000 * 1000))
    const ibps = (logs[i].BytesIn - logs[i - 1].BytesIn) / dt
    const obps = (logs[i].BytesIn - logs[i - 1].BytesIn) / dt
    const ipps = (logs[i].PacketsIn - logs[i - 1].PacketsIn) / dt
    const opps = (logs[i].PacketsOut - logs[i - 1].PacketsOut) / dt
    const ieps = (logs[i].ErrorsIn - logs[i - 1].ErrorsIn) / dt
    const oeps = (logs[i].ErrorsOut - logs[i - 1].ErrorsOut) / dt

    option.series[0].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, ibps],
    })
    option.series[1].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, obps],
    })
    option.series[2].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, ipps],
    })
    option.series[3].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, ieps],
    })
    option.series[4].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, opps],
    })
    option.series[5].data.push({
      name: echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
      value: [t, oeps],
    })
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showSFlowIFCounter', showSFlowIFCounter)
  inject('showSFlowCpuCounter', showSFlowCpuCounter)
  inject('showSFlowMemCounter', showSFlowMemCounter)
  inject('showSFlowDiskCounter', showSFlowDiskCounter)
  inject('showSFlowNetCounter', showSFlowNetCounter)
}
