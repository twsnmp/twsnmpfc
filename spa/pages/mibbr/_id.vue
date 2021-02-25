<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        MIBブラウザー - {{ node.Name }} - {{ mibget.Name }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-alert v-model="error" type="error" dense dismissible>
        MIBを取得できませんでした
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="mibs"
        :search="search"
        dense
        :loading="$fetchState.pending || wait"
        loading-text="Loading... Please wait"
        class="log"
      >
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="mibGetDialog = true">
          <v-icon>mdi-file-find</v-icon>
          取得
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="mibGetDialog" persistent width="600px">
      <v-card max-height="90%">
        <v-card-title>
          <span class="headline">取得したMIBの選択</span>
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
          <v-btn color="normal" dark @click="mibGetDialog = false">
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
    const r = await this.$axios.$get('/api/mibbr/' + this.$route.params.id)
    if (!r) {
      return
    }
    this.node = r.Node
    this.mibget.NodeID = r.Node.ID
    this.mibtree = r.MIBTree
  },
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
      headers: [
        { text: 'Index', value: 'Index' },
        { text: '名前', value: 'Name' },
        { text: '値', value: 'Value' },
      ],
      mibs: [],
      searchMIBTree: '',
      mibGetDialog: false,
      error: false,
      wait: false,
    }
  },
  methods: {
    doMIBGet() {
      this.mibGetDialog = false
      this.wait = true
      this.$axios
        .post('/api/mibbr', this.mibget)
        .then((r) => {
          this.mibs = r.data
          let i = 1
          this.mibs.forEach((e) => {
            e.Index = i++
          })
          this.wait = false
        })
        .catch((e) => {
          this.error = true
          this.wait = false
          this.mibs = []
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
