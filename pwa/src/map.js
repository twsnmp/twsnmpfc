import P5 from 'p5'
import { getState,getIconCode } from './common'
import { gauge, bar, line } from './echarts/drawitem'

let mapSize = 0
let mapSizeX = 2500
let mapSizeY = 5000
let mapP5

let mapRedraw = true

let nodes = {}
let lines = []
let items = {}
let networks = {}
let backImage = {
  X: 0,
  Y: 0,
  Width: 0,
  Height: 0,
  Path: '',
  Color: 23,
  Image: null,
}

let fontSize = 12
let iconSize = 32
let showNodeInfo = false

const imageMap = {}

export const setMAP = (m) => {
  const ms = m.MapConf.MapSize || 0
  if (!mapP5 || ms !== mapSize) {
    mapSize = m.MapConf.MapSize
    initMap(mapSize)
  }
  nodes = m.Nodes
  lines = m.Lines || []
  items = m.Items || {}
  networks = m.Networks || {}
  backImage = m.MapConf.BackImage
  fontSize = m.MapConf.FontSize || 12
  iconSize = m.MapConf.IconSize || 24
  backImage.Image = null
  if (backImage.Path && mapP5) {
    mapP5.loadImage('APIURL' + '/backimage', (img) => {
      backImage.Image = img
      mapRedraw = true
    })
  }
  for (const k in nodes) {
    if(nodes[k].Image && !imageMap[nodes[k].Image] && mapP5) {
      mapP5.loadImage('APIURL' + '/imageIcon/' + nodes[k].Image, (img) => {
        imageMap[nodes[k].Image] = img
        mapRedraw = true
      })
    }
  }
  for (const k in items) {
    if (items[k].H < 10) {
      items[k].H = 100
    }
    if (items[k].W < 10) {
      items[k].W = 100
    }
    switch (items[k].Type) {
      case 3: // Image
        if (!imageMap[items[k].Path] && mapP5) {
          mapP5.loadImage('APIURL' + '/image/' + items[k].Path, (img) => {
            imageMap[items[k].Path] = img
            mapRedraw = true
          })
        }
        break
      case 2: // Text
      case 4: // Polling
        items[k].W = items[k].Size * items[k].Text.length
        items[k].H = items[k].Size
        break
      case 5: // Gauge
        items[k].H = items[k].Size * 10
        items[k].W = items[k].Size * 10
        break
      case 6: // New Gauge
        items[k].W = items[k].H
        mapP5.loadImage(
          gauge(
            items[k].Text || '',
            items[k].Value || 0,
            backImage.Color || 23
          ),
          (img) => {
            imageMap[k] = img
            mapRedraw = true
          }
        )
        break
      case 7: // Bar
        items[k].W = items[k].H * 4
        mapP5.loadImage(
          bar(
            items[k].Text || '',
            items[k].Color || 'white',
            items[k].Value || 0,
            backImage.Color || 23
          ),
          (img) => {
            imageMap[k] = img
            mapRedraw = true
          }
        )
        break
      case 8: // Line
        items[k].W = items[k].H * 4
        mapP5.loadImage(
          line(
            items[k].Text || '',
            items[k].Color || 'white',
            items[k].Values || [],
            backImage.Color || 23
          ),
          (img) => {
            imageMap[k] = img
            mapRedraw = true
          }
        )
        break
    }
  }
  mapRedraw = true
}

const initMap = (ms) => {
  switch (ms) {
    case 2:
    case 3:
    case 1:
      mapSizeX = 2500
      mapSizeY = 5000
      break
    case 4:
      mapSizeX = 2894
      mapSizeY = 4093
      break
    case 5:
      mapSizeX = 4093
      mapSizeY = 2894
      break
    default:
      mapSizeX = 2500
      mapSizeY = 5000
      break
  }
}

const getStateColor = (state) => {
  const s = getState(state)
  return s ? s.color : 'gray'
}

const getLineColor = (state) => {
  if (state === 'high' || state === 'low' || state === 'warn') {
    return getStateColor(state)
  }
  return 250
}

const getLinePos = (id,polling) => {
  if (id && id.startsWith('NET:')) {
    const a = id.split(":")
    if(a.length !== 2 ) {
      return undefined
    }
    const net = networks[a[1]]
    if (!net || !net.Ports) {
      return undefined
    }
    let pi =  -1
    for(let i = 0; i <  net.Ports.length;i++) {
      if (net.Ports[i].ID === polling) {
        pi = i
        break
      }
    }
    if (pi < 0 ) {
      return undefined
    }
    return {
      X: net.X + net.Ports[pi].X * 45 + 10 + 20,
      Y: net.Y + net.Ports[pi].Y * 55 + fontSize + 20 + 10
    }
  }
  if (!nodes[id]) {
    return undefined
  }
  return {
    X:nodes[id].X,
    Y:nodes[id].Y + 6,
  }
}

let scale = 1.0

const drawBackImage = (p5) => {
  p5.background(backImage.Color || 23)
  if (backImage.Image) {
    if (backImage.Width) {
      p5.image(
        backImage.Image,
        backImage.X,
        backImage.Y,
        backImage.Width,
        backImage.Height
      )
    } else {
      p5.image(backImage.Image, backImage.X, backImage.Y)
    }
  }
}

const drawNetworks = (p5, portImage) => {
  for (const k in networks) {
    p5.push()
    p5.translate(networks[k].X, networks[k].Y)
    if (networks[k].Error !== '') {
      p5.stroke('#cc3300')
    } else {
      p5.stroke('#999')
    }
    p5.fill('rgba(23,23,23,0.9)')
    p5.rect(0, 0, networks[k].W, networks[k].H)
    p5.stroke('#999')
    p5.textFont('Roboto')
    p5.textSize(fontSize)
    p5.fill('#eee')
    p5.text(networks[k].Name, 5, fontSize + 5)
    if (!networks[k].Ports || networks[k].Ports.length < 1) {
      if (networks[k].Error !== '') {
        p5.fill('#cc3300')
        p5.text(networks[k].Error, 15, fontSize * 2 + 15)
      } else {
        p5.fill('#11ee00')
        p5.text('構成を分析中...', 15, fontSize * 2 + 15)
      }
    } else {
      p5.textSize(6)
      for (const p of networks[k].Ports) {
        const x = p.X * 45 + 10
        const y = p.Y * 55 + fontSize + 15
        p5.image(portImage, x, y, 40, 40)
        p5.fill(p.State === 'up' ? '#11ee00' : ' #999')
        p5.circle(x + 4, y + 4, 8)
        p5.fill('#eee')
        p5.text(p.Name, x, y + 40 + 10)
      }
    }
    p5.pop()
  }
}

const drawLines = (p5) => {
  for (const k in lines) {
    const lp1 = getLinePos(lines[k].NodeID1, lines[k].PollingID1)
    if (!lp1) {
      continue
    }
    const lp2 = getLinePos(lines[k].NodeID2, lines[k].PollingID2)
    if (!lp2) {
      continue
    }
    const x1 = lp1.X
    const x2 = lp2.X
    const y1 = lp1.Y
    const y2 = lp2.Y
    const xm = (x1 + x2) / 2
    const ym = (y1 + y2) / 2
    p5.push()
    p5.strokeWeight(lines[k].Width || 1)
    p5.stroke(getStateColor(lines[k].State1))
    p5.line(x1, y1, xm, ym)
    p5.stroke(getStateColor(lines[k].State2))
    p5.line(xm, ym, x2, y2)
    if (lines[k].Info) {
      const color = getLineColor(lines[k].State)
      const dx = Math.abs(x1 - x2)
      const dy = Math.abs(y1 - y2)
      p5.textFont('Roboto')
      p5.textSize(fontSize)
      p5.fill(color)
      if (dx === 0 || dy / dx > 0.8) {
        p5.text(lines[k].Info, xm + 10, ym)
      } else {
        p5.text(lines[k].Info, xm - dx / 4, ym + 20)
      }
    }
    p5.pop()
  }
}

const drawItems = (p5) => {
  for (const k in items) {
    p5.push()
    p5.translate(items[k].X, items[k].Y)
    switch (items[k].Type) {
      case 0: // rect
        p5.fill(items[k].Color)
        p5.stroke('rgba(23,23,23,0.9)')
        p5.rect(0, 0, items[k].W, items[k].H)
        break
      case 9: // Group(枠)
        p5.fill('rgba(23,23,23,0.01)')
        p5.strokeWeight(2)
        p5.stroke(items[k].Color)
        p5.rect(0, 0, items[k].W, items[k].H)
        if (items[k].Text) {
          p5.textSize(items[k].Size || 12)
          p5.fill(250)
          p5.noStroke()
          p5.textAlign(p5.RIGHT, p5.BOTTOM)
          p5.text(items[k].Text, items[k].W - 5, items[k].H - 5)
        }
        break
      case 10: // Group(塗りつぶし)
        p5.fill(items[k].Color)
        p5.noStroke()
        p5.rect(0, 0, items[k].W, items[k].H)
        if (items[k].Text) {
          p5.textSize(items[k].Size || 12)
          p5.fill(250)
          p5.noStroke()
          p5.textAlign(p5.RIGHT, p5.BOTTOM)
          p5.text(items[k].Text, items[k].W - 5, items[k].H - 5)
        }
        break
      case 1: // ellipse
        p5.fill(items[k].Color)
        p5.stroke('rgba(23,23,23,0.9)')
        p5.ellipse(items[k].W / 2, items[k].H / 2, items[k].W, items[k].H)
        break
      case 2: // text
      case 4: // Polling
        p5.textSize(items[k].Size || 12)
        p5.fill(items[k].Color)
        p5.text(
          items[k].Text,
          0,
          0,
          items[k].Size * items[k].Text.length + 10,
          items[k].Size + 10
        )
        break
      case 3: // Image
        if (imageMap[items[k].Path]) {
          p5.image(imageMap[items[k].Path], 0, 0, items[k].W, items[k].H)
        }
        break
      case 5: {
        const x = items[k].W / 2
        const y = items[k].H / 2
        const r0 = items[k].W
        const r1 = items[k].W - items[k].Size
        const r2 = items[k].W - items[k].Size * 4
        p5.noStroke()
        p5.fill('#eee')
        p5.arc(x, y, r0, r0, 5 * p5.QUARTER_PI, -p5.QUARTER_PI)
        if (items[k].Value > 0) {
          p5.fill(items[k].Color)
          p5.arc(
            x,
            y,
            r0,
            r0,
            5 * p5.QUARTER_PI,
            -p5.QUARTER_PI - (p5.HALF_PI - (p5.HALF_PI * items[k].Value) / 100)
          )
        }
        p5.fill(backImage.Color || 23)
        p5.arc(x, y, r1, r1, -p5.PI, 0)
        p5.textAlign(p5.CENTER)
        p5.textSize(8)
        p5.fill('#fff')
        p5.text(items[k].Value + '%', x, y - 10)
        p5.textSize(items[k].Size)
        p5.text(items[k].Text || '', x, y + 5)
        p5.fill('#e31a1c')
        const angle = -p5.QUARTER_PI + (p5.HALF_PI * items[k].Value) / 100
        const x1 = x + (r1 / 2) * p5.sin(angle)
        const y1 = y - (r1 / 2) * p5.cos(angle)
        const x2 = x + (r2 / 2) * p5.sin(angle) + 5 * p5.cos(angle)
        const y2 = y - (r2 / 2) * p5.cos(angle) + 5 * p5.sin(angle)
        const x3 = x + (r2 / 2) * p5.sin(angle) - 5 * p5.cos(angle)
        const y3 = y - (r2 / 2) * p5.cos(angle) - 5 * p5.sin(angle)
        p5.triangle(x1, y1, x2, y2, x3, y3)
      }
      break
      case 6: // New Gauge,Line,Bar
      case 7:
      case 8:
        if (imageMap[k]) {
          p5.image(imageMap[k], 0, 0, items[k].W, items[k].H)
        }
        break
    }
    p5.pop()
  }
}

const drawNodes = (p5) => {
  for (const k in nodes) {
    const icon = getIconCode(nodes[k].Icon)
    p5.push()
    p5.translate(nodes[k].X, nodes[k].Y)
    if (nodes[k].Image && imageMap[nodes[k].Image]) {
      const img = imageMap[nodes[k].Image]
      const imgW = 48
      const imgH = img.width > 0 ? imgW * (img.height / img.width) : imgW
      p5.tint(getStateColor(nodes[k].State))
      p5.image(img, -imgW / 2, -imgH / 2, imgW, imgH)
      p5.noTint()
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.textFont('Roboto')
      p5.textSize(fontSize)
      p5.fill(250)
      p5.text(nodes[k].Name, 0, imgH / 2 + fontSize / 2)
      if (showNodeInfo) {
        p5.textSize(fontSize-2)
        p5.text(nodes[k].IP, 0, imgH / 2 + fontSize / 2 + fontSize)
      }
    } else {
      p5.textFont('Material Design Icons')
      p5.textSize(iconSize)
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, 0, 0)
      p5.textFont('Roboto')
      p5.textSize(fontSize)
      p5.fill(250)
      p5.text(nodes[k].Name, 0, iconSize)
      if (showNodeInfo) {
        p5.textSize(fontSize-2)
        p5.text(nodes[k].IP, 0, iconSize + fontSize)
      }
    }
    p5.pop()
  }
}

export const setShowNodeInfo = (s) => {
  showNodeInfo = s
  mapRedraw = true
}

const mapMain = (p5) => {
  let portImage
  p5.preload = () => {
    portImage = p5.loadImage('/images/port.png')
  }

  p5.setup = () => {
    const c = p5.createCanvas(mapSizeX, mapSizeY)
    if(c && c.canvas) {
      if( (c.canvas.width * c.canvas.height) > 16777216) {
        console.log("resize canvas",c.canvas.width , c.canvas.height);
        p5.resizeCanvas(1000,1000);
        scale = 0.8;
      }
    }
  }

  p5.draw = () => {
    if (!mapRedraw) {
      return
    }
    if (scale !== 1.0) {
      p5.scale(scale)
    }
    mapRedraw = false
    drawBackImage(p5)
    drawItems(p5)
    drawNetworks(p5, portImage)
    drawLines(p5)
    drawNodes(p5)
  }

  p5.keyTyped = () => {
    switch (p5.key) {
      case '+': {
        scale += 0.05
        if (scale > 3.0) {
          scale = 3.0
        }
        mapRedraw = true
        break
      }
      case '-': {
        scale -= 0.05
        if (scale < 0.1) {
          scale = 0.1
        }
        mapRedraw = true
        break
      }
    }
  }
}


export const showMAP = (div) => {
  mapRedraw = true
  if (mapP5) {
    mapP5.remove();
    mapP5 = undefined;
  }
  mapP5 = new P5(mapMain, div)
}
