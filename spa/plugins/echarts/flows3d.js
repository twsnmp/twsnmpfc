import * as echarts from 'echarts'
import 'echarts-gl'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

let chart
const showFlows3DChart = (div, flows) => {
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(document.getElementById(div))
  const option = {
    backgroundColor: '#000',
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上'],
    },
    globe: {
      baseTexture: '/images/world.topo.bathy.200401.jpg',
      heightTexture: '/images/bathymetry_bw_composite_4k.jpg',
      shading: 'lambert',
      light: {
        ambient: {
          intensity: 0.4,
        },
        main: {
          intensity: 0.4,
        },
      },
      viewControl: {
        autoRotate: false,
      },
    },
    series: [
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '32以下',
        lineStyle: {
          width: 2,
          color: '#e31a1c',
          opacity: 0.8,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '33-41',
        lineStyle: {
          width: 2,
          color: '#fb9a99',
          opacity: 0.5,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '42-50',
        lineStyle: {
          width: 1,
          color: '#dfdf22',
          opacity: 0.3,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '51-66',
        lineStyle: {
          width: 1,
          color: '#a6cee3',
          opacity: 0.1,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '67以上',
        lineStyle: {
          width: 1,
          color: '#1f78b4',
          opacity: 0.1,
        },
        data: [],
      },
    ],
  }
  if (!flows) {
    return
  }
  let count = 0
  flows.forEach((f) => {
    if (count > 20000) {
      return
    }
    if (!f.ServerLatLong && !f.ClientLatLong) {
      return
    }
    const s = getLatLong(f.ServerLatLong)
    const c = getLatLong(f.ClientLatLong)
    const si = getScoreIndex(f.Score) - 1
    if (si > 2 && count > 1000) {
      return
    }
    count++
    option.series[si].data.push([c, s])
  })
  chart.setOption(option)
  chart.resize()
}

const getLatLong = (loc) => {
  if (!loc) {
    return [139.548088, 35.856222]
  }
  const a = loc.split(',')
  if (a.length !== 2) {
    return [139.548088, 35.856222]
  }
  return [a[1], a[0]]
}

export default (context, inject) => {
  inject('showFlows3DChart', showFlows3DChart)
}
