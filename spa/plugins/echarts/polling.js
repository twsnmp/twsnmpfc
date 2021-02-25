import * as echarts from 'echarts'

let chart

const makePollingChart = (div) => {
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
          return echarts.format.formatTime('MM/dd hh:mm', date)
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
        hoverAnimation: false,
        data: [],
      },
    ],
  }
  chart.setOption(option)
  chart.resize()
}

const showPollingChart = (polling, logs, ent) => {
  const data = []
  const dp = getDispParams(polling, ent)
  logs.forEach((l) => {
    const t = new Date(l.Time / (1000 * 1000))
    const ts = echarts.format.formatTime('yyyy/MM/dd hh:mm:ss', t)
    let numVal = l.NumVal
    if (ent !== '') {
      numVal = getNumVal(ent, l.StrVal)
    }
    numVal *= dp.mul
    data.push({
      name: ts,
      value: [new Date(l.Time / (1000 * 1000)), numVal],
    })
  })
  chart.setOption({
    yAxis: {
      name: dp.axis,
    },
    series: [
      {
        data,
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

const setDataList = (s, numValEntList) => {
  try {
    JSON.parse(s, (k, v) => {
      if (
        !isNaN(parseFloat(v)) &&
        numValEntList.findIndex((e) => e.value === k) === -1
      ) {
        numValEntList.push({
          text: getChartModeName(k),
          value: k,
        })
      }
    })
  } catch (e) {}
}

const getNumVal = (ent, s) => {
  let ret = 0.0
  try {
    JSON.parse(s, (k, v) => {
      if (k === ent) {
        const nv = parseFloat(v)
        if (!isNaN(nv)) {
          ret = nv
        }
      }
    })
  } catch (e) {}
  return ret
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
}

const snmpChartDispInfo = (polling) => {
  const a = polling.Polling.split('|')
  if (a.length < 4) {
    return {
      mul: 1.0,
      axis: '',
    }
  }
  const p = a[a.length - 1].split(',')
  if (p.length !== 2) {
    return {
      mul: 1.0,
      axis: '',
    }
  }
  return {
    mul: 1.0 * p[0],
    axis: p[1],
  }
}

const getDispParams = (p, ent) => {
  if (ent === '') {
    // 数値の場合はポーリングの種類から選ぶ
    switch (p.Type) {
      case 'ping':
        if (p.Polling === 'line') {
          ent = 'speed'
          break
        }
        ent = 'rtt'
        break
      case 'tcp':
      case 'http':
      case 'https':
      case 'dns':
      case 'ntp':
        ent = 'rtt'
        break
      case 'sysloguser':
        ent = 'successRate'
        break
      case 'snmp':
        return snmpChartDispInfo(p)
      default:
        ent = 'none'
        break
    }
  }
  const r = chartDispInfo[ent]
  if (r) {
    return r
  }
  return {
    mul: 1.0,
    axis: '',
  }
}

export default (context, inject) => {
  inject('makePollingChart', makePollingChart)
  inject('showPollingChart', showPollingChart)
  inject('setDataList', setDataList)
}