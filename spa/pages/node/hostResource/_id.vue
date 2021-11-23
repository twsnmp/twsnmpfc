<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> ホストリソース - {{ node.Name }} </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab">
          <v-tab key="system">システム</v-tab>
          <v-tab key="storage">ストレージ</v-tab>
          <v-tab key="device">デバイス</v-tab>
          <v-tab key="fs">ファイルシステム</v-tab>
          <v-tab key="process">プロセス</v-tab>
        </v-tabs>
        <v-tabs-items v-model="tab">
          <v-tab-item key="system">
            <v-data-table
              :headers="systemHeaders"
              :items="system"
              sort-by="Index"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.actions`]="{ item }">
                <v-icon
                  v-if="item.Polling"
                  small
                  @click="editSystemPolling(item.Polling)"
                >
                  mdi-card-plus
                </v-icon>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="storage">
            <v-data-table
              :headers="storageHeaders"
              :items="storage"
              sort-by="Type"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.Rate`]="{ item }">
                <span :class="getRateColor(item.Rate)">
                  {{ item.Rate.toFixed(1) }}
                </span>
              </template>
              <template #[`item.Type`]="{ item }">
                {{ getStorageTypeName(item.Type) }}
              </template>
              <template #[`item.actions`]="{ item }">
                <v-icon small @click="editStoragePolling(item)">
                  mdi-card-plus
                </v-icon>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="device">
            <v-data-table
              :headers="deviceHeaders"
              :items="device"
              sort-by="Type"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.Status`]="{ item }">
                <span :class="getStatusColor(item.Status)">
                  {{ getStatusName(item.Status) }}
                </span>
              </template>
              <template #[`item.Type`]="{ item }">
                {{ getDeviceTypeName(item.Type) }}
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="fs">
            <v-data-table
              :headers="fsHeaders"
              :items="fs"
              sort-by="Index"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.Type`]="{ item }">
                {{ getFSTypeName(item.Type) }}
              </template>
              <template #[`item.Access`]="{ item }">
                {{ getAccess(item.Access) }}
              </template>
              <template #[`item.Bootable`]="{ item }">
                {{ getTrueFalse(item.Bootable) }}
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="process">
            <v-data-table
              :headers="processHeaders"
              :items="process"
              sort-by="PID"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
              class="log"
            >
              <template #[`item.Status`]="{ item }">
                <span :class="getStatusColor(item.Status)">
                  {{ getStatusName(item.Status) }}
                </span>
              </template>
              <template #[`item.actions`]="{ item }">
                <v-icon small @click="editProcessStatusPolling(item)">
                  mdi-eye
                </v-icon>
                <v-icon small @click="editProcessCPUPolling(item)">
                  mdi-cpu-64-bit
                </v-icon>
                <v-icon small @click="editProcessMemPolling(item)">
                  mdi-memory
                </v-icon>
              </template>
            </v-data-table>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="500px">
      <v-card>
        <v-card-title> ポーリング設定 </v-card-title>
        <v-alert v-model="addError" color="error" dense dismissible>
          ポーリングを追加できませんでした
        </v-alert>
        <v-card-text>
          <v-text-field v-model="polling.Name" label="名前"></v-text-field>
          <v-select v-model="polling.Level" :items="$levelList" label="レベル">
          </v-select>
          <v-text-field
            v-if="hasTh"
            v-model="thValue"
            label="閾値"
          ></v-text-field>
          <v-slider
            v-model="polling.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="3600"
            min="5"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="polling.PollInt"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="polling.Timeout"
            label="タイムアウト(Sec)"
            class="align-center"
            max="10"
            min="1"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="polling.Timeout"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="polling.Retry"
            label="リトライ回数"
            class="align-center"
            max="5"
            min="0"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="polling.Retry"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-select
            v-model="polling.LogMode"
            :items="$logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="addPolling">
            <v-icon>mdi-content-save</v-icon>
            追加
          </v-btn>
          <v-btn color="normal" dark @click="editDialog = false">
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
      node: {},
      tab: 'system',
      system: [],
      storage: [],
      device: [],
      process: [],
      fs: [],
      systemHeaders: [
        { text: 'No', value: 'Index', width: '10%' },
        { text: '項目', value: 'Name', width: '32%' },
        { text: '値', value: 'Value', width: '50%' },
        { text: '操作', value: 'actions', width: '8%' },
      ],
      storageHeaders: [
        { text: '種別', value: 'Type', width: '20%' },
        { text: '説明', value: 'Descr', width: '32%' },
        { text: 'サイズ', value: 'Size', width: '10%' },
        { text: '使用量', value: 'Used', width: '10%' },
        { text: '使用率', value: 'Rate', width: '10%' },
        { text: '単位', value: 'Unit', width: '10%' },
        { text: '操作', value: 'actions', width: '8%' },
      ],
      deviceHeaders: [
        { text: '状態', value: 'Status', width: '10%' },
        { text: 'インデックス', value: 'Index', width: '10%' },
        { text: '種別', value: 'Type', width: '30%' },
        { text: '説明', value: 'Descr', width: '40%' },
        { text: 'エラー', value: 'Errors', width: '10%' },
      ],
      fsHeaders: [
        { text: 'マウント', value: 'Mount', width: '30%' },
        { text: 'リモート', value: 'Remote', width: '30%' },
        { text: '種別', value: 'Type', width: '20%' },
        { text: 'アクセス', value: 'Access', width: '10%' },
        { text: 'ブート', value: 'Bootable', width: '10%' },
      ],
      processHeaders: [
        { text: '状態', value: 'Status', width: '8%' },
        { text: 'PID', value: 'PID', width: '8%' },
        { text: '種別', value: 'Type', width: '10%' },
        { text: '名前', value: 'Name', width: '15%' },
        { text: 'パス', value: 'Path', width: '15%' },
        { text: 'パラメータ', value: 'Param', width: '20%' },
        { text: 'CPU', value: 'CPU', width: '8%' },
        { text: 'Mem', value: 'Mem', width: '8%' },
        { text: '操作', value: 'actions', width: '8%' },
      ],
      editDialog: false,
      hasTh: false,
      addError: false,
      thValue: '',
      polling: {
        NodeID: '',
        Name: '',
        Level: 'low',
        Type: '',
        Mode: '',
        Filter: '',
        Script: '',
        Timeout: 3,
        Retry: 1,
        LogMode: 1,
        PollInt: 60,
      },
    }
  },
  async fetch() {
    const r = await this.$axios.$get(
      '/api/node/hostResource/' + this.$route.params.id
    )
    if (!r || !r.Node) {
      return
    }
    this.node = r.Node
    this.system = r.HostResource.System
    this.storage = r.HostResource.Storage
    this.storage.forEach((s) => {
      if (s.Size > 0) {
        s.Rate = (100.0 * s.Used) / s.Size
      } else {
        s.Rate = 0.0
      }
    })
    this.device = r.HostResource.Device
    this.fs = r.HostResource.FileSystem
    this.process = r.HostResource.Process
  },
  methods: {
    getStorageTypeName(t) {
      switch (t) {
        case 'hrStorageCompactDisc':
          return 'CDドライブ'
        case 'hrStorageRemovableDisk':
          return 'リムーバブル'
        case 'hrStorageFloppyDisk':
          return 'フロッピー'
        case 'hrStorageRamDisk':
          return 'RAMディスク'
        case 'hrStorageFlashMemory':
          return 'フラッシュメモリ'
        case 'hrStorageNetworkDisk':
          return 'ネットワーク'
        case 'hrStorageFixedDisk':
          return '固定ディスク'
        case 'hrStorageVirtualMemory':
          return '仮想メモリ'
        case 'hrStorageRam':
          return '実メモリ'
      }
      return 'その他'
    },
    getStatusColor(s) {
      switch (s) {
        case 'Running':
          return 'light-blue--text'
        case 'Runnable':
        case 'Testing':
          return 'lime--text'
        case 'Invalid':
        case 'NotRunnable':
        case 'Down':
          return 'red--text'
      }
      return 'gray--text'
    },
    getStatusName(s) {
      switch (s) {
        case 'Running':
          return '動作中'
        case 'Runnable':
          return '動作待'
        case 'Testing':
          return 'テスト中'
        case 'NotRunnable':
          return '起動待'
        case 'Invalid':
        case 'Down':
          return '停止'
      }
      return '不明'
    },
    getRateColor(r) {
      if (r < 80.0) {
        return 'gray--text'
      } else if (r < 90.0) {
        return 'yellow--text'
      }
      return 'red--text'
    },
    getDeviceTypeName(t) {
      return t.replace('hrDevice', '')
    },
    getFSTypeName(t) {
      return t.replace('hrFS', '')
    },
    getTrueFalse(v) {
      if (v === 1) {
        return 'はい'
      }
      return 'いいえ'
    },
    getAccess(v) {
      if (v === 1) {
        return '読み書き'
      }
      return '読み出しのみ'
    },
    editSystemPolling(p) {
      this.polling = {
        ID: '',
        Name: '',
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: '',
        Params: '',
        Filter: '',
        Extractor: '',
        Script: '',
        Level: 'low',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      if (p.includes('hrSystemDate')) {
        this.polling.Name = 'システム時刻の監視'
        this.polling.Mode = 'hrSystemDate'
        this.polling.Script = 'diff < $thValue'
        this.hasTh = true
      } else if (p.includes('hrSystemNumUsers')) {
        this.polling.Name = 'ログインユーザー数の監視'
        this.polling.Mode = 'get'
        this.polling.Params = 'hrSystemNumUsers.0'
        this.polling.Script = 'hrSystemNumUsers < $thValue'
        this.hasTh = true
      } else if (p.includes('hrSystemProcesses')) {
        this.polling.Name = '稼働プロセス数の監視'
        this.polling.Mode = 'get'
        this.polling.Params = 'hrSystemProcesses.0'
        this.polling.Script = 'hrSystemProcesses < $thValue'
        this.hasTh = true
      } else if (p.includes('hrProcessorLoad')) {
        this.polling.Name = 'CPU使用率の監視'
        this.polling.Mode = 'get'
        this.polling.Params = p
        this.polling.Script = 'hrProcessorLoad < $thValue'
        this.hasTh = true
      } else {
        return
      }
      this.editDialog = true
    },
    editStoragePolling(i) {
      this.polling = {
        ID: '',
        Name: 'ストレージ"' + i.Descr + '"の使用率監視',
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'get',
        Params: 'hrStorageSize.' + i.Index + ',hrStorageUsed.' + i.Index,
        Filter: '',
        Extractor: '',
        Script:
          's = hrStorageSize;' +
          'u = hrStorageUsed;' +
          's && ((100.0 * u ) / s < $thValue)',
        Level: 'low',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.hasTh = true
      this.editDialog = true
    },
    editProcessStatusPolling(i) {
      this.polling = {
        ID: '',
        Name: 'プロセス"' + i.Name + '"の状態監視',
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'get',
        Params: 'hrSWRunStatus.' + i.PID,
        Filter: '',
        Extractor: '',
        Script: 'hrSWRunStatus == 1 || hrSWRunStatus == 2',
        Level: 'low',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.hasTh = false
      this.editDialog = true
    },
    editProcessCPUPolling(i) {
      this.polling = {
        ID: '',
        Name: 'プロセス"' + i.Name + '"のCPU使用量監視',
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'ps',
        Params: 'hrSWRunPerfCPU.' + i.PID,
        Filter: '',
        Extractor: '',
        Script: 'hrSWRunPerfCPU_PS < $thValue',
        Level: 'info',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.hasTh = true
      this.editDialog = true
    },
    editProcessMemPolling(i) {
      this.polling = {
        ID: '',
        Name: 'プロセス"' + i.Name + '"のメモリー使用量監視',
        NodeID: this.node.ID,
        Type: 'snmp',
        Mode: 'ps',
        Params: 'hrSWRunPerfMem.' + i.PID,
        Filter: '',
        Extractor: '',
        Script: 'hrSWRunPerfMem_PS < $thValue',
        Level: 'info',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.hasTh = true
      this.editDialog = true
    },
    addPolling() {
      const tmpScript = this.polling.Script
      this.addError = false
      if (this.hasTh) {
        this.polling.Script = this.polling.Script.replace(
          '$thValue',
          this.thValue
        )
      }
      this.$axios
        .post('/api/polling/add', this.polling)
        .then(() => {
          this.editDialog = false
        })
        .catch((e) => {
          this.polling.Script = tmpScript
          this.addError = true
        })
    },
  },
}
</script>
