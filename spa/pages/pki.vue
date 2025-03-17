<template>
  <v-row>
    <v-card v-if="caValid" min-width="1000px" width="100%">
      <v-card-title primary-title> PKI 証明証リスト </v-card-title>
      <v-data-table
        :headers="headers"
        :items="certs"
        dense
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.Status`]="{ item }">
          <v-icon
            :color="
              item.Status == 'revoked'
                ? '#e31a1c'
                : item.Status == 'expired'
                ? '#fb9a99'
                : '#1f78b4'
            "
          >
            mdi-certificate
          </v-icon>
          {{
            item.Status == 'revoked'
              ? '失効'
              : item.Status == 'expired'
              ? '期限切れ'
              : '有効'
          }}
        </template>
        <template #[`item.Created`]="{ item }">
          {{ formatTime(item.Created) }}
        </template>
        <template #[`item.Expire`]="{ item }">
          {{ formatTime(item.Expire) }}
        </template>
        <template #[`item.Revoked`]="{ item }">
          {{ formatTime(item.Revoked) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="exportCert(item.ID)"> mdi-download </v-icon>
          <v-icon
            v-if="item.Type != 'system' && item.Status == 'valid'"
            color="red"
            small
            @click="showRevokeCert(item.ID)"
          >
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-select
                v-model="conf.Status"
                :items="statusList"
                label="status"
              >
              </v-select>
            </td>
            <td>
              <v-text-field v-model="conf.ID" label="ID"></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.Subject"
                label="Subject"
              ></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.Node" label="Node"></v-text-field>
            </td>
            <td></td>
            <td></td>
            <td></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="info" dark @click="csrDialog = true">
          <v-icon>mdi-book-lock-outline</v-icon>
          証明証要求(CSR)作成
        </v-btn>
        <v-btn color="info" dark @click="certDialog = true">
          <v-icon>mdi-certificate</v-icon>
          証明証発行
        </v-btn>
        <v-btn color="info" dark @click="showPKIControl">
          <v-icon>mdi-toggle-switch</v-icon>
          PKIサーバー制御
        </v-btn>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_PKICert_List.csv"
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
          name="TWSNMP_FC_PKICert_List.xls"
          header="TWSNMP FCが発行した証明書リスト"
          worksheet="証明書"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="error" dark @click="destroyDialog = true">
          <v-icon>mdi-trash</v-icon>
          CA破棄
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-card v-else min-width="600px" width="95%" class="mx-auto">
      <v-form>
        <v-card-title primary-title> PKI CAの構築 </v-card-title>
        <v-alert v-model="errorCreateCA" color="error" dense dismissible>
          CAを構築できません
        </v-alert>
        <v-card-text>
          <v-select
            v-model="createCAReq.RootCAKeyType"
            :items="keyTypeList"
            label="鍵の種類"
          />
          <v-text-field v-model="createCAReq.Name" label="名前" />
          <v-text-field v-model="createCAReq.SANs" label="DNS名" />
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="createCAReq.AcmeBaseURL"
                label="ACMEサーバーの基本URL"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="createCAReq.AcmePort"
                label="ACMEポート"
                type="number"
                min="1"
                max="65534"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                v-model="createCAReq.HttpBaseURL"
                label="CRL/OCSP/SCEPサーバー基本URL"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="createCAReq.HttpPort"
                label="CRL/OCSP/SCEPサーバーポート"
                type="number"
                min="1"
                max="65534"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-text-field
                v-model="createCAReq.RootCATerm"
                label="CA証明書の期間(年)"
                type="number"
                min="1"
                max="100"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="createCAReq.CrlInterval"
                label="CRLの更新間隔(時間)"
                type="number"
                min="1"
                max="192"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="createCAReq.CertTerm"
                label="証明の期間(時間)"
                type="number"
                min="1"
                max="8760"
              />
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :loading="wait"
            :disabled="wait"
            color="primary"
            dark
            @click="createCA"
          >
            <v-icon>mdi-key-plus</v-icon>
            CA構築
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
    <v-dialog v-model="destroyDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">CS破棄</span>
        </v-card-title>
        <v-alert v-model="errorDestroyCA" color="error" dense dismissible>
          CAを破棄できませんでした
        </v-alert>
        <v-card-text> CAを破棄しましすか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="destroyCA()">
            <v-icon>mdi-trash</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="destroyDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="revokeDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">証明書の失効</span>
        </v-card-title>
        <v-card-text> 選択した証明書を失効しましすか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="revokeCert()">
            <v-icon>mdi-trash</v-icon>
            失効
          </v-btn>
          <v-btn color="normal" @click="revokeDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-snackbar v-model="doneRevoke" absolute centered color="primary">
      証明書を失効しました
    </v-snackbar>
    <v-dialog v-model="csrDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> 証明書要求(CSR)の作成 </v-card-title>
        <v-alert v-model="errorCreateCSR" color="error" dense dismissible>
          CSRを作成できませんでした
        </v-alert>
        <v-card-text>
          <v-select
            v-model="csrReq.KeyType"
            :items="keyTypeList"
            label="鍵の種類"
          />
          <v-text-field
            v-model="csrReq.CommonName"
            label="名前(CN)"
          ></v-text-field>
          <v-text-field
            v-model="csrReq.Sans"
            label="DNS名(SANs)"
          ></v-text-field>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="csrReq.Country"
                label="国コード(C)"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="csrReq.Province"
                label="州/都道府県名(ST)"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="csrReq.Locality"
                label="市町村名(L)"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="csrReq.Organization"
                label="組織名(O)"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="csrReq.OrganizationalUnit"
                label="組織単位(OU)"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :loading="waitCSR"
            :disabled="waitCSR"
            color="primary"
            dark
            @click="createCSR"
          >
            <v-icon>mdi-content-save</v-icon>
            作成
          </v-btn>
          <v-btn color="normal" dark @click="csrDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-snackbar v-model="doneCSR" absolute centered color="primary">
      証明書要求を作成しました
    </v-snackbar>
    <v-dialog v-model="certDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> 証明書の発行 </v-card-title>
        <v-alert v-model="errorCreateCert" color="error" dense dismissible>
          証明書を発行できませんでした
        </v-alert>
        <v-card-text>
          <v-file-input label="証明書要求(CSR)" @change="selectCSRFile">
          </v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :loading="waitCert"
            :disabled="waitCert"
            color="primary"
            dark
            @click="createCert"
          >
            <v-icon>mdi-content-save</v-icon>
            発行
          </v-btn>
          <v-btn color="normal" dark @click="certDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-snackbar v-model="doneCert" absolute centered color="primary">
      証明書を発行しました
    </v-snackbar>
    <v-dialog v-model="pkiControlDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> PKIサーバー制御 </v-card-title>
        <v-alert v-model="errorPKIControl" color="error" dense dismissible>
          PKIサーバーの設定を変更できませんでした
        </v-alert>
        <v-alert
          dense
          outlined
          :type="
            pkiControl.AcmeStatus.indexOf('error') != -1 ? 'error' : 'info'
          "
        >
          ACMEサーバーの状態 : {{ pkiControl.AcmeStatus }}
        </v-alert>
        <v-alert
          dense
          outlined
          :type="
            pkiControl.HttpStatus.indexOf('error') != -1 ? 'error' : 'info'
          "
        >
          CRL/OCSP/SCEPサーバーの状態 : {{ pkiControl.HttpStatus }}
        </v-alert>
        <v-card-text>
          <v-row>
            <v-col>
              <v-switch
                v-model="pkiControl.EnableAcme"
                label="ACMEサーバー"
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-model="pkiControl.EnableHttp"
                label="CRL/OCSP/SCEPサーバー"
              ></v-switch>
            </v-col>
          </v-row>
          <v-text-field
            v-model="pkiControl.AcmeBaseURL"
            label="ACMEサーバーの基本URL"
          />
          <v-row>
            <v-col>
              <v-text-field
                v-model="pkiControl.CrlInterval"
                label="CRLの更新間隔(時間)"
                type="number"
                min="1"
                max="192"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="pkiControl.CertTerm"
                label="証明の期間(時間)"
                type="number"
                min="1"
                max="8760"
              />
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :loading="waitPKIControl"
            :disabled="waitPKIControl"
            color="primary"
            dark
            @click="updatePkiControl"
          >
            <v-icon>mdi-content-save</v-icon>
            変更
          </v-btn>
          <v-btn color="normal" dark @click="pkiControlDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import { saveAs } from 'file-saver'
export default {
  data() {
    return {
      errorCreateCA: false,
      errorDestroyCA: false,
      caValid: false,
      wait: false,
      csrDialog: false,
      pkiControlDialog: false,
      destroyDialog: false,
      revokeDialog: false,
      doneRevoke: false,
      certDialog: false,
      selectedCertID: '',
      certs: [],
      createCAReq: {},
      csrReq: {
        KeyType: 'rsa-4096',
        CommonName: '',
        OrganizationalUnit: '',
        Organization: '',
        Locality: '',
        Province: '',
        Country: '',
        Sans: '',
      },
      waitCSR: false,
      errorCreateCSR: false,
      doneCSR: false,
      csrFile: undefined,
      waitCert: false,
      errorCreateCert: false,
      doneCert: false,
      pkiControl: {
        EnableAcme: false,
        EnableHttp: false,
        AcmeBaseURL: '',
        CertTerm: 86400,
        CrlInterval: 24,
        AcmeStatus: '',
        HttpStatus: '',
      },
      waitPKIControl: false,
      errorPKIControl: false,
      headers: [
        {
          text: '状態',
          value: 'Status',
          width: '10%',
          filter: (value) => {
            if (this.conf.Status === '') return true
            return this.conf.Status === value
          },
        },
        {
          text: 'ID',
          value: 'ID',
          width: '10%',
          filter: (value) => {
            if (!this.conf.ID) return true
            return value.includes(this.conf.ID)
          },
        },
        {
          text: 'Subject',
          value: 'Subject',
          width: '18%',
          filter: (value) => {
            if (!this.conf.Subject) return true
            return value.includes(this.conf.Subject)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: '関連ノード',
          value: 'Node',
          width: '15%',
          filter: (value) => {
            if (!this.conf.Node) return true
            return value.includes(this.conf.Node)
          },
        },
        {
          text: '開始',
          value: 'Created',
          width: '12%',
        },
        {
          text: '終了',
          value: 'Expire',
          width: '12%',
        },
        {
          text: '失効',
          value: 'Revoked',
          width: '12%',
        },
        { text: '操作', value: 'actions', width: '8%' },
      ],
      conf: {
        Status: '',
        ID: '',
        Subject: '',
        Node: '',
        sortBy: 'Created',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      statusList: [
        { text: 'すべて', value: '' },
        { text: '有効', value: 'valid' },
        { text: '期限切れ', value: 'expired' },
        { text: '失効', value: 'revoked' },
      ],
      keyTypeList: [
        { text: 'RSA 2048bits', value: 'rsa-2048' },
        { text: 'RSA 4096bits', value: 'rsa-4096' },
        { text: 'RSA 8192bits', value: 'rsa-8192' },
        { text: 'ECDSA P224', value: 'ecdsa-224' },
        { text: 'ECDSA P256', value: 'ecdsa-256' },
        { text: 'ECDSA P384', value: 'ecdsa-384' },
        { text: 'ECDSA P512', value: 'ecdsa-512' },
      ],
    }
  },
  async fetch() {
    this.caValid = await this.$axios.$get('/api/pki/hasCA')
    if (this.caValid) {
      this.certs = await this.$axios.$get('/api/pki/certs')
    } else {
      this.createCAReq = await this.$axios.$get('/api/pki/createCA')
    }
  },
  created() {
    const c = this.$store.state.pki.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('pki/setConf', this.conf)
  },
  methods: {
    async revoke(item) {
      await this.$axios.post('/api/pki/revoke/' + item.ID, {})
      this.$fetch()
    },
    async createCA() {
      this.wait = true
      const r = await this.$axios.post('/api/pki/createCA', this.createCAReq)
      if (!r || r.status !== 200) {
        this.errorCreateCA = true
      }
      this.wait = false
      this.$fetch()
    },
    async destroyCA() {
      const r = await this.$axios.post('/api/pki/destroyCA', {})
      if (!r || r.status !== 200) {
        this.errorDestroyCA = true
        return
      }
      this.destroyDialog = false
      this.$fetch()
    },
    async createCSR() {
      this.waitCSR = true
      const r = await this.$axios.post('/api/pki/createCSR', this.csrReq, {
        responseType: 'blob',
      })
      this.waitCSR = false
      if (!r || r.status !== 200) {
        this.csrError = true
        return
      }
      const fn = this.$timeFormat(new Date(), '{yyyy}{MM}{dd}{HH}{mm}')
      saveAs(r.data, 'csr_' + fn + '.zip')
      this.csrDialog = false
      this.doneCSR = true
    },
    async showPKIControl() {
      const r = await this.$axios.$get('/api/pki/control')
      if (r) {
        this.pkiControl = r
      }
      this.pkiControlDialog = true
    },
    async updatePkiControl() {
      this.waitPKIControl = true
      const r = await this.$axios.post('/api/pki/control', this.pkiControl)
      if (!r || r.status !== 200) {
        this.errorPKIControl = true
        return
      }
      this.waitPKIControl = false
      this.pkiControlDialog = false
    },
    async exportCert(id) {
      const r = await this.$axios.get('/api/pki/cert/' + id, {
        responseType: 'blob',
      })
      if (!r || r.status !== 200) {
        return
      }
      const fn = this.$timeFormat(new Date(), '{yyyy}{MM}{dd}{HH}{mm}')
      saveAs(r.data, 'crt_' + fn + '.pem')
    },
    showRevokeCert(id) {
      this.selectedCertID = id
      this.revokeDialog = true
    },
    async revokeCert() {
      if (this.selectedCertID) {
        await this.$axios.delete('/api/pki/revoke/' + this.selectedCertID)
        this.$fetch()
      }
      this.revokeDialog = false
      this.doneRevoke = true
    },
    selectCSRFile(f) {
      this.csrFile = f
    },
    async createCert() {
      const formData = new FormData()
      formData.append('file', this.csrFile)
      this.errorCreateCert = false
      this.waitCert = true
      const r = await this.$axios.post('/api/pki/createCRT', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      })
      this.waitCert = false
      if (!r || r.status !== 200) {
        this.errorCreateCert = true
        return
      }
      this.certDialog = false
      this.$fetch()
      this.doneCert = true
    },
    makeExports() {
      const exports = []
      if (!this.certs) {
        return exports
      }
      this.certs.forEach((e) => {
        if (!this.filterCert(e)) {
          return
        }
        exports.push({
          状態: e.Status,
          ID: e.ID,
          Subject: e.Subject,
          関連ノード: e.Node,
          開始: this.formatTime(e.Created),
          終了: this.formatTime(e.Expire),
          失効: this.formatTime(e.Revoked),
        })
      })
      return exports
    },
    filterCert(e) {
      if (this.conf.Staus && this.conf.Status !== e.Status) {
        return false
      }
      if (this.conf.ID && !e.ID.includes(this.conf.ID)) {
        return false
      }
      if (this.conf.Subject && !e.Subject.includes(this.conf.Subject)) {
        return false
      }
      if (this.conf.Node && !e.Node.includes(this.conf.Node)) {
        return false
      }
      return true
    },
    formatTime(t) {
      if (t < 1) {
        return ''
      }
      return this.$timeFormat(
        new Date(t / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
    },
  },
}
</script>
