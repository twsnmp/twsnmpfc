import * as echarts from 'echarts'
import 'echarts-gl'
import { getScoreIndex } from '~/plugins/echarts/utils.js'

const showFlows3DChart = (div, flows) => {
  const chart = echarts.init(document.getElementById(div))
  const option = {
    backgroundColor: '#000',
    legend: {
      top: 15,
      textStyle: {
        fontSize: 10,
        color: '#ccc',
      },
      data: ['32以下', '33-41', '42-50', '51-66', '67以上', '調査中'],
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
          width: 1,
          color: '#e31a1c',
          opacity: 0.1,
        },
        data: [],
      },
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '33-41',
        lineStyle: {
          width: 1,
          color: '#fb9a99',
          opacity: 0.1,
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
          opacity: 0.1,
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
      {
        type: 'lines3D',
        coordinateSystem: 'globe',
        blendMode: 'lighter',
        name: '調査中',
        lineStyle: {
          width: 1,
          color: '#999',
          opacity: 0.1,
        },
        data: [],
      },
    ],
  }
  if (!flows) {
    return
  }

  flows.forEach((f) => {
    if (option.series[0].data.length > 5000) {
      return
    }
    if (!f.ServerLatLong && !f.ClientLatLong) {
      return
    }
    const s = getLatLong(f.ServerLatLong)
    const c = getLatLong(f.ClientLatLong)
    const si = getScoreIndex(f.Score) - 1
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
