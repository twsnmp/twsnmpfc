<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        アドレス分析
        <v-spacer></v-spacer>
        <v-text-field
          v-model="addr"
          append-icon="mdi-magnify"
          label="アドレス"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="info"
        :items-per-page="15"
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
        <v-btn color="primary" dark @click="$fetch()">
          <v-icon>mdi-magnify</v-icon>
          調査
        </v-btn>
        <v-btn v-if="latLong" color="primary" dark @click="showGoogleMap">
          <v-icon>mdi-google-maps</v-icon>
          地図
        </v-btn>
        <download-excel
          :data="info"
          type="csv"
          name="TWSNMP_FC_Addr_Info.csv"
          header="TWSNMP FC Address Info"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="info"
          type="xls"
          name="TWSNMP_FC_Addr_Info.xls"
          header="TWSNMP FC Address Info"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
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
      addr: '',
      headers: [
        { text: '状態', value: 'Level', width: '10%' },
        { text: '項目', value: 'Title', width: '40%' },
        { text: '値', value: 'Value', width: '50%' },
      ],
      info: [],
      latLong: '',
    }
  },
  async fetch() {
    this.latLong = ''
    if (!this.addr) {
      this.addr = this.$route.params.addr
    }
    if (!this.addr) {
      return
    }
    this.info = await this.$axios.$get('/api/report/address/' + this.addr)
    if (this.info) {
      this.info.forEach((e) => {
        if (e.Title === '位置' && e.Value.includes(',')) {
          const a = e.Value.split(',')
          if (a.length > 2 && a[0] !== 'LOCAL') {
            this.latLong = a[1] + ',' + a[2]
          }
        }
      })
    }
  },
  methods: {
    showGoogleMap() {
      if (!this.latLong) {
        return
      }
      const url = `https://www.google.com/maps/search/?api=1&query=${this.latLong}`
      window.open(url, '_blank')
    },
  },
}
</script>
