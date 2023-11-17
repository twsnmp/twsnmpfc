<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Windowsログオン
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="logon"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
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
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.target" label="Target"></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.computer"
                label="Computer"
              ></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ip" label="From IP"></v-text-field>
            </td>
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
            <v-list-item @click="openGraph">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>グラフ分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openScatter3DChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>３D集計</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Windows_Logon_List.csv"
          header="TWSNMP FCで作成したWindowsログオン状況リスト"
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
          name="TWSNMP_FC_Windows_Logon_List.csv"
          header="TWSNMP FCで作成したWindowsログオン状況リスト"
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
          name="TWSNMP_FC_Windows_Logon_List.xls"
          header="TWSNMP FCで作成したWindowsログオン状況リスト"
          worksheet="Windowsログオン状況"
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
          <span class="headline">レポート削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          ログオン情報を削除できません
        </v-alert>
        <v-card-text> 選択した項目を削除しますか？ </v-card-text>
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
          <span class="headline">信用スコア再計算</span>
        </v-card-title>
        <v-alert v-model="resetError" color="error" dense dismissible>
          信用スコアを再計算できません
        </v-alert>
        <v-card-text> 信用スコアを再計算しますか？ </v-card-text>
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
    <v-dialog v-model="infoDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> ログオン状況 </v-card-title>
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
                <td>ログオン先</td>
                <td>{{ selected.Target }}</td>
              </tr>
              <tr>
                <td>コンピュータ</td>
                <td>{{ selected.Computer }}</td>
              </tr>
              <tr>
                <td>接続元</td>
                <td>{{ selected.IP }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>ログオン回数</td>
                <td>{{ selected.Logon }}</td>
              </tr>
              <tr>
                <td>ログオフ回数</td>
                <td>{{ selected.Logoff }}</td>
              </tr>
              <tr>
                <td>失敗回数</td>
                <td>{{ selected.Failed }}</td>
              </tr>
              <tr>
                <td>ログオン種別</td>
                <td>
                  <v-virtual-scroll
                    height="80"
                    item-height="20"
                    :items="selected.LogonTypeList"
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
              <tr>
                <td>失敗理由</td>
                <td>
                  <v-virtual-scroll
                    height="80"
                    item-height="20"
                    :items="selected.FailedCodeList"
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
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="graphDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          ログオン状況（グラフ分析）
          <v-spacer></v-spacer>
          <v-select
            v-model="graphType"
            :items="graphTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateGraph"
          ></v-select>
        </v-card-title>
        <div
          id="graphChart"
          style="width: 95vw; height: 80vh; overflow: hidden; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="graphDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="scatter3DChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">ログオン状況（3D集計）</span>
        </v-card-title>
        <div
          id="scatter3DChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="scatter3DChartDialog = false">
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
        { text: '信用スコア', value: 'Score', width: '8%' },
        {
          text: 'ログオン先',
          value: 'Target',
          width: '16%',
          filter: (value) => {
            if (!this.conf.target) return true
            return value.includes(this.conf.target)
          },
        },
        {
          text: 'コンピュータ',
          value: 'Computer',
          width: '14%',
          filter: (value) => {
            if (!this.conf.computer) return true
            return value.includes(this.conf.computer)
          },
        },
        {
          text: '接続元',
          value: 'IP',
          width: '14%',
          filter: (value) => {
            if (!this.conf.ip) return true
            return value.includes(this.conf.ip)
          },
        },
        { text: '回数', value: 'Count', width: '6%' },
        { text: 'ログオン', value: 'Logon', width: '7%' },
        { text: 'ログオフ', value: 'Logoff', width: '7%' },
        { text: '失敗', value: 'Failed', width: '6%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      logon: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      resetDialog: false,
      resetError: false,
      scatter3DChartDialog: false,
      graphDialog: false,
      conf: {
        target: '',
        computer: '',
        ip: '',
        sortBy: 'Score',
        sortDesc: true,
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
    this.logon = await this.$axios.$get('/api/report/WinLogon')
    if (!this.logon) {
      return
    }
    this.logon.forEach((e) => {
      e.First = this.$timeFormat(
        new Date(e.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
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
    const c = this.$store.state.report.twwinlog.winLogon
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twwinlog/setWinLogon', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/WinLogon/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
          this.deleteDialog = false
        })
        .catch((e) => {
          this.deleteError = true
        })
    },
    doReset() {
      this.resetError = false
      this.$axios
        .post('/api/report/WinLogon/reset', {})
        .then((r) => {
          this.$fetch()
          this.resetDialog = false
        })
        .catch((e) => {
          this.resetError = true
        })
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openInfoDialog(item) {
      this.selected = item
      if (!this.selected.LogonTypeList) {
        this.selected.LogonTypeList = this.mapToItemList(
          this.selected.LogonType
        )
      }
      if (!this.selected.FailedCodeList) {
        this.selected.FailedCodeList = this.mapToItemList(
          this.selected.FailedCode
        )
      }
      this.infoDialog = true
    },
    mapToItemList(m) {
      const l = []
      Object.keys(m).forEach((k) => {
        l.push({
          title: k,
          value: m[k],
        })
      })
      l.sort((a, b) => {
        if (a.value > b.value) return -1
        if (a.value < b.value) return 1
        return 0
      })
      return l
    },
    openGraph() {
      this.graphDialog = true
      this.$nextTick(() => {
        this.updateGraph()
      })
    },
    updateGraph() {
      this.$showWinLogonGraph(
        'graphChart',
        this.logon,
        this.conf,
        this.graphType
      )
    },
    openScatter3DChart() {
      this.scatter3DChartDialog = true
      this.$nextTick(() => {
        this.$showWinLogonScatter3DChart(
          'scatter3DChart',
          this.logon,
          this.conf
        )
      })
    },
    makeExports() {
      const exports = []
      this.logon.forEach((e) => {
        if (!this.$filterWinLogon(e, this.conf)) {
          return
        }
        exports.push({
          ログオン先: e.Target,
          コンピュータ: e.Computer,
          接続元: e.IP,
          回数: e.Count,
          ログオン: e.Logon,
          ログオフ: e.Logoff,
          失敗: e.Failed,
          信用スコア: e.Score,
          ペナリティー: e.Penalty,
          初回日時: e.First,
          最終日時: e.Last,
        })
      })
      return exports
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
  },
}
</script>
