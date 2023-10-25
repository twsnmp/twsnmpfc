import P5 from 'p5'

let ports = []
let cw = 1000
let ch = 500
let power = false
let rotate = false

const vpanelMain = (p) => {
  const PORT_SIZE = 150
  const LED_SIZE = 10
  const LED_XOFFSET = 300
  const LED_YOFFSET = 75
  let portImage
  let font
  let width
  let height
  let depth
  let r = 0
  p.preload = () => {
    font = p.loadFont('/fonts/inconsolata.ttf')
  }
  p.setup = () => {
    p.createCanvas(cw, ch, p.WEBGL)
    p.frameRate(10)
    portImage = p.loadImage('/images/port.png')
    p.textFont(font, 24)
    p.camera(100, -500, 2000, 0, 0, 0)
  }

  p.draw = () => {
    width = (ports.length < 16 ? ports.length : 16) * PORT_SIZE + PORT_SIZE
    height =
      Math.ceil(ports.length / 16) * ((5 * PORT_SIZE) / 4) + PORT_SIZE / 2
    if (ports.length === 0) {
      height = (5 * PORT_SIZE) / 4
    }
    depth = ports.length < 8 ? 800 : 1000
    // 背景色
    p.background(128)
    // カメラの制御
    p.orbitControl()
    if (rotate) {
      r += 0.1
    }
    p.rotateY(r)
    p.noStroke()
    p.textAlign(p.CENTER)
    p.push()
    // p.texture(box)
    p.fill(50, 50, 50)
    p.box(width, height, depth)
    for (let i = 0; i < ports.length; i++) {
      const x = i % 16
      const y = Math.floor(i / 16)
      p.push()
      // Port
      p.translate(
        -width / 2 + x * PORT_SIZE + PORT_SIZE,
        -height / 2 + y * ((5 * PORT_SIZE) / 4) + (3 * PORT_SIZE) / 4 + 10,
        depth / 2 + 1
      )
      p.texture(portImage)
      p.plane(150, 150)
      p.push()
      // Link up LED
      p.translate(2 * LED_SIZE - PORT_SIZE / 2, 2 * LED_SIZE - PORT_SIZE / 2, 0)
      p.fill(ports[i].State === 'up' ? '#11ee00' : ' #999')
      p.sphere(LED_SIZE)
      p.pop()
      p.push()
      // Speed LED
      p.translate(
        -2 * LED_SIZE + PORT_SIZE / 2,
        2 * LED_SIZE - PORT_SIZE / 2,
        0
      )
      p.fill(
        ports[i].Speed > 0 && ports[i].Speed < 1000 * 1000 * 1000
          ? '#eeaa00'
          : ' #999'
      )
      p.sphere(LED_SIZE)
      p.pop()
      p.push()
      p.fill('#ccc')
      p.text(i + 1 + '', 0, (3 * PORT_SIZE) / 4 - 10)
      p.pop()
      p.pop()
      // 裏面
      p.push()
      p.translate(
        width / 2 - i * LED_SIZE * 3 - LED_XOFFSET,
        -height / 2 + LED_YOFFSET,
        -depth / 2 - 1
      )
      p.push()
      // LED
      p.fill(ports[i].State === 'up' ? '#11ee00' : ' #999')
      p.sphere(LED_SIZE)
      p.pop()
      p.push()
      p.fill('#ccc')
      p.rotateY(p.radians(180.0))
      p.text(i + 1 + '', 0, 60)
      p.pop()
      p.pop()
    }
    p.push()
    p.translate(width / 2 - 100, -height / 2 + 50, -depth / 2 - 1)
    p.push()
    // LED
    p.fill(power ? '#2211ff' : '0x999')
    p.sphere(LED_SIZE)
    p.pop()
    p.push()
    p.fill('#ccc')
    p.rotateY(p.radians(180.0))
    p.text('POWER', 0, 60)
    p.pop()
    p.pop()
    p.pop()
  }
}

const setVPanel = (po, pw, r) => {
  ports = po
  power = pw
  rotate = r
}

const makeVPanel = (div) => {
  const d = document.getElementById(div)
  cw = d.clientWidth || 1000
  ch = d.clientHeight || 500
  new P5(vpanelMain, div) // eslint-disable-line no-new
}

export default (context, inject) => {
  inject('makeVPanel', makeVPanel)
  inject('setVPanel', setVPanel)
}
