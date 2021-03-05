import * as echarts from 'echarts'

let chart

const showServicePieChart = (div, data) => {
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
      trigger: 'item',
      formatter: '{a} <br/>{b} : {c} ({d}%)',
    },
    legend: {
      orient: 'vertical',
      left: 'right',
      data: [],
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
    },
    series: [
      {
        name: 'サービス',
        type: 'pie',
        radius: '75%',
        center: ['45%', '50%'],
        label: {
          fontSize: 10,
          color: '#ccc',
        },
        data: [],
      },
    ],
  }
  data.forEach((e) => {
    if (option.legend.data.length > 10) {
      return
    }
    option.legend.data.push(e.title)
    option.series[0].data.push({
      name: e.title,
      value: e.value,
    })
  })
  chart.setOption(option)
  chart.resize()
}

export default (context, inject) => {
  inject('showServicePieChart', showServicePieChart)
}
