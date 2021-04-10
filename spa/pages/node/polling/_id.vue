<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        ポーリング - {{ node.Name }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        ポーリングを削除できませんでした
      </v-alert>
      <v-alert v-model="updateError" color="error" dense dismissible>
        ポーリングを変更できませんでした
      </v-alert>
      <v-data-table
        v-model="selectedPollings"
        :headers="headers"
        :items="pollings"
        :search="search"
        item-key="ID"
        show-select
        dense
        :items-per-page="15"
        sort-by="State"
        sort-asec
      >
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template v-slot:[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small @click="$router.push({ path: '/polling/' + item.ID })">
            mdi-eye
          </v-icon>
          <v-icon
            v-if="item.LogMode > 1"
            small
            @click="$router.push({ path: '/report/ai/' + item.ID })"
          >
            mdi-brain
          </v-icon>
          <v-icon small @click="editPollingFunc(item)"> mdi-pencil </v-icon>
          <v-icon small @click="deletePollingFunc(item)"> mdi-delete </v-icon>
          <v-icon small @click="copyPolling(item)"> mdi-content-copy </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="addPolling">
          <v-icon>mdi-plus</v-icon>
          追加
        </v-btn>
        <v-btn color="primary" dark @click="showAutoAddDialog">
          <v-icon>mdi-brain</v-icon>
          自動追加
        </v-btn>
        <v-btn
          v-if="hasSelectedPollings"
          color="primary"
          dark
          @click="setPollingLevelDialog = true"
        >
          <v-icon>mdi-cog</v-icon>
          レベル変更
        </v-btn>
        <v-btn
          v-if="hasSelectedPollings"
          color="primary"
          dark
          @click="setPollingLogModeDialog = true"
        >
          <v-icon>mdi-alert</v-icon>
          ログモード変更
        </v-btn>
        <v-btn
          v-if="hasSelectedPollings"
          color="error"
          @click="deleteSelectedPollingDialog = true"
        >
          <v-icon>mdi-delete</v-icon>
          一括削除
        </v-btn>
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
    <v-dialog v-model="editDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング編集</span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="editPolling.Name" label="名前"></v-text-field>
          <v-select
            v-model="editPolling.Level"
            :items="$levelList"
            label="レベル"
          >
          </v-select>
          <v-select v-model="editPolling.Type" :items="$typeList" label="種別">
          </v-select>
          <v-text-field
            v-model="editPolling.Mode"
            label="モード"
          ></v-text-field>
          <v-text-field
            v-model="editPolling.Params"
            label="パラメータ"
          ></v-text-field>
          <v-text-field
            v-model="editPolling.Filter"
            label="フィルター"
          ></v-text-field>
          <v-text-field
            v-model="editPolling.Extractor"
            label="抽出パターン"
          ></v-text-field>
          <v-textarea
            v-model="editPolling.Script"
            label="判定スクリプト"
            clearable
            rows="3"
            clear-icon="mdi-close-circle"
          ></v-textarea>
          <v-slider
            v-model="editPolling.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="600"
            min="60"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="editPolling.PollInt"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="editPolling.Timeout"
            label="タイムアウト(Sec)"
            class="align-center"
            max="10"
            min="1"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="editPolling.Timeout"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-slider
            v-model="editPolling.Retry"
            label="リトライ回数"
            class="align-center"
            max="5"
            min="0"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="editPolling.Retry"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-select
            v-model="editPolling.LogMode"
            :items="$logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="editIndex === -1"
            color="primary"
            dark
            @click="showTemplateDialog"
          >
            <v-icon>mdi-content-copy</v-icon>
            テンプレート
          </v-btn>
          <v-btn color="primary" dark @click="doUpdatePolling">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="closeEdit">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング削除</span>
        </v-card-title>
        <v-card-text>
          ポーリング{{ deletePolling.Name }}を削除しますか？
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeletePolling">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="closeDelete">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog
      v-model="deleteSelectedPollingDialog"
      persistent
      max-width="500px"
    >
      <v-card>
        <v-card-title>
          <span class="headline">ポーリング一括削除</span>
        </v-card-title>
        <v-card-text> 選択したポーリングを全て削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="deletePollings">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteSelectedPollingDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="setPollingLevelDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">レベル変更</span>
        </v-card-title>
        <v-card-text>
          <v-select v-model="newLevel" :items="$levelList" label="レベル">
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setPollingLevel">
            <v-icon>mdi-content-</v-icon>
            変更
          </v-btn>
          <v-btn color="normal" @click="setPollingLevelDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="setPollingLogModeDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ログモード変更</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-model="newLogMode"
            :items="$logModeList"
            label="ログモード"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setPollingLogMode">
            <v-icon>mdi-content-</v-icon>
            変更
          </v-btn>
          <v-btn color="normal" @click="setPollingLogModeDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="templateDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span v-if="autoAdd" class="headline">テンプレートから自動追加</span>
          <span v-else class="headline">テンプレートから選択</span>
          <v-spacer></v-spacer>
          <v-text-field
            v-model="searchTemplate"
            append-icon="mdi-magnify"
            label="検索"
            single-line
            hide-details
          ></v-text-field>
        </v-card-title>
        <v-data-table
          v-model="selectedTemplate"
          :headers="headersTemplate"
          :items="templates"
          :single-select="!autoAdd"
          :search="searchTemplate"
          item-key="ID"
          show-select
          :items-per-page="15"
          sort-by="Type"
          dense
        >
          <template v-slot:[`item.Level`]="{ item }">
            <v-icon :color="$getStateColor(item.Level)">{{
              $getStateIconName(item.Level)
            }}</v-icon>
            {{ $getStateName(item.Level) }}
          </template>
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn v-if="autoAdd" color="primary" @click="doAutoAddPolling">
            <v-icon>mdi-plus</v-icon>
            追加
          </v-btn>
          <v-btn v-else color="primary" @click="selectTemplate">
            <v-icon>md-check</v-icon>
            選択
          </v-btn>
          <v-btn color="normal" @click="templateDialog = false">
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
  async fetch() {
    const r = await this.$axios.$get(
      '/api/node/polling/' + this.$route.params.id
    )
    this.node = r.Node
    if (r.Pollings) {
      this.pollings = r.Pollings
      this.pollings.forEach((e) => {
        const t = new Date(e.LastTime / (1000 * 1000))
        e.TimeStr = this.$timeFormat(t)
      })
    }
  },
  data() {
    return {
      node: {},
      editDialog: false,
      deleteDialog: false,
      editIndex: -1,
      deleteIndex: -1,
      deleteError: false,
      updateError: false,
      editPolling: {},
      deletePolling: {},
      search: '',
      headers: [
        { text: '状態', value: 'State', width: '15%' },
        { text: '名前', value: 'Name', width: '30%' },
        { text: 'レベル', value: 'Level', width: '14%' },
        { text: '種別', value: 'Type', width: '14%' },
        { text: '最終実施', value: 'TimeStr', width: '15%' },
        { text: '操作', value: 'actions', width: '12%' },
      ],
      pollings: [],
      autoAdd: false,
      searchTemplate: '',
      headersTemplate: [
        {
          text: '名前',
          value: 'Name',
          width: '25%',
        },
        {
          text: 'レベル',
          value: 'Level',
          width: '15%',
          filter: this.filterAutoAdd,
        },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
        },
        { text: 'モード', value: 'Mode', width: '10%' },
        {
          text: '説明',
          value: 'Descr',
          width: '40%',
        },
      ],
      selectedTemplate: [],
      templateDialog: false,
      templates: [],
      selectedPollings: [],
      deleteSelectedPollingDialog: false,
      setPollingLevelDialog: false,
      newLevel: 'off',
      setPollingLogModeDialog: false,
      newLogMode: 0,
    }
  },
  computed: {
    hasSelectedPollings() {
      return this.selectedPollings.length > 0
    },
  },
  methods: {
    editPollingFunc(item) {
      this.editIndex = this.pollings.indexOf(item)
      this.editPolling = Object.assign({}, item)
      this.editDialog = true
    },
    copyPolling(item) {
      this.editIndex = -1
      this.editPolling = Object.assign({}, item)
      this.editPolling.ID = ''
      this.editDialog = true
    },
    addPolling() {
      this.editIndex = -1
      this.editPolling = {
        Name: '',
        NodeID: this.node.ID,
        Type: 'ping',
        Mode: '',
        Params: '',
        Filter: '',
        Extractor: '',
        Script: '',
        Level: 'low',
        PollInt: 60,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.editDialog = true
    },
    deletePollingFunc(item) {
      this.deleteIndex = this.pollings.indexOf(item)
      this.deletePolling = Object.assign({}, item)
      this.deleteDialog = true
    },
    doDeletePolling() {
      this.pollings.splice(this.deleteIndex, 1)
      this.$axios
        .post('/api/pollings/delete', [this.deletePolling.ID])
        .catch((e) => {
          this.$fetch()
          this.deleteError = true
        })
      this.closeDelete()
    },
    deletePollings() {
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.$axios
        .post('/api/pollings/delete', ids)
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.deleteError = true
        })
      this.deleteSelectedPollingDialog = false
    },
    setPollingLevel() {
      if (!this.newLevel) {
        return
      }
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.$axios
        .post('/api/pollings/setlevel', {
          IDs: ids,
          Level: this.newLevel,
        })
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.updateError = true
        })
      this.setPollingLevelDialog = false
    },
    setPollingLogMode() {
      const ids = []
      this.selectedPollings.forEach((p) => {
        ids.push(p.ID)
      })
      this.$axios
        .post('/api/pollings/setlogmode', {
          IDs: ids,
          LogMode: this.newLogMode,
        })
        .then(() => {
          this.$fetch()
          this.selectedPollings = []
        })
        .catch((e) => {
          this.$fetch()
          this.updateError = true
        })
      this.setPollingLogModeDialog = false
    },
    closeDelete() {
      this.deleteDialog = false
      this.$nextTick(() => {
        this.deleteIndex = -1
      })
    },
    closeEdit() {
      this.editDialog = false
      this.$nextTick(() => {
        this.editIndex = -1
      })
    },
    doUpdatePolling() {
      if (this.editIndex > -1) {
        Object.assign(this.pollings[this.editIndex], this.editPolling)
        this.$axios.post('/api/polling/update', this.editPolling).catch((e) => {
          this.updateError = true
          this.$fetch()
        })
      } else {
        this.$axios
          .post('/api/polling/add', this.editPolling)
          .then(() => {
            this.$fetch()
          })
          .catch((e) => {
            this.updateError = true
            this.$fetch()
          })
      }
      this.closeEdit()
    },
    async showTemplateDialog() {
      this.autoAdd = false
      this.selectedTemplate = []
      const r = await this.$axios.$get('/api/polling/template')
      if (r) {
        this.templates = r
        this.templateDialog = true
      }
    },
    async showAutoAddDialog() {
      this.selectedTemplate = []
      const r = await this.$axios.$get('/api/polling/template')
      if (r) {
        this.templates = r
        this.autoAdd = true
        this.templateDialog = true
      }
    },
    doAutoAddPolling() {
      if (!this.selectedTemplate || this.selectedTemplate.length < 1) {
        return
      }
      const p = {
        NodeID: this.node.ID,
        PollingTemplateIDs: [],
      }
      for (let i = 0; i < this.selectedTemplate.length; i++) {
        p.PollingTemplateIDs.push(this.selectedTemplate[i].ID)
      }
      this.$axios
        .post('/api/polling/auto', p)
        .then(() => {
          this.$fetch()
          this.templateDialog = false
        })
        .catch((e) => {
          this.updateError = true
          this.$fetch()
        })
    },
    selectTemplate() {
      if (!this.selectedTemplate || this.selectedTemplate.length !== 1) {
        return
      }
      this.editPolling.Name = this.selectedTemplate[0].Name
      this.editPolling.Type = this.selectedTemplate[0].Type
      this.editPolling.Mode = this.selectedTemplate[0].Mode
      this.editPolling.Params = this.selectedTemplate[0].Params
      this.editPolling.Filter = this.selectedTemplate[0].Filter
      this.editPolling.Extractor = this.selectedTemplate[0].Extractor
      this.editPolling.Script = this.selectedTemplate[0].Script
      this.templateDialog = false
    },
    filterAutoAdd(value, search, item) {
      return !(this.autoAdd && item.AutoMode === 'disable')
    },
  },
}
</script>
