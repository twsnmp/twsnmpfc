<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        サーバー証明書
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="certs"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
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
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.target" label="Target"></v-text-field>
            </td>
            <td></td>
            <td>
              <v-text-field
                v-model="conf.subject"
                label="Subject"
              ></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.issuer" label="Issuer"></v-text-field>
            </td>
            <td colspan="4"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :data="certs"
          type="csv"
          name="TWSNMP_FC_Cert_List.csv"
          header="TWSNMP FC Cert List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="certs"
          type="xls"
          name="TWSNMP_FC_Cert_List.xls"
          header="TWSNMP FC Cert List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="primary" dark @click="addDialog = true">
          <v-icon>mdi-plus</v-icon>
          追加
        </v-btn>
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
          <span class="headline">サーバー証明書削除</span>
        </v-card-title>
        <v-card-text> 選択したサーバー証明書を削除しますか？ </v-card-text>
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
        <v-card-text> サーバー証明書の信用度を再計算しますか？ </v-card-text>
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
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー証明書情報</span>
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
                <td>ターゲット</td>
                <td>{{ selected.Target }}:{{ selected.Port }}</td>
              </tr>
              <tr>
                <td>証明内容</td>
                <td>{{ selected.Subject }}</td>
              </tr>
              <tr>
                <td>発行者</td>
                <td>{{ selected.Issuer }}</td>
              </tr>
              <tr>
                <td>有効期間</td>
                <td>
                  {{ selected.NotBeforeDate }} - {{ selected.NotAfterDate }}
                </td>
              </tr>
              <tr>
                <td>シリアル番号</td>
                <td>{{ selected.SerialNumber }}</td>
              </tr>
              <tr>
                <td>検証済み</td>
                <td>{{ selected.Verify }}</td>
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
                <td>エラー</td>
                <td>{{ selected.Error }}</td>
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
          <v-btn color="error" dark @click="doDelete">
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
    <v-dialog v-model="addDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー追加</span>
        </v-card-title>
        <v-alert v-model="addError" color="error" dense dismissible>
          対象サーバーを追加できません。
        </v-alert>
        <v-card-text>
          <v-text-field v-model="add.Target" label="ターゲット"></v-text-field>
          <v-text-field v-model="add.Port" label="ポート番号"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doAdd">
            <v-icon>mdi-content-save</v-icon>
            追加
          </v-btn>
          <v-btn color="normal" dark @click="addDialog = false">
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
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'ターゲット',
          value: 'Target',
          width: '20%',
          filter: (value) => {
            if (!this.conf.target) return true
            return value.includes(this.conf.target)
          },
        },
        { text: 'ポート', value: 'Port', width: '5%' },
        {
          text: '証明内容',
          value: 'Subject',
          width: '18%',
          filter: (value) => {
            if (!this.conf.subject) return true
            return value.includes(this.conf.subject)
          },
        },
        {
          text: '発行者',
          value: 'Issuer',
          width: '17%',
          filter: (value) => {
            if (!this.conf.issuer) return true
            return value.includes(this.conf.issuer)
          },
        },
        { text: '検証', value: 'Verify', width: '5%' },
        { text: '期限', value: 'NotAfterDate', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      certs: [],
      selected: {},
      infoDialog: false,
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      add: { Target: '', Port: 443 },
      addDialog: false,
      addError: false,
      conf: {
        target: '',
        subject: '',
        issuer: '',
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.certs = await this.$axios.$get('/api/report/cert')
    if (!this.certs) {
      return
    }
    this.certs.forEach((c) => {
      c.First = this.$timeFormat(
        new Date(c.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      c.Last = this.$timeFormat(
        new Date(c.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      if (c.NotAfter) {
        c.NotAfterDate = this.$timeFormat(
          new Date(c.NotAfter * 1000),
          '{yyyy}/{MM}/{dd}'
        )
        c.NotBeforeDate = this.$timeFormat(
          new Date(c.NotBefore * 1000),
          '{yyyy}/{MM}/{dd}'
        )
      } else {
        c.NotAfterDate = ''
        c.NotBeforeDate = ''
      }
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.cert.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/cert/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/cert/' + this.selected.ID)
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
        .post('/api/report/cert/reset', {})
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
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
    openAddDialog() {
      this.addDialog = true
    },
    doAdd() {
      const url = '/api/report/cert'
      this.add.Port *= 1
      this.$axios
        .post(url, this.add)
        .then(() => {
          this.addDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.addError = true
        })
    },
  },
}
</script>
