import P5 from 'p5'

const MAP_SIZE_X = 2500
const MAP_SIZE_Y = 5000
let mapRedraw = true
let readOnly = false

let mapCallBack

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

const selectedNodes = []
let selectedItem = ""

const iconCodeMap = {}
const imageMap = {}

const _ImageP5 = new P5()

/* eslint prettier/prettier: 0 */
const setIconCodeMap = (list) => {
  list.forEach((e) => {
    iconCodeMap[e.value] = String.fromCodePoint(e.code)
  })
  iconCodeMap.unknown = String.fromCodePoint(0xF0A39)
}

const setIconToMap = (e) => {
  iconCodeMap[e.Icon] = String.fromCodePoint(e.Code)
} 

const setMAP = (m,url,ro) => {
  if (!url || url === '/') {
    url = window.location.origin
  }
  readOnly = ro
  nodes = m.Nodes
  lines = m.Lines
  items = m.Items || {}
  backImage = m.MapConf.BackImage
  backImage.Image = null
  if (backImage.Path){
    _ImageP5.loadImage(url+'/backimage',(img)=>{
      backImage.Image = img
      mapRedraw = true
    })
  }

  for(const k in items) {
    if (items[k].Type === 3 && !imageMap[items[k].Path]) {
      _ImageP5.loadImage(url+'/image/' + items[k].Path,(img)=>{
        imageMap[items[k].Path] = img
        mapRedraw = true
      })
    } 
  }
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
  let dragMode  = 0 // 0 : None , 1: Select , 2 :Move
  const draggedNodes = []
  const draggedItems = []
  let clickInCanvas = false
  p5.setup = () => {
    const c = p5.createCanvas(MAP_SIZE_X, MAP_SIZE_Y)
    c.mousePressed(canvasMousePressed)
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
    for (const k in items) {
      p5.push()
      p5.translate(items[k].X, items[k].Y)
      if (selectedItem === items[k].ID ) {
        p5.fill('rgba(23,23,23,0.9)')
        p5.stroke('#ccc')
        const w =  items[k].Type === 2 ?  (items[k].Size *  items[k].Text.length) + 10 : items[k].W +10
        const h =  items[k].Type === 2 ?  items[k].Size + 10 : items[k].H +10
        p5.rect(-5, -5, w, h)
      }
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
        p5.textSize(items[k].Size || 12)
        p5.fill(items[k].Color)
        p5.text(items[k].Text, 0, 0,items[k].Size *  items[k].Text.length + 10, items[k].Size + 10)
        break
      case 3: // Image
        if (imageMap[items[k].Path]) {
          p5.image(imageMap[items[k].Path],0,0,items[k].W,items[k].H)
        }
        break
      }
      p5.pop()
    }
    for (const k in nodes) {
      const icon = getIconCode(nodes[k].Icon)
      p5.push()
      p5.translate(nodes[k].X, nodes[k].Y)
      if (selectedNodes.includes(nodes[k].ID)) {
        p5.fill('rgba(23,23,23,0.9)')
        p5.stroke(getStateColor(nodes[k].State))
        p5.rect(-24, -24, 48, 48)
      } else {
        p5.fill('rgba(23,23,23,0.9)')
        p5.stroke('rgba(23,23,23,0.9)')
        p5.rect(-12, -12, 24, 24)
      }
      p5.textFont('Material Design Icons')
      p5.textSize(32)
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, 0, 0)
      p5.textFont('Roboto')
      p5.textSize(12)
      p5.fill(250)
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
    if (readOnly) {
      return true
    }
    if (dragMode === 0) {
      if (selectedNodes.length > 0 || selectedItem !== "" ){
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

  const canvasMousePressed = () => {
    if (readOnly) {
      return true
    }
    clickInCanvas = true
    mapRedraw = true
    if (
      p5.keyIsDown(p5.SHIFT) &&
      selectedNodes.length === 1 &&
      setSelectNode(true)
    ) {
      editLine()
      selectedNodes.length = 0
      return false
    } else  {
      setSelectNode(false)
      setSelectItem()
    }
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    startMouseX = p5.mouseX
    startMouseY = p5.mouseY
    dragMode = 0
    return false
  }

  p5.mouseReleased = (e) => {
    if (readOnly) {
      return true
    }
    mapRedraw = true
    if(!clickInCanvas){
      selectedNodes.length = 0
      selectedItem = ""
      return true
    }
    if(p5.mouseButton === p5.RIGHT && selectedNodes.length < 2 ) {
      if (mapCallBack) {
        mapCallBack({
          Cmd: 'contextMenu',
          Node: selectedNodes[0] || '',
          Item: selectedItem || '',
          x: p5.winMouseX,
          y: p5.winMouseY,
        })
      }
    }
    clickInCanvas = false 
    if (dragMode === 0) {
      return false
    }
    if (dragMode === 1) {
      dragMode = 0
      return false
    }
    if (draggedNodes.length > 0) {
      updateNodesPos()
    }
    if (draggedItems.length > 0) {
      updateItemsPos()
    }
    return false
  }

  p5.keyReleased = () => {
    if (readOnly) {
      return true
    }
    if (p5.keyCode === p5.DELETE || p5.keyCode === p5.BACKSPACE) {
      // Delete
      if (selectedNodes.length > 0){
        deleteNodes()
      }
    }
    if (p5.keyCode === p5.ENTER) {
      p5.doubleClicked()
    }
    return true
  }

  p5.doubleClicked = () => {
    if (selectedNodes.length === 1 ){
      nodeDoubleClicked()
    } else if (selectedItem !== "" ){
      itemDoubleClicked()
    }
    return true
  }
  const checkNodePos = (n) => {
    if (n.X < 16) {
      n.X = 16
    }
    if (n.Y < 16) {
      n.Y = 16
    }
    if (n.X > MAP_SIZE_X) {
      n.X = MAP_SIZE_X - 16
    }
    if (n.Y > MAP_SIZE_Y) {
      n.Y = MAP_SIZE_Y - 16
    }
  }
  const checkItemPos = (i) => {
    if (i.X < 16) {
      i.X = 16
    }
    if (i.Y < 16) {
      i.Y = 16
    }
    if (i.X > MAP_SIZE_X - i.W) {
      i.X = MAP_SIZE_X - i.W
    }
    if (i.Y > MAP_SIZE_Y - i.H) {
      i.Y = MAP_SIZE_Y - i.H
    }
  }
  const dragMoveNodes = () => {
    selectedNodes.forEach((id) => {
      if (nodes[id]) {
        nodes[id].X += p5.mouseX - lastMouseX
        nodes[id].Y += p5.mouseY - lastMouseY
        checkNodePos(nodes[id])
        if (!draggedNodes.includes(id)) {
          draggedNodes.push(id)
        }
      }
    })
    if (selectedItem !== "" && items[selectedItem] ) {
      items[selectedItem].X += p5.mouseX - lastMouseX
      items[selectedItem].Y += p5.mouseY - lastMouseY
      checkItemPos(items[selectedItem])
      if (!draggedItems.includes(selectedItem)) {
        draggedItems.push(selectedItem)
      }
    }
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
  // 描画アイテムを選択する
  const setSelectItem = () => {
    for (const k in items) {
      const w =  items[k].Type === 2 ?  (items[k].Size *  items[k].Text.length) + 10 : items[k].W +10
      const h =  items[k].Type === 2 ?  items[k].Size + 10 : items[k].H +10
      if (
        items[k].X + w > p5.mouseX &&
        items[k].X - 10 < p5.mouseX &&
        items[k].Y + h > p5.mouseY &&
        items[k].Y - 10 < p5.mouseY
      ) {
        selectedItem = items[k].ID
        return
      }
    }
    selectedItem = ""
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
  // 描画アイテムの位置を保存する
  const updateItemsPos = () => {
    const list  = []
    draggedItems.forEach((id) => {
      if (items[id]) {
        // 位置を保存するノード
        list.push({
          ID: id,
          X: items[id].X,
          Y: items[id].Y,
        })
      }
    })
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'updateItemsPos',
        Param: list,
      })
    }
    draggedItems.length = 0
  }
  // nodeをダブルクリックした場合
  const nodeDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'nodeDoubleClicked',
        Param: selectedNodes[0],
      })
    }
  }
  // itemをダブルクリックした場合
  const itemDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'itemDoubleClicked',
        Param: selectedItem,
      })
    }
  }
  // lineの編集
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

let  hasMAP = false
let  contextMenu = true
const showMAP = (div) => {
  mapRedraw = true
  contextMenu = false
  if (hasMAP) {
    return
  }
  document.oncontextmenu = (e) => {
    if (!contextMenu) {
      e.preventDefault()
    }
  }  
  new P5(mapMain, div) // eslint-disable-line no-new
  hasMAP = true
}

const setMapContextMenu = (e) => {
  contextMenu = e
}

const refreshMAP = () => {
  if (mapCallBack) {
    mapCallBack({
      Cmd: 'refresh',
      Param: '',
    })
  }
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
  inject('setMAP', setMAP)
  inject('setCallback', setCallback)
  inject('selectNode', selectNode)
  inject('refreshMAP', refreshMAP)
  inject('setMapContextMenu', setMapContextMenu)
  inject('setIconToMap', setIconToMap)
}