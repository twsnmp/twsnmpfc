<template>
  <div class="dashboard-page">
    <v-app-bar dense app flat color="background" class="dashboard-bar">
      <v-toolbar-title class="font-weight-bold text-subtitle-1 text-sm-h6 mr-2">
        <v-icon left color="primary">mdi-view-dashboard</v-icon>
        {{ mapTitle }}
      </v-toolbar-title>

      <!-- 状態別ノード数 KPI 表示 -->
      <div
        v-if="map && map.Nodes"
        class="kpi-chips d-flex align-center flex-wrap"
      >
        <v-chip small outlined class="mr-1" title="全ノード数">
          全ノード: <strong>{{ nodeStats.total }}</strong>
        </v-chip>
        <v-chip
          v-if="nodeStats.high > 0"
          small
          color="#e31a1c"
          dark
          class="mr-1 font-weight-bold"
        >
          <v-icon left x-small>mdi-alert-circle</v-icon>
          重度: {{ nodeStats.high }}
        </v-chip>
        <v-chip
          v-if="nodeStats.low > 0"
          small
          color="#fb9a99"
          dark
          class="mr-1 font-weight-bold text--darken-4"
        >
          <v-icon left x-small>mdi-alert-circle</v-icon>
          軽度: {{ nodeStats.low }}
        </v-chip>
        <v-chip
          v-if="nodeStats.warn > 0"
          small
          color="#dfdf22"
          dark
          class="mr-1 font-weight-bold black--text"
        >
          <v-icon left x-small color="black">mdi-alert</v-icon>
          注意: {{ nodeStats.warn }}
        </v-chip>
        <v-chip
          v-if="nodeStats.normal > 0"
          small
          color="#33a02c"
          dark
          class="mr-1"
        >
          <v-icon left x-small>mdi-check-circle</v-icon>
          正常: {{ nodeStats.normal }}
        </v-chip>
        <v-chip
          v-if="nodeStats.repair > 0"
          small
          color="#1f78b4"
          dark
          class="mr-1"
        >
          復帰: {{ nodeStats.repair }}
        </v-chip>
        <v-chip
          v-if="nodeStats.other > 0"
          small
          color="#777777"
          dark
          class="mr-1"
        >
          他: {{ nodeStats.other }}
        </v-chip>
      </div>

      <v-spacer />

      <!-- ズーム操作コントロール -->
      <div class="zoom-controls d-flex align-center mr-2">
        <v-tooltip bottom>
          <template #activator="{ on, attrs }">
            <v-btn icon small v-bind="attrs" v-on="on" @click="zoomOut">
              <v-icon small>mdi-magnify-minus-outline</v-icon>
            </v-btn>
          </template>
          <span>縮小</span>
        </v-tooltip>
        <v-btn
          text
          x-small
          class="px-1 text-caption font-weight-bold"
          title="クリックで100%にリセット"
          @click="zoomReset"
        >
          {{ currentZoom }}%
        </v-btn>
        <v-tooltip bottom>
          <template #activator="{ on, attrs }">
            <v-btn icon small v-bind="attrs" v-on="on" @click="zoomIn">
              <v-icon small>mdi-magnify-plus-outline</v-icon>
            </v-btn>
          </template>
          <span>拡大</span>
        </v-tooltip>
      </div>

      <v-divider vertical class="my-2 mr-2"></v-divider>

      <!-- アクションボタン群 -->
      <v-tooltip bottom>
        <template #activator="{ on, attrs }">
          <v-btn
            icon
            color="primary"
            v-bind="attrs"
            v-on="on"
            @click="downloadImage"
          >
            <v-icon>mdi-download</v-icon>
          </v-btn>
        </template>
        <span>マップ画像をダウンロード (PNG)</span>
      </v-tooltip>
      <v-tooltip bottom>
        <template #activator="{ on, attrs }">
          <v-btn
            icon
            :loading="loading"
            v-bind="attrs"
            v-on="on"
            @click="fetchData"
          >
            <v-icon>mdi-cached</v-icon>
          </v-btn>
        </template>
        <span>最新データに更新</span>
      </v-tooltip>
      <v-tooltip bottom>
        <template #activator="{ on, attrs }">
          <v-btn
            icon
            :color="autoRefresh ? 'primary' : 'default'"
            v-bind="attrs"
            v-on="on"
            @click="autoRefresh = !autoRefresh"
          >
            <v-icon>{{ autoRefresh ? 'mdi-sync' : 'mdi-sync-off' }}</v-icon>
          </v-btn>
        </template>
        <span>自動更新: {{ autoRefresh ? '有効 (60秒毎)' : '無効' }}</span>
      </v-tooltip>
      <v-btn icon @click="toggleFullscreen">
        <v-icon>{{
          isFullscreen ? 'mdi-fullscreen-exit' : 'mdi-fullscreen'
        }}</v-icon>
      </v-btn>
    </v-app-bar>

    <!-- エラー表示 -->
    <v-alert v-if="errorMessage" type="error" prominent class="ma-4">
      {{ errorMessage }}
    </v-alert>

    <!-- 初回ローディング表示 -->
    <div
      v-if="initialLoading"
      class="d-flex flex-column justify-center align-center loading-container"
    >
      <v-progress-circular
        indeterminate
        color="primary"
        size="64"
        class="mb-4"
      ></v-progress-circular>
      <div class="text-subtitle-1 text--secondary">
        マップデータを読み込んでいます...
      </div>
    </div>

    <!-- マップ描画コンテナ -->
    <div
      v-show="!errorMessage && !initialLoading"
      id="public-map"
      ref="mapContainer"
    ></div>

    <v-snackbar v-model="snackbar" timeout="2000" color="info">
      {{ snackbarText }}
    </v-snackbar>
  </div>
</template>

<script>
export default {
  name: 'PublicDashboardPage',
  auth: false,
  layout: 'dashboard',
  data() {
    return {
      map: null,
      loading: false,
      initialLoading: true,
      errorMessage: '',
      lastUpdateTime: '',
      autoRefresh: true,
      refreshTimer: null,
      currentZoom: 100,
      isFullscreen: false,
      snackbar: false,
      snackbarText: '',
    }
  },
  computed: {
    mapTitle() {
      if (this.map && this.map.MapConf && this.map.MapConf.MapName) {
        return this.map.MapConf.MapName
      }
      return 'TWSNMP FC ダッシュボード'
    },
    nodeStats() {
      const stats = {
        total: 0,
        high: 0,
        low: 0,
        warn: 0,
        normal: 0,
        repair: 0,
        other: 0,
      }
      if (!this.map || !this.map.Nodes) {
        return stats
      }
      for (const k in this.map.Nodes) {
        stats.total++
        const s = this.map.Nodes[k].State || 'normal'
        if (stats[s] !== undefined) {
          stats[s]++
        } else {
          stats.other++
        }
      }
      return stats
    },
  },
  watch: {
    autoRefresh(val) {
      if (val) {
        this.startAutoRefresh()
      } else {
        this.stopAutoRefresh()
      }
    },
  },
  async mounted() {
    // アイコンフォントと状態色の初期化（必須）
    if (this.$setIconCodeMap && this.$iconList) {
      this.$setIconCodeMap(this.$iconList)
    }
    if (this.$setStateColorMap && this.$stateList) {
      this.$setStateColorMap(this.$stateList)
    }
    if (this.$setMapContextMenu) {
      this.$setMapContextMenu(false)
    }

    // ウェブフォント（MDI）の読み込み完了を待機してから描画
    if (document.fonts && document.fonts.ready) {
      await document.fonts.ready
    }

    await this.fetchData()
    this.initialLoading = false

    if (this.autoRefresh) {
      this.startAutoRefresh()
    }
    window.addEventListener('resize', this.resizeMap)
  },
  beforeDestroy() {
    this.stopAutoRefresh()
    window.removeEventListener('resize', this.resizeMap)
  },
  methods: {
    async fetchData() {
      const key = this.$route.params.key
      if (!key) {
        this.errorMessage = '公開キーが指定されていません。'
        this.initialLoading = false
        return
      }
      this.loading = true
      try {
        const res = await this.$axios.$get('/public/api/map/' + key)
        this.map = res
        this.errorMessage = ''
        if (this.map.LastUpdate) {
          const t = new Date(this.map.LastUpdate * 1000)
          this.lastUpdateTime = this.$timeFormat(t)
        }
        // カスタムアイコン定義の登録
        if (this.map.Icons && Array.isArray(this.map.Icons)) {
          this.map.Icons.forEach((icon) => {
            if (this.$setIcon) {
              this.$setIcon(icon)
            }
            if (this.$setIconToMap) {
              this.$setIconToMap(icon)
            }
          })
        }

        // マップ描画
        this.$nextTick(() => {
          this.$showMAP(
            'public-map',
            this.map,
            this.$axios.defaults.baseURL,
            true // 読み取り専用 (readOnly)
          )
          if (this.$getMapScale) {
            this.currentZoom = this.$getMapScale()
          }
          this.resizeMap()
        })
      } catch (err) {
        if (
          err.response &&
          (err.response.status === 403 || err.response.status === 404)
        ) {
          this.errorMessage =
            '公開ダッシュボードにアクセスできません。URLのキーが無効か、公開設定が無効になっています。'
        } else {
          this.errorMessage = 'マップデータの取得に失敗しました。'
        }
      } finally {
        this.loading = false
      }
    },
    zoomIn() {
      if (this.$zoomMap) {
        this.currentZoom = this.$zoomMap(0.1)
      }
    },
    zoomOut() {
      if (this.$zoomMap) {
        this.currentZoom = this.$zoomMap(-0.1)
      }
    },
    zoomReset() {
      if (this.$resetMapZoom) {
        this.currentZoom = this.$resetMapZoom()
      }
    },
    downloadImage() {
      const fileName =
        (this.mapTitle.replace(/\s+/g, '_') || 'TWSNMPFC-MAP') + '.png'
      if (this.$saveMapImage) {
        this.$saveMapImage(fileName)
        this.snackbarText = 'マップ画像をダウンロードしました: ' + fileName
        this.snackbar = true
      }
    },
    resizeMap() {
      const el = document.getElementById('public-map')
      if (el) {
        const topOffset = 48
        el.style.height = window.innerHeight - topOffset + 'px'
      }
    },
    startAutoRefresh() {
      this.stopAutoRefresh()
      this.refreshTimer = setInterval(() => {
        this.fetchData()
      }, 60 * 1000)
    },
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    },
    toggleFullscreen() {
      if (!document.fullscreenElement) {
        document.documentElement
          .requestFullscreen()
          .then(() => {
            this.isFullscreen = true
            this.resizeMap()
          })
          .catch(() => {})
      } else if (document.exitFullscreen) {
        document
          .exitFullscreen()
          .then(() => {
            this.isFullscreen = false
            this.resizeMap()
          })
          .catch(() => {})
      }
    },
  },
}
</script>

<style scoped>
.dashboard-page {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}
.dashboard-bar {
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}
.kpi-chips {
  max-width: 50vw;
  overflow-x: auto;
}
.zoom-controls {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
  padding: 2px 4px;
}
.loading-container {
  height: calc(100vh - 48px);
}
#public-map {
  width: 100%;
  overflow: scroll;
}
</style>
