<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> パネル表示 - {{ node.Name }} </v-card-title>
      <v-card-text>
        <div
          id="vpanel"
          style="width: 98vw; height: 40vh; margin: 0 auto"
        ></div>
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
          <template #[`item.Speed`]="{ item }">
            {{ formatSpeed(item.Speed) }}
          </template>
          <template #[`item.OutPacktes`]="{ item }">
            {{ formatCount(item.OutPacktes) }}
          </template>
          <template #[`item.OutBytes`]="{ item }">
            {{ formatBytes(item.OutBytes) }}
          </template>
          <template #[`item.OutError`]="{ item }">
            {{ formatCount(item.OutError) }}
          </template>
          <template #[`item.InPacktes`]="{ item }">
            {{ formatCount(item.InPacktes) }}
          </template>
          <template #[`item.InBytes`]="{ item }">
            {{ formatBytes(item.InBytes) }}
          </template>
          <template #[`item.InError`]="{ item }">
            {{ formatCount(item.InError) }}
          </template>
          <template #[`item.State`]="{ item }">
            <v-icon :color="$getStateColor(item.State)">{{
              $getStateIconName(item.State)
            }}</v-icon>
            {{ $getStateName(item.State) }}
          </template>
          <template #[`item.actions`]="{ item }">
            <v-icon small @click="editIFStatePolling(item)">
              mdi-lan-check
            </v-icon>
            <v-icon small @click="editTrafPolling(item)"> mdi-gauge </v-icon>
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
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="showPortChart('out')">
              <v-list-item-icon>
                <v-icon>mdi-arrow-right-circle</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>送信量</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="showPortChart('in')">
              <v-list-item-icon>
                <v-icon>mdi-arrow-left-circle</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>受信量</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Port.csv"
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
          name="TWSNMP_FC_Port.xls"
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
        <v-btn
          color="normal"
          dark
          @click="$router.push({ path: '/node/polling/' + node.ID })"
        >
          <v-icon>mdi-lan-check</v-icon>
          ポーリング
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title> ポーリング設定 </v-card-title>
        <v-alert v-model="addError" color="error" dense dismissible>
          ポーリングを追加できませんでした
        </v-alert>
        <v-card-text>
          <v-text-field v-model="polling.Name" label="名前"></v-text-field>
          <v-select v-model="polling.Level" :items="$levelList" label="レベル">
          </v-select>
          <v-slider
            v-model="polling.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="3600"
            min="5"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="polling.PollInt"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="polling.Timeout"
            label="タイムアウト(Sec)"
            class="align-center"
            max="60"
            min="1"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="polling.Timeout"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="polling.Retry"
            label="リトライ回数"
            class="align-center"
            max="20"
            min="0"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="polling.Retry"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-select
            v-model="polling.LogMode"
            :items="$logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="addPolling">
            <v-icon>mdi-content-save</v-icon>
            追加
          </v-btn>
          <v-btn color="normal" dark @click="editDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="chartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">{{ chartTitle }}</span>
        </v-card-title>
        <div id="chart" style="width: 95vw; height: 80vh; margin: 0 auto"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="chartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
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
        { text: 'Index', value: 'Index', width: '6%' },
        { text: '状態', value: 'State', width: '8%' },
        { text: '名前', value: 'Name', width: '12%' },
        { text: 'Speed', value: 'Speed', width: '8%' }, // 40
        { text: 'Tx Pkt', value: 'OutPacktes', width: '9%' },
        { text: 'Tx Bytes', value: 'OutBytes', width: '8%' },
        { text: 'Tx Err', value: 'OutError', width: '6%' },
        { text: 'Rx Pkt', value: 'InPacktes', width: '9%' },
        { text: 'Rx Byte', value: 'InBytes', width: '8%' },
        { text: 'Rx Err', value: 'InError', width: '6%' },
        { text: '操作', value: 'actions', width: '8%' },
      ],
      editDialog: false,
      addError: false,
      thValue: '',
      polling: {},
      exportTitle: '',
      exportSheet: '',
      chartTitle: '',
      chartDialog: false,
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
    editIFStatePolling(i) {
      this.polling = {
        ID: '',
        Name: 'インターフェイス監視 ' + i.Index,
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'ifOperStatus',
        Params: i.Index + '',
        Filter: '',
        Extractor: '',
        Script: '',
        Level: 'low',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.editDialog = true
    },
    editTrafPolling(i) {
      this.polling = {
        ID: '',
        Name: 'SNMP通信量測定 ' + i.Index,
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'traffic',
        Params: i.Index + '',
        Filter: '',
        Extractor: '',
        Script: '',
        Level: 'info',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.editDialog = true
    },
    addPolling() {
      const tmpScript = this.polling.Script
      this.addError = false
      if (this.hasTh) {
        this.polling.Script = this.polling.Script.replace(
          '$thValue',
          this.thValue
        )
      }
      this.$axios
        .post('/api/polling/add', this.polling)
        .then(() => {
          this.editDialog = false
        })
        .catch((e) => {
          this.polling.Script = tmpScript
          this.addError = true
        })
    },
    makeExports() {
      const exports = []
      this.exportTitle = this.node.Name + 'のポート情報'
      this.exportSheet = 'ポート情報'
      this.ports.forEach((e) => {
        exports.push({
          No: e.No,
          Index: e.Index,
          状態: e.State,
          名前: e.Name,
          速度: e.Speed,
          送信パケット: e.OutPacktes,
          送信バイト: e.OutBytes,
          送信エラー: e.OutError,
          受信パケット: e.InPacktes,
          受信バイト: e.InBytes,
          受信エラー: e.InError,
        })
      })
      return exports
    },
    showPortChart(dir) {
      if (dir === 'in') {
        this.chartTitle = 'ポート単位の受信バイト数'
      } else {
        this.chartTitle = 'ポート単位の送信バイト数'
      }
      this.chartDialog = true
      let max = 0
      this.$nextTick(() => {
        const list = []
        this.ports.forEach((e) => {
          const v = dir === 'in' ? e.InBytes : e.OutBytes
          if (max < v) {
            max = v
          }
          list.push({
            Name: e.Name + '(' + e.Index + ')',
            Value: v,
          })
        })
        list.sort((a, b) => {
          if (a.Value < b.Value) return -1
          if (a.Value > b.Value) return 1
          return 0
        })
        while (list.length > 50) {
          list.shift()
        }
        this.$showHrBarChart(
          'chart',
          dir === 'in' ? '受信バイト数' : '送信バイト数',
          'Bytes',
          list,
          max
        )
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatSpeed(n) {
      return numeral(n).format('0b') + 'PS'
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
