<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Wifiアクセスポイント
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="wifiAP"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        class="log"
      >
        <template #[`item.LastRSSI`]="{ item }">
          <v-icon :color="$getRSSIColor(item.LastRSSI)">{{
            $getRSSIIconName(item.LastRSSI)
          }}</v-icon>
          {{ item.LastRSSI }}
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
              <v-text-field v-model="conf.bssid" label="BSSID"> </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ssid" label="SSID"> </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.host" label="Host"></v-text-field>
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
            <v-list-item @click="openRSSITime3DChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>RSSI時間変化 3Dグラフ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openRSSILoc3DChart">
              <v-list-item-icon>
                <v-icon>mdi-home-map-marker</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>RSSI位置変化 3Dグラフ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Wifi_AP_List.csv"
          header="TWSNMP FCで作成したWifiアクセスポイントリスト"
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
          name="TWSNMP_FC_Wifi_AP_List.csv"
          header="TWSNMP FCで作成したWifiアクセスポイントリスト"
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
          name="TWSNMP_FC_Wifi_AP_List.xls"
          header="TWSNMP FCで作成したWifiアクセスポイントリスト"
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
          <span class="headline">レポート削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          アクセスポイントを削除できません
        </v-alert>
        <v-card-text> 選択したアクセスポイントを削除しますか？ </v-card-text>
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
          <span class="headline">Wifiアクセスポイント情報</span>
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
                <td>現在の信号レベル</td>
                <td>
                  <v-icon :color="$getRSSIColor(selected.LastRSSI)">{{
                    $getRSSIIconName(selected.LastRSSI)
                  }}</v-icon>
                  {{ selected.LastRSSI }}
                </td>
              </tr>
              <tr>
                <td>送信元ホスト</td>
                <td>{{ selected.Host }}</td>
              </tr>
              <tr>
                <td>BSSID</td>
                <td>{{ selected.BSSID }}</td>
              </tr>
              <tr>
                <td>SSID</td>
                <td>{{ selected.SSID }}</td>
              </tr>
              <tr>
                <td>チャネル</td>
                <td>{{ selected.Channel }}</td>
              </tr>
              <tr>
                <td>ベンダー</td>
                <td>{{ selected.Vendor }}</td>
              </tr>
              <tr>
                <td>情報</td>
                <td>{{ selected.Info }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>変化数</td>
                <td>{{ selected.Change }}</td>
              </tr>
              <tr>
                <td>RSSI測定数</td>
                <td>{{ selected.RSSI.length }}</td>
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
    <v-dialog v-model="rssiTime3DDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">Wifi AP RSSI時間変化 3Dグラフ</span>
        </v-card-title>
        <div
          id="rssiTime3DChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="rssiTime3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="rssiLoc3DDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">Wifi AP RSSI位置変化 3Dグラフ</span>
        </v-card-title>
        <div
          id="rssiLoc3DChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="rssiLoc3DDialog = false">
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
        { text: 'RSSI', value: 'LastRSSI', width: '10%' },
        {
          text: 'BSSID',
          value: 'BSSID',
          width: '15%',
          filter: (value) => {
            if (!this.conf.bssid) return true
            return value.includes(this.conf.bssid)
          },
        },
        {
          text: 'SSID',
          value: 'SSID',
          width: '20%',
          filter: (value) => {
            if (!this.conf.ssid) return true
            return value.includes(this.conf.ssid)
          },
        },
        {
          text: '送信元ホスト',
          value: 'Host',
          width: '20%',
          filter: (value) => {
            if (!this.conf.host) return true
            return value.includes(this.conf.host)
          },
        },
        { text: '回数', value: 'Count', width: '10%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      wifiAP: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      conf: {
        host: '',
        bsid: '',
        ssid: '',
        sortBy: 'LastRSSI',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      rssiTime3DDialog: false,
      rssiLoc3DDialog: false,
    }
  },
  async fetch() {
    this.wifiAP = await this.$axios.$get('/api/report/WifiAP')
    if (!this.wifiAP) {
      return
    }
    this.wifiAP.forEach((e) => {
      e.First = this.$timeFormat(
        new Date(e.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      e.LastRSSI =
        e.RSSI && e.RSSI.length > 0 ? e.RSSI[e.RSSI.length - 1].Value * 1 : 0
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
    const c = this.$store.state.report.twsensor.wifiAP
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twsensor/setWifiAP', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/WifiAP/' + this.selected.ID)
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
    openRSSITime3DChart() {
      this.rssiTime3DDialog = true
      this.$nextTick(() => {
        this.$showRSSITime3DChart(
          'rssiTime3DChart',
          true,
          this.wifiAP,
          this.conf
        )
      })
    },
    openRSSILoc3DChart() {
      this.rssiLoc3DDialog = true
      this.$nextTick(() => {
        this.$showRSSILoc3DChart('rssiLoc3DChart', true, this.wifiAP, this.conf)
      })
    },
    makeExports() {
      const exports = []
      this.wifiAP.forEach((d) => {
        if (!this.$filterWifiAP(d, this.conf)) {
          return
        }
        exports.push({
          SSID: d.SSID,
          BSSID: d.BSSID,
          送信元ホスト: d.Host,
          チャネル: d.Channel,
          信号レベル: d.LastRSSI,
          ベンダー: d.Vendor,
          情報: d.Info,
          回数: d.Count,
          初回日時: d.First,
          最終日時: d.Last,
        })
      })
      return exports
    },
  },
}
</script>
