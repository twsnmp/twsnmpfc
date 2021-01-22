<template>
  <div>
    <v-card>
      <v-card-title>
        ポーリングリスト
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-alert :value="deleteError" type="error" dense dismissible>
        ポーリングを削除できませんでした
      </v-alert>
      <v-alert :value="updateError" type="error" dense dismissible>
        ポーリングを変更できませんでした
      </v-alert>
      <v-data-table :headers="headers" :items="pollings" :search="search" dense>
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="getStateColor(item.State)">{{
            getStateIconName(item.State)
          }}</v-icon>
          {{ getStateName(item.State) }}
        </template>
        <template v-slot:[`item.Level`]="{ item }">
          <v-icon :color="getStateColor(item.Level)">{{
            getStateIconName(item.Level)
          }}</v-icon>
          {{ getStateName(item.Level) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small class="mr-2" @click="editPollingFunc(item)">
            mdi-pencil
          </v-icon>
          <v-icon small @click="deletePollingFunc(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
    </v-card>
    <v-dialog v-model="editDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング編集</span>
        </v-card-title>
        <v-card-text>
          <v-combobox
            v-model="editPolling.NodeID"
            :items="nodeList"
            label="ノード"
          ></v-combobox>
          <v-text-field v-model="editPolling.Name" label="名前"></v-text-field>
          <v-select v-model="editPolling.Type" :items="typeList" label="種別">
          </v-select>
          <v-select
            v-model="editPolling.Level"
            :items="levelList"
            label="レベル"
          >
          </v-select>
          <v-text-field
            v-model="editPolling.Polling"
            label="定義"
          ></v-text-field>
          <v-slider
            v-model="editPolling.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="600"
            min="60"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="editPolling.PollInt"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="editPolling.Timeout"
            label="タイムアウト(Sec)"
            class="align-center"
            max="10"
            min="1"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="editPolling.Timeout"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="editPolling.Retry"
            label="リトライ回数"
            class="align-center"
            max="5"
            min="0"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="editPolling.Retry"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-select
            v-model="editPolling.LogMode"
            :items="logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="closeEdit">キャンセル</v-btn>
          <v-btn color="primary" dark @click="doUpdatePolling">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング削除</span>
        </v-card-title>
        <v-card-text>
          ポーリング{{ deletePolling.Name }}を削除しますか？
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="closeDelete">キャンセル</v-btn>
          <v-btn color="error" @click="doDeletePolling">削除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
export default {
  async fetch() {
    const r = await this.$axios.$get('/api/pollings')
    this.nodeList = r.NodeList
    const nodeMap = {}
    r.NodeList.forEach((e) => {
      nodeMap[e.value] = e.text
    })
    this.pollings = r.Pollings
    this.pollings.forEach((e) => {
      e.NodeName = nodeMap[e.NodeID]
      const t = new Date(e.LastTime / (1000 * 1000))
      e.LastTime = t.toLocaleString()
    })
  },
  data() {
    return {
      editDialog: false,
      deleteDialog: false,
      editIndex: -1,
      deleteIndex: -1,
      deleteError: false,
      updateError: false,
      editPolling: {},
      deletePolling: {},
      search: '',
      headers: [
        {
          text: '状態',
          value: 'State',
        },
        { text: 'ノード', value: 'NodeName' },
        { text: '名前', value: 'Name' },
        { text: 'レベル', value: 'Level' },
        { text: '種別', value: 'Type' },
        { text: '定義', value: 'Polling' },
        { text: '最終実施', value: 'LastTime' },
        { text: '数値', value: 'LastVal' },
        { text: '操作', value: 'actions' },
      ],
      nodeList: [],
      pollings: [],
      levelList: [
        {
          text: '重度',
          value: 'high',
        },
        {
          text: '軽度',
          value: 'low',
        },
        {
          text: '注意',
          value: 'warn',
        },
        {
          text: '情報',
          value: 'info',
        },
      ],
      typeList: [
        { text: 'PING', value: 'ping' },
        { text: 'SNMP', value: 'snmp' },
        { text: 'TCP', value: 'tcp' },
        { text: 'HTTP', value: 'http' },
        { text: 'HTTPS', value: 'https' },
        { text: 'TLS', value: 'tls' },
        { text: 'DNS', value: 'dns' },
        { text: 'NTP', value: 'ntp' },
        { text: 'SYSLOG', value: 'syslog' },
        { text: 'SYSLOG PRI', value: 'syslogpri' },
        { text: 'SYSLOG Device', value: 'syslogdevice' },
        { text: 'SYSLOG User', value: 'sysloguser' },
        { text: 'SYSLOG Flow', value: 'syslogflow' },
        { text: 'SNMP TRAP', value: 'trap' },
        { text: 'NetFlow', value: 'netflow' },
        { text: 'IPFIX', value: 'ipfix' },
        { text: 'Command', value: 'cmd' },
        { text: 'SSH', value: 'ssh' },
        { text: 'TWSNMP', value: 'twsnmp' },
        { text: 'VMware', value: 'vmware' },
      ],
      logModeList: [
        { text: '記録しない', value: 0 },
        { text: '常に記録', value: 1 },
        { text: '状態変化時のみ記録', value: 2 },
        { text: 'AI分析', value: 3 },
      ],
    }
  },
  methods: {
    getStateColor(state) {
      switch (state) {
        case 'high':
          return 'red'
        case 'low':
          return 'pink'
        case 'warn':
          return 'yellow'
        case 'repair':
          return 'blue'
        case 'normal':
          return 'green'
        default:
          return 'gray'
      }
    },
    getStateName(state) {
      switch (state) {
        case 'high':
          return '重度'
        case 'low':
          return '軽度'
        case 'warn':
          return '注意'
        case 'repair':
          return '復帰'
        case 'normal':
          return '正常'
        default:
          return '不明'
      }
    },
    getStateIconName(state) {
      switch (state) {
        case 'high':
          return 'mdi-alert-circle'
        case 'low':
          return 'mdi-alert-circle'
        case 'warn':
          return 'mdi-alert'
        case 'repair':
          return 'mdi-autorenew'
        case 'normal':
          return 'mdi-information'
        default:
          return 'comment-question-outline'
      }
    },
    editPollingFunc(item) {
      this.editIndex = this.pollings.indexOf(item)
      this.editPolling = Object.assign({}, item)
      this.editDialog = true
    },
    deletePollingFunc(item) {
      this.deleteIndex = this.pollings.indexOf(item)
      this.deletePolling = Object.assign({}, item)
      this.deleteDialog = true
    },
    doDeletePolling() {
      this.pollings.splice(this.deleteIndex, 1)
      this.$axios
        .post('/api/polling/delete', { ID: this.deletePolling.ID })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.closeDelete()
    },
    closeDelete() {
      this.deleteDialog = false
      this.$nextTick(() => {
        this.deleteIndex = -1
      })
    },
    closeEdit() {
      this.editDialog = false
      this.$nextTick(() => {
        this.editIndex = -1
      })
    },
    doUpdatePolling() {
      if (this.editIndex > -1) {
        Object.assign(this.pollings[this.editIndex], this.editPolling)
        this.$axios.post('/api/polling/update', this.editPolling).catch((e) => {
          this.updateError = true
          this.$fetch()
        })
      }
      this.closeEdit()
    },
  },
}
</script>
