<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Syslog
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
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
          type="xls"
          name="TWSNMP_FC_Sysog.xls"
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
        <v-btn color="info" @click="aiAssistDialog = true">
          <v-icon>mdi-brain</v-icon>
          AIアシスト
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="500px">
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
            <v-menu
              ref="stMenu"
              v-model="stMenuShow"
              :close-on-content-click="false"
              :return-value.sync="filter.StartTime"
              transition="scale-transition"
              offset-y
              :nudge-right="40"
              max-width="290px"
              min-width="290px"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.StartTime"
                  label="開始時刻"
                  prepend-icon="mdi-clock-time-four-outline"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-time-picker
                v-if="stMenuShow"
                v-model="filter.StartTime"
                full-width
                @click:minute="$refs.stMenu.save(filter.StartTime)"
              ></v-time-picker>
            </v-menu>
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
            <v-menu
              ref="etMenu"
              v-model="etMenuShow"
              :close-on-content-click="false"
              :return-value.sync="filter.EndTime"
              transition="scale-transition"
              offset-y
              :nudge-right="40"
              max-width="290px"
              min-width="290px"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.EndTime"
                  label="終了時刻"
                  prepend-icon="mdi-clock-time-four-outline"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-time-picker
                v-if="etMenuShow"
                v-model="filter.EndTime"
                full-width
                dark
                @click:minute="$refs.etMenu.save(filter.EndTime)"
              ></v-time-picker>
            </v-menu>
          </v-row>
          <v-text-field
            v-model="filter.Type"
            label="種別（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.Host"
            label="ホスト名（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.Tag"
            label="タグ（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.Message"
            label="メッセージ（パイプライン正規表現）"
          ></v-text-field>
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
            type="xls"
            name="TWSNMP_FC_Sysog_Extract.xls"
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
    <v-dialog v-model="histogramDialog" persistent max-width="950px">
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
        <div id="histogram" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="histogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="clusterDialog" persistent max-width="950px">
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
        <div id="cluster" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="clusterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="host3DDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          ホスト別ログ(3D)
          <v-spacer></v-spacer>
        </v-card-title>
        <div id="host3d" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="host3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="hostListDialog" persistent max-width="950px">
      <v-card style="width: 100%">
        <v-card-title>
          ホスト別ログ
          <v-spacer></v-spacer>
        </v-card-title>
        <v-card-text>
          <div id="hostList" style="width: 900px; height: 500px"></div>
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
    <v-dialog v-model="extractHistogramDialog" persistent max-width="950px">
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
        <div id="histogram" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="extractHistogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extractClusterDialog" persistent max-width="950px">
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
        <div id="cluster" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="extractClusterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="extract3DDialog" persistent max-width="1050px">
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
          <div id="extract3d" style="width: 1000px; height: 750px"></div>
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
    <v-dialog v-model="extractTopListDialog" persistent max-width="950px">
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
          <div id="extractTopList" style="width: 900px; height: 500px"></div>
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
    <v-dialog v-model="aiAssistDialog" persistent max-width="1200px">
      <v-card style="width: 100%">
        <v-card-title> AIアシスト分析 </v-card-title>
        <v-card-text>
          <div
            v-if="hasAIErrorChart"
            id="aiAssist"
            style="width: 1200px; height: 200px"
          ></div>
          <v-data-table
            v-model="selectedAILogs"
            :headers="aiAssistHeaders"
            :items="logs"
            sort-by="TimeStr"
            sort-desc
            dense
            item-key="ID"
            show-select
            :loading="aiProcessing"
            loading-text="AI Thinking... Please wait"
            class="log"
          >
            <template #[`item.Level`]="{ item }">
              <v-icon :color="$getStateColor(item.Level)">{{
                $getStateIconName(item.Level)
              }}</v-icon>
              {{ $getStateName(item.Level) }}
            </template>
            <template #[`body.append`]>
              <tr>
                <td></td>
                <td colspan="3">
                  <v-switch
                    v-model="aiFilter.hasAIResult"
                    label="結果があるもの"
                  ></v-switch>
                </td>
                <td></td>
                <td>
                  <v-text-field
                    v-model="aiFilter.host"
                    label="Host"
                  ></v-text-field>
                </td>
                <td>
                  <v-text-field
                    v-model="aiFilter.tag"
                    label="Tag"
                  ></v-text-field>
                </td>
                <td>
                  <v-text-field
                    v-model="aiFilter.msg"
                    label="Message"
                  ></v-text-field>
                </td>
              </tr>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="hasSelectedAILogs"
            color="info"
            dark
            @click="setAIClassDialog = true"
          >
            <v-icon>mdi-teach</v-icon>
            教育
          </v-btn>
          <v-btn color="primary" dark @click="doAIAssist">
            <v-icon>mdi-brain</v-icon>
            分析
          </v-btn>
          <v-btn color="normal" dark @click="aiAssistDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="setAIClassDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">分類を教える</span>
        </v-card-title>
        <v-card-text>
          <v-text-field v-model="AIClass" dense></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="setAIClass">
            <v-icon>mdi-content-save</v-icon>
            設定
          </v-btn>
          <v-btn color="normal" @click="setAIClassDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="fftDialog" persistent max-width="1050px">
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
          <div id="FFTChart" style="width: 1000px; height: 500px"></div>
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
    <v-dialog v-model="fft3DDialog" persistent max-width="1050px">
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
          <div id="FFTChart3D" style="width: 1000px; height: 600px"></div>
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
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
  data() {
    return {
      filterDialog: false,
      sdMenuShow: false,
      stMenuShow: false,
      edMenuShow: false,
      etMenuShow: false,
      nodeList: [],
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
      },
      headers: [
        { text: '状態', value: 'Level', width: '8%' },
        { text: '日時', value: 'TimeStr', width: '17%' },
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
          width: '45%',
          filter: (value) => {
            if (!this.conf.msg) return true
            return value.includes(this.conf.msg)
          },
        },
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
      aiAssistDialog: false,
      aiFilter: {
        host: '',
        tag: '',
        msg: '',
        hasAIResult: false,
      },
      hasAIErrorChart: false,
      aiProcessing: false,
      setAIClassDialog: false,
      selectedAILogs: [],
      AIClass: '',
      aiAssistHeaders: [
        { text: '状態', value: 'Level', width: '13%' },
        { text: '教育', value: 'AIClass', width: '8%' },
        {
          text: 'AI回答',
          value: 'AIResult',
          width: '8%',
          filter: (value) => {
            if (!this.aiFilter.hasAIResult) return true
            return value && value !== ''
          },
        },
        { text: '日時', value: 'TimeStr', width: '13%' },
        {
          text: 'ホスト名',
          value: 'Host',
          width: '10%',
          filter: (value) => {
            if (!this.aiFilter.host) return true
            return value.includes(this.aiFilter.host)
          },
        },
        {
          text: 'タグ',
          value: 'Tag',
          width: '10%',
          filter: (value) => {
            if (!this.aiFilter.tag) return true
            return value.includes(this.aiFilter.tag)
          },
        },
        {
          text: 'メッセージ',
          value: 'Message',
          width: '40%',
          filter: (value) => {
            if (!this.aiFilter.msg) return true
            return value.includes(this.aiFilter.msg)
          },
        },
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
    }
  },
  async fetch() {
    this.fftMap = null
    const r = await this.$axios.$post('/api/log/syslog', this.filter)
    if (!r) {
      return
    }
    this.logs = r.Logs ? r.Logs : []
    let id = 0
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}')
      e.AIClass = ''
      e.AIResult = ''
      e.ID = id++
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
    this.$showLogLevelChart(this.logs)
    if (this.filterExtractorList.length < 1) {
      const groks = await this.$axios.$get('/api/conf/grok')
      if (groks) {
        this.filterExtractorList = [{ text: '指定しない', value: '' }]
        groks.forEach((g) => {
          this.filterExtractorList.push({
            text: g.Name,
            value: g.ID,
          })
        })
        this.filterExtractorList.sort((a, b) => {
          if (a.value < b.value) {
            return -1
          }
          if (a.value < b.value) {
            return 1
          }
          return 0
        })
      }
    }
    this.extractDatas = []
    this.extractHeader = []
    this.numExtractTypeList = []
    this.strExtractTypeList = []
    if (r.ExtractHeader.length < 1 || r.ExtractDatas.length < 1) {
      return
    }
    r.ExtractHeader.forEach((col) => {
      this.extractHeader.push({
        text: col,
        value: col,
      })
    })
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
  },
  computed: {
    hasSelectedAILogs() {
      return this.selectedAILogs.length > 0
    },
  },
  created() {
    const c = this.$store.state.log.logs.syslog
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  mounted() {
    this.$makeLogLevelChart('logCountChart')
    this.$showLogLevelChart(this.logs)
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
    doFilter() {
      if (this.filter.StartDate !== '' && this.filter.StartTime === '') {
        this.filter.StartTime = '00:00'
      }
      if (this.filter.EndDate !== '' && this.filter.EndTime === '') {
        this.filter.EndTime = '23:59'
      }
      this.filterDialog = false
      this.$fetch()
    },
    showHistogram() {
      this.histogramDialog = true
      this.$nextTick(() => {
        this.updateHistogram()
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
    doAIAssist() {
      this.hasAIErrorChart = false
      this.aiProcessing = true
      this.$syslogAIAssist(this.getFilteredSyslog()).then(() => {
        this.aiFilter.hasAIResult = true
        this.aiProcessing = false
        this.hasAIErrorChart = true
        this.$nextTick(() => {
          this.$showSyslogAIAssistChart('aiAssist')
        })
      })
    },
    setAIClass() {
      this.selectedAILogs.forEach((l) => {
        l.AIClass = this.AIClass
      })
      this.setAIClassDialog = false
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
  },
}
</script>
