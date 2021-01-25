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
export default {
  mounted() {
    this.$setIconCodeMap(this.$iconList)
    this.$setStateColorMap(this.$stateList)
    this.$showMAP('map')
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
