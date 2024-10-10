<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        gNMIツール - {{ node.Name }}
        <v-spacer></v-spacer>
        <v-text-field v-model="gnmi.Target" label="ターゲット(IP:Port)" />
        <v-combobox
          v-model="gnmi.Path"
          :items="history"
          append-icon="mdi-magnify"
          label="Path"
          dense
        ></v-combobox>
      </v-card-title>
      <v-alert v-model="error" color="error" dense dismissible>
        取得できませんでした
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
        class="gnmi"
        :items-per-page="conf.itemsPerPage"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.path" label="path"></v-text-field>
            </td>
            <td></td>
            <td>
              <v-text-field v-model="conf.value" label="value"></v-text-field>
            </td>
          </tr>
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            v-if="!node.ID.startsWith('NET:')"
            small
            @click="addPolling(item)"
          >
            mdi-card-plus
          </v-icon>
          <v-icon small @click="copyOne(item)"> mdi-content-copy </v-icon>
          <v-icon small @click="setPath(item)">
            mdi-arrow-up-thin-circle-outline
          </v-icon>
        </template>
      </v-data-table>
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          v-if="items.length > 0"
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_gNMI.csv"
          :header="exportHeader"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          v-if="items.length > 0"
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_gNMI.csv"
          :header="exportHeader"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          v-if="items.length > 0"
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_gNMI.xls"
          :header="exportHeader"
          worksheet="MIB"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn v-if="items.length > 0" color="primary" dark @click="copy">
          <v-icon>mdi-copy</v-icon>
          コピー
        </v-btn>
        <v-btn color="info" dark @click="showCapabilities">
          <v-icon>mdi-information-outline</v-icon>
          Capabilities
        </v-btn>
        <v-btn color="info" dark @click="showYangURL">
          <v-icon>mdi-search-web</v-icon>
          YANG情報
        </v-btn>
        <v-btn v-if="gnmi.Path" color="primary" dark @click="doGet">
          <v-icon>mdi-file-find</v-icon>
          取得
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="capabilitiesDialog" persistent width="98vw">
      <v-card max-height="95%">
        <v-card-title>
          Capabilities
          <v-spacer></v-spacer>
          <p class="gnmiCap">
            Version:{{ capabilities.Version }} / Ecnode:{{
              capabilities.Encodings
            }}
          </p>
        </v-card-title>
        <v-data-table
          :headers="capabilitiesHeader"
          :items="capabilities.Models"
          :items-per-page="10"
          dense
        >
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="capabilitiesDialog = false">
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
                label="ターゲット"
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
          <v-text-field v-model="polling.Filter" label="パス"></v-text-field>
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
      gnmi: {
        NodeID: '',
        Target: '',
        Path: '',
        Mode: 'get',
      },
      search: '',
      headers: [
        {
          text: 'Path',
          value: 'Path',
          width: '55%',
          filter: (value) => {
            if (!this.conf.path) return true
            return value.includes(this.conf.path)
          },
        },
        { text: 'Index', value: 'Index', width: '10%' },
        {
          text: '値',
          value: 'Value',
          width: '25%',
          filter: (value) => {
            if (!this.conf.value) return true
            return value.includes(this.conf.value)
          },
        },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      items: [],
      capabilitiesDialog: false,
      capabilities: {
        Version: '',
        Encodings: '',
        Models: '',
      },
      searchCapabilities: '',
      capabilitiesHeader: [
        { text: '識別名', value: 'name', width: '65%' },
        { text: '組織', value: 'organization', width: '25%' },
        { text: 'バージョン', value: 'version', width: '10%' },
      ],
      error: false,
      wait: false,
      conf: {
        path: '',
        value: '',
        search: '',
        history: '',
        itemsPerPage: 15,
      },
      history: [],
      options: {},
      exportHeader: '',
      copyError: false,
      copyDone: false,
      polling: {},
      addPollingDialog: false,
      addPollingError: false,
      addPollingDone: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/gnmi/' + this.$route.params.id)
    if (!r) {
      return
    }
    this.node = r.Node
    this.gnmi.NodeID = r.Node.ID
    this.gnmi.Target = r.Node.IP + ':57400'
  },
  created() {
    const c = this.$store.state.gnmi.conf
    if (c && c.itemsPerPage) {
      Object.assign(this.conf, c)
      this.history = c.history.split(',')
      this.history = this.history.filter((e) => e !== '')
    }
  },
  beforeDestroy() {
    this.conf.history = this.history.join(',')
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('gnmi/setConf', this.conf)
  },
  methods: {
    doGet() {
      this.exportHeader =
        'TWSNMP FCで' +
        this.node.Name +
        'から取得した' +
        this.gnmi.Path +
        'の情報'
      this.capabilitiesDialog = false
      this.items = []
      this.wait = true
      this.error = false
      this.gnmi.Mode = 'Get'
      this.nekoNo = 1
      this.waitAnimation()
      this.$axios
        .post('/api/gnmi', this.gnmi)
        .then((r) => {
          this.items = r.data
          this.wait = false
          this.history = this.history.filter((e) => e !== this.gnmi.Path)
          this.history.unshift(this.gnmi.Path)
        })
        .catch((e) => {
          this.error = true
          this.wait = false
          this.items = []
        })
    },
    showCapabilities() {
      this.wait = true
      this.error = false
      this.gnmi.Mode = 'Capabilities'
      this.nekoNo = 1
      this.waitAnimation()
      this.$axios
        .post('/api/gnmi', this.gnmi)
        .then((r) => {
          this.capabilities = r.data
          this.wait = false
          this.capabilitiesDialog = true
        })
        .catch((e) => {
          this.error = true
          this.wait = false
          this.capabilities = []
        })
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
    makeExports() {
      const exports = []
      this.items.forEach((e) => {
        if (this.conf.path && !e.Path.includes(this.conf.path)) {
          return
        }
        if (this.conf.value && !e.Value.includes(this.conf.value)) {
          return
        }
        exports.push({
          Path: e.Path,
          値: e.Value,
        })
      })
      return exports
    },
    copy() {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const list = []
      list.push('Path\t値')
      this.items.forEach((e) => {
        if (this.conf.path && !e.Path.includes(this.conf.path)) {
          return
        }
        if (this.conf.value && !e.Value.includes(this.conf.value)) {
          return
        }
        list.push([e.Path, e.Value].join('\t'))
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
    copyOne(e) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      navigator.clipboard.writeText(e.Path + '=' + e.Value).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    addPolling(e) {
      if (e.Path === '') {
        this.addPollingError = true
        return
      }
      this.polling = {
        ID: '',
        Name: e.Path + 'の監視',
        NodeID: this.node.ID,
        Type: 'gnmi',
        Mode: 'get',
        Params: this.gnmi.Target,
        Filter: e.Path,
        Extractor: '',
        Script: `
var value = JSON.parse(data);
value == "${e.Value}";`,
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
    setPath(e) {
      this.gnmi.Path = e.Path
    },
    showYangURL() {
      window.open('https://github.com/YangModels/yang', '_blank')
    },
  },
}
</script>

<style>
.gnmi td {
  word-wrap: break-word;
}
.gnmiCap {
  font-size: 12px;
}
</style>
