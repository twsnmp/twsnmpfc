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
      </v-data-table>
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
      <v-card-actions>
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
    <v-dialog v-model="mibTreeDialog" persistent width="800px">
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
  </v-row>
</template>

<script>
export default {
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
          if (!this.mibget.Name.includes('Table')) {
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
        { text: 'インデックス', value: 'Index' },
        {
          text: '名前',
          value: 'Name',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: '値',
          value: 'Value',
          filter: (value) => {
            if (!this.conf.value) return true
            return value.includes(this.conf.value)
          },
        },
      ]
      this.items = this.mibs
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
          if (!names.includes(base)) {
            names.push(base)
          }
          if (!indexes.includes(index)) {
            indexes.push(index)
            rows.push([index])
          }
          const r = indexes.indexOf(index)
          if (r >= 0) {
            rows[r].push(val)
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
        const d = { Index: e[0] }
        for (let i = 1; i < e.length; i++) {
          d[names[i - 1]] = e[i]
        }
        this.items.push(d)
      })
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
  },
}
</script>

<style>
.mibbr td {
  word-wrap: break-word;
}
</style>
