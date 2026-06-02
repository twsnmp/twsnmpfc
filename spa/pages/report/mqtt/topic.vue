<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        MQTTトピック
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        MQTT統計の削除に失敗しました
      </v-alert>
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーに失敗しました
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
      <v-snackbar v-model="addPollingDone" absolute centered color="primary">
        ポーリングを追加しました
      </v-snackbar>
      <v-data-table
        :headers="headers"
        :items="mqttStats"
        :items-per-page="15"
        sort-by="ClientID"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template #[`item.Count`]="{ item }">
          {{ formatCount(item.Count) }}
        </template>
        <template #[`item.Bytes`]="{ item }">
          {{ formatBytes(item.Bytes) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small title="トピック名をコピー" @click="copyTopic(item)">
            mdi-content-copy
          </v-icon>
          <v-icon
            v-if="!readOnly"
            small
            title="ポーリング追加"
            @click="editPolling(item)"
          >
            mdi-card-plus
          </v-icon>
          <v-icon
            v-if="!readOnly"
            small
            color="red"
            title="削除"
            @click="openDeleteDialog(item)"
          >
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field
                v-model="filter.clientID"
                label="Client ID"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="filter.remote"
                label="Remote"
              ></v-text-field>
            </td>
            <td>
              <v-text-field v-model="filter.topic" label="Topic"></v-text-field>
            </td>
            <td colspan="5"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_MQTT_Topic_List.csv"
          header="TWSNMP FCのMQTTトピックリスト"
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
          name="TWSNMP_FC_MQTT_Topic_List.csv"
          header="TWSNMP FCのMQTTトピックリスト"
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
          name="TWSNMP_FC_MQTT_Topic_List.xls"
          header="TWSNMP FCのMQTTトピックリスト"
          worksheet="MQTTトピック"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn v-if="!readOnly" color="error" dark @click="openDeleteAllDialog">
          <v-icon>mdi-delete</v-icon>
          全て削除
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
          <span class="headline">MQTT統計削除</span>
        </v-card-title>
        <v-card-text> 選択したトピックの統計情報を削除しますか？ </v-card-text>
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
    <v-dialog v-model="deleteAllDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">MQTT統計全削除</span>
        </v-card-title>
        <v-card-text> 全てのMQTTトピック統計情報を削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteAll">
            <v-icon>mdi-delete</v-icon>
            全て削除
          </v-btn>
          <v-btn color="normal" @click="deleteAllDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editPollingDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>MQTTポーリング追加</v-card-title>
        <v-alert v-model="addPollingError" color="error" dense dismissible>
          ポーリングを追加できませんでした
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.NodeID"
                :items="nodeList"
                label="ノード"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field v-model="polling.Name" label="名前"></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.Level"
                :items="$levelList"
                label="レベル"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Type"
                readonly
                label="種別"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Mode"
                readonly
                label="モード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="polling.Params"
                label="MQTTサーバーURL (空欄時は tcp://[ノードIP]:1883)"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Filter"
                readonly
                label="トピック"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-textarea
                v-model="polling.Script"
                label="判定スクリプト"
                rows="3"
              ></v-textarea>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-slider
                v-model="polling.PollInt"
                label="ポーリング間隔(Sec)"
                class="align-center"
                max="86400"
                min="5"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="polling.PollInt"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doAddPolling">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="editPollingDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
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
        { text: '状態', value: 'State', width: '8%' },
        {
          text: 'クライアントID',
          value: 'ClientID',
          width: '15%',
          filter: (value) => {
            if (!this.filter.clientID) return true
            return value.includes(this.filter.clientID)
          },
        },
        {
          text: '接続元',
          value: 'Remote',
          width: '10%',
          filter: (value) => {
            if (!this.filter.remote) return true
            return value.includes(this.filter.remote)
          },
        },
        {
          text: 'トピック',
          value: 'Topic',
          width: '20%',
          filter: (value) => {
            if (!this.filter.topic) return true
            return value.includes(this.filter.topic)
          },
        },
        { text: '回数', value: 'Count', width: '5%' },
        { text: 'バイト数', value: 'Bytes', width: '8%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%', sortable: false },
      ],
      filter: {
        clientID: '',
        remote: '',
        topic: '',
      },
      mqttStats: [],
      selected: {},
      deleteDialog: false,
      deleteAllDialog: false,
      deleteError: false,
      editPollingDialog: false,
      addPollingError: false,
      addPollingDone: false,
      copyDone: false,
      copyError: false,
      polling: {},
      nodeList: [],
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/report/mqtt')
    if (!r) {
      return
    }
    this.mqttStats = r
    this.mqttStats.forEach((s) => {
      s.First = this.$timeFormat(
        new Date(s.First / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      s.Last = this.$timeFormat(
        new Date(s.Last / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
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
        .delete('/api/report/mqtt/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    doDeleteAll() {
      this.$axios
        .delete('/api/report/mqtt/all')
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteAllDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openDeleteAllDialog() {
      this.deleteAllDialog = true
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
    makeExports() {
      const exports = []
      this.mqttStats.forEach((e) => {
        if (!this.filterMqtt(e)) {
          return
        }
        exports.push({
          状態: e.State,
          クライアントID: e.ClientID,
          接続元: e.Remote,
          トピック: e.Topic,
          回数: e.Count,
          バイト数: e.Bytes,
          初回日時: e.First,
          最終日時: e.Last,
        })
      })
      return exports
    },
    filterMqtt(e) {
      if (this.filter.clientID && !e.ClientID.includes(this.filter.clientID)) {
        return false
      }
      if (this.filter.remote && !e.Remote.includes(this.filter.remote)) {
        return false
      }
      if (this.filter.topic && !e.Topic.includes(this.filter.topic)) {
        return false
      }
      return true
    },
    copyTopic(item) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      navigator.clipboard.writeText(item.Topic).then(
        () => {
          this.copyDone = true
          setTimeout(() => {
            this.copyDone = false
          }, 3000)
        },
        () => {
          this.copyError = true
        }
      )
    },
    async editPolling(item) {
      if (this.nodeList.length < 1) {
        const r = await this.$axios.$get('/api/nodes')
        if (r) {
          r.forEach((n) => {
            this.nodeList.push({ text: n.Name, value: n.ID, ip: n.IP })
          })
        }
      }
      let nodeID = ''
      for (let j = 0; j < this.nodeList.length; j++) {
        if (this.nodeList[j].ip === item.Remote) {
          nodeID = this.nodeList[j].value
          break
        }
      }
      if (!nodeID) {
        for (let j = 0; j < this.nodeList.length; j++) {
          const ip = this.nodeList[j].ip
          const name = this.nodeList[j].text.toLowerCase()
          if (
            ip === '127.0.0.1' ||
            ip === 'localhost' ||
            name === 'localhost' ||
            name.includes('twsnmp')
          ) {
            nodeID = this.nodeList[j].value
            break
          }
        }
      }
      if (!nodeID && this.nodeList.length > 0) {
        nodeID = this.nodeList[0].value
      }
      this.polling = {
        ID: '',
        Name: 'mqtt:' + item.Topic,
        NodeID: nodeID,
        Type: 'mqtt',
        Mode: 'subscribe',
        Params: 'tcp://127.0.0.1:1883',
        Filter: item.Topic,
        Extractor: '',
        Script: '',
        Level: 'low',
        PollInt: 600,
        Timeout: 5,
        Retry: 0,
        LogMode: 0,
      }
      this.editPollingDialog = true
    },
    doAddPolling() {
      this.$axios
        .post('/api/polling/add', this.polling)
        .then(() => {
          this.editPollingDialog = false
          this.addPollingDone = true
          setTimeout(() => {
            this.addPollingDone = false
          }, 3000)
        })
        .catch((e) => {
          this.addPollingError = true
        })
    },
  },
}
</script>
