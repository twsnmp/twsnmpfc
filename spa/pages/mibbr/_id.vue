<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        MIBブラウザー - {{ node.Name }}
        <v-spacer></v-spacer>
        <v-combobox
          v-model="mibget.Name"
          :items="history"
          append-icon="mdi-magnify"
          label="オブジェクト名"
          dense
        ></v-combobox>
      </v-card-title>
      <v-alert v-model="error" color="error" dense dismissible>
        MIBを取得できませんでした
      </v-alert>
      <v-overlay
        absolute
        :value="overlay"
        color="rgb(179,179,179)"
        opacity="0.8"
      >
        <v-img :src="neko"></v-img>
      </v-overlay>
      <v-data-table
        :headers="headers"
        :items="items"
        :search="conf.search"
        dense
        :loading="$fetchState.pending || wait"
        loading-text="Loading... Please wait"
        class="mibbr"
        sort-by="Index"
        :items-per-page="conf.itemsPerPage"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template v-if="!tableMode" #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.value" label="value"></v-text-field>
            </td>
          </tr>
        </template>
        <template v-else #[`body.append`]>
          <tr>
            <td colspan="3">
              <v-text-field v-model="conf.search" label="filter"></v-text-field>
            </td>
          </tr>
        </template>
        <template v-if="!tableMode" #[`item.actions`]="{ item }">
          <v-icon small @click="addPolling(item)"> mdi-card-plus </v-icon>
          <v-icon small @click="copyOneMIB(item)"> mdi-content-copy </v-icon>
        </template>
      </v-data-table>
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
      <v-snackbar v-model="addPollingDone" absolute centered color="primary">
        ポーリング作成しました
      </v-snackbar>
      <v-card-actions>
        <v-select
          v-model="mode"
          :items="modeList"
          label="モード"
          @change="updateTable"
        >
        </v-select>
        <v-spacer></v-spacer>
        <v-switch v-model="mibget.Raw" label="生データ"></v-switch>
        <v-spacer></v-spacer>
        <download-excel
          v-if="mibs.length > 0"
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_MIB.csv"
          :header="exportHeader"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          v-if="mibs.length > 0"
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_MIB.csv"
          :header="exportHeader"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          v-if="mibs.length > 0"
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_MIB.xls"
          :header="exportHeader"
          worksheet="MIB"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn v-if="mibs.length > 0" color="primary" dark @click="copyMIB">
          <v-icon>mdi-copy</v-icon>
          コピー
        </v-btn>
        <v-btn v-if="mibs.length > 0" color="info" dark @click="resultMIBTree">
          <v-icon>mdi-file-tree</v-icon>
          結果MIBツリー
        </v-btn>
        <v-btn color="info" dark @click="mibTreeDialog = true">
          <v-icon>mdi-file-tree</v-icon>
          MIBツリー
        </v-btn>
        <v-btn v-if="mibget.Name" color="primary" dark @click="doMIBGet">
          <v-icon>mdi-file-find</v-icon>
          取得
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="mibTreeDialog" persistent width="70vw">
      <v-card max-height="95%">
        <v-card-title> MIBツリー </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="mibget.Name"
            label="オブジェクト名"
          ></v-text-field>
          <v-text-field
            v-model="searchMIBTree"
            label="オブジェクト名の検索"
          ></v-text-field>
          <div style="height: 350px; overflow: auto">
            <v-treeview
              ref="tree"
              :items="mibtree"
              item-key="oid"
              :search="searchMIBTree"
              hoverable
              activatable
              dense
              :open.sync="conf.mibTreeOpen"
              @update:active="selectMIB"
            >
              <template #prepend="{ item, open }">
                <v-icon v-if="item.children.length > 0">
                  {{ open ? 'mdi-folder-open' : 'mdi-folder' }}
                </v-icon>
                <v-icon v-else :color="getIconColor(item.MIBInfo)">
                  {{ getMIBIcon(item.MIBInfo) }}
                </v-icon>
              </template>
              <template #label="{ item }">
                {{
                  item.MIBInfo
                    ? `${item.name}(${item.oid}: ${item.MIBInfo.Type} )`
                    : `${item.name}(${item.oid})`
                }}
              </template>
            </v-treeview>
          </div>
          <div style="height: 160px; overflow: auto">
            <pre style="margin: 10px; background-color: #333">{{
              mibInfoText
            }}</pre>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="searchMIBTree.length > 2"
            color="normal"
            dark
            @click="openCloseMibTree"
          >
            <v-icon>mdi-file-tree</v-icon>
            {{ mibTreeOpened ? 'MIBツリーを閉じる' : 'MIBツリーを開く' }}
          </v-btn>
          <v-btn color="primary" dark @click="doMIBGet">
            <v-icon>mdi-file-find</v-icon>
            取得
          </v-btn>
          <v-btn color="normal" dark @click="mibTreeDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="resultMibTreeDialog" persistent width="70vw">
      <v-card max-height="95%">
        <v-card-title> 取得結果MIBツリー </v-card-title>
        <v-card-text v-if="resultMibTreeWait">
          <v-progress-linear
            v-model="resultMibTreeProgress"
            height="50"
            color="teal"
          >
            <template #default>
              分析中.. {{ resultMibTreeProgress.toFixed(2) }} %
            </template>
          </v-progress-linear>
        </v-card-text>
        <v-card-text v-if="!resultMibTreeWait">
          <v-text-field
            v-model="searchResultMIBTree"
            label="オブジェクト名の検索"
          ></v-text-field>
          <div style="height: 350px; overflow: auto">
            <v-treeview
              ref="tree"
              :items="resultMibtree"
              item-key="oid"
              :search="searchResultMIBTree"
              hoverable
              activatable
              dense
              @update:active="selectMIB"
            >
              <template #prepend="{ item, open }">
                <v-icon v-if="item.children.length > 0">
                  {{ open ? 'mdi-folder-open' : 'mdi-folder' }}
                </v-icon>
                <v-icon v-else :color="getIconColor(item.MIBInfo)">
                  {{ getMIBIcon(item.MIBInfo) }}
                </v-icon>
              </template>
              <template #label="{ item }">
                {{
                  item.MIBInfo
                    ? `${item.name}(${item.oid}: ${item.MIBInfo.Type} ) : ${item.count}`
                    : `${item.name}(${item.oid}) : ${item.count}`
                }}
              </template>
            </v-treeview>
          </div>
          <div style="height: 160px; overflow: auto">
            <pre style="margin: 10px; background-color: #333">{{
              mibInfoText
            }}</pre>
          </div>
        </v-card-text>
        <v-card-actions v-if="!resultMibTreeWait">
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doMIBGet">
            <v-icon>mdi-file-find</v-icon>
            取得
          </v-btn>
          <v-btn
            v-if="missingList"
            color="error"
            dark
            @click="missingDialog = true"
          >
            <v-icon>mdi-playlist-check</v-icon>
            不足している拡張MIB
          </v-btn>
          <v-btn color="normal" dark @click="resultMibTreeDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
        <v-card-actions v-if="resultMibTreeWait">
          <v-spacer></v-spacer>
          <v-btn color="error" dark @click="stopResultMibTree = true">
            <v-icon>mdi-cancel</v-icon>
            中止
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="missingDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline">不足している拡張MIB</span>
          <v-spacer></v-spacer>
        </v-card-title>
        <v-data-table
          :headers="missingHeader"
          :items="missingList"
          :items-per-page="10"
          dense
        >
          <template #[`item.actions`]="{ item }">
            <v-icon small @click="searchExtMIB(item)"> mdi-search-web </v-icon>
          </template>
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="missingDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="addPollingDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> ポーリング追加 </v-card-title>
        <v-alert v-model="addPollingError" color="error" dense dismissible>
          ポーリングを追加できませんでした
        </v-alert>
        <v-card-text>
          <v-text-field v-model="polling.Name" label="名前"></v-text-field>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.Level"
                :items="$levelList"
                label="レベル"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Type"
                label="種別"
                readonly
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Mode"
                label="モード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="polling.Params"
                label="取得するMIB"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                v-model="polling.LogMode"
                :items="$logModeList"
                label="ログモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <label>判定スクリプト</label>
          <prism-editor
            v-model="polling.Script"
            class="script"
            :highlight="highlighter"
            line-numbers
          ></prism-editor>
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
            <v-col>
              <v-slider
                v-model="polling.Timeout"
                label="タイムアウト(Sec)"
                class="align-center"
                max="60"
                min="1"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="polling.Timeout"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
            <v-col>
              <v-slider
                v-model="polling.Retry"
                label="リトライ回数"
                class="align-center"
                max="20"
                min="0"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="polling.Retry"
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
          <v-row dense>
            <v-col>
              <label>障害時アクション</label>
              <prism-editor
                v-model="polling.FailAction"
                class="script"
                :highlight="actionHighlighter"
                line-numbers
              ></prism-editor>
            </v-col>
            <v-col>
              <label>復帰時アクション</label>
              <prism-editor
                v-model="polling.RepairAction"
                class="script"
                :highlight="actionHighlighter"
                line-numbers
              ></prism-editor>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="savePolling">
            <v-icon>mdi-content-save</v-icon>
            追加
          </v-btn>
          <v-btn color="normal" dark @click="addPollingDialog = false">
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
      nekoNo: 1,
      overlay: false,
      neko: '/images/neko_anm1.png',
      node: {
        ID: '',
        Name: '',
      },
      mibtree: [],
      mibget: {
        NodeID: '',
        Name: '',
        OID: '',
        Raw: false,
      },
      search: '',
      headers: [],
      items: [],
      mibs: [],
      searchMIBTree: '',
      mibTreeDialog: false,
      resultMibtree: [],
      searchResultMIBTree: '',
      resultMibTreeDialog: false,
      resultMibTreeWait: false,
      resultMibTreeProgress: 0.0,
      stopResultMibTree: false,
      missingDialog: false,
      missingList: [],
      missingHeader: [
        { text: '名前', value: 'name', width: '30%' },
        { text: 'OID', value: 'oid', width: '50%' },
        { text: '件数', value: 'count', width: '10%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      error: false,
      wait: false,
      conf: {
        name: '',
        value: '',
        search: '',
        history: '',
        mibTreeOpen: [],
        itemsPerPage: 15,
      },
      history: [],
      options: {},
      tableMode: false,
      exportHeader: '',
      mibInfoText: '',
      copyError: false,
      copyDone: false,
      polling: {},
      addPollingDialog: false,
      addPollingError: false,
      addPollingDone: false,
      mibTreeOpened: false,
      extractorList: [],
      mode: 'auto',
      modeList: [
        { text: '自動', value: 'auto' },
        { text: 'スカラー', value: 'scalar' },
        { text: 'テーブル', value: 'table' },
      ],
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/mibbr/' + this.$route.params.id)
    if (!r) {
      return
    }
    this.node = r.Node
    this.mibget.NodeID = r.Node.ID
    this.mibtree = r.MIBTree
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
  created() {
    const c = this.$store.state.mibbr.conf
    if (c && c.itemsPerPage) {
      Object.assign(this.conf, c)
      this.history = c.history.split(',')
      this.history = this.history.filter((e) => e !== '')
      if (!this.conf.mibTreeOpen) {
        this.conf.mibTreeOpen = []
      }
    }
  },
  beforeDestroy() {
    this.conf.history = this.history.join(',')
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('mibbr/setConf', this.conf)
  },
  methods: {
    doMIBGet() {
      this.exportHeader =
        'TWSNMP FCで' +
        this.node.Name +
        'から取得した' +
        this.mibget.Name +
        'のMIB情報'
      this.mibTreeDialog = false
      this.resultMibTreeDialog = false
      this.resultMibtree = []
      this.headers = []
      this.items = []
      this.wait = true
      this.error = false
      this.nekoNo = 1
      this.waitAnimation()
      this.$axios
        .post('/api/mibbr', this.mibget)
        .then((r) => {
          this.mibs = r.data
          let i = 1
          this.mibs.forEach((e) => {
            e.Index = i++
          })
          if (!this.isTable()) {
            this.showList()
          } else {
            this.showTable()
          }
          this.wait = false
          this.history = this.history.filter((e) => e !== this.mibget.Name)
          this.history.unshift(this.mibget.Name)
        })
        .catch((e) => {
          this.error = true
          this.wait = false
          this.mibs = []
        })
    },
    isTable() {
      return (
        this.mode === 'table' ||
        (this.mode === 'auto' && this.mibget.Name.endsWith('Table'))
      )
    },
    waitAnimation() {
      if (!this.wait) {
        if (this.error) {
          this.neko = '/images/neko_ng.png'
        } else {
          this.neko = '/images/neko_ok.png'
        }
        setTimeout(() => {
          this.overlay = false
        }, 2000)
        return
      }
      this.overlay = true
      this.neko = '/images/neko_anm' + this.nekoNo + '.png'
      this.nekoNo++
      if (this.nekoNo > 7) {
        this.nekoNo = 1
      }
      this.timer = setTimeout(() => this.waitAnimation(), 200)
    },
    showList() {
      this.tableMode = false
      this.conf.search = ''
      this.headers = [
        { text: 'インデックス', value: 'Index', width: '10%' },
        {
          text: '名前',
          value: 'Name',
          width: '20%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: '値',
          value: 'Value',
          width: '60%',
          filter: (value) => {
            if (!this.conf.value) return true
            return value.includes(this.conf.value)
          },
        },
        { text: '操作', value: 'actions', width: '10%' },
      ]
      if (this.mode === 'auto') {
        this.items = this.mibs
      } else {
        this.items = this.mibs.filter(
          (e) => e.Name.endsWith('.0') && e.Name.split('.').length === 2
        )
      }
    },
    showTable() {
      this.tableMode = true
      const names = []
      const indexes = []
      const rows = []
      this.mibs.forEach((e) => {
        const name = e.Name
        const val = e.Value
        const i = name.indexOf('.')
        if (i > 0) {
          const base = name.substring(0, i)
          const index = name.substring(i + 1)
          if (index === '0') {
            return
          }
          if (!names.includes(base)) {
            names.push(base)
          }
          if (!indexes.includes(index)) {
            indexes.push(index)
            const m = new Map()
            m.set('Index', index)
            rows.push(m)
          }
          const r = indexes.indexOf(index)
          if (r >= 0) {
            rows[r].set(base, val)
          }
        }
      })
      this.headers = [
        {
          text: 'Index',
          value: 'Index',
        },
      ]
      names.forEach((e) => {
        this.headers.push({
          text: e,
          value: e,
        })
      })
      this.items = []
      rows.forEach((e) => {
        const d = {}
        this.headers.forEach((h) => {
          d[h.value] = e.get(h.value) || ''
        })
        this.items.push(d)
      })
    },
    updateTable() {
      if (this.isTable()) {
        this.showTable()
      } else {
        this.showList()
      }
    },
    selectMIB(s) {
      if (s && s.length === 1) {
        this.mibget.OID = s[0]
        this.mibget.Name = this.findMIB(s[0], this.mibtree)
      }
    },
    findMIB(oid, list) {
      for (let i = 0; i < list.length; i++) {
        if (list[i].oid === oid) {
          this.mibInfoText = ''
          if (list[i].MIBInfo) {
            this.mibInfoText += `OID  : ${list[i].MIBInfo.OID}\n`
            this.mibInfoText += `Stats: ${list[i].MIBInfo.Status}\n`
            this.mibInfoText += `Type : ${list[i].MIBInfo.Type}\n`
            this.mibInfoText += list[i].MIBInfo.Units
              ? `Units : ${list[i].MIBInfo.Units}\n`
              : ''
            this.mibInfoText += list[i].MIBInfo.Index
              ? `Index : ${list[i].MIBInfo.Index}\n`
              : ''
            this.mibInfoText += list[i].MIBInfo.Defval
              ? `DefVal : ${list[i].MIBInfo.Defval}\n`
              : ''
            this.mibInfoText += `Description :\n${list[i].MIBInfo.Description}\n`
          }
          return list[i].name
        }
        if (list[i].children) {
          const n = this.findMIB(oid, list[i].children)
          if (n) {
            return n
          }
        }
      }
      return null
    },
    makeExports() {
      const exports = []
      this.items.forEach((e) => {
        if (this.tableMode) {
          if (this.conf.search) {
            const s = Object.values(e).join(' ')
            if (!s.includes(this.conf.search)) {
              return
            }
          }
          exports.push(e)
        } else {
          if (this.conf.name && !e.Name.includes(this.conf.name)) {
            return
          }
          if (this.conf.value && !e.Value.includes(this.conf.value)) {
            return
          }
          exports.push({
            インデックス: e.Index,
            名前: e.Name,
            値: e.Value,
          })
        }
      })
      return exports
    },
    getMIBIcon(mibInfo) {
      if (mibInfo) {
        if (mibInfo.Type.startsWith('Counter')) {
          return 'mdi-counter'
        }
        if (mibInfo.Type.startsWith('ObjectIdent')) {
          return 'mdi-file-tree'
        }
        if (mibInfo.Type.startsWith('Time')) {
          return 'mdi-clock'
        }
        if (mibInfo.Type.startsWith('Int')) {
          return 'mdi-counter'
        }
        if (mibInfo.Type.includes('String')) {
          return 'mdi-code-string'
        }
        if (mibInfo.Type.startsWith('Gau')) {
          return 'mdi-speedometer'
        }
        if (
          mibInfo.Type.startsWith('Trap') ||
          mibInfo.Type.startsWith('Noti')
        ) {
          return 'mdi-alert-circle'
        }
        return 'mdi-information'
      }
      return 'mdi-folder'
    },
    getIconColor(mibInfo) {
      if (mibInfo && mibInfo.Type.startsWith('Noti')) {
        return 'red'
      }
      return ''
    },
    copyMIB() {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const list = []
      const l = []
      if (this.tableMode) {
        this.headers.forEach((h) => {
          l.push(h.value)
        })
        list.push(l.join('\t'))
      } else {
        list.push('インデックス,名前,値')
      }
      this.items.forEach((e) => {
        if (this.tableMode) {
          if (this.conf.search) {
            const s = Object.values(e).join(' ')
            if (!s.includes(this.conf.search)) {
              return
            }
          }
          l.length = 0
          this.headers.forEach((h) => {
            l.push(e[h.value])
          })
          list.push(l.join('\t'))
        } else {
          if (this.conf.name && !e.Name.includes(this.conf.name)) {
            return
          }
          if (this.conf.value && !e.Value.includes(this.conf.value)) {
            return
          }
          list.push([e.Index, e.Name, e.Value].join('\t'))
        }
      })
      const s = list.join('\n')
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    copyOneMIB(e) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      navigator.clipboard.writeText(e.Name + '=' + e.Value).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    addPolling(e) {
      const n = e.Name.split('.')[0] || ''
      if (n === '') {
        this.addPollingError = true
        return
      }
      this.polling = {
        ID: '',
        Name: e.Name + 'の監視',
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'get',
        Params: e.Name,
        Filter: '',
        Extractor: '',
        Script: n + '==' + e.Value,
        Level: 'off',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.addPollingDialog = true
    },
    savePolling() {
      this.$axios
        .post('/api/polling/add', this.polling)
        .then(() => {
          this.addPollingDialog = false
          this.addPollingDone = true
        })
        .catch((e) => {
          this.addPollingError = true
        })
    },
    highlighter(code) {
      return highlight(code, languages.js)
    },
    actionHighlighter(code) {
      return highlight(code, {
        property:
          /[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}/,
        string: /(wol|mail|line|chat|wait|cmd)/,
        number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,
        keyword: /\b(?:false|true|up|down)\b/,
      })
    },
    openCloseMibTree() {
      this.mibTreeOpened = !this.mibTreeOpened
      if (this.$refs.tree) {
        this.$refs.tree.updateAll(this.mibTreeOpened)
      }
    },
    async resultMIBTree() {
      if (this.resultMibtree.length > 0 && !this.stopResultMibTree) {
        this.resultMibTreeDialog = true
        return
      }
      this.stopResultMibTree = false
      this.resultMibTreeDialog = true
      this.resultMibTreeWait = true
      this.resultMibTreeProgress = 0.0
      let i = 0
      const nameMap = new Map()
      const missingMap = new Map()
      for (const e of this.mibs) {
        if (this.stopResultMibTree) {
          break
        }
        if (i % 500 === 0) {
          await new Promise((resolve) => {
            setTimeout(resolve, 0)
          })
          this.resultMibTreeProgress = (100.0 * i) / this.mibs.length
        }
        i++
        const a = e.Name.split('.')
        if (a.length < 2) {
          continue
        }
        const t = this.getTreePath(a[0], this.mibtree)
        if (t) {
          if (this.hasChildren(t[0], this.mibtree)) {
            const mn = t[0] + '.' + a[1]
            if (missingMap.has(mn)) {
              missingMap.set(mn, missingMap.get(mn) + 1)
            } else {
              missingMap.set(mn, 1)
            }
          }
          for (const n of t) {
            if (nameMap.has(n)) {
              nameMap.set(n, nameMap.get(n) + 1)
            } else {
              nameMap.set(n, 1)
            }
          }
        }
      }
      this.resultMibTreeWait = false
      this.resultMibtree = this.makeTreeData(nameMap, this.mibtree)
      this.missingList.length = 0
      missingMap.forEach((v, k) => {
        this.missingList.push({
          name: k,
          oid: this.getOID(k, this.mibtree),
          count: v,
        })
      })
    },
    hasChildren(name, list) {
      for (let i = 0; i < list.length; i++) {
        if (list[i].name === name) {
          return list[i].children.length > 0
        }
        if (list[i].children) {
          if (this.hasChildren(name, list[i].children)) {
            return true
          }
        }
      }
      return false
    },
    getOID(name, list) {
      const a = name.split('.')
      let idx = ''
      if (a.length > 1) {
        name = a[0]
        idx = '.' + a[1]
      }
      for (let i = 0; i < list.length; i++) {
        if (list[i].name === name) {
          return list[i].oid + idx
        }
        if (list[i].children) {
          const oid = this.getOID(name, list[i].children)
          if (oid) {
            return oid + idx
          }
        }
      }
      return ''
    },
    searchExtMIB(item) {
      const oid = item.oid.startsWith('.')
        ? item.oid.replace('.', '')
        : item.oid
      const url =
        'https://mibbrowser.online/mibdb_search.php?search=' +
        oid +
        '&userdropdown=anymatch'
      window.open(url, '_blank')
    },
    getTreePath(name, list) {
      for (let i = 0; i < list.length; i++) {
        if (list[i].name === name) {
          return [name]
        }
        if (list[i].children) {
          const n = this.getTreePath(name, list[i].children)
          if (n) {
            n.push(list[i].name)
            return n
          }
        }
      }
      return undefined
    },
    makeTreeData(nameMap, list) {
      const r = []
      for (let i = 0; i < list.length; i++) {
        if (nameMap.has(list[i].name)) {
          const e = {
            name: list[i].name,
            oid: list[i].oid,
            children: [],
            MIBInfo: list[i].MIBInfo,
            count: nameMap.get(list[i].name),
          }
          if (list[i].children) {
            const cl = this.makeTreeData(nameMap, list[i].children)
            if (cl) {
              for (const c of cl) {
                e.children.push(c)
              }
            }
          }
          r.push(e)
        }
      }
      return r
    },
  },
}
</script>

<style>
.mibbr td {
  word-wrap: break-word;
}
</style>
