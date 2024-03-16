<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Syslog
        <v-spacer></v-spacer>
        <span class="text-caption">
          {{ ft }}から{{ lt }} {{ count }} / {{ process }}件
        </span>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 20vh"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        dense
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        @dblclick:row="copyLog"
      >
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="editPolling(item)"> mdi-card-plus </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td></td>
            <td>
              <v-text-field v-model="conf.pri" label="pri"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.host" label="host"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.tag" label="tag"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.msg" label="message"></v-text-field>
            </td>
          </tr>
        </template>
      </v-data-table>
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="filterDialog = true">
          <v-icon>mdi-magnify</v-icon>
          検索条件
        </v-btn>
        <v-btn v-if="filter.NextTime > 0" color="info" dark @click="nextLog">
          <v-icon>mdi-page-next</v-icon>
          続きを検索
        </v-btn>
        <download-excel
          :fetch="makeSyslogExports"
          type="csv"
          name="TWSNMP_FC_Syslog.csv"
          header="TWSNMP FCのSyslog"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeSyslogExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_Syslog.csv"
          header="TWSNMP FCのSyslog"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeSyslogExports"
          type="xls"
          name="TWSNMP_FC_Syslog.xls"
          header="TWSNMP FCのSyslog"
          worksheet="Syslog"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-menu v-if="logs" offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="showHistogram">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒストグラム </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showCluster">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> クラスター分析 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHost">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ホスト別ログ </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHost3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ホスト別ログ(3D) </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="updateFFT">
              <v-list-item-icon>
                <v-icon>mdi-chart-bell-curve</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>FFT分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="updateFFT3D">
              <v-list-item-icon>
                <v-icon>mdi-rotate-3d</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>FFT分析(3D)</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHeatmap">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒートマップ </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn
          v-if="extractDatas.length > 0"
          color="primary"
          @click="extractDialog = true"
        >
          <v-icon>mdi-view-list</v-icon>
          抽出情報
        </v-btn>
        <v-btn color="normal" dark @click="doFilter()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">検索条件</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-model="filter.Level"
            :items="$filterEventLevelList"
            label="状態"
          ></v-select>
          <v-row justify="space-around">
            <v-menu
              ref="sdMenu"
              v-model="sdMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.StartDate"
                  label="開始日"
                  prepend-icon="mdi-calendar"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="filter.StartDate"
                no-title
                dark
                scrollable
                @input="sdMenuShow = false"
              >
              </v-date-picker>
            </v-menu>
            <v-text-field
              v-model="filter.StartTime"
              label="開始時刻"
              prepend-icon="mdi-clock-time-four-outline"
              type="time"
            ></v-text-field>
            <v-icon
              @click="
                filter.StartDate = ''
                filter.StartTime = ''
              "
            >
              mdi-close
            </v-icon>
            <v-icon
              @click="
                const t = Date.now()
                filter.StartDate = $timeFormat(t, '{yyyy}-{MM}-{dd}')
                filter.StartTime = '00:00'
              "
            >
              mdi-calendar-today
            </v-icon>
          </v-row>
          <v-row justify="space-around">
            <v-menu
              ref="edMenu"
              v-model="edMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.EndDate"
                  label="終了日"
                  prepend-icon="mdi-calendar"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="filter.EndDate"
                no-title
                dark
                scrollable
                @input="edMenuShow = false"
              >
              </v-date-picker>
            </v-menu>
            <v-text-field
              v-model="filter.EndTime"
              label="終了時刻"
              prepend-icon="mdi-clock-time-four-outline"
              type="time"
            ></v-text-field>
            <v-icon
              @click="
                filter.EndDate = ''
                filter.EndTime = ''
              "
            >
              mdi-close
            </v-icon>
            <v-icon
              @click="
                const t = Date.now() + 3600 * 24 * 1000
                filter.EndDate = $timeFormat(t, '{yyyy}-{MM}-{dd}')
                filter.EndTime = '00:00'
              "
            >
              mdi-calendar-today
            </v-icon>
          </v-row>
          <label>種別（正規表現）</label>
          <prism-editor
            v-model="filter.Type"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>ホスト名（正規表現）</label>
          <prism-editor
            v-model="filter.Host"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>タグ（正規表現）</label>
          <prism-editor
            v-model="filter.Tag"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>メッセージ（パイプライン正規表現）</label>
          <prism-editor
            v-model="filter.Message"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <v-autocomplete
            v-model="filter.Extractor"
            :items="filterExtractorList"
            label="抽出パターン"
          ></v-autocomplete>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doFilter">
            <v-icon>mdi-magnify</v-icon>
            検索
          </v-btn>
          <v-btn color="normal" dark @click="filterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extractDialog" persistent max-width="90%">
      <v-card>
        <v-card-title>
          <span class="headline">抽出した情報</span>
          <v-spacer></v-spacer>
          <v-text-field
            v-model="searchExtract"
            append-icon="mdi-magnify"
            label="検索"
            single-line
            hide-details
          ></v-text-field>
        </v-card-title>
        <v-data-table
          :headers="extractHeader"
          :items="extractDatas"
          :search="searchExtract"
          :items-per-page="15"
          sort-by="TimeStr"
          dense
        >
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <download-excel
            :fetch="makeExtractExports"
            type="csv"
            name="TWSNMP_FC_Syslog_Extract.csv"
            header="TWSNMP FCでSyslogから抽出した情報"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeExtractExports"
            type="csv"
            :escape-csv="false"
            name="TWSNMP_FC_Syslog_Extract.csv"
            header="TWSNMP FCでSyslogから抽出した情報"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV(NO ESC)
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeExtractExports"
            type="xls"
            name="TWSNMP_FC_Syslog_Extract.xls"
            header="TWSNMP FCでSyslogから抽出した情報"
            worksheet="Syslogから抽出"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-menu offset-y>
            <template #activator="{ on, attrs }">
              <v-btn color="primary" dark v-bind="attrs" v-on="on">
                <v-icon>mdi-chart-line</v-icon>
                グラフと集計
              </v-btn>
            </template>
            <v-list>
              <v-list-item
                v-if="numExtractTypeList.length > 0"
                @click="showExtractHistogram"
              >
                <v-list-item-icon
                  ><v-icon>mdi-chart-histogram</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title> ヒストグラム </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-if="numExtractTypeList.length > 0"
                @click="showExtractCluster"
              >
                <v-list-item-icon
                  ><v-icon>mdi-chart-scatter-plot</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title> クラスター分析 </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-if="strExtractTypeList.length > 0"
                @click="showExtractTopList"
              >
                <v-list-item-icon
                  ><v-icon>mdi-format-list-numbered</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title> 項目別集計リスト </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-if="
                  numExtractTypeList.length > 0 && strExtractTypeList.length > 0
                "
                @click="showExtract3D"
              >
                <v-list-item-icon
                  ><v-icon>mdi-chart-scatter-plot</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title> 項目別集計3Dグラフ </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-menu>
          <v-btn color="normal" @click="extractDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="histogramDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ヒストグラム
          <v-spacer></v-spacer>
          <v-select
            v-model="histogramType"
            :items="histogramTypeList"
            label="集計項目"
            single-line
            hide-details
            @change="updateHistogram"
          ></v-select>
        </v-card-title>
        <div
          id="histogram"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="histogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="clusterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          クラスター分析
          <v-spacer></v-spacer>
          <v-text-field
            v-model="cluster"
            label="クラスター数"
            @change="updateCluster"
          ></v-text-field>
        </v-card-title>
        <div
          id="cluster"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="clusterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="host3DDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ホスト別ログ(3D)
          <v-spacer></v-spacer>
        </v-card-title>
        <div
          id="host3d"
          style="width: 95vw; height: 40vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="host3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="hostListDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ホスト別ログ
          <v-spacer></v-spacer>
        </v-card-title>
        <v-card-text>
          <div
            id="hostList"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
          <v-data-table
            :headers="hostListHeader"
            :items="hostList"
            sort-by="Total"
            sort-desc
            dense
          >
            <template #[`item.Total`]="{ item }">
              {{ formatCount(item.Total) }}
            </template>
            <template #[`body.append`]>
              <tr>
                <td>
                  <v-text-field v-model="hostName" label="name"></v-text-field>
                </td>
                <td colspan="5"></td>
              </tr>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <download-excel
            :fetch="makeHostExports"
            type="csv"
            name="TWSNMP_FC_Syslog_Host_List.csv"
            header="TWSNMP FCへSyslogを送信するホストリスト"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeHostExports"
            type="csv"
            :escape-csv="false"
            name="TWSNMP_FC_Syslog_Host_List.csv"
            header="TWSNMP FCへSyslogを送信するホストリスト"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV(NO ESC)
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeHostExports"
            type="xls"
            name="TWSNMP_FC_Syslog_Host_List.xls"
            header="TWSNMP FCへSyslogを送信するホストリスト"
            worksheet="Syslog送信ホスト"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-btn color="normal" dark @click="hostListDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extractHistogramDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          抽出情報のヒストグラム分析
          <v-spacer></v-spacer>
          <v-select
            v-model="extractHistogramType"
            :items="numExtractTypeList"
            label="集計項目"
            single-line
            hide-details
            @change="updateExtractHistogram"
          ></v-select>
        </v-card-title>
        <div
          id="histogram"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="extractHistogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extractClusterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          抽出情報のクラスター分析
          <v-spacer></v-spacer>
          <v-text-field
            v-model="cluster"
            label="クラスター数"
            @change="updateExtractCluster"
          ></v-text-field>
        </v-card-title>
        <div
          id="cluster"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="extractClusterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extract3DDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title> 抽出情報の項目別3Dグラフ </v-card-title>
        <v-card-text>
          <v-row>
            <v-col>
              <v-select
                v-model="extract3DTypeX"
                :items="strExtractTypeList"
                label="X軸項目"
                single-line
                hide-details
                @change="updateExtract3D"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="extract3DTypeZ"
                :items="numExtractTypeList"
                label="Z軸項目"
                single-line
                hide-details
                @change="updateExtract3D"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="extract3DTypeColor"
                :items="numExtractTypeList"
                label="色項目"
                single-line
                hide-details
                @change="updateExtract3D"
              ></v-select>
            </v-col>
          </v-row>
          <div
            id="extract3d"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="extract3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extractTopListDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          抽出情報の項目別集計
          <v-spacer></v-spacer>
          <v-select
            v-model="extractTopListType"
            :items="strExtractTypeList"
            label="集計項目"
            single-line
            hide-details
            @change="updateExtractTopList"
          ></v-select>
        </v-card-title>
        <v-card-text>
          <div
            id="extractTopList"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
          <v-data-table
            :headers="extractTopListHeader"
            :items="extractTopList"
            sort-by="Total"
            sort-desc
            dense
          >
            <template #[`item.Total`]="{ item }">
              {{ formatCount(item.Total) }}
            </template>
            <template #[`body.append`]>
              <tr>
                <td>
                  <v-text-field
                    v-model="extractTopName"
                    label="name"
                  ></v-text-field>
                </td>
                <td colspan="5"></td>
              </tr>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <download-excel
            :fetch="makeExtractTopExports"
            type="csv"
            name="TWSNMP_FC_Syslog_Extract_Top_List.csv"
            header="TWSNMP FC Syslogで抽出したデータの上位リスト"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeExtractTopExports"
            type="csv"
            :escape-csv="false"
            name="TWSNMP_FC_Syslog_Extract_Top_List.csv"
            header="TWSNMP FC Syslogで抽出したデータの上位リスト"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV(NO ESC)
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeExtractTopExports"
            type="xls"
            name="TWSNMP_FC_Syslog_Extract_Top.xls"
            header="TWSNMP FC Syslogで抽出したデータの上位リスト"
            worksheet="Syslog抽出データ上位"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-btn color="normal" dark @click="extractTopListDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="fftDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> Syslog - FFT分析 </span>
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col>
              <v-select
                v-model="fftType"
                :items="fftTypeList"
                label="周波数/周期"
                single-line
                hide-details
                @change="updateFFT"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="fftHost"
                :items="fftHostList"
                label="ホスト"
                single-line
                hide-details
                @change="updateFFT"
              ></v-select>
            </v-col>
          </v-row>
          <div
            id="FFTChart"
            style="width: 95vw; height: 50vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="fftDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="fft3DDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> Syslog - FFT分析(3D) </span>
          <v-spacer></v-spacer>
          <v-select
            v-model="fftType"
            :items="fftTypeList"
            label="周波数/周期"
            single-line
            hide-details
            @change="updateFFT3D"
          ></v-select>
        </v-card-title>
        <v-card-text>
          <div
            id="FFTChart3D"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="fft3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="heatmapDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> Syslog - ヒートマップ </span>
        </v-card-title>
        <v-card-text>
          <div
            id="heatmap"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="heatmapDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editPollingDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> ポーリング追加 </v-card-title>
        <v-alert v-model="addPollingError" color="error" dense dismissible>
          ポーリングを変更できませんでした
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.NodeID"
                :items="nodeList"
                label="ノード"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field v-model="polling.Name" label="名前"></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.Level"
                :items="$levelList"
                label="レベル"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Type"
                readonly
                label="種別"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Mode"
                label="モード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="polling.Params"
                label="送信元ホスト"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-slider
                v-model="polling.PollInt"
                label="ポーリング間隔(Sec)"
                class="align-center"
                max="3600"
                min="5"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="polling.PollInt"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
            <v-col>
              <v-select
                v-model="polling.LogMode"
                :items="$logModeList"
                label="ログモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <label>フィルター</label>
          <prism-editor
            v-model="polling.Filter"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>判定スクリプト</label>
          <prism-editor
            v-model="polling.Script"
            class="script"
            :highlight="highlighter"
            line-numbers
          ></prism-editor>
          <v-row dense>
            <v-col>
              <label>障害時アクション</label>
              <prism-editor
                v-model="polling.FailAction"
                class="script"
                :highlight="actionHighlighter"
                line-numbers
              ></prism-editor>
            </v-col>
            <v-col>
              <label>復帰時アクション</label>
              <prism-editor
                v-model="polling.RepairAction"
                class="script"
                :highlight="actionHighlighter"
                line-numbers
              ></prism-editor>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doAddPolling">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="editPollingDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
import { PrismEditor } from 'vue-prism-editor'
import 'vue-prism-editor/dist/prismeditor.min.css'
import { highlight, languages } from 'prismjs/components/prism-core'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-regex'
import 'prismjs/themes/prism-tomorrow.css'

export default {
  components: {
    PrismEditor,
  },
  data() {
    return {
      count: 0,
      process: 0,
      ft: '',
      lt: '',
      filterDialog: false,
      sdMenuShow: false,
      edMenuShow: false,
      filter: {
        Level: '',
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        Host: '',
        Type: '',
        Tag: '',
        Message: '',
        Extractor: '',
        NextTime: 0,
        Filter: 0,
      },
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        { text: '状態', value: 'Level', width: '8%' },
        {
          text: '日時',
          value: 'TimeStr',
          width: '17%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.conf.pri) return true
            return value.includes(this.conf.pri)
          },
        },
        {
          text: 'ホスト名',
          value: 'Host',
          width: '10%',
          filter: (value) => {
            if (!this.conf.host) return true
            return value.includes(this.conf.host)
          },
        },
        {
          text: 'タグ',
          value: 'Tag',
          width: '10%',
          filter: (value) => {
            if (!this.conf.tag) return true
            return value.includes(this.conf.tag)
          },
        },
        {
          text: 'メッセージ',
          value: 'Message',
          width: '40%',
          filter: (value) => {
            if (!this.conf.msg) return true
            return value.includes(this.conf.msg)
          },
        },
        { text: `操作`, value: `actions`, width: '5%' },
      ],
      logs: [],
      extractDialog: false,
      searchExtract: '',
      extractDatas: [],
      extractHeader: [],
      filterExtractorList: [],
      conf: {
        pri: '',
        host: '',
        tag: '',
        msg: '',
        sortBy: 'TimeStr',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      histogramDialog: false,
      histogramType: 'Severity',
      histogramTypeList: [
        { text: 'Severity', value: 'Severity' },
        { text: 'Facility', value: 'Facility' },
        { text: 'Priority', value: 'Priority' },
      ],
      clusterDialog: false,
      cluster: 2,
      hostListHeader: [
        {
          text: 'ホスト名',
          value: 'Name',
          width: '40%',
          filter: (value) => {
            if (!this.hostName) return true
            return value.includes(this.hostName)
          },
        },
        { text: '総数', value: 'Total', width: '10%' },
        { text: '重度', value: 'High', width: '10%' },
        { text: '軽度', value: 'Low', width: '10%' },
        { text: '注意', value: 'Warn', width: '10%' },
        { text: '情報', value: 'Info', width: '10%' },
        { text: 'デバッグ', value: 'Debug', width: '10%' },
      ],
      hostList: [],
      hostListDialog: false,
      hostName: '',
      host3DDialog: false,
      extractHistogramDialog: false,
      extractHistogramType: '',
      extractClusterDialog: false,
      extract3DDialog: false,
      extractTopListDialog: false,
      extractTopList: [],
      extractHisgramType: '',
      extract3DTypeX: '',
      extract3DTypeZ: '',
      extract3DTypeColor: '',
      extractTopListType: '',
      numExtractTypeList: [],
      strExtractTypeList: [],
      extractTopName: '',
      extractTopListHeader: [
        {
          text: '項目名',
          value: 'Name',
          width: '70%',
          filter: (value) => {
            if (!this.extractTopName) return true
            return value.includes(this.extractTopName)
          },
        },
        { text: '総数', value: 'Total', width: '30%' },
      ],
      fftDialog: false,
      fftMap: null,
      fftType: 't',
      fftHost: '',
      fftHostList: [],
      fftTypeList: [
        { text: '周期(Sec)', value: 't' },
        { text: '周波数(Hz)', value: 'hz' },
      ],
      fft3DDialog: false,
      copyDone: false,
      copyError: false,
      heatmapDialog: false,
      startTimeOp: '',
      endTimeOp: '',
      editPollingDialog: false,
      addPollingError: false,
      polling: {},
      nodeList: [],
    }
  },
  async fetch() {
    this.fftMap = null
    const r = await this.$axios.$post('/api/log/syslog', this.filter)
    if (!r) {
      return
    }
    if (this.filter.NextTime === 0) {
      this.logs = []
      this.extractDatas = []
      this.extractHeader = []
      this.numExtractTypeList = []
      this.strExtractTypeList = []
      if (this.conf.page > 1) {
        this.options.page = this.conf.page
        this.conf.page = 1
      }
    }
    this.count = r.Filter
    this.process += r.Process
    this.logs = this.logs.concat(r.Logs ? r.Logs : [])
    let id = 0
    this.ft = ''
    let lt
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}')
      e.ID = id++
      if (this.ft === '') {
        this.ft = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}')
      }
      lt = t
    })
    if (this.ft === '') {
      if (this.filter.StartDate === '') {
        this.ft = this.$timeFormat(
          new Date(new Date() - 3600 * 1000),
          '{yyyy}/{MM}/{dd} {HH}:{mm}'
        )
      } else {
        this.ft =
          this.filter.StartDate + ' ' + (this.filter.StartTime || '00:00')
      }
    }
    if (lt) {
      this.lt = this.$timeFormat(lt, '{yyyy}/{MM}/{dd} {HH}:{mm}')
    } else if (this.filter.EndDate === '') {
      this.ft = this.$timeFormat(new Date(), '{yyyy}/{MM}/{dd} {HH}:{mm}')
    } else {
      this.ft = this.filter.EndDate + ' ' + (this.filter.EndtTime || '23:59')
    }
    this.$showLogLevelChart('logCountChart', this.logs, this.zoomCallBack)
    if (this.filterExtractorList.length < 1) {
      const groks = await this.$axios.$get('/api/conf/grok')
      if (groks) {
        this.filterExtractorList = [
          { text: '指定しない', value: '' },
          { text: 'SPLUNK風', value: 're_splunk' },
          { text: '自動数値', value: 're_number' },
          { text: 'JSON', value: 're_json' },
          { text: '自動(IP,MAC,EMail)', value: 're_other' },
        ]
        groks.forEach((g) => {
          this.filterExtractorList.push({
            text: g.Name,
            value: g.ID,
          })
        })
      }
    }
    if (r.ExtractHeader.length < 1 || r.ExtractDatas.length < 1) {
      this.checkNextlog(r)
      return
    }
    if (this.filter.NextTime === 0) {
      r.ExtractHeader.forEach((col) => {
        this.extractHeader.push({
          text: col,
          value: col,
        })
      })
    }
    let firstData = true
    r.ExtractDatas.forEach((row) => {
      if (row.length !== r.ExtractHeader.length) {
        return
      }
      const e = {}
      for (let i = 0; i < r.ExtractHeader.length; i++) {
        e[r.ExtractHeader[i]] = row[i]
        if (firstData && r.ExtractHeader[i] !== 'TimeStr') {
          if (isNaN(Number(row[i]))) {
            this.strExtractTypeList.push({
              text: r.ExtractHeader[i],
              value: r.ExtractHeader[i],
            })
          } else {
            this.numExtractTypeList.push({
              text: r.ExtractHeader[i],
              value: r.ExtractHeader[i],
            })
          }
        }
      }
      firstData = false
      this.extractDatas.push(e)
    })
    this.checkNextlog(r)
  },
  created() {
    const c = this.$store.state.log.logs.syslog
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  mounted() {
    this.$showLogLevelChart('logCountChart', this.logs)
    window.addEventListener('resize', this.$resizeLogLevelChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogLevelChart)
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('log/logs/setSyslog', this.conf)
  },
  methods: {
    zoomCallBack(st, et) {
      this.zoom.st = st
      this.zoom.et = et
    },
    doFilter() {
      if (this.filter.StartDate !== '' && this.filter.StartTime === '') {
        this.filter.StartTime = '00:00'
      }
      if (this.filter.EndDate !== '' && this.filter.EndTime === '') {
        this.filter.EndTime = '23:59'
      }
      this.filterDialog = false
      this.filter.NextTime = 0
      this.filter.Filter = 0
      this.count = 0
      this.process = 0
      this.limit = 0
      this.$fetch()
    },
    checkNextlog(r) {
      if (r.NextTime === 0) {
        return
      }
      this.limit = r.Limit
      this.filter.NextTime = r.NextTime
      this.filter.Filter = r.Filter
    },
    nextLog() {
      if (this.limit > 3 && this.filter.Filter >= this.limit) {
        this.logs.splice(0, this.limit / 4)
        this.filter.Filter = this.logs.length
      }
      this.$fetch()
    },
    showHistogram() {
      this.histogramDialog = true
      this.$nextTick(() => {
        this.updateHistogram()
      })
    },
    showHeatmap() {
      this.heatmapDialog = true
      this.$nextTick(() => {
        this.$showLogHeatmap('heatmap', this.logs)
      })
    },
    updateHistogram() {
      this.$showSyslogHistogram(
        'histogram',
        this.getFilteredSyslog(),
        this.histogramType
      )
    },
    showCluster() {
      this.clusterDialog = true
      this.$nextTick(() => {
        this.updateCluster()
      })
    },
    updateCluster() {
      this.$showSyslogCluster(
        'cluster',
        this.getFilteredSyslog(),
        this.cluster * 1
      )
    },
    showHost3D() {
      this.host3DDialog = true
      this.$nextTick(() => {
        this.$showSyslogHost3D('host3d', this.getFilteredSyslog())
      })
    },
    showHost() {
      this.hostList = this.$getSyslogHostList(this.getFilteredSyslog())
      this.hostListDialog = true
      this.$nextTick(() => {
        this.$showSyslogHost('hostList', this.hostList)
      })
    },
    showExtractHistogram() {
      if (
        this.extractHistogramType === '' &&
        this.numExtractTypeList.length > 1
      ) {
        this.extractHistogramType = this.numExtractTypeList[0].value
      }
      this.extractHistogramDialog = true
      this.$nextTick(() => {
        this.updateExtractHistogram()
      })
    },
    updateExtractHistogram() {
      this.$showSyslogExtractHistogram(
        'histogram',
        this.getFilteredExtractData(),
        this.extractHistogramType
      )
    },
    showExtractCluster() {
      this.extractClusterDialog = true
      this.$nextTick(() => {
        this.updateExtractCluster()
      })
    },
    updateExtractCluster() {
      this.$showSyslogExtractCluster(
        'cluster',
        this.getFilteredExtractData(),
        this.numExtractTypeList,
        this.cluster * 1
      )
    },
    showExtract3D() {
      if (this.extract3DTypeX === '' && this.strExtractTypeList.length > 0) {
        this.extract3DTypeX = 'Host'
      }
      if (this.extract3DTypeZ === '' && this.numExtractTypeList.length > 0) {
        this.extract3DTypeZ = this.numExtractTypeList[0].value
      }
      if (
        this.extract3DTypeColor === '' &&
        this.numExtractTypeList.length > 0
      ) {
        this.extract3DTypeColor = this.numExtractTypeList[0].value
      }
      this.extract3DDialog = true
      this.$nextTick(() => {
        this.updateExtract3D()
      })
    },
    updateExtract3D() {
      this.$showSyslogExtract3D(
        'extract3d',
        this.getFilteredExtractData(),
        this.extract3DTypeX,
        this.extract3DTypeZ,
        this.extract3DTypeColor
      )
    },
    showExtractTopList() {
      if (this.extractTopListType === '') {
        this.extractTopListType = this.strExtractTypeList[0].value
      }
      this.extractTopListDialog = true
      this.$nextTick(() => {
        this.updateExtractTopList()
      })
    },
    updateExtractTopList() {
      this.extractTopList = this.$getSyslogExtractTopList(
        this.getFilteredExtractData(),
        this.extractTopListType
      )
      this.$showSyslogExtractTopList('extractTopList', this.extractTopList)
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    updateFFT() {
      this.fftDialog = true
      if (!this.fftMap) {
        this.fftMap = this.$getSyslogFFTMap(this.getFilteredSyslog())
        this.fftHostList = []
        this.fftMap.forEach((e) => {
          this.fftHostList.push({ text: e.Name, value: e.Name })
        })
        this.fftHost = 'Total'
      }
      this.$nextTick(() => {
        this.$showSyslogFFT('FFTChart', this.fftMap, this.fftHost, this.fftType)
      })
    },
    updateFFT3D() {
      this.fft3DDialog = true
      if (!this.fftMap) {
        this.fftMap = this.$getSyslogFFTMap(this.getFilteredSyslog())
        this.fftHostList = []
        this.fftMap.forEach((e) => {
          this.fftHostList.push({ text: e.Name, value: e.Name })
        })
        this.fftHost = 'Total'
      }
      this.$nextTick(() => {
        this.$showSyslogFFT3D('FFTChart3D', this.fftMap, this.fftType)
      })
    },
    copyLog(me, p) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const s =
        p.item.TimeStr +
        ' ' +
        p.item.Type +
        ' ' +
        p.item.Tag +
        ' ' +
        p.item.Message
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    makeSyslogExports() {
      const exports = []
      this.logs.forEach((e) => {
        if (!this.filterSyslog(e)) {
          return
        }
        exports.push({
          状態: this.$getStateName(e.Level),
          記録日時: e.TimeStr,
          種別: e.Type,
          ホスト名: e.Host,
          タグ: e.Tag,
          メッセージ: e.Message,
        })
      })
      return exports
    },
    getFilteredSyslog() {
      const ret = []
      if (!this.logs) {
        return ret
      }
      this.logs.forEach((e) => {
        if (!this.filterSyslog(e)) {
          return
        }
        ret.push(e)
      })
      return ret
    },
    filterSyslog(e) {
      if (this.conf.pri && !e.Type.includes(this.conf.pri)) {
        return false
      }
      if (this.conf.host && !e.Host.includes(this.conf.host)) {
        return false
      }
      if (this.conf.tag && !e.Tag.includes(this.conf.tag)) {
        return false
      }
      if (this.conf.msg && !e.Message.includes(this.conf.msg)) {
        return false
      }
      return true
    },
    makeExtractExports() {
      const exports = []
      this.extractDatas.forEach((e) => {
        if (!this.filterExtract(e)) {
          return
        }
        exports.push(e)
      })
      return exports
    },
    filterExtract(e) {
      if (this.searchExtract) {
        const s = Object.values(e).join(' ')
        if (s.includes(this.searchExtract)) {
          return false
        }
      }
      return true
    },
    getFilteredExtractData() {
      const ret = []
      this.extractDatas.forEach((e) => {
        if (!this.filterExtract(e)) {
          return
        }
        ret.push(e)
      })
      return ret
    },
    makeExtractTopExports() {
      const exports = []
      this.extractTopList.forEach((e) => {
        if (this.extractTopName && !e.Name.includes(this.extractTopName)) {
          return
        }
        exports.push({
          項目名: e.Name,
          総数: e.Total,
        })
      })
      return exports
    },
    makeHostExports() {
      const exports = []
      this.hostList.forEach((e) => {
        if (this.hostName && !e.Name.includes(this.hostName)) {
          return
        }
        exports.push({
          ホスト名: e.Name,
          総数: e.Total,
          重度: e.High,
          軽度: e.Low,
          注意: e.Warn,
          情報: e.Info,
          デバッグ: e.Debug,
        })
      })
      return exports
    },
    async editPolling(i) {
      if (this.nodeList.length < 1) {
        const r = await this.$axios.$get('/api/nodes')
        r.forEach((n) => {
          this.nodeList.push({ text: n.Name, value: n.ID, ip: n.IP })
        })
      }
      let ip = i.Host
      const a = ip.split('(')
      if (a.length > 1) {
        ip = a[0]
      }
      let nodeID
      for (let j = 0; j < this.nodeList.length; j++) {
        if (this.nodeList[j].text === i.Host) {
          nodeID = this.nodeList[j].value
          break
        } else if (this.nodeList[j].ip === ip) {
          nodeID = this.nodeList[j].value
          break
        } else if (this.nodeList[j].text.startsWith(i.Host)) {
          nodeID = this.nodeList[j].value
        } else if (!nodeID) {
          const lcName = this.nodeList[j].text.toLowerCase()
          const lcHost = i.Host.toLowerCase()
          if (lcName === lcHost || lcName.includes(lcHost)) {
            nodeID = this.nodeList[j].value
          }
        }
      }
      let filter = i.Message.replace(/[-/\\^$*+?.()|[\]{}]/g, '\\$&')
      if (i.Tag) {
        filter += `[\\s\\S\\n]*` + 'tag\\s*' + i.Tag
      }
      this.polling = {
        ID: '',
        Name: 'syslog監視',
        NodeID: nodeID,
        Type: 'syslog',
        Mode: 'count',
        Params: i.Host,
        Filter: filter,
        Extractor: '',
        Script: 'count < 1',
        Level: 'low',
        PollInt: 600,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.editPollingDialog = true
    },
    doAddPolling() {
      this.$axios
        .post('/api/polling/add', this.polling)
        .then(() => {
          this.editPollingDialog = false
        })
        .catch((e) => {
          this.addPollingError = true
        })
    },
    highlighter(code) {
      return highlight(code, languages.js)
    },
    regexHighlighter(code) {
      return highlight(code, languages.regex)
    },
    actionHighlighter(code) {
      return highlight(code, {
        property:
          /[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}/,
        string: /(wol|mail|line|chat|wait|cmd)/,
        number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,
        keyword: /\b(?:false|true|up|down)\b/,
      })
    },
  },
}
</script>

<style>
.script {
  height: 100px;
  overflow: auto;
  margin-top: 5px;
  margin-bottom: 5px;
}

.filter {
  overflow: auto;
  margin-top: 5px;
  margin-bottom: 5px;
}
</style>
