<template>
  <v-row justify="center">
    <v-card style="width: 100%">
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
      <v-alert v-model="deleteError" type="error" dense dismissible>
        ノードを削除できませんでした
      </v-alert>
      <v-alert v-model="updateError" type="error" dense dismissible>
        ノードを変更できませんでした
      </v-alert>
      <v-data-table :headers="headers" :items="nodes" :search="search" dense>
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getIconName(item.Icon)
          }}</v-icon>
          {{ $getStateName(item.State) }}
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
          <span class="headline">ノード設定</span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="editNode.Name" label="名前"></v-text-field>
          <v-text-field v-model="editNode.IP" label="IPアドレス"></v-text-field>
          <v-select v-model="editNode.Icon" :items="$iconList" label="アイコン">
          </v-select>
          <v-select
            v-model="editNode.AddrMode"
            :items="$addrModeList"
            label="アドレスモード"
          >
          </v-select>
          <v-select
            v-model="editNode.SnmpMode"
            :items="$snmpModeList"
            label="SNMPモード"
          >
          </v-select>
          <v-text-field
            v-model="editNode.Community"
            label="Community"
          ></v-text-field>
          <v-text-field v-model="editNode.User" label="ユーザー"></v-text-field>
          <v-text-field
            v-model="editNode.Password"
            type="password"
            label="パスワード"
          ></v-text-field>
          <v-text-field
            v-model="editNode.PublicKey"
            label="公開鍵"
          ></v-text-field>
          <v-text-field v-model="editNode.URL" label="URL"></v-text-field>
          <v-text-field v-model="editNode.Descr" label="説明"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doUpdateNode">
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
    <v-dialog v-model="deleteDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード削除</span>
        </v-card-title>
        <v-card-text> ノード{{ deleteNode.Name }}を削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteNode">
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
  </v-row>
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
        { text: '状態', value: 'State', width: '10%' },
        { text: '名前', value: 'Name', width: '20%' },
        { text: 'IPアドレス', value: 'IP', width: '10%' },
        { text: 'MACアドレス', value: 'MAC', width: '10%' },
        { text: '説明', value: 'Descr', width: '40%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      nodes: [],
    }
  },
  methods: {
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
