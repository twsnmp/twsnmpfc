<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        イベントログ - {{ node.Name }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
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
      search: '',
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        { text: '状態', value: 'Level', width: '10%' },
        {
          text: '発生日時',
          value: 'TimeStr',
          width: '15%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
        { text: '種別', value: 'Type', width: '10%' },
        { text: 'イベント', value: 'Event', width: '50%' },
      ],
      logs: [],
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/node/log/' + this.$route.params.id)
    this.node = r.Node
    if (r.Logs) {
      this.logs = r.Logs
      this.logs.forEach((e) => {
        const t = new Date(e.Time / (1000 * 1000))
        e.TimeStr = this.$timeFormat(t)
      })
    }
    this.$showLogLevelChart('logCountChart', this.logs, this.zoomCallBack)
  },
  mounted() {
    this.$showLogLevelChart('logCountChart', this.logs)
  },
  methods: {
    zoomCallBack(st, et) {
      this.zoom.st = st
      this.zoom.et = et
    },
  },
}
</script>
