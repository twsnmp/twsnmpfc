import P5 from 'p5'

const MAP_SIZE_X = 2500
const MAP_SIZE_Y = 5000
let mapRedraw = true

let mapCallBack

let nodes = {}
let lines = []
const selectedNodes = []

const iconCodeMap = {}

/* eslint prettier/prettier: 0 */
const setIconCodeMap = (list) => {
  list.forEach((e) => {
    iconCodeMap[e.value] = String.fromCodePoint(e.code)
  })
  iconCodeMap.unknown = String.fromCodePoint(0xF0A39)
}

const setNodes = (n) => {
  nodes = n
  mapRedraw = true
}

const setLines = (l) => {
  lines = l
  mapRedraw = true
}

const setCallback = (cb) => {
  mapCallBack = cb
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

const mapMain = (p5) => {
  let startMouseX
  let startMouseY
  let lastMouseX
  let lastMouseY
  let dragMode  = 0 // 0 : None , 1: Select , 2 :Move
  const draggedNodes = []

  p5.setup = () => {
    p5.createCanvas(MAP_SIZE_X, MAP_SIZE_Y)
  }

  p5.draw = () => {
    if (! mapRedraw){
      return
    }
    mapRedraw = false
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
      p5.strokeWeight(1)
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
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, 0, 0)
      p5.textFont('Arial')
      p5.textSize(12)
      p5.fill(0)
      p5.text(nodes[k].Name, 0, 32)
      p5.pop()
    }
    if (dragMode === 1) {
      let x = startMouseX
      let y = startMouseY
      let w = lastMouseX  - startMouseX
      let h = lastMouseY  - startMouseY
      if (startMouseX > lastMouseX){
        x = lastMouseX
        w = startMouseX - lastMouseX
      }
      if (startMouseY > lastMouseY){
        y = lastMouseY
        h = startMouseY - lastMouseY
      }
      p5.push()
      p5.fill('rgba(250,250,250,0.6)')
      p5.stroke(0)
      p5.rect(x,y,w,h)
      p5.pop()
    } 
  }

  p5.mouseDragged = () => {
    if (
      p5.winMouseX < 32 ||
      p5.winMouseY < 32 ||
      p5.winMouseY > p5.windowHeight * 0.75
    ) {
      return true
    }
    if (dragMode === 0) {
      if (selectedNodes.length > 0 ){
        dragMode = 2
      } else {
        dragMode = 1
      }
    }
    if (dragMode === 1) {
      dragSelectNodes()
    } else if (dragMode === 2 && lastMouseX) {
      dragMoveNodes()
    }
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    return true
  }

  p5.mousePressed = () => {
    if (
      p5.keyIsDown(p5.SHIFT) &&
      selectedNodes.length === 1 &&
      setSelectNode(true)
    ) {
      editLine()
      selectedNodes.length = 0
      return true
    } else if (setSelectNode(false)) {
      mapRedraw = true
    }
    console.log("mousePressed",p5.mouseY)
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    startMouseX = p5.mouseX
    startMouseY = p5.mouseY
    dragMode = 0
    return true
  }

  p5.mouseClicked = () => {
    console.log("mouseClicked",p5.mouseY)
    return false
  }

  p5.mouseReleased = () => {
    if( dragMode === 0) {
      return false
    }
    if( dragMode === 1) {
      dragMode = 0
      mapRedraw = true
      return false
    }
    if (draggedNodes.length === 0) {
      return false
    }
    updateNodesPos()
  }

  p5.keyReleased = () => {
    if (!p5.focused) {
      return false
    }
    if (p5.keyCode === p5.DELETE || p5.keyCode === p5.BACKSPACE) {
      // Delete
      deleteNodes()
    }
    if (p5.keyCode === p5.ENTER) {
      p5.doubleClicked()
    }
    return true
  }

  p5.doubleClicked = () => {
    // Show Node Info
    showNode()
    return true
  }

  const dragMoveNodes = () => {
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
    mapRedraw = true
  }

  const dragSelectNodes = () => {
    selectedNodes.length = 0
    const sx = startMouseX < lastMouseX ? startMouseX : lastMouseX
    const sy = startMouseY < lastMouseY ? startMouseY : lastMouseY
    const lx = startMouseX > lastMouseX ? startMouseX : lastMouseX
    const ly = startMouseY > lastMouseY ? startMouseY : lastMouseY
    for (const k in nodes) {
      if (
        nodes[k].X > sx &&
        nodes[k].X < lx &&
        nodes[k].Y > sy &&
        nodes[k].Y < ly
      ) {
        selectedNodes.push(nodes[k].ID)
      }
    }
    mapRedraw = true
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
  // ノードを削除する
  const deleteNodes = () => {
    if (mapCallBack){
      mapCallBack({
        Cmd: 'deleteNodes',
        Param: selectedNodes,
      })
      selectedNodes.length = 0
    }
  }
  // Nodeの位置を保存する
  const updateNodesPos = () => {
    const list  = []
    draggedNodes.forEach((id) => {
      if (nodes[id]) {
        // 位置を保存するノード
        list.push({
          ID: id,
          X: nodes[id].X,
          Y: nodes[id].Y,
        })
      }
    })
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'updateNodesPos',
        Param: list,
      })
    }
    draggedNodes.length = 0
  }
  const showNode = () => {
    if (selectedNodes.length !== 1 ){
      return
    }
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'showNode',
        Param: selectedNodes[0],
      })
    }
  }
  const editLine = () => {
    if (selectedNodes.length !== 2 ){
      return
    }
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'editLine',
        Param: selectedNodes,
      })
    }
    selectedNodes.length = 0
    mapRedraw = true
  }
}

const showMAP = (div) => {
  new P5(mapMain, div) // eslint-disable-line no-new
}

const selectNode = (id) => {
  if( nodes[id] ) {
    selectedNodes.length = 0
    selectedNodes.push(id)
    mapRedraw = true
  }
}

export default (context, inject) => {
  inject('showMAP', showMAP)
  inject('setIconCodeMap', setIconCodeMap)
  inject('setStateColorMap', setStateColorMap)
  inject('setNodes', setNodes)
  inject('setLines', setLines)
  inject('setCallback', setCallback)
  inject('selectNode', selectNode)
}
