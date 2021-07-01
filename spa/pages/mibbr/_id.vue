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
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="info" dark @click="mibTreeDialog = true">
          <v-icon>mdi-file-tree</v-icon>
          MIBツリー
        </v-btn>
        <v-btn v-if="mibget.Name" color="primary" dark @click="doMIBGet">
          <v-icon>mdi-file-find</v-icon>
          取得
        </v-btn>
        <download-excel
          :data="items"
          type="csv"
          name="TWSNMP_FC_MIB.csv"
          header="TWSNMP FC MIB"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="items"
          type="xls"
          name="TWSNMP_FC_MIB.xls"
          header="TWSNMP FC MIB"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="mibTreeDialog" persistent width="600px">
      <v-card max-height="90%">
        <v-card-title>
          <span class="headline">MIBツリー</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="mibget.Name"
            label="オブジェクト名"
          ></v-text-field>
          <v-text-field
            v-model="searchMIBTree"
            label="オブジェクト名の検索"
          ></v-text-field>
          <div style="height: 300px; overflow: auto">
            <v-treeview
              :items="mibtree"
              item-key="oid"
              :search="searchMIBTree"
              hoverable
              activatable
              dense
              @update:active="selectMIB"
            ></v-treeview>
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
      node: {
        ID: '',
        Name: '',
      },
      mibtree: [],
      mibget: {
        NodeID: '',
        Name: '',
        OID: '',
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
        itemsPerPage: 15,
      },
      history: [],
      options: {},
      tableMode: false,
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
    }
  },
  beforeDestroy() {
    this.conf.history = this.history.join(',')
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('mibbr/setConf', this.conf)
  },
  methods: {
    doMIBGet() {
      this.mibTreeDialog = false
      this.headers = []
      this.items = []
      this.wait = true
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
          if (!this.history.includes(this.mibget.Name)) {
            this.history.push(this.mibget.Name)
          }
        })
        .catch((e) => {
          this.error = true
          this.wait = false
          this.mibs = []
        })
    },
    showList() {
      this.tableMode = false
      this.conf.search = ''
      this.headers = [
        { text: 'Index', value: 'Index' },
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
  },
}
</script>

<style>
.mibbr td {
  word-wrap: break-word;
}
</style>
