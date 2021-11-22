<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> ホストリソース - {{ node.Name }} </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab">
          <v-tab key="system">システム</v-tab>
          <v-tab key="storage">ストレージ</v-tab>
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
                <span :class="getProcStatusColor(item.Status)">
                  {{ getProcStatusName(item.Status) }}
                </span>
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
      process: [],
      systemHeaders: [
        { text: 'No', value: 'Index', width: '10%' },
        { text: '項目', value: 'Name', width: '30%' },
        { text: '値', value: 'Value', width: '50%' },
        { text: '操作', value: 'Action', width: '10%' },
      ],
      storageHeaders: [
        { text: '種別', value: 'Type', width: '30%' },
        { text: '説明', value: 'Descr', width: '30%' },
        { text: 'サイズ', value: 'Size', width: '10%' },
        { text: '使用量', value: 'Used', width: '10%' },
        { text: '使用率', value: 'Rate', width: '10%' },
        { text: '操作', value: 'Action', width: '10%' },
      ],
      processHeaders: [
        { text: '状態', value: 'Status', width: '10%' },
        { text: 'PID', value: 'PID', width: '10%' },
        { text: '種別', value: 'Type', width: '10%' },
        { text: '名前', value: 'Name', width: '30%' },
        { text: 'CPU', value: 'CPU', width: '10%' },
        { text: 'メモリー', value: 'Mem', width: '10%' },
        { text: '操作', value: 'Action', width: '10%' },
      ],
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
    getProcStatusColor(s) {
      switch (s) {
        case 'Running':
          return 'light-blue--text'
        case 'Runnable':
          return 'lime--text'
        case 'Invalid':
        case 'NotRunnable':
          return 'red--text'
      }
      return 'gray--text'
    },
    getProcStatusName(s) {
      switch (s) {
        case 'Running':
          return '実行中'
        case 'Runnable':
          return '実行待'
        case 'NotRunnable':
          return '起動待'
        case 'Invalid':
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
  },
}
</script>
