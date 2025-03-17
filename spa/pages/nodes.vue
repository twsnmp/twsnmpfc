<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
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
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getIconName(item.Icon)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            small
            @click="$router.push({ path: '/node/polling/' + item.ID })"
          >
            mdi-lan-check
          </v-icon>
          <v-icon small @click="$router.push({ path: '/node/log/' + item.ID })">
            mdi-calendar-check
          </v-icon>
          <v-icon small @click="$router.push({ path: '/ping/' + item.IP })">
            mdi-check-network
          </v-icon>
          <v-icon
            v-if="item.Community || item.User"
            small
            @click="$router.push({ path: '/mibbr/' + item.ID })"
          >
            mdi-eye
          </v-icon>
          <v-icon
            v-if="item.GNMIUser"
            small
            @click="$router.push({ path: '/gnmi/' + item.ID })"
          >
            mdi-eye-check
          </v-icon>
          <v-icon
            small
            @click="$router.push({ path: '/node/vpanel/' + item.ID })"
          >
            mdi-apps-box
          </v-icon>
          <v-icon
            v-if="item.Community || item.User"
            small
            @click="$router.push({ path: '/node/hostResource/' + item.ID })"
          >
            mdi-gauge
          </v-icon>
          <v-icon
            v-if="item.Community || item.User"
            small
            @click="$router.push({ path: '/node/rmon/' + item.ID })"
          >
            mdi-minus-network
          </v-icon>
          <v-icon small @click="editNodeFunc(item)"> mdi-pencil </v-icon>
          <v-icon small color="red" @click="deleteNodeFunc(item)">
            mdi-delete
          </v-icon>
          <v-icon v-if="item.MAC" small @click="doWOL(item.ID)">
            mdi-alarm
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-select v-model="conf.state" :items="stateList" label="state">
              </v-select>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ip" label="ip"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.descr" label="descr"></v-text-field>
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
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Node_List.csv"
          header="TWSNMP FCで管理するノードリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_Node_List.csv"
          header="TWSNMP FCで管理するノードリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_Node_List.xls"
          header="TWSNMP FCで管理するノードリスト"
          worksheet="ノード"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="info" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-list</v-icon>
              マップ表示項目
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="networkDialog = true">
              <v-list-item-icon><v-icon>mdi-lan</v-icon> </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ネットワーク </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="itemDialog = true">
              <v-list-item-icon><v-icon>mdi-drawing</v-icon> </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 描画アイテム </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="lineDialog = true">
              <v-list-item-icon><v-icon>mdi-link</v-icon> </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ライン </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-snackbar v-model="wolError" absolute centered color="error">
          Wake on LANパケットを送信できません
        </v-snackbar>
        <v-snackbar v-model="wolDone" absolute centered color="primary">
          Wake on LANパケットを送信しました
        </v-snackbar>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> ノード設定 </v-card-title>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field v-model="editNode.Name" label="名前"></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.IP"
                label="IPアドレス"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.MAC"
                label="MACアドレス"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                v-model="editNode.AddrMode"
                :items="$addrModeList"
                label="アドレスモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editNode.Icon"
                :items="$iconList"
                label="アイコン"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-icon x-large style="margin-top: 10px; margin-left: 10px">
                {{ $getIconName(editNode.Icon) }}
              </v-icon>
            </v-col>
            <v-col>
              <v-switch
                v-model="editNode.AutoAck"
                label="復帰時に自動確認"
                dense
              >
              </v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="8">
              <v-autocomplete
                v-model="editNode.Image"
                :items="imageIconList"
                label="イメージ"
                dense
                @change="selectImageIcon"
              >
              </v-autocomplete>
            </v-col>
            <v-col></v-col>
            <v-col>
              <v-img v-if="editNode.Image" :width="48" :src="imageIcon"></v-img>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editNode.SnmpMode"
                :items="$snmpModeList"
                label="SNMPモード"
              >
              </v-select>
            </v-col>
            <v-col v-if="editNode.SnmpMode == ''">
              <v-text-field
                v-model="editNode.Community"
                label="Community"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editNode.User"
                autocomplete="username"
                label="ユーザー"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.Password"
                autocomplete="new-password"
                type="password"
                label="パスワード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-text-field
            v-model="editNode.PublicKey"
            label="公開鍵"
          ></v-text-field>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editNode.GNMIPort"
                label="gNMI Port"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.GNMIEncoding"
                label="gNMI Encoding"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.GNMIUser"
                autocomplete="username"
                label="gNMI ユーザー"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.GNMIPassword"
                autocomplete="new-password"
                type="password"
                label="gNMI パスワード"
              ></v-text-field>
            </v-col>
          </v-row>
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
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
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
    <v-dialog v-model="deleteSelectedDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">{{ deleteTypeName }}削除</span>
        </v-card-title>
        <v-card-text>
          選択した{{ deleteTypeName }}を削除しますか？
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteSelected">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteSelectedDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="networkDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline">ネットワーク</span>
          <v-spacer></v-spacer>
        </v-card-title>
        <v-data-table
          v-model="deleteList"
          :headers="headersNetworks"
          :items="networks"
          show-select
          item-key="ID"
          :items-per-page="20"
          dense
          :footer-props="{
            'items-per-page-options': [10, 20, 30],
          }"
        >
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="deleteList.length > 0"
            color="error"
            @click="deleteMapItem('network')"
          >
            <v-icon>mdi-trash</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="networkDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="itemDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline">描画アイテム</span>
          <v-spacer></v-spacer>
        </v-card-title>
        <v-data-table
          v-model="deleteList"
          :headers="headersItem"
          :items="items"
          show-select
          item-key="ID"
          :items-per-page="20"
          dense
          :footer-props="{
            'items-per-page-options': [10, 20, 30],
          }"
        >
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="deleteList.length > 0"
            color="error"
            @click="deleteMapItem('item')"
          >
            <v-icon>mdi-trash</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="itemDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="lineDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline">ライン</span>
          <v-spacer></v-spacer>
        </v-card-title>
        <v-data-table
          v-model="deleteList"
          :headers="headersLine"
          :items="lines"
          show-select
          item-key="ID"
          :items-per-page="20"
          dense
          :footer-props="{
            'items-per-page-options': [10, 20, 30],
          }"
        >
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="deleteList.length > 0"
            color="error"
            @click="deleteMapItem('line')"
          >
            <v-icon>mdi-trash</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="lineDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
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
          width: '8%',
          filter: (value) => {
            if (this.conf.state === '') return true
            const l = this.$levelNum(value)
            return this.conf.state >= 4
              ? this.conf.state === l
              : this.conf.state >= l
          },
          sort: (a, b) => {
            const al = this.$levelNum(a)
            const bl = this.$levelNum(b)
            return al - bl
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '15%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '12%',
          filter: (value) => {
            if (!this.conf.ip) return true
            return value.includes(this.conf.ip)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: 'MACアドレス',
          value: 'MAC',
          width: '25%',
          filter: (value) => {
            if (!this.conf.mac) return true
            return value.includes(this.conf.mac)
          },
        },
        {
          text: '説明',
          value: 'Descr',
          width: '25%',
          filter: (value) => {
            if (!this.conf.descr) return true
            return value.includes(this.conf.descr)
          },
        },
        { text: '操作', value: 'actions', width: '18%' },
      ],
      nodes: [],
      conf: {
        state: '',
        name: '',
        ip: '',
        mac: '',
        descr: '',
        sortBy: 'State',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      stateList: [
        { text: 'すべて', value: '' },
        { text: '重度', value: 0 },
        { text: '軽度以上', value: 1 },
        { text: '注意以上', value: 2 },
        { text: '復帰以上', value: 3 },
        { text: '正常', value: 4 },
        { text: '不明', value: 5 },
      ],
      wolError: false,
      wolDone: false,
      imageIconList: [],
      imageIcon: '',
      networkDialog: false,
      networks: [],
      headersNetworks: [
        {
          text: '名前',
          value: 'Name',
          width: '20%',
        },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '10%',
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: '説明',
          value: 'Descr',
          width: '70%',
        },
      ],
      items: [],
      itemDialog: false,
      headersItem: [
        {
          text: '種別',
          value: 'TypeName',
          width: '20%',
        },
        {
          text: 'X',
          value: 'X',
          width: '10%',
        },
        {
          text: 'Y',
          value: 'Y',
          width: '10%',
        },
        {
          text: '文字列',
          value: 'Info',
          width: '70%',
        },
      ],
      drawItemList: [
        '矩形',
        '楕円',
        'ラベル',
        'イメージ',
        'ポーリング結果(テキスト)',
        'ポーリング結果(ゲージ)',
        'ポーリング結果(新ゲージ)',
        'ポーリング結果(バー)',
        'ポーリング結果(ライン)',
      ],
      lines: [],
      lineDialog: false,
      headersLine: [
        {
          text: 'ノード1',
          value: 'Node1Name',
          width: '25%',
        },
        {
          text: 'ポーリング1',
          value: 'Polling1Name',
          width: '25%',
        },
        {
          text: 'ノード2',
          value: 'Node2Name',
          width: '25%',
        },
        {
          text: 'ポーリング2',
          value: 'Polling2Name',
          width: '25%',
        },
      ],
      deleteList: [],
      deleteType: '',
      deleteTypeName: '',
      deleteSelectedDialog: false,
    }
  },
  async fetch() {
    const map = await this.$axios.$get('/api/map')
    if (!map) {
      return
    }
    this.nodes = []
    for (const k in map.Nodes) {
      this.nodes.push(map.Nodes[k])
    }
    this.networks = []
    for (const k in map.Networks) {
      this.networks.push(map.Networks[k])
    }
    this.items = []
    for (const id in map.Items) {
      this.items.push({
        ID: id,
        X: map.Items[id].X,
        Y: map.Items[id].Y,
        TypeName:
          map.Items[id].Type >= 0 && map.Items[id].Type < 9
            ? this.drawItemList[map.Items[id].Type]
            : '不明',
        Info:
          map.Items[id].Text || map.Items[id].Path || map.Items[id].PollingID,
      })
    }
    this.lines = []
    for (const l of map.Lines) {
      this.lines.push({
        ID: l.ID,
        Node1Name: this.getNodeName(l.NodeID1, map),
        Polling1Name: this.getPollingName(l.NodeID1, l.PollingID1, map),
        Node2Name: this.getNodeName(l.NodeID2, map),
        Polling2Name: this.getPollingName(l.NodeID2, l.PollingID2, map),
      })
    }
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
    this.getImageIconList()
  },
  created() {
    const c = this.$store.state.nodes.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('nodes/setConf', this.conf)
  },
  methods: {
    editNodeFunc(item) {
      this.editIndex = this.nodes.indexOf(item)
      this.editNode = Object.assign({}, item)
      this.selectImageIcon()
      this.editDialog = true
    },
    deleteNodeFunc(item) {
      this.deleteIndex = this.nodes.indexOf(item)
      this.deleteNode = Object.assign({}, item)
      this.deleteDialog = true
    },
    doDeleteNode() {
      this.deleteError = false
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
      this.updateError = false
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
    makeExports() {
      const exports = []
      this.nodes.forEach((e) => {
        if (!this.filterNode(e)) {
          return
        }
        exports.push({
          名前: e.Name,
          IPアドレス: e.IP,
          MACアドレス: e.MAC,
          状態: this.$getStateName(e.State),
          説明: e.Descr,
        })
      })
      return exports
    },
    filterNode(e) {
      if (this.conf.state && this.conf.state !== e.State) {
        return false
      }
      if (this.conf.name && !e.Name.includes(this.conf.name)) {
        return false
      }
      if (this.conf.ip && !e.IP.includes(this.conf.ip)) {
        return false
      }
      if (this.conf.mac && !e.MAC.includes(this.conf.mac)) {
        return false
      }
      if (this.conf.descr && !e.Descr.includes(this.conf.descr)) {
        return false
      }
      return true
    },
    doWOL(id) {
      this.$axios
        .post('/api/wol/' + id)
        .then(() => {
          this.wolDone = true
          this.$fetch()
        })
        .catch((e) => {
          this.wolError = true
        })
    },
    async getImageIconList() {
      if (this.imageIconList.length > 0) {
        return
      }
      const l = await this.$axios.$get('/api/imageIconList')
      this.imageIconList = [{ text: 'なし', value: '' }]
      for (const i of l) {
        this.imageIconList.push({
          text: i,
          value: i,
        })
      }
    },
    selectImageIcon() {
      this.imageIcon = ''
      if (!this.editNode.Image) {
        return
      }
      const url =
        this.$axios.defaults.baseURL && this.$axios.defaults.baseURL !== '/'
          ? this.$axios.defaults.baseURL
          : window.location.origin
      this.imageIcon = url + '/imageIcon/' + this.editNode.Image
    },
    deleteMapItem(t) {
      this.deleteType = t
      switch (t) {
        case 'network':
          this.networkDialog = false
          this.deleteTypeName = 'ネットワーク'
          break
        case 'line':
          this.lineDialog = false
          this.deleteTypeName = 'ライン'
          break
        case 'item':
          this.itemDialog = false
          this.deleteTypeName = '描画アイテム'
          break
        default:
          return
      }
      if (this.deleteList.length > 0) {
        this.deleteSelectedDialog = true
      }
    },
    async doDeleteSelected() {
      switch (this.deleteType) {
        case 'network':
          for (const n of this.deleteList) {
            await this.$axios.delete('/api/network/' + n.ID)
          }
          break
        case 'line':
          for (const l of this.deleteList) {
            await this.$axios.delete('/api/line/' + l.ID)
          }
          break
        case 'item':
          await this.$axios.post(
            '/api/nodes/delete_items',
            this.deleteList.map((e) => e.ID)
          )
          break
      }
      this.deleteSelectedDialog = false
      this.deleteList = []
      this.$fetch()
    },
    getNodeName(id, map) {
      if (id.startsWith('NET:')) {
        const a = id.split(':')
        const net = map.Networks[a[1]]
        if (net) {
          return net.Name
        }
      }
      const node = map.Nodes[id]
      if (node) {
        return node.Name
      }
      return ''
    },
    getPollingName(nid, pid, map) {
      const pollings = map.Pollings[nid]
      if (pollings) {
        for (const polling of pollings) {
          if (pid === polling.ID) {
            return polling.Name
          }
        }
      }
      return pid
    },
  },
}
</script>
