<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        アイコン
        <v-spacer></v-spacer>
      </v-card-title>
      <v-alert v-model="deleteError" color="error" dense dismissible>
        アイコンの登録を削除できませんでした
      </v-alert>
      <v-alert v-model="updateError" color="error" dense dismissible>
        アイコンを登録できませんでした
      </v-alert>
      <v-data-table :headers="headers" :items="icons" dense sort-by="Text">
        <template #[`item.Icon`]="{ item }">
          <v-icon>{{ item.Icon }}</v-icon>
          {{ item.Icon }}
        </template>
        <template #[`item.Code`]="{ item }">
          0x{{ item.Code.toString(16) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="editIcon(item)"> mdi-pencil </v-icon>
          <v-icon small @click="deleteIcon(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="addIcon">
          <v-icon>mdi-plus</v-icon>
          追加
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="editDialog" persistent max-width="500px">
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
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">アイコン削除</span>
        </v-card-title>
        <v-card-text>
          {{ selected.Text }}({{ selected.Icon }})を削除しますか？
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
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      editDialog: false,
      deleteDialog: false,
      deleteError: false,
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
    }
  },
  async fetch() {
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
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/conf/icon/' + this.selected.Icon)
        .then(() => {
          this.$fetch()
          this.deleteDialog = false
          this.$delIcon(this.selected.Icon)
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
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
  },
}
</script>
