<template>
  <v-row>
    <v-card class="mx-auto">
      <v-card-title>
        管理マップ - {{ map.MapConf.MapName }}
        <v-spacer></v-spacer>
        <v-row dense>
          <v-col cols="10">
            <v-select
              v-model="selectedNodeID"
              :items="nodeList"
              label="ノード"
              single-line
              hide-details
              @change="$selectNode"
            ></v-select>
          </v-col>
          <v-col align-self="end">
            <v-btn icon @click="$fetch()">
              <v-icon>mdi-cached</v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-alert :value="deleteError" type="error" dense dismissible>
        ノードを削除できませんでした
      </v-alert>
      <div id="map"></div>
    </v-card>
    <v-dialog v-model="deleteDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード削除</span>
        </v-card-title>
        <v-card-text> 選択したノードを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="deleteDialog = false">キャンセル</v-btn>
          <v-btn color="error" @click="doDeleteNode">削除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.map = await this.$axios.$get('/api/map')
    this.$setNodes(this.map.Nodes)
    this.$setLines(this.map.Lines)
    Object.keys(this.map.Nodes).forEach((k) => {
      this.nodeList.push({
        text: this.map.Nodes[k].Name,
        value: k,
      })
    })
  },
  data() {
    return {
      nodeDialog: false,
      lineDialog: false,
      deleteDialog: false,
      deleteError: false,
      lineError: false,
      selectedNodeID: '',
      showNode: {},
      editLine: {},
      nodeList: [],
      deleteNodes: [],
      map: {
        Nodes: {},
        Lines: [],
        MapConf: { MapName: '' },
      },
    }
  },
  mounted() {
    this.$setIconCodeMap(this.$iconList)
    this.$setStateColorMap(this.$stateList)
    this.$setCallback(this.callback)
    this.$showMAP('map')
  },
  methods: {
    callback(r) {
      switch (r.Cmd) {
        case 'updateNodesPos':
          this.$axios.post('/api/map/update', r.Param)
          break
        case 'deleteNodes':
          this.deleteNodes = Array.from(r.Param)
          this.deleteDialog = true
          break
      }
    },
    doDeleteNode() {
      this.$axios.post('/api/map/delete', this.deleteNodes).then(() => {
        this.$fetch()
      })
      this.deleteDialog = false
    },
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
