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
      <v-alert v-if="deleteError" color="error" dense dismissible>
        環境センサーを削除できません
      </v-alert>
      <v-alert v-if="setNameError" color="error" dense dismissible>
        環境センサーの名前を変更できません
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
            <v-list-item @click="openEnv3DChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>時系列3D</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list>
            <v-list-item @click="openEnv2DChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-line</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>時系列2D</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeSensorExports"
          type="csv"
          name="TWSNMP_FC_Env_Monitor_List.csv"
          header="TWSNMP FCで作成した環境センサーリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeSensorExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_Env_Monitor_List.csv"
          header="TWSNMP FCで作成した環境センサーリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeSensorExports"
          type="xls"
          name="TWSNMP_FC_Env_Monitor_List.xls"
          header="TWSNMP FCで作成した環境センサーリスト"
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
                <td>送信元ホスト</td>
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
          <download-excel
            :fetch="makeDataExports"
            type="csv"
            name="TWSNMP_FC_Env_Data_List.csv"
            :header="makeDataHeader"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeDataExports"
            type="csv"
            :escape-csv="false"
            name="TWSNMP_FC_Env_Data_List.csv"
            :header="makeDataHeader"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV(NO ESC)
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeDataExports"
            type="xls"
            name="TWSNMP_FC_Env_Data_List.xls"
            :header="makeDataHeader"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="env3DDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">{{ chartTitle }}</span>
          <v-spacer></v-spacer>
          <v-select
            v-model="selectedType"
            :items="$envTypeList"
            label="項目"
            single-line
            hide-details
            @change="change3DType"
          ></v-select>
        </v-card-title>
        <div
          id="env3DChart"
          style="width: 95vw; height: 80vh; margin: 0 atuo"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="env3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="env2DDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">{{ chartTitle }}</span>
          <v-spacer></v-spacer>
          <v-select
            v-model="selectedType"
            :items="$envTypeList"
            label="項目"
            single-line
            hide-details
            @change="change2DType"
          ></v-select>
        </v-card-title>
        <div
          id="env2DChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="env2DDialog = false">
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
          text: '送信元ホスト',
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
      env2DDialog: false,
      chartTitle: '',
      selectedType: 'Temp',
      setNameDialog: false,
      setNameError: false,
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
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
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
  computed: {
    readOnly() {
      return this.$store.state.map.readOnly
    },
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
    openEditNameDialog(item) {
      this.selected = item
      this.setNameDialog = true
      this.setNameError = false
    },
    setName() {
      const req = {
        Type: 'env',
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
    openEnv3DChart() {
      this.env3DDialog = true
      this.change3DType()
    },
    change3DType() {
      if (!this.selectedType) {
        this.selectedType = 'Temp'
      }
      this.chartTitle = this.$getEnvName(this.selectedType)
      this.$nextTick(() => {
        this.$showEnv3DChart(
          'env3DChart',
          this.selectedType,
          this.envMonitor,
          this.conf
        )
      })
    },
    openEnv2DChart() {
      this.env2DDialog = true
      this.change2DType()
    },
    change2DType() {
      if (!this.selectedType) {
        this.selectedType = 'Temp'
      }
      this.chartTitle = this.$getEnvName(this.selectedType)
      this.$nextTick(() => {
        this.$showEnv2DChart(
          'env2DChart',
          this.selectedType,
          this.envMonitor,
          this.conf
        )
      })
    },
    makeSensorExports() {
      const exports = []
      this.envMonitor.forEach((d) => {
        if (!this.$filterEnvMon(d, this.conf)) {
          return
        }
        exports.push({
          アドレス: d.Address,
          名前: d.Name,
          送信元ホスト: d.Host,
          信号レベル: d.LastRSSI,
          データ数: d.DataCount,
          回数: d.Count,
          初回日時: d.First,
          最終日時: d.Last,
        })
      })
      return exports
    },
    makeDataExports() {
      const exports = []
      this.selected.EnvData.forEach((d) => {
        exports.push({
          記録日時: this.$timeFormat(
            new Date(d.Time / (1000 * 1000)),
            '{yyyy}/{MM}/{dd} {HH}:{mm}'
          ),
          気温: d.Temp,
          湿度: d.Humidity,
          照度: d.Illuminance,
          気圧: d.BarometricPressure,
          騒音: d.Sound,
          総揮発性有機化合物: d.ETVOC,
          二酸化炭素濃度: d.ECo2,
          電池残量: d.Battery,
        })
      })
      return exports
    },
    makeDataHeader() {
      return (
        'TWSNMP FCで作成した環境センサー(' + this.selected.Address + ')のデータ'
      )
    },
  },
}
</script>
