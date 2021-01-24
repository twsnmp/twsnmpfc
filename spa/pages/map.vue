<template>
  <v-card class="mx-auto">
    <div id="map"></div>
    <blockquote class="blockquote">
      &#8220;ここにマップを表示する予定！！&#8221;
      <footer>
        <small>
          <em>&mdash;TWSNMP FC</em>
        </small>
      </footer>
    </blockquote>
  </v-card>
</template>

<script>
import P5 from 'p5'

export default {
  mounted() {
    const script = function (p5) {
      let speed = 2
      let posX = 0
      p5.setup = (_) => {
        p5.createCanvas(2500, 5000)
        p5.ellipse(p5.width / 2, p5.height / 2, 500, 500)
      }

      p5.draw = (_) => {
        p5.background(0)
        const degree = p5.frameCount * 3
        const y = p5.sin(p5.radians(degree)) * 50
        p5.push()
        p5.translate(0, p5.height / 2)
        p5.ellipse(posX, y, 50, 50)
        p5.pop()

        posX += speed
        if (posX > p5.width || posX < 0) {
          speed *= -1
        }
      }
    }

    new P5(script, 'map') // eslint-disable-line no-new
  },
}
</script>

<style>
#map {
  width: 100%;
  height: 800px;
  overflow: scroll;
}
</style>
