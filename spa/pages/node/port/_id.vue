<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> ポートリスト - {{ node.Name }} </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab">
          <v-tab key="tcp">TCP</v-tab>
          <v-tab key="udp">UDP</v-tab>
        </v-tabs>
        <v-tabs-items v-model="tab">
          <v-tab-item key="tcp">
            <v-data-table
              :headers="portHeaders"
              :items="tcpPorts"
              sort-by="Port"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.actions`]="{ item }">
                <v-icon
                  v-if="item.Polling"
                  small
                  @click="editSystemPolling(item.Polling)"
                >
                  mdi-card-plus
                </v-icon>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="udp">
            <v-data-table
              :headers="portHeaders"
              :items="udpPorts"
              sort-by="Port"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.actions`]="{ item }">
                <v-icon
                  v-if="item.Polling"
                  small
                  @click="editSystemPolling(item.Polling)"
                >
                  mdi-card-plus
                </v-icon>
              </template>
            </v-data-table>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          :name="'TWSNMP_FC' + exportSheet + '.csv'"
          :header="exportTitle"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          :name="'TWSNMP_FC' + exportSheet + '.xls'"
          :header="exportTitle"
          :worksheet="exportSheet"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
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
      tab: 0,
      tcpPorts: [],
      udpPorts: [],
      portHeaders: [
        { text: 'ポート番号', value: 'Port', width: '10%' },
        { text: 'アドレス', value: 'Address', width: '20%' },
        { text: 'プロセス', value: 'Process', width: '20%' },
        { text: '情報', value: 'Descr', width: '50%' },
      ],
      exportTitle: '',
      exportSheet: '',
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/node/port/' + this.$route.params.id)
    if (!r || !r.Node) {
      return
    }
    this.node = r.Node
    this.tcpPorts = r.TcpPorts
    this.udpPorts = r.UdpPorts
  },
  methods: {
    makeExports() {
      const exports = []
      if (this.tab === 0) {
        this.exportTitle = this.node.Name + 'のTCPポート'
        this.exportSheet = '_TCP_Port'
        this.tcpPorts.forEach((p) => {
          exports.push({
            ポート番号: p.Port,
            アドレス: p.Address,
            プロセス: p.Process,
            情報: p.Descr,
          })
        })
      } else {
        this.exportTitle = this.node.Name + 'のUDPポート'
        this.exportSheet = '_UDP_Port'
        this.udpPorts.forEach((p) => {
          exports.push({
            ポート番号: p.Port,
            アドレス: p.Address,
            プロセス: p.Process,
            情報: p.Descr,
          })
        })
      }
      return exports
    },
  },
}
</script>
