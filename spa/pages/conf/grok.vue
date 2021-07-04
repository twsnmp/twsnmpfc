<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="95%">
      <v-card-title primary-title> 抽出パターン(Grok)設定 </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        抽出パターン(Grok)を取得できません
      </v-alert>
      <v-alert v-model="exportError" color="error" dense dismissible>
        抽出パターンのエクスポートに失敗しました
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="grok"
        sort-by="ID"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
      >
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="editGrok(item)"> mdi-pencil </v-icon>
          <v-icon small @click="deleteGrok(item)"> mdi-delete </v-icon>
          <v-icon small @click="copyGrok(item)"> mdi-content-copy </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="addGrok">
          <v-icon>mdi-plus</v-icon>
          追加
        </v-btn>
        <v-btn color="primary" dark @click="importDialog = true">
          <v-icon>mdi-upload</v-icon>
          インポート
        </v-btn>
        <v-btn color="primary" dark @click="exportGrok">
          <v-icon>mdi-download</v-icon>
          エクスポート
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-btn color="normal" dark to="/map">
          <v-icon>mdi-lan</v-icon>
          マップ
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="80%">
      <v-card>
        <v-card-title>
          <span class="headline"> 抽出パターン(Grok)編集 </span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="selected.ID" label="ID"></v-text-field>
          <v-text-field v-model="selected.Name" label="名前"></v-text-field>
          <v-textarea
            v-model="selected.Descr"
            label="説明"
            clearable
            rows="3"
            clear-icon="mdi-close-circle"
          ></v-textarea>
          <v-text-field
            v-model="selected.Pat"
            label="パターン(Grok)"
          ></v-text-field>
          <v-text-field v-model="selected.Ok" label="正常値"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doUpdateGrok">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="testDialog = true">
            <v-icon>mdi-check</v-icon>
            テスト
          </v-btn>
          <v-btn color="normal" @click="editDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="testDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline"> 抽出パターン(Grok)テスト </span>
        </v-card-title>
        <v-alert v-model="testError" color="error" dense dismissible>
          抽出パターンのテストに失敗しました
        </v-alert>
        <v-alert v-model="testNoData" color="lime" dense dismissible>
          抽出したデータはありません。
        </v-alert>
        <v-card-text>
          <v-text-field
            v-model="selected.Pat"
            label="パターン(Grok)"
          ></v-text-field>
          <v-textarea
            v-model="testData"
            label="テストデータ"
            clearable
            rows="5"
            clear-icon="mdi-close-circle"
          ></v-textarea>
        </v-card-text>
        <v-card-subtitle> 抽出結果 </v-card-subtitle>
        <v-card-text>
          <v-data-table
            :headers="extractHeader"
            :items="extractDatas"
            :items-per-page="10"
            dense
          >
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doTestGrok">
            <v-icon>mdi-content-save</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="testDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">抽出パターン削除</span>
        </v-card-title>
        <v-card-text>
          抽出パターン{{ selected.Name }}を削除しますか？
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteGrok">
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
    <v-dialog v-model="importDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">抽出パターンのインポート</span>
        </v-card-title>
        <v-alert v-model="importError" color="error" dense dismissible>
          抽出パターンのインポートに失敗しました
        </v-alert>
        <v-card-text>
          <v-file-input label="パターン定義ファイル" @change="selectGrokFile">
          </v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="importGrok">
            <v-icon>mdi-content-save</v-icon>
            インポート
          </v-btn>
          <v-btn color="normal" @click="importDialog = false">
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
        { text: 'ID', value: 'ID', width: '15%' },
        { text: '名前', value: 'Name', width: '20%' },
        { text: '説明', value: 'Descr', width: '50%' },
        { text: '操作', value: 'actions', width: '15%' },
      ],
      grok: [],
      selected: {
        ID: '',
        Name: '',
        Descr: '',
        Pat: '',
        Ok: '',
      },
      editDialog: false,
      updateError: false,
      exportError: false,
      deleteDialog: false,
      deleteError: false,
      importDialog: false,
      importError: false,
      grokFile: null,
      testError: false,
      testNoData: false,
      testDialog: false,
      testData: '',
      extractHeader: [],
      extractDatas: [],
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/conf/grok')
    if (!r) {
      return
    }
    this.grok = r
  },
  methods: {
    copyGrok(item) {
      this.selected = item
      this.selected.ID += '-Copy'
      this.editDialog = true
    },
    editGrok(item) {
      this.selected = item
      this.editDialog = true
    },
    addGrok() {
      this.selected = {
        ID: 'New',
        Name: '',
        Descr: '',
        Pat: '',
        Ok: '',
      }
      this.editDialog = true
    },
    deleteGrok(item) {
      this.selected = item
      this.deleteDialog = true
    },
    doUpdateGrok() {
      this.$axios
        .post('/api/conf/grok', this.selected)
        .then(() => {
          this.editDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.updateError = true
          this.$fetch()
        })
    },
    doTestGrok() {
      this.testNoData = false
      this.testError = false
      this.$axios
        .post('/api/test/grok', { Pat: this.selected.Pat, Data: this.testData })
        .then((resp) => {
          this.setExtractData(resp.data)
        })
        .catch((e) => {
          this.testError = true
        })
    },
    setExtractData(r) {
      this.extractDatas = []
      this.extractHeader = []
      if (r.ExtractHeader.length < 1 || r.ExtractDatas.length < 1) {
        this.testNoData = true
        return
      }
      r.ExtractHeader.forEach((col) => {
        this.extractHeader.push({
          text: col,
          value: col,
        })
      })
      r.ExtractDatas.forEach((row) => {
        if (row.length !== r.ExtractHeader.length) {
          return
        }
        const e = {}
        for (let i = 0; i < r.ExtractHeader.length; i++) {
          e[r.ExtractHeader[i]] = row[i]
        }
        this.extractDatas.push(e)
      })
    },
    exportGrok() {
      this.$axios
        .get('/api/export/grok', {
          responseType: 'blob',
        })
        .then((response) => {
          const blob = new Blob([response.data], { type: 'text/yaml' })
          const url = (window.URL || window.webkitURL).createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = 'twsnmpfc_grok.yml'
          // aタグ要素を画面に一時的に追加する
          document.body.appendChild(a)
          a.click()
          // 不要になったら削除.
          document.body.removeChild(a)
        })
        .catch((e) => {
          this.exportError = true
        })
    },
    selectGrokFile(f) {
      this.grokFile = f
    },
    importGrok() {
      const formData = new FormData()
      formData.append('file', this.grokFile)
      this.$axios
        .$post('/api/import/grok', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })
        .then((r) => {
          this.importDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.importError = true
          this.$fetch()
        })
    },
    doDeleteGrok() {
      this.$axios
        .delete('/api/conf/grok/' + this.selected.ID)
        .then(() => {
          this.deleteDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
    },
  },
}
</script>
