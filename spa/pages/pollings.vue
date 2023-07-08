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
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
        <template #[`item.LogMode`]="{ item }">
          {{ $getLogModeName(item.LogMode) }}
        </template>
        <template #[`item.actions`]="{ item }">
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
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-select v-model="conf.state" :items="stateList" label="State">
              </v-select>
            </td>
            <td>
              <v-text-field v-model="conf.node" label="node"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
            </td>
            <td></td>
            <td>
              <v-select v-model="conf.polltype" :items="typeList" label="type">
              </v-select>
            </td>
            <td></td>
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
          color="primary"
          dark
          @click="setPollingParamsDialog = true"
        >
          <v-icon>mdi-tune</v-icon>
          パラメータ変更
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
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Polling_List.csv"
          header="TWSNMP FCのポーリングリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_Polling_List.xls"
          header="TWSNMP FCのポーリングリスト"
          worksheet="ポーリング"
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
    <v-dialog v-model="editDialog" persistent max-width="900px">
      <v-card>
        <v-card-title> ポーリング設定 </v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-select
                v-if="editIndex === -1"
                v-model="editPolling.NodeID"
                :items="nodeList"
                label="ノード"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editPolling.Name"
                label="名前"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editPolling.Level"
                :items="$levelList"
                label="レベル"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="editPolling.Type"
                :items="$typeList"
                label="種別"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editPolling.Mode"
                label="モード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editPolling.Params"
                label="パラメータ"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editPolling.Filter"
                label="フィルター"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editPolling.Extractor"
                :items="extractorList"
                label="抽出パターン"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="editPolling.LogMode"
                :items="$logModeList"
                label="ログモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <label>判定スクリプト</label>
          <prism-editor
            v-model="editPolling.Script"
            class="script"
            :highlight="highlighter"
            line-numbers
          ></prism-editor>
          <v-slider
            v-model="editPolling.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="3600"
            min="5"
            hide-details
          >
            <template #append>
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
            <template #append>
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
            <template #append>
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
    <v-dialog v-model="setPollingParamsDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリングパラメータ設定</span>
        </v-card-title>
        <v-card-text>
          <v-slider
            v-model="newPollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="3600"
            min="5"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="newPollInt"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="newTimeout"
            label="タイムアウト(Sec)"
            class="align-center"
            max="10"
            min="1"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="newTimeout"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="newRetry"
            label="リトライ回数"
            class="align-center"
            max="5"
            min="0"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="newRetry"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setPollingParams()">
            <v-icon>mdi-content-</v-icon>
            変更
          </v-btn>
          <v-btn color="normal" @click="setPollingParamsDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="templateDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline">テンプレートから選択</span>
          <v-spacer></v-spacer>
        </v-card-title>
        <v-data-table
          v-model="selectedTemplate"
          :headers="headersTemplate"
          :items="templates"
          single-select
          item-key="ID"
          show-select
          :items-per-page="15"
          sort-by="Type"
          dense
        >
          <template #[`item.Level`]="{ item }">
            <v-icon :color="$getStateColor(item.Level)">{{
              $getStateIconName(item.Level)
            }}</v-icon>
            {{ $getStateName(item.Level) }}
          </template>
          <template #[`body.append`]>
            <tr>
              <td></td>
              <td>
                <v-text-field v-model="tempName" label="name"></v-text-field>
              </td>
              <td></td>
              <td>
                <v-select v-model="tempType" :items="typeList" label="type">
                </v-select>
              </td>
              <td></td>
              <td></td>
            </tr>
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
import { PrismEditor } from 'vue-prism-editor'
import 'vue-prism-editor/dist/prismeditor.min.css'
import { highlight, languages } from 'prismjs/components/prism-core'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-javascript'
import 'prismjs/themes/prism-tomorrow.css'
export default {
  components: {
    PrismEditor,
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
          width: '10%',
          filter: (value) => {
            if (!this.conf.state) return true
            return this.conf.state === value
          },
        },
        {
          text: 'ノード',
          value: 'NodeName',
          width: '18%',
          filter: (value) => {
            if (!this.conf.node) return true
            return value.includes(this.conf.node)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '18%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        { text: 'レベル', value: 'Level', width: '13%' },
        {
          text: '種別',
          value: 'Type',
          width: '8%',
          filter: (value) => {
            if (!this.conf.polltype) return true
            return this.conf.polltype === value
          },
        },
        { text: 'ログ', value: 'LogMode', width: '6%' },
        { text: '最終実施', value: 'TimeStr', width: '15%' },
        { text: '操作', value: 'actions', width: '12%' },
      ],
      nodeList: [],
      pollings: [],
      tempType: '',
      tempName: '',
      headersTemplate: [
        {
          text: '名前',
          value: 'Name',
          width: '25%',
          filter: (value) => {
            if (!this.tempName) return true
            return value.includes(this.tempName)
          },
        },
        { text: 'レベル', value: 'Level', width: '15%' },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.tempType) return true
            return this.tempType === value
          },
        },
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
      setPollingParamsDialog: false,
      newPollInt: 60,
      newTimeout: 1,
      newRetry: 1,
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
        { text: 'ARP Log', value: 'arplog' },
        { text: 'Command', value: 'cmd' },
        { text: 'SSH', value: 'ssh' },
        { text: 'Report', value: 'report' },
        { text: 'TWSNMP', value: 'twsnmp' },
        { text: 'VMware', value: 'vmware' },
        { text: 'LXI', value: 'lxi' },
      ],
      conf: {
        state: '',
        node: '',
        name: '',
        polltype: '',
        sortBy: 'State',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
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
      e.TimeStr = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    })
    if (this.extractorList.length < 1) {
      this.extractorList = [
        {
          text: '',
          ID: '',
        },
        {
          text: 'goqueryによるデータ取得',
          value: 'goquery',
        },
        {
          text: 'getBodyによるデータ取得',
          value: 'getBody',
        },
      ]
      const groks = await this.$axios.$get('/api/conf/grok')
      if (groks) {
        groks.forEach((g) => {
          this.extractorList.push({
            text: g.Name,
            value: g.ID,
          })
        })
      }
    }
  },
  computed: {
    hasSelectedPollings() {
      return this.selectedPollings.length > 0
    },
  },
  created() {
    const c = this.$store.state.pollings.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('pollings/setConf', this.conf)
  },
  methods: {
    highlighter(code) {
      return highlight(code, languages.js)
    },
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
      this.deleteError = false
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
      this.deleteError = false
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
      this.updateError = false
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
      this.updateError = false
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
    setPollingParams() {
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.updateError = false
      this.$axios
        .post('/api/pollings/setParams', {
          IDs: ids,
          Timeout: this.newTimeout,
          Retry: this.newRetry,
          PollInt: this.newPollInt,
        })
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.updateError = true
        })
      this.setPollingParamsDialog = false
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
      this.updateError = false
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
    makeExports() {
      const exports = []
      this.pollings.forEach((e) => {
        if (!this.filterPolling(e)) {
          return
        }
        exports.push({
          ノード: e.NodeName,
          名前: e.Name,
          種別: e.Type,
          モード: e.Mode,
          ログ: this.$getLogModeName(e.LogMode),
          レベル: this.$getStateName(e.Level),
          状態: this.$getStateName(e.State),
          最終実施日時: e.TimeStr,
        })
      })
      return exports
    },
    filterPolling(e) {
      if (this.conf.state && this.conf.state !== e.State) {
        return false
      }
      if (this.conf.polltype && this.conf.Type !== e.polltype) {
        return false
      }
      if (this.conf.name && !e.Name.includes(this.conf.name)) {
        return false
      }
      if (this.conf.node && !e.NodeName.includes(this.conf.node)) {
        return false
      }
      return true
    },
  },
}
</script>

<style>
.script {
  height: 100px;
  overflow: auto;
  margin-top: 5px;
  margin-bottom: 5px;
}
</style>
