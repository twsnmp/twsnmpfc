<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        MQTTクライアント
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        MQTT統計の削除に失敗しました
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="mqttClients"
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
        <template #[`item.TopicsCount`]="{ item }">
          {{ formatCount(item.TopicsCount) }}
        </template>
        <template #[`item.actions`]="{ item }">
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
            <td colspan="5"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_MQTT_Client_List.csv"
          header="TWSNMP FCのMQTTクライアントリスト"
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
          name="TWSNMP_FC_MQTT_Client_List.csv"
          header="TWSNMP FCのMQTTクライアントリスト"
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
          name="TWSNMP_FC_MQTT_Client_List.xls"
          header="TWSNMP FCのMQTTクライアントリスト"
          worksheet="MQTTクライアント"
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
          <span class="headline">MQTTクライアント統計削除</span>
        </v-card-title>
        <v-card-text>
          選択したクライアントの統計情報を全て削除しますか？
        </v-card-text>
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
        <v-card-text> 全てのMQTT統計情報を削除しますか？ </v-card-text>
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
          width: '18%',
          filter: (value) => {
            if (!this.filter.clientID) return true
            return value.includes(this.filter.clientID)
          },
        },
        {
          text: '接続元',
          value: 'Remote',
          width: '15%',
          filter: (value) => {
            if (!this.filter.remote) return true
            return value.includes(this.filter.remote)
          },
        },
        { text: 'トピック数', value: 'TopicsCount', width: '8%' },
        { text: '回数合計', value: 'Count', width: '8%' },
        { text: 'バイト数合計', value: 'Bytes', width: '10%' },
        { text: '初回', value: 'First', width: '13%' },
        { text: '最終', value: 'Last', width: '13%' },
        { text: '操作', value: 'actions', width: '8%', sortable: false },
      ],
      filter: {
        clientID: '',
        remote: '',
      },
      mqttClients: [],
      selected: {},
      deleteDialog: false,
      deleteAllDialog: false,
      deleteError: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/report/mqtt')
    if (!r) {
      return
    }
    const clientMap = {}
    r.forEach((s) => {
      if (!clientMap[s.ClientID]) {
        clientMap[s.ClientID] = {
          ClientID: s.ClientID,
          Remote: s.Remote,
          State: s.State,
          Count: 0,
          Bytes: 0,
          Topics: new Set(),
          FirstTime: s.First,
          LastTime: s.Last,
          IDs: [],
        }
      }
      const c = clientMap[s.ClientID]
      c.Count += s.Count
      c.Bytes += s.Bytes
      c.Topics.add(s.Topic)
      c.IDs.push(s.ID)
      if (s.First < c.FirstTime) {
        c.FirstTime = s.First
      }
      if (s.Last > c.LastTime) {
        c.LastTime = s.Last
        c.Remote = s.Remote
        c.State = s.State
      }
    })

    this.mqttClients = Object.values(clientMap).map((c) => {
      return {
        ClientID: c.ClientID,
        Remote: c.Remote,
        State: c.State,
        Count: c.Count,
        Bytes: c.Bytes,
        TopicsCount: c.Topics.size,
        First: this.$timeFormat(
          new Date(c.FirstTime / (1000 * 1000)),
          '{yyyy}/{MM}/{dd} {HH}:{mm}'
        ),
        Last: this.$timeFormat(
          new Date(c.LastTime / (1000 * 1000)),
          '{yyyy}/{MM}/{dd} {HH}:{mm}'
        ),
        IDs: c.IDs,
      }
    })
  },
  computed: {
    readOnly() {
      return this.$store.state.map.readOnly
    },
  },
  methods: {
    async doDelete() {
      this.deleteDialog = false
      this.deleteError = false
      for (const id of this.selected.IDs) {
        try {
          await this.$axios.delete('/api/report/mqtt/' + id)
        } catch (e) {
          this.deleteError = true
        }
      }
      this.$fetch()
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
      this.mqttClients.forEach((e) => {
        if (!this.filterMqtt(e)) {
          return
        }
        exports.push({
          状態: e.State,
          クライアントID: e.ClientID,
          接続元: e.Remote,
          トピック数: e.TopicsCount,
          回数合計: e.Count,
          バイト数合計: e.Bytes,
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
      return true
    },
  },
}
</script>
