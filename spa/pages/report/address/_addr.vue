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
        >
        </v-text-field>
        <v-menu v-if="history.length > 0" offset-y>
          <template #activator="{ on, attrs }">
            <v-btn
              class="mt-3 ml-2"
              color="normal"
              dark
              v-bind="attrs"
              v-on="on"
            >
              <v-icon>mdi-history</v-icon>
            </v-btn>
          </template>
          <v-list>
            <v-list-item
              v-for="(a, index) in history"
              :key="index"
              @click="addr = a"
            >
              <v-list-item-title>{{ a }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="info"
        :items-per-page="20"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
        @dblclick:row="copy1Info"
      >
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
      </v-data-table>
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
      <v-card-actions>
        <v-switch
          v-model="blackList"
          class="ml-2"
          label="DNS Black List"
        ></v-switch>
        <v-switch v-model="noCache" class="ml-2" label="No Cache"></v-switch>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="$fetch()">
          <v-icon>mdi-magnify</v-icon>
          調査
        </v-btn>
        <v-btn v-if="isIP" color="primary" dark @click="showVirusTotal">
          <v-icon>mdi-virus</v-icon>
          VirusTotal
        </v-btn>
        <v-btn v-if="latLong" color="primary" dark @click="showGoogleMap">
          <v-icon>mdi-google-maps</v-icon>
          地図
        </v-btn>
        <v-btn v-if="info.length > 0" color="info" dark @click="copyInfo()">
          <v-icon>mdi-content-copy</v-icon>
          コピー
        </v-btn>
        <download-excel
          v-if="info.length > 0"
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Addr_Info.csv"
          header="TWSNMP FCによるアドレス調査結果"
          class="v-btn ml-0"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          v-if="info.length > 0"
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_Addr_Info.csv"
          header="TWSNMP FCによるアドレス調査結果"
          class="v-btn ml-0"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          v-if="info.length > 0"
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_Addr_Info.xls"
          header="TWSNMP FCによるアドレス調査結果"
          worksheet="調査結果"
          class="v-btn ml-0 pl-0 pr-1"
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
      history: [],
      copyDone: false,
      copyError: false,
      blackList: false,
      noCache: false,
      isIP: false,
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
    this.isIP = this.addr.includes('.')
    if (!this.history.includes(this.addr)) {
      this.history.push(this.addr)
    }
    let url = '/api/report/address/' + this.addr
    if (this.blackList) {
      url += '?dnsbl=true'
    }
    if (this.noCache) {
      url += this.blackList ? '&' : '?'
      url += 'noCache=true'
    }
    this.info = await this.$axios.$get(url)
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
    showVirusTotal() {
      if (!this.isIP) {
        return
      }
      const url = `https://www.virustotal.com/gui/ip-address/${this.addr}`
      window.open(url, '_blank')
    },
    copyInfo() {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const a = []
      this.info.forEach((e) => {
        a.push(e.Title + ',' + e.Value)
      })
      if (a.length < 1) {
        return
      }
      const s = a.join('\n')
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    copy1Info(me, p) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const s = p.item.Title + ' ' + p.item.Value
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    makeExports() {
      const exports = []
      this.info.forEach((e) => {
        exports.push({
          状態: this.$getStateName(e.Level),
          項目: e.Title,
          値: e.Value,
        })
      })
      return exports
    },
  },
}
</script>
