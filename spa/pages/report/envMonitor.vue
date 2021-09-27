<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        環境センサー
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="envMonitor"
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
              <v-text-field v-model="conf.address" label="Address">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="Name"></v-text-field>
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
            <v-list-item @click="openEnv3DChart('Temp')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>気温</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('Humidity')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>湿度</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('Illuminance')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>照度</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('BarometricPressure')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>気圧</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('Sound')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>騒音</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('ETVOC')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>総揮発性有機化合物</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('ECo2')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>二酸化炭素濃度</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv3DChart('Battery')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>電池</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :data="envMonitor"
          type="csv"
          name="TWSNMP_FC_Env_Monitor_List.csv"
          header="TWSNMP FC Env Monitor List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="envMonitor"
          type="xls"
          name="TWSNMP_FC_Env_Monitor_List.xls"
          header="TWSNMP FC Env Monitor List"
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
          <span class="headline">環境センサー情報</span>
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
                <td>名前</td>
                <td>{{ selected.Name }}</td>
              </tr>
              <tr>
                <td>受信ホスト</td>
                <td>{{ selected.Host }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>データ数</td>
                <td>{{ selected.EnvData.length }}</td>
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
    <v-dialog v-model="env3DDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline">{{ chartTitle }}</span>
        </v-card-title>
        <div id="env3DChart" style="width: 1000px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="env3DDialog = false">
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
          text: 'アドレス',
          value: 'Address',
          width: '15%',
          filter: (value) => {
            if (!this.conf.address) return true
            return value.includes(this.conf.address)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '15%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: '受信ホスト',
          value: 'Host',
          width: '15%',
          filter: (value) => {
            if (!this.conf.host) return true
            return value.includes(this.conf.host)
          },
        },
        { text: 'データ数', value: 'DataCount', width: '10%' },
        { text: '回数', value: 'Count', width: '10%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      envMonitor: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      conf: {
        host: '',
        address: '',
        name: '',
        sortBy: 'DataCount',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      env3DDialog: false,
      chartTitle: '',
    }
  },
  async fetch() {
    this.envMonitor = await this.$axios.$get('/api/report/EnvMonitor')
    if (!this.envMonitor) {
      return
    }
    this.envMonitor.forEach((e) => {
      e.First = this.$timeFormat(
        new Date(e.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.DataCount = e.EnvData.length
      e.LastRSSI =
        e.EnvData && e.EnvData.length > 0
          ? e.EnvData[e.EnvData.length - 1].RSSI * 1
          : 0
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.twsensor.envMonitor
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twsensor/setEnvMonitor', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/EnvMonitor/' + this.selected.ID)
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
    openEnv3DChart(type) {
      this.chartTitle = this.$getEnvName(type)
      this.env3DDialog = true
      this.$nextTick(() => {
        this.$showEnv3DChart('env3DChart', type, this.envMonitor)
      })
    },
  },
}
</script>
