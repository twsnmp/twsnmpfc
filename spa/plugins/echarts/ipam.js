import * as echarts from 'echarts'

let chart

const showIPAMHeatmap = (div, ranges) => {
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
    grid: {
      left: '12%',
      right: '5%',
      top: 30,
      buttom: 0,
    },
    toolbox: {
      iconStyle: {
        color: '#ccc',
      },
      feature: {
        saveAsImage: { name: 'twsnmp_' + div },
      },
    },
    tooltip: {},
    xAxis: {
      type: 'category',
      name: '%',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 12,
        margin: 3,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 8,
        margin: 3,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      data: [],
    },
    yAxis: {
      type: 'category',
      name: 'IP範囲',
      nameTextStyle: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLabel: {
        color: '#ccc',
        fontSize: 10,
        margin: 2,
      },
      axisLine: {
        lineStyle: {
          color: '#ccc',
        },
      },
      data: [],
    },
    visualMap: {
      min: 0,
      max: 0,
      textStyle: {
        color: '#ccc',
        fontSize: 8,
      },
      calculable: true,
      realtime: false,
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
    series: [
      {
        name: 'IP Usage',
        type: 'heatmap',
        data: [],
        emphasis: {
          itemStyle: {
            borderColor: '#ccc',
            borderWidth: 1,
          },
        },
        progressive: 1000,
        animation: false,
      },
    ],
  }
  for (let x = 0; x < 100; x++) {
    option.xAxis.data.push(x)
  }
  for (let y = ranges.length - 1; y >= 0; y--) {
    const r = ranges[y]
    option.yAxis.data.push(r.Range)
    for (let x = 0; x < 100; x++) {
      const c = r.UsedIP[x]
      option.series[0].data.push([x, ranges.length - y - 1, c])
      if (option.visualMap.max < c) {
        option.visualMap.max = c
      }
    }
  }
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showIPAMHeatmap', showIPAMHeatmap)
}
