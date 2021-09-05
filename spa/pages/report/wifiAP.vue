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
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
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
        <download-excel
          :data="wifiAP"
          type="csv"
          name="TWSNMP_FC_Wifi_AP_List.csv"
          header="TWSNMP FC Wifi AP List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="wifiAP"
          type="xls"
          name="TWSNMP_FC_Wifi_AP_List.xls"
          header="TWSNMP FC Wifi AP List"
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
                <td>センサー</td>
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
                <td>情報</td>
                <td>{{ selected.Info }}</td>
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
    }
  },
  async fetch() {
    this.wifiAP = await this.$axios.$get('/api/report/WifiAP')
    if (!this.wifiAP) {
      return
    }
    this.wifiAP.forEach((e) => {
      e.First = this.$timeFormat(
        new Date(e.FirstTime * 1000),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime * 1000),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.LastRSSI =
        e.RSSI && e.RSSI.length > 0 ? e.RSSI[e.RSSI.length - 1].Value * 1 : 0
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
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
      this.$axios
        .delete('/api/report/WifiAP/' + this.selected.ID)
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