const vmapUsage = [
  {
    gt: 0,
    lte: 90,
    color: '#0062f7',
  },
  {
    gt: 90,
    lte: 100,
    color: '#FBFB0F',
  },
  {
    gt: 100,
    color: '#FF2000',
  },
]

const chartParams = {
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
    vmap: vmapUsage,
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

export const getChartParams = (ent) => {
  const r = chartParams[ent]
  if (r) {
    return r
  }
  return {
    mul: 1.0,
    axis: '',
  }
}
