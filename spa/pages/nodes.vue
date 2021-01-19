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
      <v-data-table :headers="headers" :items="nodes" :search="search" dense>
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="getStateColor(item.State)">{{
            getIconName(item.Icon)
          }}</v-icon>
          {{ getStateName(item.State) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small class="mr-2" @click="editItem(item)">
            mdi-pencil
          </v-icon>
          <v-icon small @click="deleteItem(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
    </v-card>
    <v-dialog v-model="dialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード編集</span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="editedItem.Name" label="名前"></v-text-field>
          <v-text-field
            v-model="editedItem.IP"
            label="IPアドレス"
          ></v-text-field>
          <v-text-field v-model="editedItem.Descr" label="説明"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="blue darken-1" text @click="close">キャンセル</v-btn>
          <v-btn color="blue darken-1" text @click="save">保存</v-btn>
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
      dialog: false,
      editedIndex: -1,
      editedItem: {
        Name: '',
        Descr: '',
        IP: '',
      },
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
    editItem(item) {
      this.editedIndex = this.nodes.indexOf(item)
      this.editedItem = Object.assign({}, item)
      this.dialog = true
    },

    deleteItem(item) {
      const index = this.nodes.indexOf(item)
      confirm('このノードを削除しますか?') && this.nodes.splice(index, 1)
    },

    close() {
      this.dialog = false
      this.$nextTick(() => {
        this.editedIndex = -1
      })
    },

    save() {
      if (this.editedIndex > -1) {
        Object.assign(this.nodes[this.editedIndex], this.editedItem)
      }
      this.close()
    },
  },
}
</script>
