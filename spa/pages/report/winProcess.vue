<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Windowsプロセス
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="process"
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
              <v-text-field v-model="conf.process" label="Process">
              </v-text-field>
            </td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :data="process"
          type="csv"
          name="TWSNMP_FC_Windows_Process_List.csv"
          header="TWSNMP FC Windows Process List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="process"
          type="xls"
          name="TWSNMP_FC_Windows_Process_List.xls"
          header="TWSNMP FC Windows Process List"
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
          <span class="headline">プロセス情報</span>
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
                <td>プロセス</td>
                <td>{{ selected.Process }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>起動</td>
                <td>{{ selected.Start }}</td>
              </tr>
              <tr>
                <td>停止</td>
                <td>{{ selected.Exit }}</td>
              </tr>
              <tr>
                <td>親プロセス</td>
                <td>{{ selected.LastParent }}</td>
              </tr>
              <tr>
                <td>関連アカウント</td>
                <td>{{ selected.LastSubject }}</td>
              </tr>
              <tr>
                <td>最終ステータス</td>
                <td>{{ selected.LastStatus }}</td>
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
          text: 'プロセス',
          value: 'Process',
          width: '35%',
          filter: (value) => {
            if (!this.conf.process) return true
            return value.includes(this.conf.process)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '開始', value: 'Start', width: '8%' },
        { text: '終了', value: 'Exit', width: '8%' },
        { text: '最終', value: 'Last', width: '16%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      process: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      conf: {
        computer: '',
        process: '',
        sortBy: 'Count',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.process = await this.$axios.$get('/api/report/WinProcess')
    if (!this.process) {
      return
    }
    this.process.forEach((e) => {
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
    const c = this.$store.state.report.twwinlog.winProcess
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twwinlog/setWinProcess', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/WinProcess/' + this.selected.ID)
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
  },
}
</script>
