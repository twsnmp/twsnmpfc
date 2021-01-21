<template>
  <div>
    <v-card>
      <v-card-title>
        ノードリスト
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
        ノードを削除できませんでした
      </v-alert>
      <v-alert :value="updateError" type="error" dense dismissible>
        ノードを変更できませんでした
      </v-alert>
      <v-data-table :headers="headers" :items="nodes" :search="search" dense>
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="getStateColor(item.State)">{{
            getIconName(item.Icon)
          }}</v-icon>
          {{ getStateName(item.State) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small class="mr-2" @click="editNodeFunc(item)">
            mdi-pencil
          </v-icon>
          <v-icon small @click="deleteNodeFunc(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
    </v-card>
    <v-dialog v-model="editDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード編集</span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="editNode.Name" label="名前"></v-text-field>
          <v-text-field v-model="editNode.IP" label="IPアドレス"></v-text-field>
          <v-text-field v-model="editNode.Descr" label="説明"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="closeEdit">キャンセル</v-btn>
          <v-btn color="primary" dark @click="doUpdateNode">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード削除</span>
        </v-card-title>
        <v-card-text> ノード{{ deleteNode.Name }}を削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="closeDelete">キャンセル</v-btn>
          <v-btn color="error" @click="doDeleteNode">削除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
export default {
  async fetch() {
    this.nodes = await this.$axios.$get('/api/nodes')
  },
  data() {
    return {
      editDialog: false,
      deleteDialog: false,
      editIndex: -1,
      deleteIndex: -1,
      deleteError: false,
      updateError: false,
      editNode: {},
      deleteNode: {},
      search: '',
      headers: [
        {
          text: '状態',
          value: 'State',
        },
        { text: '名前', value: 'Name' },
        { text: 'IPアドレス', value: 'IP' },
        { text: 'MACアドレス', value: 'MAC' },
        { text: '説明', value: 'Descr' },
        { text: '操作', value: 'actions' },
      ],
      nodes: [],
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
    getIconName(icon) {
      switch (icon) {
        case 'desktop':
          return 'mdi-desktop-mac'
        case 'hdd':
          return 'mdi-router-network'
        default:
          return 'comment-question-outline'
      }
    },
    editNodeFunc(item) {
      this.editIndex = this.nodes.indexOf(item)
      this.editNode = Object.assign({}, item)
      this.editDialog = true
    },
    deleteNodeFunc(item) {
      this.deleteIndex = this.nodes.indexOf(item)
      this.deleteNode = Object.assign({}, item)
      this.deleteDialog = true
    },
    doDeleteNode() {
      this.nodes.splice(this.deleteIndex, 1)
      this.$axios
        .post('/api/node/delete', { ID: this.deleteNode.ID })
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
    doUpdateNode() {
      if (this.editIndex > -1) {
        Object.assign(this.nodes[this.editIndex], this.editNode)
        this.$axios.post('/api/node/update', this.editNode).catch((e) => {
          this.updateError = true
          this.$fetch()
        })
      }
      this.closeEdit()
    },
  },
}
</script>
