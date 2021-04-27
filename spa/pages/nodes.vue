<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        ノードリスト
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        ノードを削除できませんでした
      </v-alert>
      <v-alert v-model="updateError" color="error" dense dismissible>
        ノードを変更できませんでした
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="nodes"
        dense
        :items-per-page="15"
        sort-by="State"
        sort-asec
      >
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getIconName(item.Icon)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon
            small
            @click="$router.push({ path: '/node/polling/' + item.ID })"
          >
            mdi-lan-check
          </v-icon>
          <v-icon small @click="$router.push({ path: '/node/log/' + item.ID })">
            mdi-calendar-check
          </v-icon>
          <v-icon small @click="editNodeFunc(item)"> mdi-pencil </v-icon>
          <v-icon small @click="deleteNodeFunc(item)"> mdi-delete </v-icon>
        </template>
        <template v-slot:[`body.append`]>
          <tr>
            <td>
              <v-select v-model="state" :items="stateList" label="state">
              </v-select>
            </td>
            <td>
              <v-text-field v-model="name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="ip" label="ip"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="descr" label="descr"></v-text-field>
            </td>
            <td></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="checkAllPolling">
          <v-icon>mdi-cached</v-icon>
          すべて再確認
        </v-btn>
        <download-excel
          :data="nodes"
          type="csv"
          name="TWSNMP_FC_Node_List.csv"
          header="TWSNMP FC Node List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="nodes"
          type="xls"
          name="TWSNMP_FC_Node_List.xls"
          header="TWSNMP FC Node List"
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
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
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
          text: '名前',
          value: 'Name',
          width: '23%',
          filter: (value) => {
            if (!this.name) return true
            return value.includes(this.name)
          },
        },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '15%',
          filter: (value) => {
            if (!this.ip) return true
            return value.includes(this.ip)
          },
        },
        {
          text: '説明',
          value: 'Descr',
          width: '35%',
          filter: (value) => {
            if (!this.descr) return true
            return value.includes(this.descr)
          },
        },
        { text: '操作', value: 'actions', width: '15%' },
      ],
      nodes: [],
      state: '',
      name: '',
      ip: '',
      descr: '',
      stateList: [
        { text: '', value: '' },
        { text: '重度', value: 'high' },
        { text: '軽度', value: 'low' },
        { text: '注意', value: 'warn' },
        { text: '正常', value: 'normal' },
        { text: '復帰', value: 'repair' },
        { text: '不明', value: 'unknown' },
      ],
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
      this.$axios.post('/api/nodes/delete', [this.deleteNode.ID]).catch((e) => {
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
    async checkAllPolling() {
      await this.$axios.$get('/api/polling/check/all')
      this.$nextTick(() => {
        this.$fetch()
      })
    },
  },
}
</script>
