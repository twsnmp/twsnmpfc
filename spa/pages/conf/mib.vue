<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        MIB管理
        <v-spacer></v-spacer>
      </v-card-title>
      <v-card-row>
        <v-text-field
          v-model="filter"
          append-icon="mdi-magnify"
          label="Search"
          single-line
          hide-details
        >
        </v-text-field>
      </v-card-row>
      <v-data-table
        :headers="headers"
        :items="mibmods"
        dense
        sort-by="Error"
        sort-desc
        :search="filter"
      >
        <template #[`item.Type`]="{ item }">
          <v-icon :color="item.Error == '' ? '#1f78b4' : '#e31a1c'">
            {{ item.Errror == '' ? 'mdi-information' : 'mdi-alert-circle' }}
          </v-icon>
          {{ item.Type == 'int' ? '組み込み' : '拡張' }}
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="info" dark @click="mibTreeDialog = true">
          <v-icon>mdi-file-tree</v-icon>
          MIBツリー
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="mibTreeDialog" persistent width="900px">
      <v-card max-height="95%">
        <v-card-title> MIBツリー </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="searchMIBTree"
            label="オブジェクト名の検索"
          ></v-text-field>
          <div style="height: 500px; overflow: auto">
            <v-treeview
              ref="tree"
              :items="mibtree"
              item-key="oid"
              :search="searchMIBTree"
              hoverable
              activatable
              dense
              @update:active="selectMIB"
            >
              <template #prepend="{ item, open }">
                <v-icon v-if="item.children.length > 0">
                  {{ open ? 'mdi-folder-open' : 'mdi-folder' }}
                </v-icon>
                <v-icon v-else :color="getIconColor(item.MIBInfo)">
                  {{ getMIBIcon(item.MIBInfo) }}
                </v-icon>
              </template>
              <template #label="{ item }">
                {{
                  item.MIBInfo
                    ? `${item.name}(${item.oid}: ${item.MIBInfo.Type} )`
                    : `${item.name}(${item.oid})`
                }}
                <v-icon small @click="copyMIB(item)"> mdi-content-copy </v-icon>
              </template>
            </v-treeview>
          </div>
          <div style="height: 160px; overflow: auto">
            <pre style="margin: 10px; background-color: #333">{{
              mibInfoText
            }}</pre>
          </div>
        </v-card-text>
        <v-snackbar v-model="copyError" absolute centered color="error">
          コピーできません
        </v-snackbar>
        <v-snackbar v-model="copyDone" absolute centered color="primary">
          コピーしました
        </v-snackbar>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="searchMIBTree.length > 2"
            color="normal"
            dark
            @click="openCloseMibTree"
          >
            <v-icon>mdi-file-tree</v-icon>
            {{ mibTreeOpened ? 'MIBツリーを閉じる' : 'MIBツリーを開く' }}
          </v-btn>
          <v-btn color="normal" dark @click="mibTreeDialog = false">
            <v-icon>mdi-cancel</v-icon>
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
          text: '種別',
          value: 'Type',
          width: '8%',
        },
        {
          text: 'モジュール名',
          value: 'Name',
          width: '15%',
        },
        {
          text: 'ファイル名',
          value: 'File',
          width: '12%',
        },
        {
          text: 'エラー',
          value: 'Error',
          width: '25%',
        },
      ],
      mibmods: [],
      filter: '',
      mibtree: [],
      searchMIBTree: '',
      mibTreeDialog: false,
      mibInfoText: '',
      copyError: false,
      copyDone: false,
      mibTreeOpened: false,
    }
  },
  async fetch() {
    this.mibmods = await this.$axios.$get('/api/mibmods')
    this.mibtree = await this.$axios.$get('/api/mibtree')
  },
  methods: {
    selectMIB(s) {
      if (s && s.length === 1) {
        this.findMIB(s[0], this.mibtree)
      }
    },
    findMIB(oid, list) {
      for (let i = 0; i < list.length; i++) {
        if (list[i].oid === oid) {
          this.mibInfoText = ''
          if (list[i].MIBInfo) {
            this.mibInfoText += `OID  : ${list[i].MIBInfo.OID}\n`
            this.mibInfoText += `Stats: ${list[i].MIBInfo.Status}\n`
            this.mibInfoText += `Type : ${list[i].MIBInfo.Type}\n`
            this.mibInfoText += list[i].MIBInfo.Units
              ? `Units : ${list[i].MIBInfo.Units}\n`
              : ''
            this.mibInfoText += list[i].MIBInfo.Index
              ? `Index : ${list[i].MIBInfo.Index}\n`
              : ''
            this.mibInfoText += list[i].MIBInfo.Defval
              ? `DefVal : ${list[i].MIBInfo.Defval}\n`
              : ''
            this.mibInfoText += `Description :\n${list[i].MIBInfo.Description}\n`
          }
          return list[i].name
        }
        if (list[i].children) {
          const n = this.findMIB(oid, list[i].children)
          if (n) {
            return n
          }
        }
      }
      return null
    },
    getMIBIcon(mibInfo) {
      if (mibInfo) {
        if (mibInfo.Type.startsWith('Counter')) {
          return 'mdi-counter'
        }
        if (mibInfo.Type.startsWith('ObjectIdent')) {
          return 'mdi-file-tree'
        }
        if (mibInfo.Type.startsWith('Time')) {
          return 'mdi-clock'
        }
        if (mibInfo.Type.startsWith('Int')) {
          return 'mdi-counter'
        }
        if (mibInfo.Type.includes('String')) {
          return 'mdi-code-string'
        }
        if (mibInfo.Type.startsWith('Gau')) {
          return 'mdi-speedometer'
        }
        if (
          mibInfo.Type.startsWith('Trap') ||
          mibInfo.Type.startsWith('Noti')
        ) {
          return 'mdi-alert-circle'
        }
        return 'mdi-information'
      }
      return 'mdi-folder'
    },
    getIconColor(mibInfo) {
      if (mibInfo && mibInfo.Type.startsWith('Noti')) {
        return 'red'
      }
      return ''
    },
    copyMIB(item) {
      if (!navigator.clipboard) {
        return
      }
      const s = `${item.name}(${item.oid})`
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    openCloseMibTree() {
      this.mibTreeOpened = !this.mibTreeOpened
      if (this.$refs.tree) {
        this.$refs.tree.updateAll(this.mibTreeOpened)
      }
    },
  },
}
</script>
