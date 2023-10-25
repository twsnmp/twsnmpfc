<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        RADIUS通信
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="radius"
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
        <template #[`item.Request`]="{ item }">
          {{ formatCount(item.Request) }}
        </template>
        <template #[`item.Challenge`]="{ item }">
          {{ formatCount(item.Challenge) }}
        </template>
        <template #[`item.Accept`]="{ item }">
          {{ formatCount(item.Accept) }}
        </template>
        <template #[`item.Reject`]="{ item }">
          {{ formatCount(item.Reject) }}
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
            <td colspan="3"></td>
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
            <v-list-item @click="openRADIUSFlowsChart">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>グラフ分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openRADIUSBarChart('Client')">
              <v-list-item-icon>
                <v-icon>mdi-earth</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>クライアント別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openRADIUSBarChart('Server')">
              <v-list-item-icon>
                <v-icon>mdi-earth</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>サーバー別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openRADIUSBarChart('ClientToServer')">
              <v-list-item-icon>
                <v-icon>mdi-earth</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>クライアント/サーバー間</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_RADIUS_List.csv"
          header="TWSNMP FCで作成したRADIUS通信リスト"
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
          name="TWSNMP_FC_RADIUS_List.xls"
          header="TWSNMP FCで作成したRADIUS通信リスト"
          worksheet="RADIUS通信"
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
          <span class="headline">RADIUS通信削除</span>
        </v-card-title>
        <v-card-text> 選択したRADIUS通信を削除しますか？ </v-card-text>
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
        <v-card-text>
          RADIUS通信レポートの信用度を再計算しますか？
        </v-card-text>
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
    <v-dialog v-model="infoDialog" persistent max-width="80vw">
      <v-card>
        <v-card-title>
          <span class="headline">RADIUS通信情報</span>
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
                <td>サーバー名</td>
                <td>{{ selected.ServerName }}</td>
              </tr>
              <tr>
                <td>サーバーIPアドレス</td>
                <td>{{ selected.Server }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ formatCount(selected.Count) }}</td>
              </tr>
              <tr>
                <td>成功率</td>
                <td>{{ selected.Rate }}</td>
              </tr>
              <tr>
                <td>Request</td>
                <td>{{ selected.Request }}</td>
              </tr>
              <tr>
                <td>Challenge</td>
                <td>{{ selected.Challenge }}</td>
              </tr>
              <tr>
                <td>Accept</td>
                <td>{{ selected.Accept }}</td>
              </tr>
              <tr>
                <td>Reject</td>
                <td>{{ selected.Reject }}</td>
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
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
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
    <v-dialog v-model="flowsChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          RADIUS通信グラフ分析
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
    <v-dialog v-model="barChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">{{ barChartTitle }}</span>
        </v-card-title>
        <div
          id="barChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="barChartDialog = false">
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
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'クライアント',
          value: 'ClientName',
          width: '15%',
          filter: (value) => {
            if (!this.conf.client) return true
            return value.includes(this.conf.client)
          },
        },
        {
          text: 'サーバー',
          value: 'ServerName',
          width: '15%',
          filter: (value) => {
            if (!this.conf.server) return true
            return value.includes(this.conf.server)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '成功率', value: 'Rate', width: '8%' },
        { text: 'Request', value: 'Request', width: '6%' },
        { text: 'Challenge', value: 'Challenge', width: '6%' },
        { text: 'Accept', value: 'Accept', width: '6%' },
        { text: 'Reject', value: 'Reject', width: '6%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      radius: [],
      conf: {
        client: '',
        server: '',
        country: '',
        service: '',
        version: '',
        cipher: '',
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      selected: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      infoDialog: false,
      flowsChartDialog: false,
      barChartDialog: false,
      barChartTitle: '',
      graphType: 'force',
      graphTypeList: [
        { text: '力学モデル', value: 'force' },
        { text: '円形', value: 'circular' },
      ],
    }
  },
  async fetch() {
    this.radius = await this.$axios.$get('/api/report/radius')
    if (!this.radius) {
      return
    }
    this.radius.forEach((r) => {
      r.First = this.$timeFormat(
        new Date(r.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      r.Last = this.$timeFormat(
        new Date(r.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const t = r.Accept * 1 + r.Reject * 1
      r.Rate = t ? (100.0 * r.Accept) / t : 0.0
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
    const c = this.$store.state.report.radius.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/radius/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/radius/' + this.selected.ID)
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
        .post('/api/report/radius/reset', {})
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
      this.infoDialog = true
    },
    openRADIUSFlowsChart() {
      this.flowsChartDialog = true
      this.$nextTick(() => {
        this.updateFlowsChart()
      })
    },
    updateFlowsChart() {
      this.$showRADIUSFlowsChart(
        'flowsChart',
        this.getFilterList(),
        this.graphType
      )
    },
    openRADIUSBarChart(type) {
      switch (type) {
        case 'Server':
          this.barChartTitle = 'RADIUSサーバー別'
          break
        case 'Client':
          this.barChartTitle = 'RADIUSクライアント別'
          break
        case 'ClientToServer':
          this.barChartTitle = 'RADIUSクライアント/サーバー間'
          break
        default:
          return
      }
      this.barChartDialog = true
      this.$nextTick(() => {
        this.$showRADIUSBarChart('barChart', type, this.getFilterList())
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    getFilterList() {
      const list = []
      this.radius.forEach((r) => {
        if (!this.filterRADIUS(r)) {
          return
        }
        list.push(r)
      })
      return list
    },
    filterRADIUS(r) {
      if (this.conf.client && !r.ClientName.includes(this.conf.client)) {
        return false
      }
      if (this.conf.server && !r.ServerName.includes(this.conf.server)) {
        return false
      }
      return true
    },
    makeExports() {
      const exports = []
      this.radius.forEach((r) => {
        if (!this.filterRADIUS(r)) {
          return
        }
        exports.push({
          クライアント名: r.ClientName,
          クライアントIP: r.Client,
          サーバー名: r.ServerName,
          サーバーIP: r.Server,
          回数: r.Count,
          成功率: r.Rate,
          Request: r.Request,
          Challenge: r.Challenge,
          Accept: r.Accept,
          Reject: r.Reject,
          初回日時: r.First,
          最終日時: r.Last,
        })
      })
      return exports
    },
  },
}
</script>
