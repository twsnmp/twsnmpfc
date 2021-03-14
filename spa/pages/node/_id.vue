<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        イベントログ - {{ node.Name }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="logSearch"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="logHeaders"
        :items="logs"
        :search="logSearch"
        sort-by="TimeStr"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template v-slot:[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
      </v-data-table>
      <v-card-title>
        ポーリング - {{ node.Name }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="pollingSearch"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-alert v-model="deleteError" type="error" dense dismissible>
        ポーリングを削除できませんでした
      </v-alert>
      <v-alert v-model="updateError" type="error" dense dismissible>
        ポーリングを変更できませんでした
      </v-alert>
      <v-data-table
        :headers="pollingHeaders"
        :items="pollings"
        :search="pollingSearch"
        dense
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
          <v-icon
            v-if="item.LogMode > 0"
            small
            @click="$router.push({ path: '/polling/' + item.ID })"
          >
            mdi-eye
          </v-icon>
          <v-icon small @click="editPollingFunc(item)"> mdi-pencil </v-icon>
          <v-icon small @click="deletePollingFunc(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="addPolling">
          <v-icon>mdi-plus</v-icon>
          追加
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
  </v-row>
</template>

<script>
export default {
  async fetch() {
    const r = await this.$axios.$get('/api/node/' + this.$route.params.id)
    this.node = r.Node
    if (r.Logs) {
      this.logs = r.Logs
      this.logs.forEach((e) => {
        const t = new Date(e.Time / (1000 * 1000))
        e.TimeStr = this.$timeFormat(t)
      })
    }
    if (r.Pollings) {
      this.pollings = r.Pollings
      this.pollings.forEach((e) => {
        const t = new Date(e.LastTime / (1000 * 1000))
        e.TimeStr = this.$timeFormat(t)
      })
    }
    this.$showLogLevelChart(this.logs)
  },
  data() {
    return {
      node: {},
      logSearch: '',
      logHeaders: [
        { text: '状態', value: 'Level', width: '10%' },
        { text: '発生日時', value: 'TimeStr', width: '15%' },
        { text: '種別', value: 'Type', width: '10%' },
        { text: 'イベント', value: 'Event', width: '50%' },
      ],
      logs: [],
      editDialog: false,
      deleteDialog: false,
      editIndex: -1,
      deleteIndex: -1,
      deleteError: false,
      updateError: false,
      editPolling: {},
      deletePolling: {},
      pollingSearch: '',
      pollingHeaders: [
        { text: '状態', value: 'State', width: '15%' },
        { text: '名前', value: 'Name', width: '30%' },
        { text: 'レベル', value: 'Level', width: '15%' },
        { text: '種別', value: 'Type', width: '15%' },
        { text: '最終実施', value: 'TimeStr', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      pollings: [],
    }
  },
  mounted() {
    this.$makeLogLevelChart('logCountChart')
    this.$showLogLevelChart(this.logs)
  },
  methods: {
    editPollingFunc(item) {
      this.editIndex = this.pollings.indexOf(item)
      this.editPolling = Object.assign({}, item)
      this.editDialog = true
    },
    addPolling() {
      this.editIndex = -1
      this.editPolling = {
        Name: '',
        NodeID: this.node.ID,
        Type: 'ping',
        Polling: '',
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
        .post('/api/polling/delete', { ID: this.deletePolling.ID })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.closeDelete()
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
  },
}
</script>
