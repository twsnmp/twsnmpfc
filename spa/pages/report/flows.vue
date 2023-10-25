<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        フロー
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="flows"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.Score)">{{
            $getScoreIconName(item.Score)
          }}</v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template #[`item.Count`]="{ item }">
          {{ formatCount(item.Count) }}
        </template>
        <template #[`item.Bytes`]="{ item }">
          {{ formatBytes(item.Bytes) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            v-if="item.ClientNodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.ClientNodeID })"
          >
            mdi-lan
          </v-icon>
          <v-icon
            v-if="item.ServerNodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.ServerNodeID })"
          >
            mdi-server
          </v-icon>
          <v-icon
            small
            @click="$router.push({ path: '/report/address/' + item.Server })"
          >
            mdi-file-find
          </v-icon>
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.client" label="client"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.server" label="server"></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.country"
                label="country"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.service"
                label="service"
              ></v-text-field>
            </td>
            <td colspan="5"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="openFlowsChart">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>グラフ分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openFlows3DChart">
              <v-list-item-icon>
                <v-icon>mdi-earth</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>地球儀</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openCountryChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>国別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openUnknownPortDialog">
              <v-list-item-icon>
                <v-icon>mdi-progress-question</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>不明ポート</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Flow_List.csv"
          header="TWSNMP FCで作成した通信フローリスト"
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
          name="TWSNMP_FC_Flow_List.xls"
          header="TWSNMP FCで作成した通信フローリスト"
          worksheet="通信フロー"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="error" dark @click="resetDialog = true">
          <v-icon>mdi-calculator</v-icon>
          再計算
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">フロー削除</span>
        </v-card-title>
        <v-card-text> フロー{{ selected.Name }}を削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDelete">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="resetDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">信用度再計算</span>
        </v-card-title>
        <v-card-text> フローレポートの信用度を再計算しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doReset">
            <v-icon>mdi-calculator</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="resetDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="flowsChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          通信フロー（グラフ分析）
          <v-spacer></v-spacer>
          <v-select
            v-model="graphType"
            :items="graphTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateFlowsChart"
          ></v-select>
        </v-card-title>
        <v-alert v-model="over" color="error" dense dismissible>
          対象の通信フローが多すぎます。フィルターしてください。
        </v-alert>
        <div
          id="flowsChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="flowsChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="flows3DChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">通信フロー（位置情報）</span>
        </v-card-title>
        <div
          id="flows3DChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="flows3DChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="countryChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">国別</span>
        </v-card-title>
        <div
          id="countryChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="countryChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="infoDialog" persistent max-width="80vw">
      <v-card>
        <v-card-title>
          <span class="headline">フロー情報</span>
        </v-card-title>
        <v-simple-table dense>
          <template #default>
            <thead>
              <tr>
                <th class="text-left">項目</th>
                <th class="text-left">値</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>クライアント名</td>
                <td>{{ selected.ClientName }}</td>
              </tr>
              <tr>
                <td>クライアントIPアドレス</td>
                <td>{{ selected.Client }}</td>
              </tr>
              <tr>
                <td>クライアント位置</td>
                <td>
                  {{ selected.ClientLocInfo }}
                  <v-btn
                    v-if="selected.ClientLatLong"
                    icon
                    dark
                    @click="showGoogleMap(selected.ClientLatLong)"
                  >
                    <v-icon color="grey">mdi-google-maps</v-icon>
                  </v-btn>
                </td>
              </tr>
              <tr>
                <td>サーバー名</td>
                <td>{{ selected.ServerName }}</td>
              </tr>
              <tr>
                <td>サーバーIPアドレス</td>
                <td>{{ selected.Server }}</td>
              </tr>
              <tr>
                <td>サーバー位置</td>
                <td>
                  {{ selected.ServerLocInfo }}
                  <v-btn
                    v-if="selected.ServerLatLong"
                    icon
                    dark
                    @click="showGoogleMap(selected.ServerLatLong)"
                  >
                    <v-icon color="grey">mdi-google-maps</v-icon>
                  </v-btn>
                </td>
              </tr>
              <tr>
                <td>信用スコア</td>
                <td>{{ selected.Score }}</td>
              </tr>
              <tr>
                <td>ペナリティー</td>
                <td>{{ selected.Penalty }}</td>
              </tr>
              <tr>
                <td>初回日時</td>
                <td>{{ selected.First }}</td>
              </tr>
              <tr>
                <td>最終日時</td>
                <td>{{ selected.Last }}</td>
              </tr>
              <tr>
                <td>記録回数</td>
                <td>{{ formatCount(selected.Count) }}</td>
              </tr>
              <tr>
                <td>通信量</td>
                <td>{{ formatBytes(selected.Bytes) }}</td>
              </tr>
              <tr>
                <td>サービス数</td>
                <td>{{ selected.ServiceCount }}</td>
              </tr>
              <tr>
                <td>サービス詳細</td>
                <td>
                  <v-virtual-scroll
                    height="200"
                    item-height="20"
                    :items="selected.ServiceList"
                  >
                    <template #default="{ item }">
                      <v-list-item>
                        <v-list-item-title>{{ item.title }}</v-list-item-title>
                        {{ formatCount(item.value) }}
                      </v-list-item>
                    </template>
                  </v-virtual-scroll>
                </td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="showServicePieChart">
            <v-icon>mdi-chart-pie</v-icon>
            サービス割合
          </v-btn>
          <v-btn color="error" dark @click="doDelete">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="servicePieChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">サービス割合</span>
        </v-card-title>
        <div
          id="servicePieChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="servicePieChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="unknownPortDialog" persistent max-width="60vw">
      <v-card>
        <v-card-title>
          <span class="headline">不明ポート一覧</span>
        </v-card-title>
        <v-alert v-model="unknownPortError" color="error" dense dismissible>
          不明ポートリストを取得できません。
        </v-alert>
        <v-card-text>
          <v-data-table
            :headers="unknownPortsHeaders"
            :items="unknownPorts"
            sort-by="Count"
            sort-asec
            dense
          >
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="unknownPortDialog = false">
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
      search: '',
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'クライアント',
          value: 'ClientName',
          width: '16%',
          filter: (value) => {
            if (!this.conf.client) return true
            return value.includes(this.conf.client)
          },
        },
        {
          text: 'サーバー',
          value: 'ServerName',
          width: '16%',
          filter: (value) => {
            if (!this.conf.server) return true
            return value.includes(this.conf.server)
          },
        },
        {
          text: '国',
          value: 'Country',
          width: '8%',
          filter: (value) => {
            if (!this.conf.country) return true
            return value.includes(this.conf.country)
          },
        },
        {
          text: 'サービス',
          value: 'ServiceInfo',
          width: '12%',
          filter: (value) => {
            if (!this.conf.service) return true
            return value.includes(this.conf.service)
          },
          sort: (a, b) => {
            const re = /\d+/
            const aa = a.match(re)
            const ba = b.match(re)
            if (!aa || !ba) return 0
            const an = aa[0] * 1
            const bn = ba[0] * 1
            if (an < bn) return -1
            if (an > bn) return 1
            return 0
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '通信量', value: 'Bytes', width: '8%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      flows: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      infoDialog: false,
      servicePieChartDialog: false,
      flowsChartDialog: false,
      flows3DChartDialog: false,
      countryChartDialog: false,
      over: false,
      unknownPortsHeaders: [
        { text: '名前', value: 'Name', width: '60%' },
        { text: '回数', value: 'Count', width: '40%' },
      ],
      unknownPorts: [],
      unknownPortDialog: false,
      unknownPortError: false,
      conf: {
        client: '',
        server: '',
        country: '',
        service: '',
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      graphType: 'force',
      graphTypeList: [
        { text: '力学モデル', value: 'force' },
        { text: '円形', value: 'circular' },
      ],
    }
  },
  async fetch() {
    this.flows = await this.$axios.$get('/api/report/flows')
    if (!this.flows) {
      return
    }
    this.flows.forEach((f) => {
      f.First = this.$timeFormat(
        new Date(f.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      f.Last = this.$timeFormat(
        new Date(f.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const sl = Object.keys(f.Services)
      f.ServiceCount = sl.length
      f.ServiceInfo = this.$getServiceInfo(sl)
      let loc = this.$getLocInfo(f.ClientLoc)
      f.ClientLatLong = loc.LatLong
      f.ClientLocInfo = loc.LocInfo
      loc = this.$getLocInfo(f.ServerLoc)
      f.ServerLatLong = loc.LatLong
      f.ServerLocInfo = loc.LocInfo
      f.Country = loc.Country
      f.Loc = f.ServerLoc
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  computed: {
    readOnly() {
      return this.$store.state.map.readOnly
    },
  },
  created() {
    const c = this.$store.state.report.flows.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/flows/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/flow/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    doReset() {
      this.$axios
        .post('/api/report/flows/reset', {})
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.resetError = true
          this.$fetch()
        })
      this.resetDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openInfoDialog(item) {
      this.selected = item
      if (!this.selected.ServiceList) {
        this.selected.ServiceList = []
        Object.keys(this.selected.Services).forEach((k) => {
          this.selected.ServiceList.push({
            title: k,
            value: this.selected.Services[k],
          })
        })
        this.selected.ServiceList.sort((a, b) => {
          if (a.value > b.value) return -1
          if (a.value < b.value) return 1
          return 0
        })
      }
      this.infoDialog = true
    },
    showServicePieChart() {
      if (!this.selected || !this.selected.ServiceList) {
        return
      }
      this.servicePieChartDialog = true
      this.$nextTick(() => {
        this.$showServicePieChart('servicePieChart', this.selected.ServiceList)
      })
    },
    openFlowsChart() {
      this.flowsChartDialog = true
      this.$nextTick(() => {
        this.updateFlowsChart()
      })
    },
    updateFlowsChart() {
      this.over = this.$showFlowsChart(
        'flowsChart',
        this.flows,
        this.conf,
        this.graphType
      )
    },
    openFlows3DChart() {
      this.flows3DChartDialog = true
      this.$nextTick(() => {
        this.$showFlows3DChart('flows3DChart', this.flows, this.conf)
      })
    },
    openCountryChart() {
      this.countryChartDialog = true
      this.$nextTick(() => {
        const list = []
        this.flows.forEach((f) => {
          if (!this.$filterFlow(f, this.conf)) {
            return
          }
          list.push(f)
        })
        this.$showCountryChart('countryChart', list)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
    showGoogleMap(latLong) {
      const url = `https://www.google.com/maps/search/?api=1&query=${latLong}`
      window.open(url, '_blank')
    },
    openUnknownPortDialog() {
      this.unknownPortDialog = true
      this.unknownPortError = false
      this.$axios
        .get('/api/report/unknownport')
        .then((r) => {
          this.unknownPorts = r.data
        })
        .catch((e) => {
          this.unknownPortError = true
        })
    },
    makeExports() {
      const exports = []
      this.flows.forEach((f) => {
        if (!this.$filterFlow(f, this.conf)) {
          return
        }
        exports.push({
          クライアント名: f.ClientName,
          クライアントIP: f.Client,
          クライアント位置: f.ClientLocInfo,
          サーバー名: f.ServerName,
          サーバーIP: f.Server,
          サーバー位置: f.ServerLocInfo,
          記録回数: f.Count,
          通信量: f.Bytes,
          サービス: f.ServiceInfo,
          サービス数: f.ServiceCount,
          信用スコア: f.Score,
          ペナリティー: f.Penalty,
          初回日時: f.First,
          最終日時: f.Last,
        })
      })
      return exports
    },
  },
}
</script>
