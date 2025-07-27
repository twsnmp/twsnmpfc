<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="95%">
      <v-card-title>
        アイコン
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="updateError" color="error" dense dismissible>
        アイコンを登録できませんでした
      </v-alert>
      <v-alert v-model="importError" color="error" dense dismissible>
        {{ importErrorMsg }}
      </v-alert>
      <v-data-table
        v-model="selectedIcons"
        show-select
        item-key="Icon"
        :headers="headers"
        :items="icons"
        dense
        sort-by="Text"
      >
        <template #[`item.Icon`]="{ item }">
          <v-icon>{{ item.Icon }}</v-icon>
          {{ item.Icon }}
        </template>
        <template #[`item.Code`]="{ item }">
          0x{{ item.Code.toString(16) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="editIcon(item)"> mdi-pencil </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="addIcon">
          <v-icon>mdi-plus</v-icon>
          追加
        </v-btn>
        <v-btn
          :loading="importWait"
          :disabled="importWait"
          color="primary"
          dark
          @click="selectIconFile"
        >
          <v-icon>mdi-upload</v-icon>
          インポート
        </v-btn>
        <input
          ref="iconFileInput"
          type="file"
          accept=".csv;.txt"
          style="display: none"
          @change="importIcon"
        />
        <download-excel
          v-if="icons.length > 0"
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_Icons.csv"
          class="v-btn ml-0"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-download</v-icon>
            エクスポート
          </v-btn>
        </download-excel>
        <v-btn
          v-if="selectedIcons.length > 0"
          color="error"
          class="ml-1"
          @click="deleteDialog = true"
        >
          <v-icon>mdi-delete</v-icon>
          削除
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title> アイコン </v-card-title>
        <v-card-text>
          <v-autocomplete
            v-model="selected.Icon"
            :items="iconList"
            dense
            :readonly="!add"
            label="アイコン"
          >
          </v-autocomplete>
          <v-icon>{{ selected.Icon }}</v-icon>
          <v-text-field v-model="selected.Text" label="名前"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doUpdate">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="editDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">アイコン削除</span>
        </v-card-title>
        <v-card-text> 選択したアイコンを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            :loading="deleteWait"
            :disabled="deleteWait"
            color="error"
            @click="doDelete"
          >
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
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      editDialog: false,
      deleteDialog: false,
      updateError: false,
      selected: {
        Icon: '',
        Text: '',
        Code: 0,
      },
      headers: [
        {
          text: 'アイコン',
          value: 'Icon',
          width: '30%',
        },
        {
          text: '名前',
          value: 'Text',
          width: '40%',
        },
        {
          text: 'コード',
          value: 'Code',
          width: '15%',
        },
        { text: '操作', value: 'actions', width: '15%' },
      ],
      icons: [],
      iconList: [],
      iconCode: new Map(),
      add: false,
      importError: false,
      importErrorMsg: '',
      importWait: false,
      deleteWait: false,
      selectedIcons: [],
    }
  },
  async fetch() {
    this.selectedIcons = []
    const r = await this.$axios.$get('/api/conf/icons')
    if (r) {
      this.icons = r
    }
    if (
      this.iconList.length < 1 &&
      document.styleSheets &&
      document.styleSheets.length > 0
    ) {
      document.styleSheets.forEach((ss) => {
        const classes = ss.rules || ss.cssRules
        if (!classes) {
          return
        }
        const re = /mdi-[^:]+/
        classes.forEach((e) => {
          if (
            e.selectorText &&
            e.selectorText.includes('::before') &&
            e.style &&
            e.style.content
          ) {
            const m = e.selectorText.match(re)
            if (m) {
              const code =
                e.style.content && e.style.content.length > 2
                  ? e.style.content.codePointAt(1)
                  : 0
              if (code !== 0) {
                this.iconList.push(m[0])
                this.iconCode.set(m[0], code)
              }
            }
          }
        })
      })
    }
  },
  methods: {
    editIcon(item) {
      this.selected = item
      this.add = false
      this.editDialog = true
    },
    addIcon(item) {
      this.selected = {
        Icon: '',
        Text: '',
        Code: 0,
      }
      this.add = true
      this.editDialog = true
    },
    deleteIcon(item) {
      this.selected = item
      this.deleteDialog = true
    },
    async doDelete() {
      this.deleteWait = true
      for (const i of this.selectedIcons) {
        await this.$axios.delete('/api/conf/icon/' + i.Icon)
      }
      this.$fetch()
      this.deleteWait = false
      this.deleteDialog = false
    },
    doUpdate() {
      this.selected.Code = this.iconCode.has(this.selected.Icon)
        ? this.iconCode.get(this.selected.Icon)
        : 0
      this.updateError = false
      this.$axios
        .post('/api/conf/icon', this.selected)
        .then(() => {
          this.editDialog = false
          this.$fetch()
          this.$setIcon(this.selected)
          this.$setIconToMap(this.selected)
        })
        .catch((e) => {
          this.updateError = true
          this.$fetch()
        })
    },
    selectIconFile() {
      if (this.$refs.iconFileInput) {
        this.$refs.iconFileInput.click()
      }
    },
    importIcon() {
      this.importErrorMsg = ''
      this.importError = false
      const iconFile = this.$refs.iconFileInput.files[0]
      if (iconFile) {
        const fileReader = new FileReader()
        fileReader.onload = async (event) => {
          if (!event.target.result) {
            return
          }
          this.importWait = true
          const lines = event.target.result.split('\n')
          const icons = []
          const errors = []
          for (const l of lines) {
            const a = l.split(iconFile.type === 'text/plain' ? ' ' : ',', 2)
            if (a.length !== 2) {
              continue
            }
            const i = a[0].trim()
            const t = a[1].trim()
            if (!i || !t || i === 'Icon' || i.startsWith('#')) {
              continue
            }
            const icon = i.replace(/([A-Z])/g, '-$1').toLowerCase()
            if (this.iconCode.has(icon)) {
              icons.push({
                Icon: icon,
                Text: t,
                Code: this.iconCode.get(icon),
              })
            } else {
              errors.push(icon)
            }
          }
          if (errors.length > 0) {
            this.importError = true
            this.importErrorMsg =
              '次のアイコンはインポートできません。 ' + errors.join(',')
            this.$refs.iconFileInput.value = ''
            this.importWait = false
            return
          }
          this.$refs.iconFileInput.value = ''
          for (const icon of icons) {
            await this.$axios.$post('/api/conf/icon', icon)
          }
          this.importWait = false
          this.$fetch()
        }
        fileReader.readAsText(iconFile)
      }
    },
    makeExports() {
      const exports = []
      this.icons.forEach((i) => {
        exports.push({
          Icon: i.Icon,
          Text: i.Text,
        })
      })
      return exports
    },
  },
}
</script>
