<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Bluetoothデバイス
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-alert v-if="deleteError" color="error" dense dismissible>
        デバイスを削除できません
      </v-alert>
      <v-alert v-if="setNameError" color="error" dense dismissible>
        デバイスの名前を変更できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="blueDevice"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        class="log"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.LastRSSI`]="{ item }">
          <v-icon :color="$getRSSIColor(item.LastRSSI)">{{
            $getRSSIIconName(item.LastRSSI)
          }}</v-icon>
          {{ item.LastRSSI }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon v-if="!readOnly" small @click="openEditNameDialog(item)">
            mdi-pencil
          </v-icon>
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.address" label="Address">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.addressType" label="AType">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="Name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.host" label="Host"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.vendor" label="Vendor"></v-text-field>
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
          <v-list>
            <v-list-item @click="openBlueVendorChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>ベンダー別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openBlueScanGraph">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>名前とアドレスの関係</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Bluetooth_Device_List.csv"
          header="TWSNMP FCで作成したBluetoothデバイスリスト"
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
          name="TWSNMP_FC_Bluetooth_Device_List.csv"
          header="TWSNMP FCで作成したBluetoothデバイスリスト"
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
          name="TWSNMP_FC_Bluetooth_Device_List.xls"
          header="TWSNMP FCで作成したBluetoothデバイスリスト"
          worksheet="Bluetoothデバイス"
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
    <v-dialog v-model="setNameDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">名前変更</span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="selected.Name" label="名前"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setName">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" @click="setNameDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="infoDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">Bluetoothデバイス情報</span>
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
                <td>アドレス</td>
                <td>{{ selected.Address }}</td>
              </tr>
              <tr>
                <td>アドレス種別</td>
                <td>{{ selected.AddressType }}</td>
              </tr>
              <tr>
                <td>名前</td>
                <td>{{ selected.Name }}</td>
              </tr>
              <tr>
                <td>送信元ホスト</td>
                <td>{{ selected.Host }}</td>
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
          <span class="headline">Bluetooth Device RSSI時間変化 3Dグラフ</span>
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
          <span class="headline">Bluetooth Device RSSI位置変化 3Dグラフ</span>
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
    <v-dialog v-model="blueScanGraphDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">Bluetooth Deviceの名前とアドレスの関係</span>
        </v-card-title>
        <div
          id="blueScanGraph"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="blueScanGraphDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="blueVendorChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">Bluetooth Deviceベンダー別</span>
        </v-card-title>
        <div
          id="blueVendorChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="blueVendorChartDialog = false">
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
        { text: 'RSSI', value: 'LastRSSI', width: '8%' },
        {
          text: 'アドレス',
          value: 'Address',
          width: '12%',
          filter: (value) => {
            if (!this.conf.address) return true
            return value.includes(this.conf.address)
          },
        },
        {
          text: 'アドレスタイプ',
          value: 'AddressType',
          width: '12%',
          filter: (value) => {
            if (!this.conf.addressType) return true
            return value.includes(this.conf.addressType)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '8%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: '送信元ホスト',
          value: 'Host',
          width: '10%',
          filter: (value) => {
            if (!this.conf.host) return true
            return value.includes(this.conf.host)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '15%',
          filter: (value) => {
            if (!this.conf.vendor) return true
            return value.includes(this.conf.vendor)
          },
        },
        { text: '回数', value: 'Count', width: '10%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      blueDevice: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      conf: {
        host: '',
        address: '',
        addressType: '',
        vendor: '',
        name: '',
        sortBy: 'LastRSSI',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      rssiTime3DDialog: false,
      rssiLoc3DDialog: false,
      blueScanGraphDialog: false,
      blueVendorChartDialog: false,
      setNameDialog: false,
      setNameError: false,
    }
  },
  async fetch() {
    this.blueDevice = await this.$axios.$get('/api/report/BlueDevice')
    if (!this.blueDevice) {
      return
    }
    this.blueDevice.forEach((e) => {
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
    const c = this.$store.state.report.twsensor.blueDevice
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twsensor/setBlueDevice', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/BlueDevice/' + this.selected.ID)
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
    openEditNameDialog(item) {
      this.selected = item
      this.setNameDialog = true
      this.setNameError = false
    },
    setName() {
      const req = {
        Type: 'device',
        ID: this.selected.ID,
        Name: this.selected.Name,
      }
      this.$axios
        .post('/api/report/BlueScan/name', req)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.setNameError = true
          this.$fetch()
        })
      this.setNameDialog = false
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
          false,
          this.blueDevice,
          this.conf
        )
      })
    },
    openRSSILoc3DChart() {
      this.rssiLoc3DDialog = true
      this.$nextTick(() => {
        this.$showRSSILoc3DChart(
          'rssiLoc3DChart',
          false,
          this.blueDevice,
          this.conf
        )
      })
    },
    openBlueScanGraph() {
      this.blueScanGraphDialog = true
      this.$nextTick(() => {
        this.$showBlueScanGraph('blueScanGraph', this.blueDevice, this.conf)
      })
    },
    openBlueVendorChart() {
      this.blueVendorChartDialog = true
      this.$nextTick(() => {
        this.$showBlueVendorChart('blueVendorChart', this.blueDevice, this.conf)
      })
    },
    makeExports() {
      const exports = []
      this.blueDevice.forEach((d) => {
        if (!this.$filterBluetoothDev(d, this.conf)) {
          return
        }
        exports.push({
          アドレス: d.Address,
          アドレス種別: d.AddressType,
          名前: d.Name,
          送信元ホスト: d.Host,
          ベンダー: d.Vendor,
          信号レベル: d.LastRSSI,
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
