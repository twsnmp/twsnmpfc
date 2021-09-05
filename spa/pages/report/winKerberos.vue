<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Windows Kerberos
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        レポートのデータを取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="kerberos"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.target" label="Target">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.computer" label="Computer">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ip" label="From IP"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.service" label="Service">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ticketType" label="Ticket Type">
              </v-text-field>
            </td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :data="kerberos"
          type="csv"
          name="TWSNMP_FC_Windows_Kerberos_List.csv"
          header="TWSNMP FC Windows Kerberos List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="kerberos"
          type="xls"
          name="TWSNMP_FC_Windows_Kerberos_List.xls"
          header="TWSNMP FC Windows Kerberos List"
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
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">レポート削除</span>
        </v-card-title>
        <v-card-text> 選択した項目を削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDelete">
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
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">Kerberos情報</span>
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
                <td>対象</td>
                <td>{{ selected.Target }}</td>
              </tr>
              <tr>
                <td>コンピュータ</td>
                <td>{{ selected.Computer }}</td>
              </tr>
              <tr>
                <td>IP</td>
                <td>{{ selected.IP }}</td>
              </tr>
              <tr>
                <td>サービス</td>
                <td>{{ selected.Service }}</td>
              </tr>
              <tr>
                <td>チケット種別</td>
                <td>{{ selected.TicketType }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>失敗</td>
                <td>{{ selected.Failed }}</td>
              </tr>
              <tr>
                <td>初回日時</td>
                <td>{{ selected.First }}</td>
              </tr>
              <tr>
                <td>最終日時</td>
                <td>{{ selected.Last }}</td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
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
      headers: [
        {
          text: '対象',
          value: 'Target',
          width: '15%',
          filter: (value) => {
            if (!this.conf.target) return true
            return value.includes(this.conf.target)
          },
        },
        {
          text: 'コンピュータ',
          value: 'Computer',
          width: '12%',
          filter: (value) => {
            if (!this.conf.computer) return true
            return value.includes(this.conf.computer)
          },
        },
        {
          text: '要求元IP',
          value: 'IP',
          width: '10%',
          filter: (value) => {
            if (!this.conf.ip) return true
            return value.includes(this.conf.ip)
          },
        },
        {
          text: 'サービス',
          value: 'Service',
          width: '13%',
          filter: (value) => {
            if (!this.conf.service) return true
            return value.includes(this.conf.service)
          },
        },
        {
          text: 'チケット種別',
          value: 'TicketType',
          width: '10%',
          filter: (value) => {
            if (!this.conf.ticketType) return true
            return value.includes(this.conf.ticketType)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '失敗', value: 'Failed', width: '8%' },
        { text: '最終', value: 'Last', width: '14%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      kerberos: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      infoDialog: false,
      conf: {
        target: '',
        computer: '',
        ip: '',
        service: '',
        ticketType: '',
        sortBy: 'Count',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.kerberos = await this.$axios.$get('/api/report/WinKerberos')
    if (!this.kerberos) {
      return
    }
    this.kerberos.forEach((e) => {
      e.First = this.$timeFormat(
        new Date(e.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      e.Last = this.$timeFormat(
        new Date(e.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.twwinlog.winKerberos
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/twwinlog/setWinKerberos', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/WinKerberos/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
  },
}
</script>