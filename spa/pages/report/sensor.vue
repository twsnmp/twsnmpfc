<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        センサー
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        センサーの削除に失敗しました
      </v-alert>
      <v-alert v-model="toggleError" color="error" dense dismissible>
        センサーの設定に失敗しました
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="sensors"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        sort-by="Name"
        :items-per-page="20"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template #[`item.Total`]="{ item }">
          {{ formatCount(item.Total) }}
        </template>
        <template #[`item.Send`]="{ item }">
          {{ formatCount(item.Send) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
          <v-icon v-if="item.Ignore" small @click="toggleSensor(item.ID)">
            mdi-play
          </v-icon>
          <v-icon v-if="!item.Ignore" small @click="toggleSensor(item.ID)">
            mdi-stop
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="filter.host" label="Host"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="filter.type" label="Type"></v-text-field>
            </td>
            <td colspan="6"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Sensor_List.csv"
          header="TWSNMP FCのセンサーリスト"
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
          name="TWSNMP_FC_Sensor_List.csv"
          header="TWSNMP FCのセンサーリスト"
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
          name="TWSNMP_FC_Sensor_List.xls"
          header="TWSNMP FCのセンサーリスト"
          worksheet="センサー"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="primary" dark @click="showTree">
          <v-icon>mdi-shuffle-disabled</v-icon>
          ツリー
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
          <span class="headline">センサー削除</span>
        </v-card-title>
        <v-card-text> 選択したセンサーを削除しますか？ </v-card-text>
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
    <v-dialog v-model="infoDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">センサー情報</span>
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
                <td>送信元</td>
                <td>{{ selected.Host }}</td>
              </tr>
              <tr>
                <td>種別</td>
                <td>{{ selected.Type }}</td>
              </tr>
              <tr>
                <td>パラメータ</td>
                <td>{{ selected.Param }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ formatCount(selected.Total) }}</td>
              </tr>
              <tr>
                <td>送信数</td>
                <td>{{ formatCount(selected.Send) }}</td>
              </tr>
              <tr>
                <td>統計履歴数</td>
                <td>{{ formatCount(selected.StatsLen) }}</td>
              </tr>
              <tr>
                <td>リソースモニタ数</td>
                <td>{{ formatCount(selected.MonitorsLen) }}</td>
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
          <v-menu
            v-if="selected.MonitorsLen > 0 || selected.StatsLen > 0"
            offset-y
          >
            <template #activator="{ on, attrs }">
              <v-btn color="primary" dark v-bind="attrs" v-on="on">
                <v-icon>mdi-chart-line</v-icon>
                グラフ表示
              </v-btn>
            </template>
            <v-list>
              <v-list-item v-if="selected.StatsLen > 0" @click="openStatsChart">
                <v-list-item-icon>
                  <v-icon>mdi-chart-bar</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>統計情報</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-if="selected.MonitorsLen > 0"
                @click="openCpuMemChart"
              >
                <v-list-item-icon><v-icon>mdi-gauge</v-icon></v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>CPU/Memory</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-if="selected.MonitorsLen > 0"
                @click="openNetChart"
              >
                <v-list-item-icon><v-icon>mdi-lan</v-icon></v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>通信量</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-if="selected.MonitorsLen > 0"
                @click="openProcChart"
              >
                <v-list-item-icon>
                  <v-icon>mdi-animation</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>プロセス</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-menu>
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="statsChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> 統計情報 </span>
        </v-card-title>
        <div
          id="statsChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="statsChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="cpuMemChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> CPU/Memory </span>
        </v-card-title>
        <div
          id="cpuMemChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="cpuMemChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="netChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> 通信量 </span>
        </v-card-title>
        <div
          id="netChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="netChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="procChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> プロセス数と負荷 </span>
        </v-card-title>
        <div
          id="procChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="procChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="treeDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          センサーツリー
          <v-spacer></v-spacer>
          <v-select
            v-model="sensorType"
            :items="sensorTypeList"
            label="タイプ"
            single-line
            hide-details
            @change="showTree"
          ></v-select>
        </v-card-title>
        <div id="tree" style="width: 95vw; height: 80vh; margin: 0 auto"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="treeDialog = false">
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
        { text: '状態', value: 'State', width: '10%' },
        {
          text: '送信元',
          value: 'Host',
          width: '15%',
          filter: (value) => {
            if (!this.filter.host) return true
            return value.includes(this.filter.host)
          },
        },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.filter.type) return true
            return value.includes(this.filter.type)
          },
        },
        {
          text: 'パラメータ',
          value: 'Param',
          width: '15%',
        },
        { text: '回数', value: 'Total', width: '7%' },
        { text: '送信数', value: 'Send', width: '7%' },
        { text: '初回', value: 'First', width: '13%' },
        { text: '最終', value: 'Last', width: '13%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      sensors: [],
      selected: {},
      filter: {
        host: '',
        type: '',
      },
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      statsChartDialog: false,
      cpuMemChartDialog: false,
      netChartDialog: false,
      procChartDialog: false,
      toggleError: false,
      treeDialog: false,
      sensorType: '',
      sensorTypeList: [],
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/report/sensors')
    if (!r) {
      return
    }
    const typeMap = {}
    this.sensors = r
    this.sensors.forEach((s) => {
      s.First = this.$timeFormat(
        new Date(s.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      s.Last = this.$timeFormat(
        new Date(s.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      typeMap[s.Type] = s.Total
    })
    this.sensorTypeList = []
    Object.keys(typeMap).forEach((k) => {
      if (!this.sensorType || k === 'syslog') {
        this.sensorType = k
      }
      this.sensorTypeList.push({ text: k, value: k })
    })
  },
  computed: {
    readOnly() {
      return this.$store.state.map.readOnly
    },
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/sensor/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    toggleSensor(id) {
      this.$axios
        .post('/api/report/sensor/' + id)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.toggleError = true
          this.$fetch()
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
    async openStatsChart() {
      const r = await this.$axios.$get(
        '/api/report/sensor/stats/' + this.selected.ID
      )
      if (!r) {
        return
      }
      this.statsChartDialog = true
      this.$nextTick(() => {
        this.$showSensorStatsChart('statsChart', r)
      })
    },
    async openCpuMemChart() {
      const r = await this.$axios.$get(
        '/api/report/sensor/monitors/' + this.selected.ID
      )
      if (!r) {
        return
      }
      this.cpuMemChartDialog = true
      this.$nextTick(() => {
        this.$showSensorCpuMemChart('cpuMemChart', r)
      })
    },
    async openNetChart() {
      const r = await this.$axios.$get(
        '/api/report/sensor/monitors/' + this.selected.ID
      )
      if (!r) {
        return
      }
      this.netChartDialog = true
      this.$nextTick(() => {
        this.$showSensorNetChart('netChart', r)
      })
    },
    async openProcChart() {
      const r = await this.$axios.$get(
        '/api/report/sensor/monitors/' + this.selected.ID
      )
      if (!r) {
        return
      }
      this.procChartDialog = true
      this.$nextTick(() => {
        this.$showSensorProcChart('procChart', r)
      })
    },
    showTree() {
      this.treeDialog = true
      this.$nextTick(() => {
        this.$showSensorTree('tree', this.sensorType, this.sensors)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    makeExports() {
      const exports = []
      this.sensors.forEach((e) => {
        if (!this.filterSensor(e)) {
          return
        }
        exports.push({
          送信元: e.Host,
          種別: e.Type,
          パラメータ: e.Param,
          回数: e.Total,
          送信数: e.Send,
          統計履歴数: e.StatsLen,
          リソースモニタ数: e.MonitorsLen,
          初回日時: e.First,
          最終日時: e.Last,
        })
      })
      return exports
    },
    filterSensor(e) {
      if (this.filter.host && !e.Host.includes(this.filter.host)) {
        return false
      }
      if (this.filter.type && !e.Type.includes(this.filter.type)) {
        return false
      }
      return true
    },
  },
}
</script>
