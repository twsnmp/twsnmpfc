<template>
  <v-row>
    <v-dialog v-model="deleteDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード削除</span>
        </v-card-title>
        <v-alert :value="deleteError" type="error" dense dismissible>
          ノードを削除できませんでした
        </v-alert>
        <v-card-text> 選択したノードを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="deleteDialog = false">キャンセル</v-btn>
          <v-btn color="error" @click="doDeleteNode">削除</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editNodeDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード設定</span>
        </v-card-title>
        <v-alert :value="editNodeError" type="error" dense dismissible>
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
          <v-btn color="normal" dark @click="editNodeDialog = false"
            >キャンセル</v-btn
          >
          <v-btn color="primary" dark @click="doUpdateNode">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="lineDialog" max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ライン編集</span>
        </v-card-title>
        <v-alert :value="lineError" type="error" dense dismissible>
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
          <v-btn color="normal" dark @click="lineDialog = false"
            >キャンセル</v-btn
          >
          <v-btn color="red" dark @click="deleteLine">切断</v-btn>
          <v-btn color="primary" dark @click="addLine">接続</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="showNodeDialog" max-width="500px">
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
          <v-btn color="primary" dark @click="showEditNodeDialog">編集</v-btn>
          <v-btn color="success" dark @click="showNodeInfoPage">詳細</v-btn>
          <v-btn color="normal" dark @click="showNodeDialog = false"
            >閉じる</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.map = await this.$axios.$get('/api/map')
    this.$setMAP(this.map)
    this.$store.commit('map/setMAP', this.map)
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
      },
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
    showEditNodeDialog() {
      this.showNodeDialog = false
      this.editNodeDialog = true
    },
    showNodeInfoPage() {
      this.$router.push({ path: '/node/' + this.editNode.ID })
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
