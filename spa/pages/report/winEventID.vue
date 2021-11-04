<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        WindowsイベントID
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="eventids"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-select v-model="conf.level" :items="levelList" label="Level">
              </v-select>
            </td>
            <td>
              <v-text-field
                v-model="conf.computer"
                label="Computer"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.channel"
                label="Channel"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.provider"
                label="Provider"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.eventID"
                label="EventID"
              ></v-text-field>
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
            <v-list-item @click="openWinEventID3DChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>3D集計</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Windows_EventID_List.csv"
          header="TWSNMP FCで作成したWindowsイベントIDリスト"
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
          name="TWSNMP_FC_Windows_EventID_List.xls"
          header="TWSNMP FCで作成したWindowsイベントIDリスト"
          worksheet="WindowsイベントID"
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
          イベントIDを削除できません
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
          <span class="headline">イベントID情報</span>
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
                <td>レベル</td>
                <td>
                  <v-icon :color="$getStateColor(selected.Level)">
                    {{ $getStateIconName(selected.Level) }}
                  </v-icon>
                  {{ $getStateName(selected.Level) }}
                </td>
              </tr>
              <tr>
                <td>コンピュータ</td>
                <td>{{ selected.Computer }}</td>
              </tr>
              <tr>
                <td>チャネル</td>
                <td>{{ selected.Channel }}</td>
              </tr>
              <tr>
                <td>プロバイダー</td>
                <td>{{ selected.Provider }}</td>
              </tr>
              <tr>
                <td>イベントID</td>
                <td>{{ selected.EventID }}</td>
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
    <v-dialog v-model="winEventIDDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline">WindowsイベントID(3D集計)</span>
        </v-card-title>
        <div id="winEventID3DChart" style="width: 1000px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="winEventIDDialog = false">
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
          text: '状態',
          value: 'Level',
          width: '10%',
          filter: (value) => {
            if (!this.conf.level) return true
            return this.conf.level === value
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
        {
          text: 'チャネル',
          value: 'Channel',
          width: '13%',
          filter: (value) => {
            if (!this.conf.channel) return true
            return value.includes(this.conf.channel)
          },
        },
        {
          text: 'プロバイダー',
          value: 'Provider',
          width: '15%',
          filter: (value) => {
            if (!this.conf.provider) return true
            return value.includes(this.conf.provider)
          },
        },
        {
          text: 'イベントID',
          value: 'EventID',
          width: '12%',
          filter: (value) => {
            if (!this.conf.eventID) return true
            return value.toString(10).includes(this.conf.eventID)
          },
        },
        { text: '回数', value: 'Count', width: '10%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      eventids: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      winEventIDDialog: false,
      conf: {
        level: '',
        computer: '',
        channel: '',
        provider: '',
        eventID: '',
        sortBy: 'Count',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      levelList: [
        { text: '', value: '' },
        { text: 'エラー', value: 'error' },
        { text: '注意', value: 'warn' },
        { text: '正常', value: 'normal' },
      ],
    }
  },
  async fetch() {
    this.eventids = await this.$axios.$get('/api/report/WinEventIDs')
    if (!this.eventids) {
      return
    }
    this.eventids.forEach((e) => {
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
    const c = this.$store.state.report.twwinlog.winEventID
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twwinlog/setWinEventID', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/WinEventID/' + this.selected.ID)
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
    openWinEventID3DChart() {
      this.winEventIDDialog = true
      this.$nextTick(() => {
        this.$showWinEventID3DChart(
          'winEventID3DChart',
          this.eventids,
          this.conf
        )
      })
    },
    makeExports() {
      const exports = []
      this.eventids.forEach((e) => {
        if (!this.$filterWinEventID(e, this.conf)) {
          return
        }
        exports.push({
          レベル: this.$getStateName(e.Level),
          コンピュータ: e.Computer,
          チャネル: e.Channel,
          プロバイダー: e.Provider,
          イベントID: e.EventID,
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
