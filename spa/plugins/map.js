import P5 from 'p5'
import { gauge,bar, line } from './echarts/drawitem'

let mapSize = 0
let mapSizeX = 2500
let mapSizeY = 5000
let mapP5
let contextMenu = true

let mapRedraw = true
let readOnly = false

let mapCallBack

let nodes = {}
let lines = []
let items = {}
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

const selectedNodes = []
const selectedItems = []

const iconCodeMap = {}
const imageMap = {}

/* eslint prettier/prettier: 0 */
const setIconCodeMap = (list) => {
  list.forEach((e) => {
    iconCodeMap[e.value] = String.fromCodePoint(e.code)
  })
  iconCodeMap.unknown = String.fromCodePoint(0xf0a39)
}

const setIconToMap = (e) => {
  iconCodeMap[e.Icon] = String.fromCodePoint(e.Code)
}

const showMAP = (div, m, url, ro) => {
  const ms = m.MapConf.MapSize || 0
  if (!mapP5 || ms !== mapSize) {
    mapSize = m.MapConf.mapSize
    initMap(div, ms)
  }
  if (!url || url === '/') {
    url = window.location.origin
  }
  readOnly = ro
  nodes = m.Nodes
  lines = m.Lines
  items = m.Items || {}
  backImage = m.MapConf.BackImage
  fontSize = m.MapConf.FontSize || 12
  iconSize = m.MapConf.IconSize || 24
  backImage.Image = null
  if (backImage.Path && mapP5) {
    mapP5.loadImage(url + '/backimage', (img) => {
      backImage.Image = img
      mapRedraw = true
    })
  }

  for (const k in items) {
    if (items[k].H < 10) {
      items[k].H = 100
    }
    if (items[k].W < 10) {
      items[k].W = 100
    }
    switch (items[k].Type) {
      case 3:
        if (!imageMap[items[k].Path] && mapP5) {
          mapP5.loadImage(url + '/image/' + items[k].Path, (img) => {
            imageMap[items[k].Path] = img
            mapRedraw = true
          })
        }
        break
      case 2:
      case 4:
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
          }
        )
        break
    }
  }
  mapRedraw = true
}

const initMap = (div, ms) => {
  if (mapP5) {
    mapP5.remove()
  }
  mapRedraw = true
  contextMenu = false
  document.oncontextmenu = (e) => {
    if (!contextMenu) {
      e.preventDefault()
    }
  }
  switch (ms) {
    case 1:
      mapSizeX = 5000
      mapSizeY = 5000
      break
    case 2:
      mapSizeX = 5000
      mapSizeY = 10000
      break
    case 3:
      mapSizeX = 10000
      mapSizeY = 10000
      break
    default:
      mapSizeX = 2500
      mapSizeY = 5000
      break
  }
  mapP5 = new P5(mapMain, div) // eslint-disable-line no-new
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

let scale = 1.0

const mapMain = (p5) => {
  let startMouseX
  let startMouseY
  let lastMouseX
  let lastMouseY
  let dragMode = 0 // 0 : None , 1: Select , 2 :Move
  const draggedNodes = []
  const draggedItems = []
  let clickInCanvas = false
  p5.setup = () => {
    const c = p5.createCanvas(mapSizeX, mapSizeY)
    c.mousePressed(canvasMousePressed)
  }

  p5.draw = () => {
    if (!mapRedraw) {
      return
    }
    if (scale !== 1.0) {
      p5.scale(scale)
    }
    mapRedraw = false
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
    for (const k in items) {
      p5.push()
      p5.translate(items[k].X, items[k].Y)
      if (selectedItems.includes(items[k].ID)) {
        p5.fill('rgba(23,23,23,0.9)')
        p5.stroke('#ccc')
        const w = items[k].W + 10
        const h = items[k].H + 10
        p5.rect(-5, -5, w, h)
      }
      switch (items[k].Type) {
        case 0: // rect
          p5.fill(items[k].Color)
          p5.stroke('rgba(23,23,23,0.9)')
          p5.rect(0, 0, items[k].W, items[k].H)
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
        case 5:
          {
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
                -p5.QUARTER_PI -
                  (p5.HALF_PI - (p5.HALF_PI * items[k].Value) / 100)
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
    for (const k in nodes) {
      const icon = getIconCode(nodes[k].Icon)
      p5.push()
      p5.translate(nodes[k].X, nodes[k].Y)
      if (selectedNodes.includes(nodes[k].ID)) {
        const w = iconSize + 16
        p5.fill('rgba(23,23,23,0.9)')
        p5.stroke(getStateColor(nodes[k].State))
        p5.rect(-w / 2, -w / 2, w, w)
      } else {
        const w = iconSize - 8
        p5.fill('rgba(23,23,23,0.9)')
        p5.stroke('rgba(23,23,23,0.9)')
        p5.rect(-w / 2, -w / 2, w, w)
      }
      p5.textFont('Material Design Icons')
      p5.textSize(iconSize)
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, 0, 0)
      p5.textFont('Roboto')
      p5.textSize(fontSize)
      p5.fill(250)
      p5.text(nodes[k].Name, 0, iconSize)
      p5.pop()
    }
    if (dragMode === 1) {
      let x = startMouseX
      let y = startMouseY
      let w = lastMouseX - startMouseX
      let h = lastMouseY - startMouseY
      if (startMouseX > lastMouseX) {
        x = lastMouseX
        w = startMouseX - lastMouseX
      }
      if (startMouseY > lastMouseY) {
        y = lastMouseY
        h = startMouseY - lastMouseY
      }
      p5.push()
      p5.fill('rgba(250,250,250,0.6)')
      p5.stroke(0)
      p5.rect(x, y, w, h)
      p5.pop()
    }
  }

  p5.mouseDragged = () => {
    if (readOnly) {
      return true
    }
    if (dragMode === 0) {
      if (selectedNodes.length > 0 || selectedItems.length > 0) {
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
    lastMouseX = p5.mouseX / scale
    lastMouseY = p5.mouseY / scale
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
    } else if (dragMode !== 3) {
      setSelectNode(false)
      setSelectItem()
    }
    lastMouseX = p5.mouseX / scale
    lastMouseY = p5.mouseY / scale
    startMouseX = p5.mouseX / scale
    startMouseY = p5.mouseY / scale
    dragMode = 0
    return false
  }

  p5.mouseReleased = (e) => {
    if (readOnly) {
      return true
    }
    mapRedraw = true
    if (!clickInCanvas) {
      selectedNodes.length = 0
      selectedItems.length = 0
      return true
    }
    if (
      p5.mouseButton === p5.RIGHT &&
      selectedNodes.length + selectedItems.length < 2
    ) {
      if (mapCallBack) {
        mapCallBack({
          Cmd: 'contextMenu',
          Node: selectedNodes[0] || '',
          Item: selectedItems[0] || '',
          x: p5.winMouseX,
          y: p5.winMouseY,
        })
      }
    }
    clickInCanvas = false
    if (dragMode === 0 || dragMode === 3) {
      dragMode = 0
      return false
    }
    if (dragMode === 1) {
      if (selectedNodes.length > 0 || selectedItems.length > 0) {
        dragMode = 3
      } else {
        dragMode = 0
      }
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

  p5.keyTyped = () => {
    if (p5.key === '+') {
      scale += 0.05
      if (scale > 3.0) {
        scale = 3.0
      }
      mapRedraw = true
    } else if (p5.key === '-') {
      scale -= 0.05
      if (scale < 0.1) {
        scale = 0.1
      }
      mapRedraw = true
    }
  }

  p5.keyReleased = () => {
    if (readOnly) {
      return true
    }
    if (p5.keyCode === p5.DELETE || p5.keyCode === p5.BACKSPACE) {
      // Delete
      if (selectedNodes.length > 0) {
        deleteNodes()
      }
    }
    if (p5.keyCode === p5.ENTER) {
      p5.doubleClicked()
    }
    return true
  }

  p5.doubleClicked = () => {
    if (selectedNodes.length === 1) {
      nodeDoubleClicked()
    } else if (selectedItems.length === 1) {
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
    if (n.X > mapSizeX) {
      n.X = mapSizeX - 16
    }
    if (n.Y > mapSizeY) {
      n.Y = mapSizeY - 16
    }
  }
  const checkItemPos = (i) => {
    if (i.X < 16) {
      i.X = 16
    }
    if (i.Y < 16) {
      i.Y = 16
    }
    if (i.X > mapSizeX - i.W) {
      i.X = mapSizeX - i.W
    }
    if (i.Y > mapSizeY - i.H) {
      i.Y = mapSizeY - i.H
    }
  }
  const dragMoveNodes = () => {
    selectedNodes.forEach((id) => {
      if (nodes[id]) {
        nodes[id].X += (p5.mouseX /scale) - lastMouseX
        nodes[id].Y += (p5.mouseY /scale) - lastMouseY
        checkNodePos(nodes[id])
        if (!draggedNodes.includes(id)) {
          draggedNodes.push(id)
        }
      }
    })
    selectedItems.forEach((id) => {
      if (items[id]) {
        items[id].X += (p5.mouseX / scale) - lastMouseX
        items[id].Y += (p5.mouseY / scale) - lastMouseY
        checkItemPos(items[id])
        if (!draggedItems.includes(id)) {
          draggedItems.push(id)
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
    selectedItems.length = 0
    for (const k in items) {
      if (
        items[k].X > sx &&
        items[k].X < lx &&
        items[k].Y > sy &&
        items[k].Y < ly
      ) {
        selectedItems.push(items[k].ID)
      }
    }
    mapRedraw = true
  }

  const setSelectNode = (bMulti) => {
    const l = selectedNodes.length
    const x = p5.mouseX / scale
    const y = p5.mouseY / scale
    for (const k in nodes) {
      if (
        nodes[k].X + 32 > x &&
        nodes[k].X - 32 < x &&
        nodes[k].Y + 32 > y &&
        nodes[k].Y - 32 < y
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
    const x = p5.mouseX / scale
    const y = p5.mouseY / scale
    for (const k in items) {
      const w =
        items[k].Type === 2
          ? items[k].Size * items[k].Text.length + 10
          : items[k].W + 10
      const h = items[k].Type === 2 ? items[k].Size + 10 : items[k].H + 10
      if (
        items[k].X + w > x &&
        items[k].X - 10 < x &&
        items[k].Y + h > y &&
        items[k].Y - 10 < y
      ) {
        if (selectedItems.includes(items[k].ID)) {
          return
        }
        selectedItems.push(items[k].ID)
        return
      }
    }
    selectedItems.length = 0
  }
  // ノードを削除する
  const deleteNodes = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'deleteNodes',
        Param: selectedNodes,
      })
      selectedNodes.length = 0
    }
  }
  // Nodeの位置を保存する
  const updateNodesPos = () => {
    const list = []
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
    const list = []
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
        Param: selectedItems[0],
      })
    }
  }
  // lineの編集
  const editLine = () => {
    if (selectedNodes.length !== 2) {
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
  if (nodes[id]) {
    selectedNodes.length = 0
    selectedNodes.push(id)
    mapRedraw = true
  }
}

export default (context, inject) => {
  inject('showMAP', showMAP)
  inject('setIconCodeMap', setIconCodeMap)
  inject('setStateColorMap', setStateColorMap)
  inject('setCallback', setCallback)
  inject('selectNode', selectNode)
  inject('refreshMAP', refreshMAP)
  inject('setMapContextMenu', setMapContextMenu)
  inject('setIconToMap', setIconToMap)
}
