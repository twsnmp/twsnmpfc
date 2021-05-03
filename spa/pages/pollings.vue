<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        ポーリングリスト
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        ポーリングを削除できませんでした
      </v-alert>
      <v-alert v-model="updateError" color="error" dense dismissible>
        ポーリングを変更できませんでした
      </v-alert>
      <v-data-table
        v-model="selectedPollings"
        :headers="headers"
        :items="pollings"
        item-key="ID"
        show-select
        dense
        :items-per-page="15"
        sort-by="State"
        sort-asec
      >
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template v-slot:[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small @click="$router.push({ path: '/polling/' + item.ID })">
            mdi-eye
          </v-icon>
          <v-icon
            v-if="item.LogMode > 1"
            small
            @click="$router.push({ path: '/report/ai/' + item.ID })"
          >
            mdi-brain
          </v-icon>
          <v-icon small @click="editPollingFunc(item)"> mdi-pencil </v-icon>
          <v-icon small @click="deletePollingFunc(item)"> mdi-delete </v-icon>
          <v-icon small @click="copyPolling(item)"> mdi-content-copy </v-icon>
        </template>
        <template v-slot:[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-select v-model="state" :items="stateList" label="level">
              </v-select>
            </td>
            <td>
              <v-text-field v-model="node" label="node"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="name" label="name"></v-text-field>
            </td>
            <td></td>
            <td>
              <v-select v-model="polltype" :items="typeList" label="type">
              </v-select>
            </td>
            <td></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="addPolling">
          <v-icon>mdi-plus</v-icon>
          追加
        </v-btn>
        <v-btn
          v-if="hasSelectedPollings"
          color="primary"
          dark
          @click="setPollingLevelDialog = true"
        >
          <v-icon>mdi-cog</v-icon>
          レベル変更
        </v-btn>
        <v-btn
          v-if="hasSelectedPollings"
          color="primary"
          dark
          @click="setPollingLogModeDialog = true"
        >
          <v-icon>mdi-alert</v-icon>
          ログモード変更
        </v-btn>
        <v-btn
          v-if="hasSelectedPollings"
          color="error"
          @click="deleteSelectedPollingDialog = true"
        >
          <v-icon>mdi-delete</v-icon>
          一括削除
        </v-btn>
        <download-excel
          :data="pollings"
          type="csv"
          name="TWSNMP_FC_Polling_List.csv"
          header="TWSNMP FC Polling List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="pollings"
          type="xls"
          name="TWSNMP_FC_Polling_List.xls"
          header="TWSNMP FC Polling List"
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
        <v-btn color="normal" dark to="/map">
          <v-icon>mdi-lan</v-icon>
          マップ
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング設定</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-if="editIndex === -1"
            v-model="editPolling.NodeID"
            :items="nodeList"
            label="ノード"
          ></v-select>
          <v-text-field v-model="editPolling.Name" label="名前"></v-text-field>
          <v-select
            v-model="editPolling.Level"
            :items="$levelList"
            label="レベル"
          >
          </v-select>
          <v-select v-model="editPolling.Type" :items="$typeList" label="種別">
          </v-select>
          <v-text-field
            v-model="editPolling.Mode"
            label="モード"
          ></v-text-field>
          <v-text-field
            v-model="editPolling.Params"
            label="パラメータ"
          ></v-text-field>
          <v-text-field
            v-model="editPolling.Filter"
            label="フィルター"
          ></v-text-field>
          <v-select
            v-model="editPolling.Extractor"
            :items="extractorList"
            label="抽出パターン"
          ></v-select>
          <v-textarea
            v-model="editPolling.Script"
            label="判定スクリプト"
            clearable
            rows="3"
            clear-icon="mdi-close-circle"
          ></v-textarea>
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
            :items="$logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="editIndex === -1"
            color="primary"
            dark
            @click="showTemplateDialog"
          >
            <v-icon>mdi-content-copy</v-icon>
            テンプレート
          </v-btn>
          <v-btn color="primary" dark @click="doUpdatePolling">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="closeEdit">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング削除</span>
        </v-card-title>
        <v-card-text>
          ポーリング{{ deletePolling.Name }}を削除しますか？
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeletePolling">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="closeDelete">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog
      v-model="deleteSelectedPollingDialog"
      persistent
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング一括削除</span>
        </v-card-title>
        <v-card-text> 選択したポーリングを全て削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="deletePollings">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteSelectedPollingDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="setPollingLevelDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">レベル変更</span>
        </v-card-title>
        <v-card-text>
          <v-select v-model="newLevel" :items="$levelList" label="レベル">
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setPollingLevel">
            <v-icon>mdi-content-</v-icon>
            変更
          </v-btn>
          <v-btn color="normal" @click="setPollingLevelDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="setPollingLogModeDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ログモード変更</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-model="newLogMode"
            :items="$logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setPollingLogMode">
            <v-icon>mdi-content-</v-icon>
            変更
          </v-btn>
          <v-btn color="normal" @click="setPollingLogModeDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="templateDialog" persistent max-width="900px">
      <v-card>
        <v-card-title>
          <span class="headline">テンプレートから選択</span>
          <v-spacer></v-spacer>
          <v-text-field
            v-model="searchTemplate"
            append-icon="mdi-magnify"
            label="検索"
            single-line
            hide-details
          ></v-text-field>
        </v-card-title>
        <v-data-table
          v-model="selectedTemplate"
          :headers="headersTemplate"
          :items="templates"
          :search="searchTemplate"
          single-select
          item-key="ID"
          show-select
          :items-per-page="15"
          sort-by="Type"
          dense
        >
          <template v-slot:[`item.Level`]="{ item }">
            <v-icon :color="$getStateColor(item.Level)">{{
              $getStateIconName(item.Level)
            }}</v-icon>
            {{ $getStateName(item.Level) }}
          </template>
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="selectTemplate">
            <v-icon>mdi-check</v-icon>
            選択
          </v-btn>
          <v-btn color="normal" @click="templateDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
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
      e.TimeStr = this.$timeFormat(t, '{MM}/{dd} {HH}:{mm}:{ss}')
    })
    if (this.extractorList.length < 1) {
      const groks = await this.$axios.$get('/api/conf/grok')
      if (groks) {
        this.extractorList = []
        groks.forEach((g) => {
          this.extractorList.push({
            text: g.Name,
            value: g.ID,
          })
        })
      }
    }
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
      extractorList: [],
      headers: [
        {
          text: '状態',
          value: 'State',
          width: '12%',
          filter: (value) => {
            if (!this.state) return true
            return this.state === value
          },
        },
        {
          text: 'ノード',
          value: 'NodeName',
          width: '18%',
          filter: (value) => {
            if (!this.node) return true
            return value.includes(this.node)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '20%',
          filter: (value) => {
            if (!this.name) return true
            return value.includes(this.name)
          },
        },
        { text: 'レベル', value: 'Level', width: '15%' },
        {
          text: '種別',
          value: 'Type',
          width: '8%',
          filter: (value) => {
            if (!this.polltype) return true
            return this.polltype === value
          },
        },
        { text: '最終実施', value: 'TimeStr', width: '15%' },
        { text: '操作', value: 'actions', width: '12%' },
      ],
      nodeList: [],
      pollings: [],
      searchTemplate: '',
      headersTemplate: [
        { text: '名前', value: 'Name', width: '25%' },
        { text: 'レベル', value: 'Level', width: '15%' },
        { text: '種別', value: 'Type', width: '10%' },
        { text: 'モード', value: 'Mode', width: '10%' },
        { text: '説明', value: 'Descr', width: '35%' },
      ],
      selectedTemplate: [],
      templateDialog: false,
      templates: [],
      selectedPollings: [],
      deleteSelectedPollingDialog: false,
      setPollingLevelDialog: false,
      newLevel: 'off',
      setPollingLogModeDialog: false,
      newLogMode: 0,
      state: '',
      node: '',
      name: '',
      polltype: '',
      stateList: [
        { text: '', value: '' },
        { text: '重度', value: 'high' },
        { text: '軽度', value: 'low' },
        { text: '注意', value: 'warn' },
        { text: '正常', value: 'normal' },
        { text: '復帰', value: 'repair' },
        { text: '不明', value: 'unknown' },
      ],
      typeList: [
        { text: '', value: '' },
        { text: 'PING', value: 'ping' },
        { text: 'SNMP', value: 'snmp' },
        { text: 'TCP', value: 'tcp' },
        { text: 'HTTP', value: 'http' },
        { text: 'TLS', value: 'tls' },
        { text: 'DNS', value: 'dns' },
        { text: 'NTP', value: 'ntp' },
        { text: 'SYSLOG', value: 'syslog' },
        { text: 'SNMP TRAP', value: 'trap' },
        { text: 'NetFlow', value: 'netflow' },
        { text: 'IPFIX', value: 'ipfix' },
        { text: 'Command', value: 'cmd' },
        { text: 'SSH', value: 'ssh' },
        { text: 'Report', value: 'report' },
        { text: 'TWSNMP', value: 'twsnmp' },
        { text: 'VMware', value: 'vmware' },
      ],
    }
  },
  computed: {
    hasSelectedPollings() {
      return this.selectedPollings.length > 0
    },
  },
  methods: {
    editPollingFunc(item) {
      this.editIndex = this.pollings.indexOf(item)
      this.editPolling = Object.assign({}, item)
      this.editDialog = true
    },
    copyPolling(item) {
      this.editIndex = -1
      this.editPolling = Object.assign({}, item)
      this.editPolling.ID = ''
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
        .post('/api/pollings/delete', [this.deletePolling.ID])
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.closeDelete()
    },
    deletePollings() {
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.$axios
        .post('/api/pollings/delete', ids)
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.deleteError = true
        })
      this.deleteSelectedPollingDialog = false
    },
    setPollingLevel() {
      if (!this.newLevel) {
        return
      }
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.$axios
        .post('/api/pollings/setlevel', {
          IDs: ids,
          Level: this.newLevel,
        })
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.updateError = true
        })
      this.setPollingLevelDialog = false
    },
    setPollingLogMode() {
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.$axios
        .post('/api/pollings/setlogmode', {
          IDs: ids,
          LogMode: this.newLogMode,
        })
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.updateError = true
        })
      this.setPollingLogModeDialog = false
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
    addPolling() {
      this.editIndex = -1
      this.editPolling = {
        ID: '',
        Name: '',
        NodeID: '',
        Type: 'ping',
        Mode: '',
        Params: '',
        Filter: '',
        Extractor: '',
        Script: '',
        Level: 'low',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.editDialog = true
    },
    doUpdatePolling() {
      if (this.editIndex > -1) {
        Object.assign(this.pollings[this.editIndex], this.editPolling)
        this.$axios.post('/api/polling/update', this.editPolling).catch((e) => {
          this.updateError = true
          this.$fetch()
        })
      } else {
        this.$axios
          .post('/api/polling/add', this.editPolling)
          .then(() => {
            this.$fetch()
          })
          .catch((e) => {
            this.updateError = true
            this.$fetch()
          })
      }
      this.closeEdit()
    },
    async showTemplateDialog() {
      this.selectedTemplate = []
      const r = await this.$axios.$get('/api/polling/template')
      if (r) {
        this.templates = r
        this.templateDialog = true
      }
    },
    selectTemplate() {
      if (!this.selectedTemplate || this.selectedTemplate.length !== 1) {
        return
      }
      this.editPolling.Name = this.selectedTemplate[0].Name
      this.editPolling.Type = this.selectedTemplate[0].Type
      this.editPolling.Mode = this.selectedTemplate[0].Mode
      this.editPolling.Params = this.selectedTemplate[0].Params
      this.editPolling.Filter = this.selectedTemplate[0].Filter
      this.editPolling.Extractor = this.selectedTemplate[0].Extractor
      this.editPolling.Script = this.selectedTemplate[0].Script
      this.templateDialog = false
    },
  },
}
</script>
