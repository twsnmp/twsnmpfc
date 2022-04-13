import P5 from 'p5'
import { getState,getIconCode } from './common'
const MAP_SIZE_X = 2500
const MAP_SIZE_Y = 5000

let mapRedraw = true
let hasMAP = false

let nodes = {}
let lines = []
let backImage = {
  X:0,
  Y:0,
  Width:0,
  Height: 0,
  Path: '',
  Color: 23,
  Image: null,
}

export const setMAP = (m) => {
  nodes = m.Nodes
  lines = m.Lines
  backImage = m.MapConf.BackImage
  backImage.Image = null
  if (backImage.Path){
    const _p5 = new P5()
    _p5.loadImage('APIURL'+'/backimage',(img)=>{
      backImage.Image = img
      mapRedraw = true
    })
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

const mapMain = (p5) => {
  let startMouseX
  let startMouseY
  let lastMouseX
  let lastMouseY
  p5.setup = () => {
    p5.createCanvas(MAP_SIZE_X, MAP_SIZE_Y)
  }

  p5.draw = () => {
    if (!mapRedraw){
      return
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
}

export const showMAP = (div) => {
  mapRedraw = true
  if (hasMAP) {
    return
  }
  new P5(mapMain, div)
  hasMAP = true
}

