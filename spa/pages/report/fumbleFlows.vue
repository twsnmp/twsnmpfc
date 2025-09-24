<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Fumbleフロー
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="fumbleFlows"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.TCPCunt`]="{ item }">
          {{ formatCount(item.TCPCount) }}
        </template>
        <template #[`item.IcmpCount`]="{ item }">
          {{ formatCount(item.IcmpCount) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.src" label="Src"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.dst" label="Dst"></v-text-field>
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
            <v-list-item @click="openFlowChart">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>グラフ分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openIPChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>IPアドレス別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Fumble_Flow_List.csv"
          header="TWSNMP FCで作成したFumbleフローリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_Fumble_Flow_List.csv"
          header="TWSNMP FCで作成したFumbleフローリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_Fumble_Flow_List.xls"
          header="TWSNMP FCで作成したFumbleフローリスト"
          worksheet="通信フロー"
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
      </v-card-actions>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">Fumbleフロー削除</span>
        </v-card-title>
        <v-card-text>
          Fumbleフロー{{ selected.ID }}を削除しますか？
        </v-card-text>
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
    <v-dialog v-model="flowChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          グラフ分析
          <v-spacer></v-spacer>
          <v-select
            v-model="graphType"
            :items="graphTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateFlowChart"
          ></v-select>
        </v-card-title>
        <div
          id="flowChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="flowChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="ipChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">IPアドレス別</span>
        </v-card-title>
        <div
          id="ipChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="ipChartDialog = false">
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
        {
          text: 'Src',
          value: 'Src',
          width: '20%',
          filter: (value) => {
            if (!this.conf.src) return true
            return value.includes(this.conf.src)
          },
        },
        {
          text: 'Dst',
          value: 'Dst',
          width: '20%',
          filter: (value) => {
            if (!this.conf.dst) return true
            return value.includes(this.conf.dst)
          },
        },
        { text: 'TCP', value: 'TCPCount', width: '10%' },
        { text: 'ICMP', value: 'IcmpCount', width: '10%' },
        { text: '初回', value: 'First', width: '15%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      fumbleFlows: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      conf: {
        src: '',
        dst: '',
        sortBy: 'TCPCount',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      flowChartDialog: false,
      ipChartDialog: false,
      graphType: 'force',
      graphTypeList: [
        { text: '力学モデル', value: 'force' },
        { text: '円形', value: 'circular' },
      ],
    }
  },
  async fetch() {
    this.fumbleFlows = await this.$axios.$get('/api/report/fumbleFlows')
    if (!this.fumbleFlows) {
      return
    }
    this.fumbleFlows.forEach((f) => {
      f.First = this.$timeFormat(
        new Date(f.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      f.Last = this.$timeFormat(
        new Date(f.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const a = f.ID.split('_')
      f.Src = a[0] || ''
      f.Dst = a[1] || ''
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
    const c = this.$store.state.report.fumbleFlows.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/fumbleFlows/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/fumbleFlow/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openFlowChart() {
      this.flowChartDialog = true
      this.$nextTick(() => {
        this.updateFlowChart()
      })
    },
    updateFlowChart() {
      this.over = this.$showFumbleFlowChart(
        'flowChart',
        this.fumbleFlows,
        this.conf,
        this.graphType
      )
    },
    openIPChart() {
      this.ipChartDialog = true
      this.$nextTick(() => {
        this.$showFumbleIPChart('ipChart', this.fumbleFlows)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    makeExports() {
      const exports = []
      this.fumbleFlows.forEach((f) => {
        if (!this.$filterFlow(f, this.conf)) {
          return
        }
        exports.push({
          Src: f.Src,
          Dst: f.Dst,
          TCP: f.TCPCount,
          ICMP: f.IcmpCount,
          初回日時: f.First,
          最終日時: f.Last,
        })
      })
      return exports
    },
  },
}
</script>
