<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        IPアドレス使用状況
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="ipamHeatmap" style="width: 100%; height: 35vh"></div>
      <v-simple-table>
        <template #default>
          <thead>
            <tr>
              <th class="text-left">IPアドレス範囲</th>
              <th class="text-left">サイズ</th>
              <th class="text-left">使用量</th>
              <th class="text-left">使用率</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in ranges" :key="item.Range">
              <td>{{ item.Range }}</td>
              <td>{{ item.Size }}</td>
              <td>{{ item.Used }}</td>
              <td>
                <v-progress-linear
                  :value="item.Usage"
                  height="25"
                  :color="getUsageColor(item.Usage)"
                >
                  {{ item.Usage.toFixed(2) }}%
                </v-progress-linear>
              </td>
            </tr>
          </tbody>
        </template>
      </v-simple-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_IPAM.csv"
          header="TWSNMP FCで作成したIPアドレス使用状況"
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
          name="TWSNMP_FC_IPAM.csv"
          header="TWSNMP FCで作成したIPアドレス使用状況"
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
          name="TWSNMP_FC_IPAM.xls"
          header="TWSNMP FCで作成したIPアドレス使用状況"
          worksheet="IPアドレス使用状況"
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
      </v-card-actions>
    </v-card>
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      ranges: [],
    }
  },
  async fetch() {
    this.ranges = await this.$axios.$get('/api/report/ipam')
    if (!this.ranges) {
      this.ranges = []
    }
    this.$showIPAMHeatmap('ipamHeatmap', this.ranges)
  },
  methods: {
    getUsageColor(u) {
      if (u < 60.0) {
        return '#4575b4'
      }
      if (u < 90.0) {
        return '#fee090'
      }
      return '#d73027'
    },
    makeExports() {
      const exports = []
      this.ranges.forEach((e) => {
        exports.push({
          IP範囲: e.Range,
          サイズ: e.Size,
          使用量: e.Used,
          使用率: e.Usage,
        })
      })
      return exports
    },
  },
}
</script>
