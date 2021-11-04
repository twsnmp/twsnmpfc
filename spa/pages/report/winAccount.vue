<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Windowsアカウント
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="account"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.subject" label="Subject">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.target" label="Target">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.computer" label="Computer">
              </v-text-field>
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
                <v-list-item-title>３Dグラフ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Windows_Account_List.csv"
          header="TWSNMP FCで作成したWindowsアカウント操作リスト"
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
          name="TWSNMP_FC_Windows_Account_List.xls"
          header="TWSNMP FCで作成したWindowsアカウント操作リスト"
          worksheet="Windowsアカウント操作"
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
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">レポート削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          アカウントを削除できません
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
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">アカウント情報</span>
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
                <td>操作アカウント</td>
                <td>{{ selected.Subject }}</td>
              </tr>
              <tr>
                <td>対象アカウント</td>
                <td>{{ selected.Target }}</td>
              </tr>
              <tr>
                <td>コンピュータ</td>
                <td>{{ selected.Computer }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>編集</td>
                <td>{{ selected.Edit }}</td>
              </tr>
              <tr>
                <td>パスワード変更</td>
                <td>{{ selected.Password }}</td>
              </tr>
              <tr>
                <td>その他</td>
                <td>{{ selected.Other }}</td>
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
    <v-dialog v-model="graphDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          アカウント（グラフ分析）
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
        <div id="graphChart" style="width: 1000px; height: 700px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="graphDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="scatter3DChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline">アカウント（3D集計）</span>
        </v-card-title>
        <div id="scatter3DChart" style="width: 1000px; height: 700px"></div>
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
export default {
  data() {
    return {
      headers: [
        {
          text: '操作アカウント',
          value: 'Subject',
          width: '15%',
          filter: (value) => {
            if (!this.conf.subject) return true
            return value.includes(this.conf.subject)
          },
        },
        {
          text: '対象アカウント',
          value: 'Target',
          width: '15%',
          filter: (value) => {
            if (!this.conf.target) return true
            return value.includes(this.conf.target)
          },
        },
        {
          text: 'コンピュータ',
          value: 'Computer',
          width: '15%',
          filter: (value) => {
            if (!this.conf.computer) return true
            return value.includes(this.conf.computer)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '編集', value: 'Edit', width: '8%' },
        { text: 'パスワード', value: 'Password', width: '8%' },
        { text: 'その他', value: 'Other', width: '8%' },
        { text: '最終', value: 'Last', width: '13%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      account: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      graphDialog: false,
      scatter3DChartDialog: false,
      conf: {
        subject: '',
        target: '',
        computer: '',
        sortBy: 'Count',
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
    this.account = await this.$axios.$get('/api/report/WinAccount')
    if (!this.account) {
      return
    }
    this.account.forEach((e) => {
      e.First = this.$timeFormat(
        new Date(e.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.twwinlog.winAccount
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twwinlog/setWinAccount', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/WinAccount/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
          this.deleteDialog = false
        })
        .catch((e) => {
          this.deleteError = true
        })
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
    openScatter3DChart() {
      this.scatter3DChartDialog = true
      this.$nextTick(() => {
        this.$showWinAccountScatter3DChart(
          'scatter3DChart',
          this.account,
          this.conf
        )
      })
    },
    openGraph() {
      this.graphDialog = true
      this.$nextTick(() => {
        this.updateGraph()
      })
    },
    updateGraph() {
      this.$showWinAccountGraph(
        'graphChart',
        this.account,
        this.conf,
        this.graphType
      )
    },
    makeExports() {
      const exports = []
      this.account.forEach((e) => {
        if (!this.$filterWinAccount(e, this.conf)) {
          return
        }
        exports.push({
          操作したアカウント: e.Subject,
          対象アカウント: e.Target,
          コンピュータ: e.Computer,
          回数: e.Count,
          編集: e.Edit,
          パスワード変更: e.Password,
          その他: e.Other,
          初回日時: e.First,
          最終日時: e.Last,
        })
      })
      return exports
    },
  },
}
</script>
