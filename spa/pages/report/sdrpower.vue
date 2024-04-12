<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        周波数別電波強度(RTL-SDR)
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        選択した電波強度の情報を削除できません
      </v-alert>
      <v-alert v-model="getError" color="error" dense dismissible>
        選択した電波強度の情報を取得できません
      </v-alert>
      <v-data-table
        v-model="selected"
        :headers="headers"
        :items="sdrPowerKeys"
        item-key="ID"
        show-select
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="20"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="filter.host" label="Host"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="filter.time" label="Time"></v-text-field>
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
            <v-list-item @click="openChart('2d')">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>周波数別電波強度(2D)</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openChart('3d')">
              <v-list-item-icon>
                <v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>周波数別電波強度(3D)</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn v-if="hasSelected" color="error" @click="deleteDialog = true">
          <v-icon>mdi-delete</v-icon>
          削除
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
    <v-dialog v-model="chartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          {{ chartTitle }}
          <v-spacer></v-spacer>
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
export default {
  data() {
    return {
      headers: [
        {
          text: 'ホスト',
          value: 'Host',
          width: '60%',
          filter: (value) => {
            if (!this.filter.host) return true
            return value.includes(this.filter.host)
          },
        },
        {
          text: '測定日時',
          value: 'TimeStr',
          width: '45%',
          filter: (value) => {
            if (!this.filter.time) return true
            return value.includes(this.filter.time)
          },
        },
      ],
      sdrPowerKeys: [],
      selected: [],
      deleteDialog: false,
      deleteError: false,
      getError: false,
      chartDialog: false,
      filter: {
        host: '',
        time: '',
      },
      chartTitle: '',
    }
  },
  async fetch() {
    this.sdrPowerKeys = await this.$axios.$get('/api/report/sdrPowerKeys')
    if (!this.sdrPowerKeys) {
      return
    }
    this.sdrPowerKeys.forEach((e) => {
      e.TimeStr = this.$timeFormat(
        new Date(e.Time * 1000),
        '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.ID = e.Host + ':' + e.Time
    })
  },
  computed: {
    hasSelected() {
      return this.selected.length > 0
    },
    readOnly() {
      return this.$store.state.map.readOnly
    },
  },
  methods: {
    doDelete() {
      this.deleteDialog = false
      this.deleteError = false
      const ids = []
      this.selected.forEach((e) => {
        ids.push(e.ID)
      })
      this.$axios
        .post('/api/report/sdrPower/delete', ids)
        .then(() => {
          this.$fetch()
          this.selected = []
        })
        .catch((e) => {
          this.$fetch()
          this.deleteError = true
        })
    },
    openDeleteDialog(item) {
      if (this.selected.length > 0) {
        this.deleteDialog = true
      }
    },
    async openChart(type) {
      this.getError = false
      if (this.selected.length < 1) {
        return
      }
      const ids = []
      this.selected.forEach((e) => {
        ids.push(e.ID)
      })
      const r = await this.$axios.$post('/api/report/sdrPowerData', ids)
      if (!r) {
        this.getError = true
        return
      }
      this.sdrPowerData = r
      this.chartDialog = true
      this.$nextTick(() => {
        if (type === '2d') {
          this.chartTitle = '周波数別電波強度(2D)'
          this.$showSdrPower2DChart('chart', this.sdrPowerData)
        } else {
          this.chartTitle = '周波数別電波強度(3D)'
          this.$showSdrPower3DChart('chart', this.sdrPowerData)
        }
      })
    },
    makeExports() {
      const exports = []
      this.sdrPowerData.forEach((l) => {
        l.forEach((e) => {
          exports.push({
            ホスト名: e.Host,
            日時: this.$timeFormat(
              new Date(e.Time * 1000),
              '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
            ),
            周波数: e.Freq,
            強度: e.Dbm,
          })
        })
      })
      return exports
    },
  },
}
</script>
