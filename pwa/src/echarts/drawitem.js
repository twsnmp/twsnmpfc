import * as echarts from 'echarts'

export const gauge = (title, val, backgroundColor) => {
  const chart = echarts.init(null, null, {
    renderer: 'svg',
    ssr: true,
    width: 1000,
    height: 1000,
  })
  const option = {
    series: [
      {
        type: 'gauge',
        axisLine: {
          lineStyle: {
            width: 30,
            color: [
              [0.8, '#1f78b4'],
              [0.9, '#dfdf22'],
              [1, '#e31a1c'],
            ],
          },
        },
        pointer: {
          itemStyle: {
            color: 'auto',
          },
        },
        axisTick: {
          distance: -30,
          length: 8,
          lineStyle: {
            color: '#fff',
            width: 2,
          },
        },
        splitLine: {
          distance: -30,
          length: 30,
          lineStyle: {
            color: '#fff',
            width: 4,
          },
        },
        axisLabel: {
          color: 'inherit',
          distance: 40,
          fontSize: 20,
        },
        detail: {
          valueAnimation: true,
          formatter: '{value}%',
          color: 'inherit',
        },
        data: [
          {
            value: val,
            name: title,
            title: {
              color: '#fff',
              fontSize: 40,
              offsetCenter: [0, '60%'],
            },
          },
        ],
      },
    ],
  }
  chart.setOption(option)
  return chart.getDataURL({ backgroundColor })
}

export const line = (title, color, values, backgroundColor) => {
  const chart = echarts.init(null, null, {
    renderer: 'svg',
    ssr: true,
    width: 400,
    height: 100,
  })
  const option = {
    title: {
      show: true,
      top: 'center',
      left: 'center',
      textAlign: 'center',
      text: title,
      textStyle: {
        fontSize: 14,
        fontWeight: 'normal',
        color: '#fff',
      },
    },
    grid: {
      top: 0,
      left: 10,
      bottom: 0,
      right: 0,
    },
    xAxis: {
      show: false,
    },
    yAxis: {
      show: false,
    },
    series: [
      {
        type: 'bar',
        color,
        data: [],
        large: true,
      },
    ],
  }
  for (let i = 0; i < values.length; i++) {
    option.series[0].data.push([i, values[i] - values[0]])
  }
  chart.setOption(option)
  return chart.getDataURL({ backgroundColor })
}

export const bar = (title, color, value, backgroundColor) => {
  const chart = echarts.init(null, null, {
    renderer: 'svg',
    ssr: true,
    width: 400,
    height: 100,
  })
  const option = {
    title: {
      show: true,
      top: 'center',
      left: 'center',
      textAlign: 'center',
      text: title + ':' + value + '%',
      textStyle: {
        fontSize: 14,
        fontWeight: 'normal',
        color: '#fff',
      },
    },
    grid: {
      top: 0,
      left: 0,
      bottom: 0,
      right: 0,
    },
    yAxis: {
      type: 'category',
      data: ['0'],
      show: false,
    },
    xAxis: {
      type: 'value',
      show: false,
      max: 100,
      min: 0,
    },
    series: [
      {
        data: [value],
        type: 'bar',
        color,
        smooth: true,
      },
    ],
  }
  chart.setOption(option)
  return chart.getDataURL({ backgroundColor })
}
