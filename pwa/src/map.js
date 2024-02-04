import P5 from 'p5'
import { getState,getIconCode } from './common'
const MAP_SIZE_X = 2500
const MAP_SIZE_Y = 5000

let mapRedraw = true

let nodes = {}
let lines = []
let items = {}
let backImage = {
  X:0,
  Y:0,
  Width:0,
  Height: 0,
  Path: '',
  Color: 23,
  Image: null,
}
const imageMap = {}
let map;


export const setMAP = (m) => {
  nodes = m.Nodes
  lines = m.Lines || []
  items = m.Items || {}
  backImage = m.MapConf.BackImage
  backImage.Image = null
  if (backImage.Path && map){
    map.loadImage('APIURL'+'/backimage',(img)=>{
      backImage.Image = img
      mapRedraw = true
    })
  }
  for(const k in items) {
    switch (items[k].Type) {
    case 3:
      if (!imageMap[items[k].Path] && map) {
        map.loadImage('APIURL'+'/image/' + items[k].Path,(img)=>{
          imageMap[items[k].Path] = img
          mapRedraw = true
        })
      }  
      break
    case 2:
    case 4:
      items[k].W = items[k].Size *  items[k].Text.length
      items[k].H = items[k].Size
      break
    case 5:
      items[k].H = items[k].Size * 10
      items[k].W = items[k].Size * 10
      break
    } 
  }
  mapRedraw = true
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

let scale = 1.0
const mapMain = (p5) => {
  p5.setup = () => {
    const c = p5.createCanvas(MAP_SIZE_X, MAP_SIZE_Y)
    if(c && c.canvas) {
      if( (c.canvas.width * c.canvas.height) > 16777216) {
        console.log("resize canvas",c.canvas.width , c.canvas.height);
        p5.resizeCanvas(1000,1000);
        scale = 0.8;
      }
    }
  }

  p5.draw = () => {
    if (!mapRedraw){
      return
    }
    if (scale != 1.0) {
      p5.scale(scale);
    }
    mapRedraw = false
    p5.background(backImage.Color || 23)
    if(backImage.Image){
      if(backImage.Width){
        p5.image(backImage.Image,backImage.X,backImage.Y,backImage.Width,backImage.Height);
      }else {
        p5.image(backImage.Image,backImage.X,backImage.Y);
      }
    }
    for (const k in lines) {
      if (!nodes[lines[k].NodeID1] || !nodes[lines[k].NodeID2]) {
        continue
      }
      const x1 = nodes[lines[k].NodeID1].X
      const x2 = nodes[lines[k].NodeID2].X
      const y1 = nodes[lines[k].NodeID1].Y + 6
      const y2 = nodes[lines[k].NodeID2].Y + 6
      const xm = (x1 + x2) / 2
      const ym = (y1 + y2) / 2
      p5.push()
      p5.strokeWeight(lines[k].Width || 1 )
      p5.stroke(getStateColor(lines[k].State1))
      p5.line(x1, y1, xm, ym)
      p5.stroke(getStateColor(lines[k].State2))
      p5.line(xm, ym, x2, y2)
      if (lines[k].Info) {
        const color = getLineColor(lines[k].State)
        p5.textFont('Roboto')
        p5.textSize(10)
        p5.fill(color)
        p5.stroke(color)
        p5.strokeWeight(1)
        p5.text(lines[k].Info, xm + 10, ym)  
      }
      p5.pop()
    }
    for (const k in items) {
      p5.push()
      p5.translate(items[k].X, items[k].Y)
      switch (items[k].Type) {
      case 0: // rect
        p5.fill(items[k].Color)
        p5.stroke('rgba(23,23,23,0.9)')
        p5.rect(0,0,items[k].W, items[k].H)
        break
      case 1: // ellipse
        p5.fill(items[k].Color)
        p5.stroke('rgba(23,23,23,0.9)')
        p5.ellipse(items[k].W/2,items[k].H/2,items[k].W, items[k].H)
        break
      case 2: // text
      case 4: // Polling
        p5.textSize(items[k].Size || 12)
        p5.fill(items[k].Color)
        p5.text(items[k].Text, 0, 0,items[k].Size *  items[k].Text.length + 10, items[k].Size + 10)
        break
      case 3: // Image
        if (imageMap[items[k].Path]) {
          p5.image(imageMap[items[k].Path],0,0,items[k].W,items[k].H)
        }
        break
      case 5: { // Gauge
          const x = items[k].W / 2
          const y = items[k].H / 2
          const r0 = items[k].W 
          const r1 = (items[k].W - items[k].Size) 
          const r2 = (items[k].W - items[k].Size *4)
          p5.noStroke()
          p5.fill('#eee')
          p5.arc(x, y, r0, r0, 5*p5.QUARTER_PI, -p5.QUARTER_PI)
          if(items[k].Value > 0){
            p5.fill(items[k].Color)
            p5.arc(x, y, r0, r0, 5*p5.QUARTER_PI, -p5.QUARTER_PI - (p5.HALF_PI - p5.HALF_PI * items[k].Value/100))
          }
          p5.fill(backImage.Color || 23)
          p5.arc(x, y, r1, r1, -p5.PI, 0)
          p5.textAlign(p5.CENTER)
          p5.textSize(8)
          p5.fill('#fff')
          p5.text( items[k].Value + '%', x, y - 10 )
          p5.textSize(items[k].Size)
          p5.text( items[k].Text || "", x, y + 5)
          p5.fill('#e31a1c')
          const angle = -p5.QUARTER_PI + (p5.HALF_PI * items[k].Value/100)
          const x1 = x + r1/2 * p5.sin(angle)
          const y1 = y - r1/2 * p5.cos(angle)
          const x2 = x + r2/2 * p5.sin(angle) + 5  * p5.cos(angle)
          const y2 = y - r2/2 * p5.cos(angle) + 5  * p5.sin(angle)
          const x3 = x + r2/2 * p5.sin(angle) - 5  * p5.cos(angle)
          const y3 = y - r2/2 * p5.cos(angle) - 5  * p5.sin(angle)
          p5.triangle(x1, y1, x2, y2, x3, y3)
        }
      }
      p5.pop()
    }
    for (const k in nodes) {
      const icon = getIconCode(nodes[k].Icon)
      p5.push()
      p5.translate(nodes[k].X, nodes[k].Y)
      p5.textFont('Material Design Icons')
      p5.textSize(32)
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, 0, 0)
      p5.textFont('monospace')
      p5.textSize(12)
      p5.fill(250)
      p5.text(nodes[k].Name, 0, 32)
      p5.pop()
    }
  }
  p5.keyTyped = () => {
    if (p5.key === '+') {
      scale +=0.05
      if (scale > 3.0) {
        scale = 3.0
      }
      mapRedraw = true
    } else if (p5.key === '-') {
      scale -=0.05
      if (scale < 0.1) {
        scale = 0.1
      }
      mapRedraw = true
    }
  }
}


export const showMAP = (div) => {
  mapRedraw = true
  if (map) {
    map.remove();
    map = undefined;
  }
  map = new P5(mapMain, div)
}

