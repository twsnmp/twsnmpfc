<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Windowsタスク
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="task"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        class="log"
      >
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.computer" label="Computer">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.task" label="Task"> </v-text-field>
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
          name="TWSNMP_FC_Windows_Task_List.csv"
          header="TWSNMP FCで作成したWindowsタスクリスト"
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
          name="TWSNMP_FC_Windows_Task_List.xls"
          header="TWSNMP FCで作成したWindowsタスクリスト"
          worksheet="Windowsタスク"
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
          <span class="headline">タスク情報</span>
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
                <td>コンピュータ</td>
                <td>{{ selected.Computer }}</td>
              </tr>
              <tr>
                <td>タスク</td>
                <td>{{ selected.TaskName }}</td>
              </tr>
              <tr>
                <td>操作アカウント</td>
                <td>{{ selected.Subject }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
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
          Windowsタスク（グラフ分析）
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
        <v-card-title> Windowsタスク（3Dグラフ） </v-card-title>
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
          text: 'コンピュータ',
          value: 'Computer',
          width: '15%',
          filter: (value) => {
            if (!this.conf.computer) return true
            return value.includes(this.conf.computer)
          },
        },
        {
          text: 'タスク',
          value: 'TaskName',
          width: '35%',
          filter: (value) => {
            if (!this.conf.task) return true
            return value.includes(this.conf.task)
          },
        },
        {
          text: '操作アカウント',
          value: 'Subject',
          width: '20%',
          filter: (value) => {
            if (!this.conf.subject) return true
            return value.includes(this.conf.subject)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      task: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      graphDialog: false,
      scatter3DChartDialog: false,
      conf: {
        computer: '',
        task: '',
        subject: '',
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
    this.task = await this.$axios.$get('/api/report/WinTask')
    if (!this.task) {
      return
    }
    this.task.forEach((e) => {
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
    const c = this.$store.state.report.twwinlog.winTask
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twwinlog/setWinTask', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/WinTask/' + this.selected.ID)
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
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
    openScatter3DChart() {
      this.scatter3DChartDialog = true
      this.$nextTick(() => {
        this.$showWinTaskScatter3DChart('scatter3DChart', this.task, this.conf)
      })
    },
    openGraph() {
      this.graphDialog = true
      this.$nextTick(() => {
        this.updateGraph()
      })
    },
    updateGraph() {
      this.$showWinTaskGraph('graphChart', this.task, this.conf, this.graphType)
    },
    makeExports() {
      const exports = []
      this.task.forEach((e) => {
        if (!this.$filterWinTask(e, this.conf)) {
          return
        }
        exports.push({
          コンピュータ: e.Computer,
          タスク名: e.TaskName,
          操作アカウント: e.Subject,
          回数: e.Count,
          初回日時: e.First,
          最終日時: e.Last,
        })
      })
      return exports
    },
  },
}
</script>
