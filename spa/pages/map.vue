<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
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
        <template #[`item.Level`]="{ item }">
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
        <v-alert v-model="deleteError" color="error" dense dismissible>
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
        <v-alert v-model="editNodeError" color="error" dense dismissible>
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
          <v-text-field
            v-model="editNode.User"
            autocomplete="off"
            label="ユーザー"
          ></v-text-field>
          <v-text-field
            v-model="editNode.Password"
            autocomplete="off"
            type="password"
            label="パスワード"
          ></v-text-field>
          <v-text-field
            v-model="editNode.PublicKey"
            label="公開鍵"
          ></v-text-field>
          <v-text-field v-model="editNode.URL" label="URL"></v-text-field>
          <v-text-field v-model="editNode.Descr" label="説明"></v-text-field>
          <v-switch
            v-if="copyFrom"
            v-model="copyPolling"
            label="ポーリングの設定も含めてコピーする"
          ></v-switch>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doUpdateNode">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn
            color="normal"
            dark
            @click="
              editNodeDialog = false
              $fetch()
            "
          >
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
        <v-alert v-model="lineError" color="error" dense dismissible>
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
          <v-select
            v-model="editLine.PollingID"
            :items="linePollingList"
            label="情報のためのポーリング"
          >
          </v-select>
          <v-text-field v-model="editLine.Info" label="情報"></v-text-field>
          <v-select
            v-model="editLine.Width"
            :items="lineWidthList"
            label="ラインの太さ"
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
    <v-dialog v-model="showNodeDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード情報</span>
        </v-card-title>
        <v-simple-table dense>
          <template #default>
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
                <td>IPv6アドレス</td>
                <td>{{ editNode.IPv6 }}</td>
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
        <v-divider></v-divider>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="copyNode">
            <v-icon>mdi-content-copy</v-icon>
            コピー
          </v-btn>
          <v-btn color="normal" dark @click="showMIBBr">
            <v-icon>mdi-eye</v-icon>
            MIB
          </v-btn>
          <v-btn color="primary" dark @click="showNodePollingPage">
            <v-icon>mdi-lan-check</v-icon>
            ポーリング
          </v-btn>
          <v-btn color="primary" dark @click="showNodeLogPage">
            <v-icon>mdi-calendar-check</v-icon>
            ログ
          </v-btn>
          <v-btn color="primary" dark @click="checkPolling(false)">
            <v-icon>mdi-cached</v-icon>
            再確認
          </v-btn>
          <v-btn color="primary" dark @click="showEditNodeDialog">
            <v-icon>mdi-pencil</v-icon>
            編集
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
    <v-menu
      v-model="showMapContextMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="addNode()">
          <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>新規ノード</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="checkPolling(true)">
          <v-list-item-icon><v-icon>mdi-cached</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>全て再確認</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item :to="discoverURL">
          <v-list-item-icon><v-icon>mdi-file-find</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>自動発見</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item to="/conf/map">
          <v-list-item-icon><v-icon>mdi-cog</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>マップ設定</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-menu
      v-model="showNodeContextMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="editNodeDialog = true">
          <v-list-item-icon><v-icon>mdi-pencil</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>編集</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="checkPolling(false)">
          <v-list-item-icon><v-icon>mdi-cached</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>再確認</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="deleteDialog = true">
          <v-list-item-icon
            ><v-icon color="red">mdi-delete</v-icon></v-list-item-icon
          >
          <v-list-item-content>
            <v-list-item-title>削除</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="copyNode()">
          <v-list-item-icon><v-icon>mdi-content-copy</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>コピー</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showMIBBr()">
          <v-list-item-icon><v-icon>mdi-eye</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>MIBブラウザー</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showNodeDialog = true">
          <v-list-item-icon><v-icon>mdi-information</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>情報</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showNodePollingPage()">
          <v-list-item-icon><v-icon>mdi-lan-check</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>ポーリング</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showNodeLogPage()">
          <v-list-item-icon>
            <v-icon>mdi-calendar-check</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>ログ</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item v-for="(url, i) in urls" :key="i" @click="openURL(url)">
          <v-list-item-icon>
            <v-icon>mdi-link</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ url }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
  </v-row>
</template>

<script>
export default {
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
        PollingID: '',
        Info: '',
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
      showMapContextMenu: false,
      showNodeContextMenu: false,
      x: 0,
      y: 0,
      copyFrom: '',
      copyPolling: false,
      lineWidthList: [
        { text: '1', value: 1 },
        { text: '2', value: 2 },
        { text: '3', value: 3 },
        { text: '4', value: 4 },
        { text: '5', value: 5 },
      ],
      urls: [],
    }
  },
  async fetch() {
    this.map = await this.$axios.$get('/api/map')
    this.$setMAP(this.map, this.$axios.defaults.baseURL)
    this.$store.commit('map/setMAP', this.map)
    this.map.Logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    const nodeID = this.$route.query.node
    if (nodeID && this.map.Nodes[nodeID]) {
      this.$selectNode(nodeID)
    }
  },
  computed: {
    pollingList1() {
      return this.pollingList(this.editLine.NodeID1, false)
    },
    pollingList2() {
      return this.pollingList(this.editLine.NodeID2, false)
    },
    linePollingList() {
      const l1 = [{ text: '設定しない', value: '' }]
      const l2 = this.pollingList(this.editLine.NodeID1, true)
      const l3 = this.pollingList(this.editLine.NodeID2, true)
      return l1.concat(l2, l3)
    },
    discoverURL() {
      return `/discover?x=${this.x}&y=${this.y}`
    },
  },
  mounted() {
    this.$setIconCodeMap(this.$iconList)
    this.$setStateColorMap(this.$stateList)
    this.$setCallback(this.callback)
    this.$showMAP('map')
  },
  beforeDestroy() {
    this.$setMapContextMenu(true)
  },
  methods: {
    nodeName(id) {
      return this.map.Nodes[id] ? this.map.Nodes[id].Name : ''
    },
    pollingList(id, lineMode) {
      const l = []
      if (!this.map.Nodes[id]) {
        return l
      }
      let nodeName = ''
      if (lineMode) {
        nodeName = this.map.Nodes[id].Name + ':'
      }
      this.map.Pollings[id].forEach((p) => {
        if (!lineMode || p.Mode === 'traffic') {
          l.push({
            text: nodeName + p.Name,
            value: p.ID,
          })
        }
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
          this.copyFrom = ''
          this.editNode = this.map.Nodes[r.Param]
          this.showNodeDialog = true
          break
        case 'contextMenu':
          this.x = r.x
          this.y = r.y
          if (!r.Node) {
            this.showMapContextMenu = true
            this.editNode.ID = ''
          } else {
            if (!this.map.Nodes[r.Node]) {
              return
            }
            this.copyFrom = ''
            this.editNode = this.map.Nodes[r.Node]
            this.urls = []
            this.editNode.URL.split(',').forEach((u) => {
              u = u.trim()
              if (u !== '') {
                this.urls.push(u)
              }
            })
            this.deleteNodes = [r.Node]
            this.showNodeContextMenu = true
          }
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
          (e.NodeID1 === p[1] && e.NodeID2 === p[0])
      )
      this.editLine = l || {
        NodeID1: p[0],
        PollingID2: '',
        NodeID2: p[1],
        PollingID1: '',
        PollingID: '',
        Info: '',
      }
      this.editLine.NodeName1 = this.nodeName(this.editLine.NodeID1)
      this.editLine.NodeName2 = this.nodeName(this.editLine.NodeID2)
      this.lineDialog = true
    },
    doDeleteNode() {
      this.$axios.post('/api/nodes/delete', this.deleteNodes).then(() => {
        this.$fetch()
      })
      this.deleteDialog = false
    },
    doUpdateNode() {
      let url = '/api/node/update'
      if (this.copyFrom && this.copyPolling) {
        url += '?from=' + this.copyFrom
      }
      this.$axios
        .post(url, this.editNode)
        .then(() => {
          this.$fetch()
          this.editNodeDialog = false
        })
        .catch((e) => {
          this.editNodeError = true
          this.$fetch()
        })
    },
    addNode() {
      this.copyFrom = ''
      this.editNode = {
        ID: '',
        Name: '新規ノード',
        IP: '',
        X: this.x,
        Y: this.y,
        Descr: '',
        Icon: 'desktop',
        MAC: '',
        SnmpMode: '',
        Community: '',
        User: '',
        Password: '',
        PublicKey: '',
        URL: '',
        Type: '',
        AddrMode: '',
      }
      this.editNodeDialog = true
    },
    deleteNode() {
      this.showNodeDialog = false
      this.deleteNodes = [this.editNode.ID]
      this.deleteDialog = true
    },
    copyNode() {
      this.showNodeDialog = false
      this.copyFrom = this.editNode.ID
      // 位置をずらして新規追加
      this.editNode.X += 64
      this.editNode.ID = ''
      this.editNode.State = 'unknown'
      this.editNode.Name += 'のコピー'
      this.editNodeDialog = true
    },
    showEditNodeDialog() {
      this.showNodeDialog = false
      this.editNodeDialog = true
    },
    showNodePollingPage() {
      this.$router.push({ path: '/node/polling/' + this.editNode.ID })
    },
    showNodeLogPage() {
      this.$router.push({ path: '/node/log/' + this.editNode.ID })
    },
    showMIBBr() {
      this.$router.push({ path: '/mibbr/' + this.editNode.ID })
    },
    openURL(url) {
      window.open(url, '_blank')
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
    checkPolling(all) {
      let id = 'all'
      if (!all) {
        if (this.editNode.ID === '') {
          return
        }
        id = this.editNode.ID
      }
      this.$axios.get('/api/polling/check/' + id).then(() => {
        this.showNodeDialog = false
        this.$fetch()
      })
    },
  },
}
</script>

<style scoped>
/* コンテキストメニューの高さ調整 */
.v-list--dense .v-list-item .v-list-item__icon {
  height: 20px;
  margin-top: 4px;
  margin-bottom: 4px;
}
.v-list--dense .v-list-item {
  min-height: 30px;
}
</style>
