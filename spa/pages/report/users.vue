<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        ユーザー
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="users"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.Score)">{{
            $getScoreIconName(item.Score)
          }}</v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            v-if="item.ServerNodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.ServerNodeID })"
          >
            mdi-lan
          </v-icon>
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.user" label="user"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.server" label="server"></v-text-field>
            </td>
            <td colspan="6"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="openUserChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>サーバー別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_User_List.csv"
          header="TWSNMP FCで作成したユーザーリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_User_List.xls"
          header="TWSNMP FCで作成したユーザーリスト"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="error" dark @click="resetDialog = true">
          <v-icon>mdi-calculator</v-icon>
          再計算
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ユーザー削除</span>
        </v-card-title>
        <v-card-text>
          ユーザー"{{ selected.UserID }}"を削除しますか？
        </v-card-text>
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
    <v-dialog v-model="resetDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">信用度再計算</span>
        </v-card-title>
        <v-card-text> ユーザーレポートの信用度を再計算しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doReset">
            <v-icon>mdi-calculator</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="resetDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="usersChartDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー別</span>
        </v-card-title>
        <div id="usersChart" style="width: 1000px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="usersChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">ユーザー情報</span>
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
                <td>ユーザーID</td>
                <td>{{ selected.UserID }}</td>
              </tr>
              <tr>
                <td>サーバー名</td>
                <td>{{ selected.ServerName }}</td>
              </tr>
              <tr>
                <td>サーバーIP</td>
                <td>{{ selected.Server }}</td>
              </tr>
              <tr>
                <td>クライアント</td>
                <td>
                  <v-virtual-scroll
                    height="200"
                    item-height="20"
                    :items="selected.ClientList"
                  >
                    <template #default="{ item }">
                      <v-list-item>
                        <v-list-item-title>
                          {{ item.title }}
                          <v-icon
                            small
                            @click="
                              $router.push({
                                path: '/report/address/' + item.title,
                              })
                            "
                          >
                            mdi-file-find
                          </v-icon>
                        </v-list-item-title>
                        {{ item.value }}
                      </v-list-item>
                    </template>
                  </v-virtual-scroll>
                </td>
              </tr>
              <tr>
                <td>ログイン回数</td>
                <td>{{ selected.Total }}</td>
              </tr>
              <tr>
                <td>成功回数</td>
                <td>{{ selected.Ok }}</td>
              </tr>
              <tr>
                <td>成功率</td>
                <td>{{ selected.Rate }}</td>
              </tr>
              <tr>
                <td>信用スコア</td>
                <td>{{ selected.Score }}</td>
              </tr>
              <tr>
                <td>ペナリティー</td>
                <td>{{ selected.Penalty }}</td>
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
          <v-btn
            color="error"
            dark
            @click="
              infoDialog = false
              deleteDialog = true
            "
          >
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
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
import * as numeral from 'numeral'
export default {
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'ユーザーID',
          value: 'UserID',
          width: '15%',
          filter: (value) => {
            if (!this.conf.user) return true
            return value.includes(this.conf.user)
          },
        },
        {
          text: 'サーバー',
          value: 'ServerName',
          width: '15%',
          filter: (value) => {
            if (!this.conf.server) return true
            return value.includes(this.conf.server)
          },
        },
        { text: 'CL数', value: 'Client', width: '8%' },
        { text: '回数', value: 'Total', width: '8%' },
        { text: '成功率%', value: 'Rate', width: '10%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      users: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      usersChartDialog: false,
      infoDialog: false,
      conf: {
        user: '',
        server: '',
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.users = await this.$axios.$get('/api/report/users')
    if (!this.users) {
      return
    }
    this.users.forEach((u) => {
      u.First = this.$timeFormat(
        new Date(u.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      u.Last = this.$timeFormat(
        new Date(u.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      if (u.Total > 0) {
        u.Rate = ((100 * u.Ok) / u.Total).toFixed(2)
      }
      const cl = u.ClientMap ? Object.keys(u.ClientMap) : 0
      u.Client = cl.length
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.users.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/users/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/user/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    doReset() {
      this.$axios
        .post('/api/report/users/reset', {})
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.resetError = true
          this.$fetch()
        })
      this.resetDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openUserChart() {
      this.usersChartDialog = true
      this.$nextTick(() => {
        this.$showUsersChart('usersChart', this.users, this.conf)
      })
    },
    openInfoDialog(item) {
      this.selected = item
      if (!this.selected.ClientList) {
        this.selected.ClientList = []
        if (!this.selected.ClientMap) {
          return
        }
        Object.keys(this.selected.ClientMap).forEach((k) => {
          this.selected.ClientList.push({
            title: k,
            value:
              this.selected.ClientMap[k].Ok +
              '/' +
              this.selected.ClientMap[k].Total,
            total: this.selected.ClientMap[k].Total,
          })
        })
        this.selected.ClientList.sort((a, b) => {
          if (a.total > b.total) return -1
          if (a.total < b.total) return 1
          return 0
        })
      }
      this.infoDialog = true
    },
    makeExports() {
      const exports = []
      this.users.forEach((u) => {
        if (!this.$filterUser(u, this.conf)) {
          return
        }
        exports.push({
          ユーザーID: u.UserID,
          サーバー名: u.ServerName,
          サーバーIP: u.Server,
          クライアント数: u.Client,
          ログイン回数: u.Total,
          成功回数: u.Ok,
          成功率: u.Rate,
          信用スコア: u.Score,
          ペナリティー: u.Penalty,
          初回日時: u.First,
          最終日時: u.Last,
        })
      })
      return exports
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
