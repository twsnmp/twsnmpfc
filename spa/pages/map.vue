<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        イベントログ
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="map.Logs"
        :search="search"
        sort-by="TimeStr"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template v-slot:[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
      </v-data-table>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" type="error" dense dismissible>
          ノードを削除できませんでした
        </v-alert>
        <v-card-text> 選択したノードを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteNode">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editNodeDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード設定</span>
        </v-card-title>
        <v-alert v-model="editNodeError" type="error" dense dismissible>
          ノードの保存に失敗しました
        </v-alert>
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
          <v-btn color="normal" dark @click="editNodeDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="lineDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ライン編集</span>
        </v-card-title>
        <v-alert v-model="lineError" type="error" dense dismissible>
          ラインの保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-text-field
            v-model="editLine.NodeName1"
            label="ノード１"
            disabled
          ></v-text-field>
          <v-select
            v-model="editLine.PollingID1"
            :items="pollingList1"
            label="ポーリング"
          >
          </v-select>
          <v-text-field
            v-model="editLine.NodeName2"
            label="ノード２"
            disabled
          ></v-text-field>
          <v-select
            v-model="editLine.PollingID2"
            :items="pollingList2"
            label="ポーリング"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" dark @click="deleteLine">
            <v-icon>mdi-lan-disconnect</v-icon>
            切断
          </v-btn>
          <v-btn color="primary" dark @click="addLine">
            <v-icon>mdi-lan-connect</v-icon>
            接続
          </v-btn>
          <v-btn color="normal" dark @click="lineDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="showNodeDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード情報</span>
        </v-card-title>
        <v-simple-table dense>
          <template v-slot:default>
            <thead>
              <tr>
                <th class="text-left">項目</th>
                <th class="text-left">値</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>名前</td>
                <td>{{ editNode.Name }}</td>
              </tr>
              <tr>
                <td>IPアドレス</td>
                <td>{{ editNode.IP }}</td>
              </tr>
              <tr>
                <td>MACアドレス</td>
                <td>{{ editNode.MAC }}</td>
              </tr>
              <tr>
                <td>説明</td>
                <td>{{ editNode.Descr }}</td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="copyNode">
            <v-icon>mdi-content-copy</v-icon>
            コピー
          </v-btn>
          <v-btn color="normal" dark @click="showMIBBr">
            <v-icon>mdi-eye</v-icon>
            MIBブラウザー
          </v-btn>
          <v-btn color="primary" dark @click="showNodeInfoPage">
            <v-icon>mdi-eye</v-icon>
            詳細...
          </v-btn>
          <v-btn color="primary" dark @click="showEditNodeDialog">
            <v-icon>mdi-pencil</v-icon>
            編集...
          </v-btn>
          <v-btn color="error" dark @click="deleteNode">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" dark @click="showNodeDialog = false">
            <v-icon>mdi-close</v-icon>
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
    this.map = await this.$axios.$get('/api/map')
    this.$setMAP(this.map, this.$axios.defaults.baseURL)
    this.$store.commit('map/setMAP', this.map)
    this.map.Logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
  },
  data() {
    return {
      showNodeDialog: false,
      editNodeDialog: false,
      editNodeError: false,
      lineDialog: false,
      lineError: false,
      deleteDialog: false,
      deleteError: false,
      selectedNodeID: '',
      showNode: {},
      editLine: {
        NodeID1: '',
        NodeID2: '',
        PollingID1: '',
        PollingID2: '',
      },
      editNode: {},
      nodeList: [],
      deleteNodes: [],
      map: {
        Nodes: {},
        Pollings: [],
        Lines: [],
        MapConf: { MapName: '' },
        Logs: [],
      },
      search: '',
      headers: [
        { text: '状態', value: 'Level', width: '10%' },
        { text: '発生日時', value: 'TimeStr', width: '15%' },
        { text: '種別', value: 'Type', width: '10%' },
        { text: '関連ノード', value: 'NodeName', width: '15%' },
        { text: 'イベント', value: 'Event', width: '50%' },
      ],
    }
  },
  computed: {
    pollingList1() {
      return this.pollingList(this.editLine.NodeID1)
    },
    pollingList2() {
      return this.pollingList(this.editLine.NodeID2)
    },
  },
  mounted() {
    this.$setIconCodeMap(this.$iconList)
    this.$setStateColorMap(this.$stateList)
    this.$setCallback(this.callback)
    this.$showMAP('map')
  },
  methods: {
    nodeName(id) {
      return this.map.Nodes[id] ? this.map.Nodes[id].Name : ''
    },
    pollingList(id) {
      const l = []
      if (!this.map.Nodes[id]) {
        return l
      }
      this.map.Pollings[id].forEach((p) => {
        l.push({
          text: p.Name,
          value: p.ID,
        })
      })
      return l
    },
    callback(r) {
      if (
        this.deleteDialog ||
        this.showNodeDialog ||
        this.lineDialog ||
        this.editNodeDialog
      ) {
        return
      }
      switch (r.Cmd) {
        case 'updateNodesPos':
          this.$axios.post('/api/map/update', r.Param)
          break
        case 'deleteNodes':
          this.deleteNodes = Array.from(r.Param)
          this.deleteDialog = true
          break
        case 'editLine':
          this.showEditLineDiaglog(r.Param)
          break
        case 'refresh':
          this.$fetch()
          break
        case 'showNode':
          if (!this.map.Nodes[r.Param]) {
            return
          }
          this.editNode = this.map.Nodes[r.Param]
          this.showNodeDialog = true
          break
        case 'addNode':
          this.editNode = r.Param
          this.editNodeDialog = true
          break
      }
    },
    showEditLineDiaglog(p) {
      if (p.length !== 2 || !this.map.Nodes[p[0]] || !this.map.Nodes[p[1]]) {
        return
      }
      const l = this.map.Lines.find(
        (e) =>
          (e.NodeID1 === p[0] && e.NodeID2 === p[1]) ||
          (e.NodeID1 === p[0] && e.NodeID2 === p[1])
      )
      this.editLine = l || {
        NodeID1: p[0],
        PollingID2: '',
        NodeID2: p[1],
        PollingID1: '',
      }
      this.editLine.NodeName1 = this.nodeName(this.editLine.NodeID1)
      this.editLine.NodeName2 = this.nodeName(this.editLine.NodeID2)
      this.lineDialog = true
    },
    doDeleteNode() {
      this.$axios.post('/api/map/delete', this.deleteNodes).then(() => {
        this.$fetch()
      })
      this.deleteDialog = false
    },
    doUpdateNode() {
      this.$axios
        .post('/api/node/update', this.editNode)
        .then(() => {
          this.$fetch()
          this.editNodeDialog = false
        })
        .catch((e) => {
          this.editNodeError = true
          this.$fetch()
        })
    },
    deleteNode() {
      this.showNodeDialog = false
      this.deleteNodes = [this.editNode.ID]
      this.deleteDialog = true
    },
    copyNode() {
      this.showNodeDialog = false
      // 位置をずらして新規追加
      this.editNode.X += 64
      this.editNode.ID = ''
      this.editNode.Name += 'のコピー'
      this.$axios
        .post('/api/node/update', this.editNode)
        .then(() => {
          this.$fetch()
        })
        .catch((e) => {
          this.$fetch()
        })
    },
    showEditNodeDialog() {
      this.showNodeDialog = false
      this.editNodeDialog = true
    },
    showNodeInfoPage() {
      this.$router.push({ path: '/node/' + this.editNode.ID })
    },
    showMIBBr() {
      this.$router.push({ path: '/mibbr/' + this.editNode.ID })
    },
    addLine() {
      this.$axios
        .post('/api/line/add', this.editLine)
        .then(() => {
          this.$fetch()
          this.lineDialog = false
        })
        .catch((e) => {
          this.lineError = true
          this.$fetch()
        })
    },
    deleteLine() {
      this.$axios
        .post('/api/line/delete', this.editLine)
        .then(() => {
          this.$fetch()
          this.lineDialog = false
        })
        .catch((e) => {
          this.lineError = true
          this.$fetch()
        })
    },
  },
}
</script>
