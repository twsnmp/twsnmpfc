<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> パネル表示 - {{ node.Name }} </v-card-title>
      <v-card-text>
        <div id="vpanel" style="width: 95%; height: 500px"></div>
        <v-data-table
          :headers="headers"
          :items="ports"
          sort-by="No"
          sort-asc
          dense
          :loading="$fetchState.pending"
          loading-text="Loading... Please wait"
          class="log"
        >
          <template #[`item.State`]="{ item }">
            <v-icon :color="$getStateColor(item.State)">{{
              $getStateIconName(item.State)
            }}</v-icon>
            {{ $getStateName(item.State) }}
          </template>
        </v-data-table>
      </v-card-text>
      <v-card-actions>
        <v-switch v-model="rotate" label="回転する" @change="showVPanel">
        </v-switch>
        <v-spacer></v-spacer>
        <v-switch
          v-model="showInternal"
          label="内部ポート表示"
          @change="showVPanel"
        >
        </v-switch>
        <v-spacer></v-spacer>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      node: {},
      ports: [],
      allPorts: [],
      power: true,
      rotate: false,
      showInternal: false,
      headers: [
        { text: 'No', value: 'No', width: '6%' },
        { text: 'Index', value: 'Index', width: '8%' },
        { text: '状態', value: 'State', width: '10%' },
        { text: '名前', value: 'Name', width: '14%' },
        { text: 'Speed', value: 'Speed', width: '10%' },
        { text: 'Tx Pkt', value: 'OutPacktes', width: '8%' },
        { text: 'Tx Bytes', value: 'OutBytes', width: '8%' },
        { text: 'Tx Error', value: 'OutError', width: '8%' },
        { text: 'Rx Pkt', value: 'InPacktes', width: '8%' },
        { text: 'Rx Byte', value: 'InBytes', width: '8%' },
        { text: 'Rx Error', value: 'InError', width: '8%' },
      ],
    }
  },
  async fetch() {
    const r = await this.$axios.$get(
      '/api/node/vpanel/' + this.$route.params.id
    )
    if (!r || !r.Node || !r.Ports) {
      return
    }
    this.node = r.Node
    this.allPorts = r.Ports
    this.power = r.Power
    this.showVPanel()
  },
  mounted() {
    this.$makeVPanel('vpanel')
  },
  methods: {
    showVPanel() {
      this.ports = []
      let i = 1
      this.allPorts.forEach((p) => {
        if (this.showInternal || p.Type === 6) {
          p.No = i++
          this.ports.push(p)
        }
      })
      this.$setVPanel(this.ports, this.power, this.rotate)
    },
  },
}
</script>
