<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> パネル表示 - {{ node.Name }} </v-card-title>
      <v-card-text>
        <div id="vpanel" style="width: 95%; height: 500px"></div>
        <v-data-table
          :headers="headers"
          :items="ports"
          sort-by="State"
          sort-desc
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
        <v-switch v-model="rotate" label="回転する" @change="setRotate">
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
      power: true,
      rotate: false,
      headers: [
        { text: 'No', value: 'Index', width: '6%' },
        { text: '状態', value: 'State', width: '10%' },
        { text: '名前', value: 'Name', width: '14%' },
        { text: 'Speed', value: 'Speed', width: '10%' },
        { text: 'Tx Pkt', value: 'OutPacktes', width: '10%' },
        { text: 'Tx Bytes', value: 'OutBytes', width: '10%' },
        { text: 'Tx Error', value: 'OutError', width: '10%' },
        { text: 'Rx Pkt', value: 'InPacktes', width: '10%' },
        { text: 'Rx Byte', value: 'InBytes', width: '10%' },
        { text: 'Rx Error', value: 'InError', width: '10%' },
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
    this.ports = r.Ports
    this.power = r.Power
    this.$setVPanel(this.ports, this.power, this.rotate)
  },
  mounted() {
    this.$makeVPanel('vpanel')
  },
  methods: {
    setRotate() {
      this.$setVPanel(this.ports, this.power, this.rotate)
    },
  },
}
</script>
