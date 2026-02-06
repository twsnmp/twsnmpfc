<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        IPv6アドレス
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="ipv6s"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
        @dblclick:row="copyIP"
      >
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.ipv6" label="IPv6"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.node" label="node"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ipv4" label="IPv4"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.vendor" label="vendor"></v-text-field>
            </td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_IPv6_List.csv"
          header="TWSNMP FCで作成したIPv6アドレスリスト"
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
          name="TWSNMP_FC_IPv6_List.csv"
          header="TWSNMP FCで作成したIPv6アドレスリスト"
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
          name="TWSNMP_FC_IPv6_List.xls"
          header="TWSNMP FCで作成したIPv6アドレスリスト"
          worksheet="IPv6アドレス"
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
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
    </v-card>
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      headers: [
        {
          text: 'IPv6アドレス',
          value: 'IPv6',
          width: '20%',
          filter: (value) => {
            if (!this.conf.ipv6) return true
            return value.includes(this.conf.ipv6)
          },
        },
        {
          text: 'ノード',
          value: 'Node',
          width: '20%',
          filter: (value) => {
            if (!this.conf.node) return true
            return value.includes(this.conf.node)
          },
        },
        {
          text: 'IPv4アドレス',
          value: 'IPv4',
          width: '20%',
          filter: (value) => {
            if (!this.conf.ipv4) return true
            return value.includes(this.conf.ipv4)
          },
        },
        {
          text: 'MACアドレス',
          value: 'MAC',
          width: '20%',
          filter: (value) => {
            if (!this.conf.mac) return true
            return value.includes(this.conf.mac)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '30%',
          filter: (value) => {
            if (!this.conf.vendor) return true
            return value.includes(this.conf.vendor)
          },
        },
      ],
      ipv6s: [],
      copyDone: false,
      copyError: false,
      conf: {
        ipv6: '',
        ipv4: '',
        node: '',
        mac: '',
        vendor: '',
        sortBy: 'IPv6',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.ipv6s = await this.$axios.$get('/api/report/ipv6s')
    if (!this.ipv6s) {
      return
    }
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.ipv6report.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/ipv6report/setConf', this.conf)
  },
  methods: {
    getFilterList() {
      const list = []
      this.ipv6s.forEach((ip) => {
        if (!this.filterIP(ip, this.conf)) {
          return
        }
        list.push(ip)
      })
      return list
    },
    doCopy(s) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    copyIP(me, p) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const s = p.item.IPv6
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
      this.ipv6s.forEach((ip) => {
        if (!this.filterIP(ip, this.conf)) {
          return
        }
        exports.push({
          IPv6アドレス: ip.IPv6,
          ノード: ip.Node,
          IPv4アドレス: ip.IPv4,
          MACアドレス: ip.MAC,
          ベンダー: ip.Vendor,
        })
      })
      return exports
    },
    filterIP(ip, filter) {
      if (filter.mac && !ip.MAC.includes(filter.mac)) {
        return false
      }
      if (filter.node && !ip.Node.includes(filter.node)) {
        return false
      }
      if (filter.ipv6 && !ip.IPv6.includes(filter.ipv6)) {
        return false
      }
      if (filter.ipv4 && !ip.IPv4.includes(filter.ipv4)) {
        return false
      }
      if (filter.vendor && !ip.Vendor.includes(filter.vendor)) {
        return false
      }
      return true
    },
  },
}
</script>
