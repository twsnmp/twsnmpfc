import P5 from 'p5'

const MAP_SIZE_X = 2500
const MAP_SIZE_Y = 5000

let nodes = {
  1: { Name: 'test', X: 100, Y: 100, Icon: 'desktop', State: 'high', ID: 1 },
  2: { Name: 'test2', X: 180, Y: 100, Icon: 'router', State: 'repair', ID: 2 },
}
let lines = [{ NodeID1: 1, NodeID2: 2, State1: 'high', State2: 'repair' }]

const iconCodeMap = {}

const setIconCodeMap = (list) => {
  list.forEach((e) => {
    iconCodeMap[e.value] = String.fromCodePoint(e.code)
  })
  iconCodeMap.unknown = String.fromCodePoint(0x1111)
}

const setMapData = (md) => {
  nodes = {}
  lines = []
}

const getIconCode = (icon) => {
  return iconCodeMap[icon] ? iconCodeMap[icon] : iconCodeMap.unknown
}

const stateColorMap = {}

const setStateColorMap = (list) => {
  list.forEach((e) => {
    stateColorMap[e.value] = e.color
  })
}

const getStateColor = (state) => {
  return stateColorMap[state] ? stateColorMap[state] : 'gray'
}

const mapMain = function (p5) {
  let lastMouseX
  let lastMouseY
  const draggedNodes = []
  const selectedNodes = []

  p5.setup = (_) => {
    p5.createCanvas(MAP_SIZE_X, MAP_SIZE_Y)
  }

  p5.draw = (_) => {
    p5.background(250)
    // if(backimg){
    //   p5.image(backimg,0,0);
    // }
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
      p5.strokeWeight(2)
      p5.stroke(getStateColor(lines[k].State1))
      p5.line(x1, y1, xm, ym)
      p5.stroke(getStateColor(lines[k].State2))
      p5.line(xm, ym, x2, y2)
      p5.pop()
    }
    for (const k in nodes) {
      const icon = getIconCode(nodes[k].Icon)
      p5.push()
      p5.translate(nodes[k].X, nodes[k].Y)
      if (selectedNodes.includes(nodes[k].ID)) {
        p5.fill('rgba(240,248,255,0.9)')
        p5.stroke(getStateColor(nodes[k].State))
        p5.rect(-24, -24, 48, 48)
      } else {
        p5.fill('rgba(250,250,250,0.8)')
        p5.stroke(250)
        p5.rect(-18, -18, 36, 36)
      }
      p5.textFont('Material Design Icons')
      p5.textSize(32)
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.fill(0)
      p5.text(icon, 0, 0)
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, -1, -1)
      p5.textFont('Arial')
      p5.textSize(12)
      p5.fill(0)
      p5.text(nodes[k].Name, 0, 32)
      p5.pop()
    }
  }
  const setSelectNode = (bMulti) => {
    const l = selectedNodes.length
    for (const k in nodes) {
      if (
        nodes[k].X + 32 > p5.mouseX &&
        nodes[k].X - 32 < p5.mouseX &&
        nodes[k].Y + 32 > p5.mouseY &&
        nodes[k].Y - 32 < p5.mouseY
      ) {
        if (selectedNodes.includes(nodes[k].ID)) {
          return false
        }
        if (!bMulti) {
          selectedNodes.length = 0
        }
        selectedNodes.push(nodes[k].ID)
        return true
      }
    }
    if (!bMulti) {
      selectedNodes.length = 0
    }
    return l !== selectedNodes.length
  }
  p5.mouseDragged = () => {
    if (
      p5.winMouseX < 32 ||
      p5.winMouseY < 32 ||
      p5.winMouseY > p5.windowHeight * 0.75
    ) {
      return true
    }
    if (lastMouseX) {
      selectedNodes.forEach((id) => {
        if (nodes[id]) {
          nodes[id].X += p5.mouseX - lastMouseX
          nodes[id].Y += p5.mouseY - lastMouseY
          if (nodes[id].X < 16) {
            nodes[id].X = 16
          }
          if (nodes[id].Y < 16) {
            nodes[id].Y = 16
          }
          if (nodes[id].X > MAP_SIZE_X) {
            nodes[id].X = MAP_SIZE_X - 16
          }
          if (nodes[id].Y > MAP_SIZE_Y) {
            nodes[id].Y = MAP_SIZE_Y - 16
          }
          if (!draggedNodes.includes(id)) {
            draggedNodes.push(id)
          }
        }
      })
      p5.redraw()
    }
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    return true
  }
  p5.mousePressed = () => {
    if (
      p5.keyIsDown(p5.SHIFT) &&
      selectedNodes.length > 0 &&
      setSelectNode(true)
    ) {
      // createEditLinePane(selectedNodes[0], selectedNodes[1]);
      selectedNodes.length = 0
      return true
    } else if (p5.keyIsDown(p5.CONTROL)) {
      if (setSelectNode(true)) {
        p5.redraw()
      }
    } else {
      setSelectNode(false)
    }
    if (p5.mouseButton === p5.RIGHT && selectedNodes.length <= 1) {
      const id = selectedNodes[0]
      if (nodes[id]) {
      }
    }
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    return true
  }
  p5.mouseClicked = () => {
    return false
  }
  p5.mouseReleased = () => {
    if (draggedNodes.length === 0) {
      return
    }
    draggedNodes.forEach((id) => {
      if (nodes[id]) {
        // ノードの位置を保存
      }
    })
    draggedNodes.length = 0
  }
  p5.keyReleased = () => {
    if (!p5.focused) {
      return false
    }
    if (p5.keyCode === p5.DELETE || p5.keyCode === p5.BACKSPACE) {
      // Delete
    }
    if (p5.keyCode === p5.ENTER) {
      p5.doubleClicked()
    }
    return true
  }
  p5.doubleClicked = () => {
    // Show Node Info
    return true
  }
}

const showMAP = (div) => {
  new P5(mapMain, div) // eslint-disable-line no-new
}

export default (context, inject) => {
  inject('showMAP', showMAP)
  inject('setIconCodeMap', setIconCodeMap)
  inject('setStateColorMap', setStateColorMap)
  inject('setMapData', setMapData)
}
