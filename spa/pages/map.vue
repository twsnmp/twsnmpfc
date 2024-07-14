<template>
  <v-row justify="center">
    <v-snackbar v-model="wolError" absolute centered color="error">
      Wake on LANパケットを送信できません
    </v-snackbar>
    <v-snackbar v-model="wolDone" absolute centered color="primary">
      Wake on LANパケットを送信しました
    </v-snackbar>
    <v-card min-width="1000px" width="100%">
      <v-data-table
        :headers="headers"
        :items="map.Logs"
        sort-by="TimeStr"
        sort-desc
        dense
        disable-pagination
        height="300"
        hide-default-footer
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
      </v-data-table>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">ノード削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          ノードを削除できませんでした
        </v-alert>
        <v-card-text> 選択したノードを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteNode">
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
    <v-dialog v-model="deleteItemDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">描画アイテム</span>
        </v-card-title>
        <v-alert v-model="deleteItemError" color="error" dense dismissible>
          描画アイテムを削除できませんでした
        </v-alert>
        <v-card-text> 選択した描画アイテムを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteItem">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteItemDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="deleteNetworkDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">ネットワーク</span>
        </v-card-title>
        <v-alert v-model="deleteNetworkError" color="error" dense dismissible>
          ネットワークを削除できませんでした
        </v-alert>
        <v-card-text> 選択したネットワークを削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDeleteNetwork">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteNetworkDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editNodeDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">ノード設定</span>
        </v-card-title>
        <v-alert v-model="editNodeError" color="error" dense dismissible>
          ノードの保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field v-model="editNode.Name" label="名前"></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.IP"
                label="IPアドレス"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-select
                v-model="editNode.AddrMode"
                :items="$addrModeList"
                label="アドレスモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editNode.Icon"
                :items="$iconList"
                label="アイコン"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-icon x-large style="margin-top: 10px; margin-left: 10px">
                {{ $getIconName(editNode.Icon) }}
              </v-icon>
            </v-col>
            <v-col>
              <v-switch
                v-model="editNode.AutoAck"
                label="復帰時に自動確認"
                dense
              >
              </v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editNode.SnmpMode"
                :items="$snmpModeList"
                label="SNMPモード"
              >
              </v-select>
            </v-col>
            <v-col v-if="editNode.SnmpMode == ''">
              <v-text-field
                v-model="editNode.Community"
                label="Community"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editNode.User"
                autocomplete="username"
                label="ユーザー"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNode.Password"
                autocomplete="new-password"
                type="password"
                label="パスワード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-text-field
            v-model="editNode.PublicKey"
            label="公開鍵"
          ></v-text-field>
          <v-text-field v-model="editNode.URL" label="URL"></v-text-field>
          <v-text-field v-model="editNode.Descr" label="説明"></v-text-field>
          <v-switch
            v-if="copyFrom"
            v-model="copyPolling"
            label="ポーリングの設定も含めてコピーする"
          ></v-switch>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doUpdateNode">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn
            color="normal"
            dark
            @click="
              editNodeDialog = false
              $fetch()
            "
          >
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editItemDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">描画アイテム</span>
        </v-card-title>
        <v-alert v-model="editItemError" color="error" dense dismissible>
          描画アイテムの保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-select
              v-model="editItem.Type"
              :items="drawItemList"
              label="描画アイテムタイプ"
            >
            </v-select>
          </v-row>
          <v-row dense>
            <v-col v-if="editItem.Type < 2 || editItem.Type == 3">
              <v-text-field
                v-model="editItem.W"
                type="number"
                step="any"
                min="0"
                max="1000"
                label="幅"
              ></v-text-field>
            </v-col>
            <v-col
              v-if="
                editItem.Type < 2 || editItem.Type == 3 || editItem.Type >= 6
              "
            >
              <v-text-field
                v-model="editItem.H"
                type="number"
                step="any"
                min="0"
                max="1000"
                label="高さ"
              ></v-text-field>
            </v-col>
            <v-col
              v-if="
                editItem.Type == 2 || (editItem.Type > 3 && editItem.Type < 6)
              "
            >
              <v-text-field
                v-model="editItem.Size"
                type="number"
                step="any"
                min="8"
                max="128"
                label="文字サイズ"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-color-picker v-if="editItem.Type < 3" v-model="editItem.Color">
          </v-color-picker>
          <v-text-field
            v-if="editItem.Type == 2"
            v-model="editItem.Text"
            label="文字列"
          >
          </v-text-field>
          <v-autocomplete
            v-if="editItem.Type >= 4"
            v-model="editItem.PollingID"
            :items="itemPollingList"
            label="ポーリング"
          >
          </v-autocomplete>
          <v-text-field
            v-if="editItem.Type >= 4"
            v-model="editItem.VarName"
            label="結果の変数名"
          >
          </v-text-field>
          <v-text-field
            v-if="editItem.Type == 4"
            v-model="editItem.Format"
            label="表示フォーマット"
          >
          </v-text-field>
          <v-text-field
            v-if="editItem.Type >= 5"
            v-model="editItem.Text"
            label="ラベル"
          >
          </v-text-field>
          <v-text-field
            v-model="editItem.Scale"
            type="number"
            label="倍率"
          ></v-text-field>
          <v-select
            v-if="editItem.Type == 3"
            v-model="editItem.Path"
            :items="map.Images"
            label="画像ファイル"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="editItem.Type == 3"
            color="primary"
            dark
            @click="imageUploadDialog = true"
          >
            <v-icon>mdi-image</v-icon>
            画像ファイルアップロード
          </v-btn>
          <v-btn
            v-if="editItem.Type == 3 && notUsedImages.length > 0"
            color="error"
            dark
            @click="imageDeleteDialog = true"
          >
            <v-icon>mdi-image</v-icon>
            画像ファイル削除
          </v-btn>
          <v-btn color="primary" dark @click="doUpdateItem">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn
            color="normal"
            dark
            @click="
              editItemDialog = false
              $fetch()
            "
          >
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editNetworkDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">ネットワーク設定</span>
        </v-card-title>
        <v-alert v-model="editNetworkError" color="error" dense dismissible>
          ネットワークの保存に失敗しました
        </v-alert>
        <v-alert v-if="editNetwork.Error != ''" color="error" dense>
          {{ editNetwork.Error }}
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editNetwork.Name"
                label="名前"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNetwork.IP"
                label="IPアドレス"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editNetwork.HPorts"
                type="number"
                min="5"
                max="100"
                label="横の最大ポート数"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-chip v-if="editNetwork.LLDP" color="primary" variant="flat">
                LLDP
              </v-chip>
              <v-chip v-if="!editNetwork.LLDP" color="error" variant="flat">
                Not LLDP
              </v-chip>
            </v-col>
            <v-col>
              <v-switch
                v-model="editNetwork.AutoConn"
                label="自動で接続先を探す"
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editNetwork.SnmpMode"
                :items="$snmpModeList"
                label="SNMPモード"
              >
              </v-select>
            </v-col>
            <v-col v-if="editNetwork.SnmpMode == ''">
              <v-text-field
                v-model="editNetwork.Community"
                label="Community"
              ></v-text-field>
            </v-col>
            <v-col v-if="editNetwork.SnmpMode == ''"></v-col>
            <v-col v-if="editNetwork.SnmpMode != ''">
              <v-text-field
                v-model="editNetwork.User"
                autocomplete="username"
                label="ユーザー"
              ></v-text-field>
            </v-col>
            <v-col v-if="editNetwork.SnmpMode != ''">
              <v-text-field
                v-model="editNetwork.Password"
                autocomplete="new-password"
                type="password"
                label="パスワード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-text-field v-model="editNetwork.URL" label="URL"></v-text-field>
          <v-textarea
            v-model="editNetwork.Descr"
            clear-icon="mdi-close-circle"
            label="説明"
            rows="3"
            clearable
          >
          </v-textarea>
          <v-text-field
            v-if="editNetwork.LLDP"
            v-model="editNetwork.SystemID"
            label="ID"
            readonly
          ></v-text-field>
          <v-data-table
            :headers="netPortHeaders"
            :items="editNetwork.Ports"
            :items-per-page="5"
            dense
          >
            <template #[`item.State`]="{ item }">
              <v-icon :color="$getStateColor(item.State)">{{
                $getStateIconName(item.State)
              }}</v-icon>
              {{ $getStateName(item.State) }}
            </template>
            <template #[`item.actions`]="{ item }">
              <v-icon small @click="topNetworkPort(item)">
                mdi-arrow-collapse-up
              </v-icon>
              <v-icon small @click="upNetworkPort(item)"> mdi-arrow-up </v-icon>
              <v-icon small @click="downNetworkPort(item)">
                mdi-arrow-down
              </v-icon>
              <v-icon small @click="bottomNetworkPort(item)">
                mdi-arrow-collapse-down
              </v-icon>
              <v-icon small @click="editNetworkPortFunc(item)">
                mdi-pencil
              </v-icon>
              <v-icon small color="red" @click="deleteNetworkPort(item)">
                mdi-delete
              </v-icon>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn v-if="editNetwork.ID" color="error" @click="researchNetwork">
            <v-icon>mdi-refresh</v-icon>
            再検索
          </v-btn>
          <v-btn
            v-if="editNetwork.ID"
            color="error"
            @click="reNumberNetworkPort"
          >
            <v-icon>mdi-order-numeric-ascending</v-icon>
            ポート再配置
          </v-btn>
          <v-btn color="primary" dark @click="doUpdateNetwork">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn
            color="normal"
            dark
            @click="
              editNetworkDialog = false
              $fetch()
            "
          >
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="editNetworkPortDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">ポート設定</span>
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="editNetworkPort.Name"
            label="名前"
          ></v-text-field>
          <v-text-field
            v-model="editNetworkPort.Polling"
            label="ポーリング"
          ></v-text-field>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editNetworkPort.X"
                type="number"
                min="0"
                max="100"
                style="width: 80px"
                label="X"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNetworkPort.Y"
                type="number"
                min="0"
                max="100"
                style="width: 80px"
                label="Y"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editNetworkPort.ID"
                label="ID"
                readonly
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="saveNetworkPort">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="editNetworkPortDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="lineDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">ライン編集</span>
        </v-card-title>
        <v-alert v-model="lineError" color="error" dense dismissible>
          ラインの保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="editLine.NodeName1"
                label="ノード１"
                disabled
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editLine.NodeName2"
                label="ノード２"
                disabled
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col style="height: 160px; overflow: auto">
              <v-list dense>
                <v-list-item-group
                  v-model="selectedLinePolling1"
                  color="primary"
                >
                  <v-list-item v-for="(item, i) in pollingList1" :key="i">
                    <v-list-item-icon>
                      <v-icon :color="$getStateColor(item.state)">
                        {{ $getStateIconName(item.state) }}
                      </v-icon>
                    </v-list-item-icon>
                    <v-list-item-content>
                      <v-list-item-title>{{ item.text }}</v-list-item-title>
                    </v-list-item-content>
                  </v-list-item>
                </v-list-item-group>
              </v-list>
            </v-col>
            <v-col style="height: 160px; overflow: auto">
              <v-list dense>
                <v-list-item-group
                  v-model="selectedLinePolling2"
                  color="primary"
                >
                  <v-list-item v-for="(item, i) in pollingList2" :key="i">
                    <v-list-item-icon>
                      <v-icon :color="$getStateColor(item.state)">
                        {{ $getStateIconName(item.state) }}
                      </v-icon>
                    </v-list-item-icon>
                    <v-list-item-content>
                      <v-list-item-title>{{ item.text }}</v-list-item-title>
                    </v-list-item-content>
                  </v-list-item>
                </v-list-item-group>
              </v-list>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editLine.PollingID"
                :items="linePollingList"
                label="情報のためのポーリング"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field v-model="editLine.Info" label="情報"></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="editLine.Width"
                :items="lineWidthList"
                label="ラインの太さ"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="editLine.Port"
                label="ポート"
              ></v-text-field>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="red" dark @click="deleteLine">
            <v-icon>mdi-lan-disconnect</v-icon>
            切断
          </v-btn>
          <v-btn color="primary" dark @click="addLine">
            <v-icon>mdi-lan-connect</v-icon>
            接続
          </v-btn>
          <v-btn color="normal" dark @click="lineDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="showNodeDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">ノード情報</span>
        </v-card-title>
        <v-snackbar v-model="wolError" absolute centered color="error">
          Wake on LANパケットを送信できません
        </v-snackbar>
        <v-snackbar v-model="wolDone" absolute centered color="primary">
          Wake on LANパケットを送信しました
        </v-snackbar>
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
                <td>状態</td>
                <td>
                  <v-icon :color="$getStateColor(editNode.State)">
                    {{ $getIconName(editNode.Icon) }}
                  </v-icon>
                  {{ $getStateName(editNode.State) }}
                </td>
              </tr>
              <tr>
                <td>名前</td>
                <td>{{ editNode.Name }}</td>
              </tr>
              <tr>
                <td>IPアドレス</td>
                <td>{{ editNode.IP }}</td>
              </tr>
              <tr>
                <td>IPv6アドレス</td>
                <td>{{ editNode.IPv6 }}</td>
              </tr>
              <tr>
                <td>MACアドレス</td>
                <td>{{ editNode.MAC }}</td>
              </tr>
              <tr>
                <td>説明</td>
                <td>{{ editNode.Descr }}</td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-divider></v-divider>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-menu offset-y>
            <template #activator="{ on, attrs }">
              <v-btn dark v-bind="attrs" v-on="on">
                <v-icon>mdi-menu</v-icon>
                操作メニュー
              </v-btn>
            </template>
            <v-list>
              <v-list-item @click="editNodeDialog = true">
                <v-list-item-icon>
                  <v-icon>mdi-pencil</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>編集</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="checkPolling(false)">
                <v-list-item-icon>
                  <v-icon>mdi-cached</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>再確認</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="deleteDialog = true">
                <v-list-item-icon>
                  <v-icon color="red">mdi-delete</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>削除</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="copyNode()">
                <v-list-item-icon>
                  <v-icon>mdi-content-copy</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>コピー</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showMIBBr()">
                <v-list-item-icon><v-icon>mdi-eye</v-icon></v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>MIBブラウザー</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showPing()">
                <v-list-item-icon>
                  <v-icon>mdi-check-network</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>PING</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showVPanelPage()">
                <v-list-item-icon>
                  <v-icon>mdi-apps-box</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>パネル</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showHostResourcePage()">
                <v-list-item-icon>
                  <v-icon>mdi-gauge</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>ホストリソース</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showRMONPage()">
                <v-list-item-icon>
                  <v-icon>mdi-minus-network</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>RMON管理</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showTcpUdpPortPage()">
                <v-list-item-icon>
                  <v-icon>mdi-power-socket-jp</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>TCP/UDPポート</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showNodePollingPage()">
                <v-list-item-icon>
                  <v-icon>mdi-lan-check</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>ポーリング</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showNodeLogPage()">
                <v-list-item-icon>
                  <v-icon>mdi-calendar-check</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>ログ</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item v-if="editNode.MAC" @click="doWOL">
                <v-list-item-icon><v-icon>mdi-alarm</v-icon></v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>Wake On LAN</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item
                v-for="(url, i) in urls"
                :key="i"
                @click="openURL(url)"
              >
                <v-list-item-icon>
                  <v-icon>mdi-link</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>{{ url }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-menu>
          <v-btn color="normal" dark @click="showNodeDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="openURLDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">アクセス先URLの選択</span>
        </v-card-title>
        <v-card-text>
          <v-radio-group v-model="selectedURL" mandatory>
            <v-radio
              v-for="(url, i) in urls"
              :key="i"
              :label="url"
              :value="url"
            ></v-radio>
          </v-radio-group>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="openURL(selectedURL)">
            <v-icon>mdi-link</v-icon>
            選択
          </v-btn>
          <v-btn color="normal" @click="openURLDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-menu
      v-model="showMapContextMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="addNode()">
          <v-list-item-icon><v-icon>mdi-plus</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>新規ノード</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="addItem()">
          <v-list-item-icon><v-icon>mdi-drawing</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>描画アイテム</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="addNetwork()">
          <v-list-item-icon><v-icon>mdi-lan</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>新規ネットワーク</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="checkPolling(true)">
          <v-list-item-icon><v-icon>mdi-cached</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>全て再確認</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item :to="discoverURL">
          <v-list-item-icon><v-icon>mdi-file-find</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>自動発見</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item to="/conf/map">
          <v-list-item-icon><v-icon>mdi-cog</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>マップ設定</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="gridDialog = true">
          <v-list-item-icon><v-icon>mdi-grid</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>グリッド整列</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-dialog v-model="gridDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">グリッド整列</span>
        </v-card-title>
        <v-card-text>
          <v-radio-group v-model="selectedGrid" mandatory>
            <v-radio label="20" value="20"></v-radio>
            <v-radio label="40" value="40"></v-radio>
            <v-radio label="80" value="80"></v-radio>
            <v-radio label="90" value="90"></v-radio>
          </v-radio-group>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="grid(false)">
            <v-icon>mdi-eye</v-icon>
            テスト
          </v-btn>
          <v-btn color="error" @click="grid(true)">
            <v-icon>mdi-grid</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="gridDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="imageUploadDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">画僧ファイルアップロード</span>
        </v-card-title>
        <v-alert v-model="imageError" color="error" dense dismissible>
          画像ファイルの保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-file-input
            label="背景画像ファイル"
            accept="image/*"
            @change="selectFile"
          >
          </v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="doUploadImage">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" @click="imageUploadDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="imageDeleteDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">画僧ファイル操作</span>
        </v-card-title>
        <v-alert v-model="imageDeleteError" color="error" dense dismissible>
          画像ファイルの削除に失敗しました
        </v-alert>
        <v-card-text>
          <v-select
            v-model="selectedImagePath"
            :items="notUsedImages"
            label="削除できる画像ファイル"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn v-if="selectedImagePath" color="error" @click="doDeleteImage">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="imageDelteDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-menu
      v-model="showNodeContextMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="editNodeDialog = true">
          <v-list-item-icon><v-icon>mdi-pencil</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>編集</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="checkPolling(false)">
          <v-list-item-icon><v-icon>mdi-cached</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>再確認</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="deleteDialog = true">
          <v-list-item-icon
            ><v-icon color="red">mdi-delete</v-icon></v-list-item-icon
          >
          <v-list-item-content>
            <v-list-item-title>削除</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="copyNode()">
          <v-list-item-icon><v-icon>mdi-content-copy</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>コピー</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showMIBBr()">
          <v-list-item-icon><v-icon>mdi-eye</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>MIBブラウザー</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showPing()">
          <v-list-item-icon>
            <v-icon>mdi-check-network</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>PING</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showNodeDialog = true">
          <v-list-item-icon><v-icon>mdi-information</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>情報</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showNodePollingPage()">
          <v-list-item-icon><v-icon>mdi-lan-check</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>ポーリング</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showNodeLogPage()">
          <v-list-item-icon>
            <v-icon>mdi-calendar-check</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>ログ</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showVPanelPage()">
          <v-list-item-icon>
            <v-icon>mdi-apps-box</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>パネル</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showHostResourcePage()">
          <v-list-item-icon>
            <v-icon>mdi-gauge</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>ホストリソース</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showRMONPage()">
          <v-list-item-icon>
            <v-icon>mdi-minus-network</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>RMON管理</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="showTcpUdpPortPage()">
          <v-list-item-icon>
            <v-icon>mdi-power-socket-jp</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>TCP/UDPポート</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item v-if="editNode.MAC" @click="doWOL">
          <v-list-item-icon><v-icon>mdi-alarm</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>Wake On LAN</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item v-for="(url, i) in urls" :key="i" @click="openURL(url)">
          <v-list-item-icon>
            <v-icon>mdi-link</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>{{ url }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-menu
      v-model="showItemContextMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="editItemDialog = true">
          <v-list-item-icon><v-icon>mdi-pencil</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>編集</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="copyDrawItem()">
          <v-list-item-icon><v-icon>mdi-content-copy</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>コピー</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="deleteItemDialog = true">
          <v-list-item-icon
            ><v-icon color="red">mdi-delete</v-icon></v-list-item-icon
          >
          <v-list-item-content>
            <v-list-item-title>削除</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-menu
      v-model="showFormatNodesMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="horizontal()">
          <v-list-item-icon>
            <v-icon>mdi-format-vertical-align-center</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>水平に整列</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="vertical()">
          <v-list-item-icon>
            <v-icon>mdi-format-horizontal-align-center</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>垂直に整列</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="circle()">
          <v-list-item-icon>
            <v-icon> mdi-circle-outline</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>円形に整列</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-menu
      v-model="showNetworkContextMenu"
      :position-x="x"
      :position-y="y"
      absolute
    >
      <v-list dense>
        <v-list-item @click="editNetworkDialog = true">
          <v-list-item-icon><v-icon>mdi-pencil</v-icon></v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>編集</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
        <v-list-item @click="deleteNetworkDialog = true">
          <v-list-item-icon
            ><v-icon color="red">mdi-delete</v-icon></v-list-item-icon
          >
          <v-list-item-content>
            <v-list-item-title>削除</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-menu>
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      showNodeDialog: false,
      editNodeDialog: false,
      editNodeError: false,
      lineDialog: false,
      lineError: false,
      deleteDialog: false,
      deleteError: false,
      editItemDialog: false,
      editItemError: false,
      deleteItemDialog: false,
      deleteItemError: false,
      editNetworkDialog: false,
      editNetworkPortDialog: false,
      editNetworkError: false,
      deleteNetworkDialog: false,
      deleteNetworkError: false,
      showNode: {},
      editLine: {
        NodeID1: '',
        NodeID2: '',
        PollingID1: '',
        PollingID2: '',
        PollingID: '',
        Info: '',
        Port: '',
      },
      selectedLinePolling1: 0,
      selectedLinePolling2: 0,
      editNode: {},
      editItem: {},
      editNetwork: {},
      editNetworkPortIndex: 0,
      editNetworkPort: {},
      deleteNodes: [],
      map: {
        Nodes: {},
        Pollings: {},
        Lines: [],
        Items: {},
        Networks: {},
        MapConf: { MapName: '' },
        Logs: [],
      },
      search: '',
      headers: [
        { text: '状態', value: 'Level', width: '10%' },
        { text: '発生日時', value: 'TimeStr', width: '15%' },
        { text: '種別', value: 'Type', width: '10%' },
        { text: '関連ノード', value: 'NodeName', width: '15%' },
        { text: 'イベント', value: 'Event', width: '50%' },
      ],
      showMapContextMenu: false,
      showNodeContextMenu: false,
      showItemContextMenu: false,
      showNetworkContextMenu: false,
      x: 0,
      y: 0,
      copyFrom: '',
      copyPolling: false,
      lineWidthList: [
        { text: '1', value: 1 },
        { text: '2', value: 2 },
        { text: '3', value: 3 },
        { text: '4', value: 4 },
        { text: '5', value: 5 },
      ],
      drawItemList: [
        { text: '矩形', value: 0 },
        { text: '楕円', value: 1 },
        { text: 'ラベル', value: 2 },
        { text: 'イメージ', value: 3 },
        { text: 'ポーリング結果(テキスト)', value: 4 },
        { text: 'ポーリング結果(ゲージ)', value: 5 },
        { text: 'ポーリング結果(新ゲージ)', value: 6 },
        { text: 'ポーリング結果(バー)', value: 7 },
        { text: 'ポーリング結果(ライン)', value: 8 },
      ],
      selectedImagePath: '',
      imageFile: '',
      imageUploadDialog: false,
      imageDeleteDialog: false,
      imageError: false,
      imageDeleteError: false,
      notUsedImages: [],
      urls: [],
      wolDone: false,
      wolError: false,
      openURLDialog: false,
      selectedURL: '',
      gridDialog: false,
      selectedGrid: 20,
      itemPollingList: [],
      showFormatNodesMenu: false,
      formatNodes: [],
      netPortHeaders: [
        { text: '状態', value: 'State' },
        { text: '名前', value: 'Name' },
        { text: 'ポーリング', value: 'Polling' },
        { text: 'X', value: 'X' },
        { text: 'Y', value: 'Y' },
        { text: '操作', value: 'actions' },
      ],
    }
  },
  async fetch() {
    this.map = await this.$axios.$get('/api/map')
    const mb = document.getElementById('map')
    const sbt = mb ? mb.scrollTop : 0
    const sbl = mb ? mb.scrollLeft : 0
    this.itemPollingList = []
    for (const k in this.map.Pollings) {
      if (!this.map.Nodes[k]) {
        continue
      }
      this.map.Pollings[k].forEach((p) => {
        try {
          const nodeName = this.map.Nodes[k].Name + ':'
          this.itemPollingList.push({
            text: nodeName + p.Name,
            value: p.ID,
          })
        } catch {}
      })
    }
    this.$showMAP(
      'map',
      this.map,
      this.$axios.defaults.baseURL,
      this.$store.state.map.readOnly
    )
    if (mb) {
      const ma = document.getElementById('map')
      if (ma) {
        ma.scrollTop = sbt
        ma.scrollLeft = sbl
      }
    }
    this.$store.commit('map/setMAP', this.map)
    this.map.Logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    const nodeID = this.$route.query.node
    if (nodeID && this.map.Nodes[nodeID]) {
      this.$selectNode(nodeID)
    }
    const usedImages = {}
    for (const k in this.map.Items) {
      if (this.map.Items[k].Type === 3) {
        usedImages[this.map.Items[k].Path] = true
      }
    }
    this.notUsedImages = []
    for (const i of this.map.Images) {
      if (!usedImages[i]) {
        this.notUsedImages.push(i)
      }
    }
  },
  computed: {
    pollingList1() {
      return this.pollingList(this.editLine.NodeID1, false)
    },
    pollingList2() {
      return this.pollingList(this.editLine.NodeID2, false)
    },
    linePollingList() {
      const l1 = [{ text: '設定しない', value: '' }]
      const l2 = this.pollingList(this.editLine.NodeID1, true)
      const l3 = this.pollingList(this.editLine.NodeID2, true)
      return l1.concat(l2, l3)
    },
    discoverURL() {
      return `/discover?x=${this.x}&y=${this.y}`
    },
  },
  mounted() {
    this.$setIconCodeMap(this.$iconList)
    this.$setStateColorMap(this.$stateList)
    this.$setCallback(this.callback)
  },
  beforeDestroy() {
    this.$setMapContextMenu(true)
  },
  methods: {
    nodeName(id) {
      return this.map.Nodes[id] ? this.map.Nodes[id].Name : ''
    },
    pollingList(id, lineMode) {
      const l = []
      if (!this.map.Nodes[id] || !this.map.Pollings[id]) {
        return l
      }
      let nodeName = ''
      if (lineMode) {
        nodeName = this.map.Nodes[id].Name + ':'
      }
      this.map.Pollings[id].forEach((p) => {
        if (!lineMode || p.Mode === 'traffic') {
          l.push({
            text: nodeName + p.Name,
            state: p.State,
            value: p.ID,
          })
        }
      })
      return l
    },
    getPollingIndex(nid, pid) {
      if (!this.map.Pollings[nid]) {
        return -1
      }
      let i = 0
      let sel = -1
      this.map.Pollings[nid].forEach((p) => {
        if (p.ID === pid) {
          sel = i
        }
        i++
      })
      return sel
    },
    callback(r) {
      if (
        this.deleteDialog ||
        this.showNodeDialog ||
        this.lineDialog ||
        this.editNodeDialog
      ) {
        return
      }
      switch (r.Cmd) {
        case 'updateNodesPos':
          this.$axios.post('/api/map/update', r.Param)
          break
        case 'updateItemsPos':
          this.$axios.post('/api/map/update_item', r.Param)
          break
        case 'updateNetworkPos':
          this.$axios.post('/api/map/update_network', r.Param)
          break
        case 'updateItem':
          this.editItem = this.map.Items[r.Param]
          this.doUpdateItem()
          break
        case 'deleteNodes':
          this.deleteNodes = Array.from(r.Param)
          this.deleteDialog = true
          break
        case 'editLine':
          this.showEditLineDiaglog(r.Param)
          break
        case 'refresh':
          this.$fetch()
          break
        case 'nodeDoubleClicked':
          if (this.map.Nodes[r.Param]) {
            const n = this.map.Nodes[r.Param]
            if (n.URL) {
              this.urls = n.URL.split(',')
              if (this.urls.length === 1) {
                this.openURL(this.urls[0])
                return
              } else if (this.urls.length > 1) {
                this.openURLDialog = true
                return
              }
            }
            this.urls = []
            this.copyFrom = ''
            this.editNode = this.map.Nodes[r.Param]
            this.showNodeDialog = true
          }
          break
        case 'itemDoubleClicked':
          if (this.map.Items[r.Param]) {
            this.editItem = this.map.Items[r.Param]
            this.editItemDialog = true
          }
          break
        case 'networkDoubleClicked':
          if (this.map.Networks[r.Param]) {
            const n = this.map.Networks[r.Param]
            if (n.URL) {
              this.urls = n.URL.split(',')
              if (this.urls.length === 1) {
                this.openURL(this.urls[0])
                return
              } else if (this.urls.length > 1) {
                this.openURLDialog = true
                return
              }
            }
            this.urls = []
            this.editNetwork = n
            this.editNetworkDialog = true
          }
          break
        case 'contextMenu':
          this.x = r.x
          this.y = r.y
          if (r.Node) {
            if (!this.map.Nodes[r.Node]) {
              return
            }
            this.copyFrom = ''
            this.editNode = this.map.Nodes[r.Node]
            this.urls = []
            this.editNode.URL.split(',').forEach((u) => {
              u = u.trim()
              if (u !== '') {
                this.urls.push(u)
              }
            })
            this.deleteNodes = [r.Node]
            this.showNodeContextMenu = true
          } else if (r.Item) {
            if (!this.map.Items[r.Item]) {
              return
            }
            this.editItem = this.map.Items[r.Item]
            this.showItemContextMenu = true
          } else if (r.Network) {
            if (!this.map.Networks[r.Network]) {
              return
            }
            this.editNetwork = this.map.Networks[r.Network]
            this.showNetworkContextMenu = true
          } else {
            this.showMapContextMenu = true
            this.editNode.ID = ''
          }
          break
        case 'formatNodes':
          this.x = r.x
          this.y = r.y
          this.formatNodes = Array.from(r.Param)
          this.showFormatNodesMenu = true
          break
      }
    },
    showEditLineDiaglog(p) {
      if (p.length !== 2 || !this.map.Nodes[p[0]] || !this.map.Nodes[p[1]]) {
        return
      }
      const l = this.map.Lines.find(
        (e) =>
          (e.NodeID1 === p[0] && e.NodeID2 === p[1]) ||
          (e.NodeID1 === p[1] && e.NodeID2 === p[0])
      )
      this.editLine = l || {
        NodeID1: p[0],
        PollingID2: '',
        NodeID2: p[1],
        PollingID1: '',
        PollingID: '',
        Info: '',
      }
      this.selectedLinePolling1 = this.getPollingIndex(
        this.editLine.NodeID1,
        this.editLine.PollingID1
      )
      this.selectedLinePolling2 = this.getPollingIndex(
        this.editLine.NodeID2,
        this.editLine.PollingID2
      )
      this.editLine.NodeName1 = this.nodeName(this.editLine.NodeID1)
      this.editLine.NodeName2 = this.nodeName(this.editLine.NodeID2)
      this.lineDialog = true
    },
    doDeleteNode() {
      this.deleteError = false
      this.$axios
        .post('/api/nodes/delete', this.deleteNodes)
        .then(() => {
          this.$fetch()
          this.deleteDialog = false
        })
        .catch((e) => {
          this.deleteError = true
        })
    },
    doDeleteItem() {
      this.deleteItemError = false
      this.$axios
        .post('/api/nodes/delete_items', [this.editItem.ID])
        .then(() => {
          this.$fetch()
          this.deleteItemDialog = false
        })
        .catch((e) => {
          this.deleteItemError = true
        })
    },
    doDeleteNetwork() {
      this.deleteNetworkError = false
      this.$axios
        .delete('/api/network/' + this.editNetwork.ID)
        .then(() => {
          this.$fetch()
          this.deleteNetworkDialog = false
        })
        .catch((e) => {
          this.deleteNetworkError = true
        })
    },
    editNetworkPortFunc(item) {
      const i = this.editNetwork.Ports.indexOf(item)
      this.editNetworkPortIndex = i
      this.editNetworkPort = {
        Name: item.Name,
        Polling: item.Polling,
        X: item.X,
        Y: item.Y,
        ID: item.ID,
        State: item.State,
      }
      this.editNetworkPortDialog = true
    },
    saveNetworkPort() {
      this.editNetwork.Ports.splice(
        this.editNetworkPortIndex,
        1,
        this.editNetworkPort
      )
      this.editNetworkPortDialog = false
    },
    deleteNetworkPort(item) {
      const i = this.editNetwork.Ports.indexOf(item)
      this.editNetwork.Ports.splice(i, 1)
    },
    upNetworkPort(item) {
      const i = this.editNetwork.Ports.indexOf(item)
      if (i <= 0) {
        return
      }
      const r = this.editNetwork.Ports.splice(i, 1)
      this.editNetwork.Ports.splice(i - 1, 0, r[0])
    },
    topNetworkPort(item) {
      const i = this.editNetwork.Ports.indexOf(item)
      if (i <= 0) {
        return
      }
      const r = this.editNetwork.Ports.splice(i, 1)
      this.editNetwork.Ports.unshift(r[0])
    },
    downNetworkPort(item) {
      const i = this.editNetwork.Ports.indexOf(item)
      if (i >= this.editNetwork.Ports.length - 1) {
        return
      }
      const r = this.editNetwork.Ports.splice(i, 1)
      this.editNetwork.Ports.splice(i + 1, 0, r[0])
    },
    bottomNetworkPort(item) {
      const i = this.editNetwork.Ports.indexOf(item)
      if (i >= this.editNetwork.Ports.length - 1) {
        return
      }
      const r = this.editNetwork.Ports.splice(i, 1)
      this.editNetwork.Ports.push(r[0])
    },
    doUpdateNode() {
      let url = '/api/node/update'
      if (this.copyFrom && this.copyPolling) {
        url += '?from=' + this.copyFrom
      }
      this.editNodeError = false
      this.$axios
        .post(url, this.editNode)
        .then(() => {
          this.$fetch()
          this.editNodeDialog = false
        })
        .catch((e) => {
          this.editNodeError = true
        })
    },
    doUpdateItem() {
      this.editItem.Size = this.editItem.Size * 1
      this.editItem.X = this.editItem.X * 1
      this.editItem.Y = this.editItem.Y * 1
      this.editItem.H = this.editItem.H * 1
      this.editItem.W = this.editItem.W * 1
      this.editItem.Scale = this.editItem.Scale * 1.0
      const url = '/api/item/update'
      this.editItemError = false
      this.$axios
        .post(url, this.editItem)
        .then(() => {
          this.$fetch()
          this.editItemDialog = false
        })
        .catch((e) => {
          this.editItemError = true
        })
    },
    doUpdateNetwork() {
      this.editNetworkError = false
      this.editNetwork.X *= 1
      this.editNetwork.Y *= 1
      this.editNetwork.H *= 1
      this.editNetwork.W *= 1
      this.editNetwork.HPorts *= 1
      this.$axios
        .post('/api/network/update', this.editNetwork)
        .then(() => {
          this.$fetch()
          this.editNetworkDialog = false
        })
        .catch((e) => {
          this.editnNetworkError = true
        })
    },
    researchNetwork() {
      this.editNetworkError = false
      this.editNetwork.X *= 1
      this.editNetwork.Y *= 1
      this.editNetwork.H *= 1
      this.editNetwork.W *= 1
      this.editNetwork.Ports = [] // ポートをクリア
      this.editNetwork.Error = '' // エラーをクリア
      this.editNetwork.LLDP = false // LLDPをクリア
      this.editNetwork.HPorts *= 1
      this.$axios
        .post('/api/network/update', this.editNetwork)
        .then(() => {
          this.$fetch()
          this.editNetworkDialog = false
        })
        .catch((e) => {
          this.editnNetworkError = true
        })
    },
    reNumberNetworkPort() {
      const ports = []
      let x = 0
      let y = 0
      const HPorts = this.editNetwork.HPorts || 24
      this.editNetwork.Ports.forEach((p) => {
        p.X = x
        p.Y = y
        x++
        if (x >= HPorts) {
          x = 0
          y++
        }
        ports.push(p)
      })
      this.editNetwork.Ports = ports
    },
    addNode() {
      const m = document.getElementById('map')
      const x = Math.trunc(m && m.scrollLeft ? this.x + m.scrollLeft : this.x)
      const y = Math.trunc(m && m.scrollTop ? this.y + m.scrollTop : this.y)
      this.copyFrom = ''
      this.editNode = {
        ID: '',
        Name: '新規ノード',
        IP: '',
        X: x,
        Y: y,
        Descr: '',
        Icon: 'desktop',
        MAC: '',
        SnmpMode: '',
        Community: '',
        User: '',
        Password: '',
        PublicKey: '',
        URL: '',
        Type: '',
        AddrMode: '',
        AutoAck: false,
      }
      this.editNodeDialog = true
    },
    addItem() {
      const m = document.getElementById('map')
      const x = Math.trunc(m && m.scrollLeft ? this.x + m.scrollLeft : this.x)
      const y = Math.trunc(m && m.scrollTop ? this.y + m.scrollTop : this.y)
      this.editItem = {
        ID: '',
        Type: 2, // Text
        X: x,
        Y: y,
        W: 100,
        H: 32,
        Text: '新しいラベル',
        Color: '#ccc',
        Size: 24,
        Scale: 1.0,
        Format: '',
        VarName: '',
      }
      this.editItemDialog = true
    },
    addNetwork() {
      const m = document.getElementById('map')
      const x = Math.trunc(m && m.scrollLeft ? this.x + m.scrollLeft : this.x)
      const y = Math.trunc(m && m.scrollTop ? this.y + m.scrollTop : this.y)
      this.editNetwork = {
        ID: '',
        Name: '',
        IP: '',
        X: x,
        Y: y,
        W: 0,
        H: 0,
        HPorts: 24,
        AutoCon: false,
        Descr: '',
        SnmpMode: this.map.MapConf.SnmpMode,
        Community: this.map.MapConf.Community,
        User: this.map.MapConf.User,
        Password: '',
        URL: '',
        Ports: [],
        Error: '',
        LLDP: false,
      }
      this.editNetworkDialog = true
    },
    copyNode() {
      this.showNodeDialog = false
      this.copyFrom = this.editNode.ID
      // 位置をずらして新規追加
      this.editNode.X += 64
      this.editNode.ID = ''
      this.editNode.State = 'unknown'
      this.editNode.Name += 'のコピー'
      this.editNodeDialog = true
    },
    copyDrawItem() {
      this.editItem.X += 64
      this.editItem.ID = ''
      this.editItem.Text += 'のコピー'
      this.editItemDialog = true
    },
    showNodePollingPage() {
      this.$router.push({ path: '/node/polling/' + this.editNode.ID })
    },
    showNodeLogPage() {
      this.$router.push({ path: '/node/log/' + this.editNode.ID })
    },
    showVPanelPage() {
      this.$router.push({ path: '/node/vpanel/' + this.editNode.ID })
    },
    showHostResourcePage() {
      this.$router.push({ path: '/node/hostResource/' + this.editNode.ID })
    },
    showRMONPage() {
      this.$router.push({ path: '/node/rmon/' + this.editNode.ID })
    },
    showTcpUdpPortPage() {
      this.$router.push({ path: '/node/port/' + this.editNode.ID })
    },
    showMIBBr() {
      this.$router.push({ path: '/mibbr/' + this.editNode.ID })
    },
    showPing() {
      this.$router.push({ path: '/ping/' + this.editNode.IP })
    },
    openURL(url) {
      this.openURLDialog = false
      window.open(url, '_blank')
    },
    addLine() {
      if (
        this.map.Pollings[this.editLine.NodeID1] &&
        this.selectedLinePolling1 >= 0 &&
        this.selectedLinePolling1 <
          this.map.Pollings[this.editLine.NodeID1].length
      ) {
        this.editLine.PollingID1 =
          this.map.Pollings[this.editLine.NodeID1][this.selectedLinePolling1].ID
      } else {
        this.editLine.PollingID1 = ''
      }
      if (
        this.map.Pollings[this.editLine.NodeID2] &&
        this.selectedLinePolling2 >= 0 &&
        this.selectedLinePolling2 <
          this.map.Pollings[this.editLine.NodeID2].length
      ) {
        this.editLine.PollingID2 =
          this.map.Pollings[this.editLine.NodeID2][this.selectedLinePolling2].ID
      } else {
        this.editLine.PollingID2 = ''
      }
      this.lineError = false
      this.$axios
        .post('/api/line/add', this.editLine)
        .then(() => {
          this.$fetch()
          this.lineDialog = false
        })
        .catch((e) => {
          this.lineError = true
        })
    },
    deleteLine() {
      this.lineError = false
      this.$axios
        .post('/api/line/delete', this.editLine)
        .then(() => {
          this.$fetch()
          this.lineDialog = false
        })
        .catch((e) => {
          this.lineError = true
        })
    },
    checkPolling(all) {
      let id = 'all'
      if (!all) {
        if (this.editNode.ID === '') {
          return
        }
        id = this.editNode.ID
      }
      this.$axios.get('/api/polling/check/' + id).then(() => {
        this.showNodeDialog = false
        this.$fetch()
      })
    },
    doWOL() {
      this.$axios
        .post('/api/wol/' + this.editNode.ID)
        .then(() => {
          this.wolDone = true
          this.$fetch()
        })
        .catch((e) => {
          this.wolError = true
        })
    },
    async grid(d) {
      const list = []
      const g = Math.max(this.selectedGrid * 1.0, 20)
      const mx = Math.ceil(2500 / g)
      const my = Math.ceil(5000 / g)
      const m = new Array(mx)
      for (let x = 0; x < m.length; x++) {
        m[x] = new Array(my)
        for (let y = 0; y < m[x].length; y++) {
          m[x][y] = false
        }
      }
      for (const id in this.map.Nodes) {
        let x = Math.max(
          Math.min(Math.ceil((this.map.Nodes[id].X * 1.0) / g), mx - 1),
          0
        )
        let y = Math.max(
          Math.min(Math.ceil((this.map.Nodes[id].Y * 1.0) / g), my - 1),
          0
        )
        while (m[x][y]) {
          x++
          if (x >= mx) {
            y++
            x = 0
            if (y >= my) {
              y = 0
              break
            }
          }
        }
        m[x][y] = true
        this.map.Nodes[id].X = x * g
        this.map.Nodes[id].Y = y * g
        list.push({
          ID: id,
          X: this.map.Nodes[id].X,
          Y: this.map.Nodes[id].Y,
        })
      }
      this.gridDialog = false
      if (d && list.length > 0) {
        await this.$axios.post('/api/map/update', list)
      }
    },
    selectFile(f) {
      this.imageFile = f
    },
    doUploadImage() {
      const formData = new FormData()
      formData.append('file', this.imageFile)
      this.imageError = false
      this.$axios
        .$post('/api/image', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })
        .then((r) => {
          this.imageUploadDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.imageError = true
          this.$fetch()
        })
    },
    doDeleteImage() {
      this.imageDeleteError = false
      this.$axios
        .delete('/api/image/' + this.selectedImagePath)
        .then((r) => {
          this.imageDeleteDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.imageDeleteError = true
          this.$fetch()
        })
    },
    async horizontal() {
      const list = []
      if (!this.formatNodes || this.formatNodes.length < 2) {
        return
      }
      this.formatNodes.sort((a, b) => {
        return this.map.Nodes[a].X - this.map.Nodes[b].X
      })
      const id0 = this.formatNodes[0]
      let idLast = ''
      let dx = this.map.Nodes[this.formatNodes[1]].X - this.map.Nodes[id0].X
      if (dx < 60) {
        dx = 60
      }
      for (const id of this.formatNodes) {
        if (id !== id0) {
          this.map.Nodes[id].Y = this.map.Nodes[id0].Y
          this.map.Nodes[id].X = this.map.Nodes[idLast].X + dx
          if (this.map.Nodes[id].X > 2500 - 40) {
            this.map.Nodes[id].X = 2500 - 40
          }
          list.push({
            ID: id,
            X: this.map.Nodes[id].X,
            Y: this.map.Nodes[id].Y,
          })
        }
        idLast = id
      }
      if (list.length > 0) {
        await this.$axios.post('/api/map/update', list)
      }
    },
    async vertical() {
      const list = []
      if (!this.formatNodes || this.formatNodes.length < 2) {
        return
      }
      this.formatNodes.sort((a, b) => {
        return this.map.Nodes[a].Y - this.map.Nodes[b].Y
      })
      const id0 = this.formatNodes[0]
      let idLast = ''
      let dy = this.map.Nodes[this.formatNodes[1]].X - this.map.Nodes[id0].X
      if (dy < 60) {
        dy = 60
      }
      for (const id of this.formatNodes) {
        if (id !== id0) {
          this.map.Nodes[id].X = this.map.Nodes[id0].X
          this.map.Nodes[id].Y = this.map.Nodes[idLast].Y + dy
          if (this.map.Nodes[id].Y > 5000 - 80) {
            this.map.Nodes[id].Y = 5000 - 80
          }
          list.push({
            ID: id,
            X: this.map.Nodes[id].X,
            Y: this.map.Nodes[id].Y,
          })
        }
        idLast = id
      }
      if (list.length > 0) {
        await this.$axios.post('/api/map/update', list)
      }
    },
    async circle() {
      const list = []
      if (!this.formatNodes || this.formatNodes.length < 2) {
        return
      }
      this.formatNodes.sort((a, b) => {
        return this.map.Nodes[a].X - this.map.Nodes[b].X
      })
      const c = 80 * this.formatNodes.length
      const r = Math.min(Math.trunc(c / 3.14 / 2), 1250 - 80)
      const cx = this.map.Nodes[this.formatNodes[0]].X + r
      let cy = this.map.Nodes[this.formatNodes[0]].Y
      if (cy - r < 0) {
        cy = 40 + r
      }
      for (let i = 0; i < this.formatNodes.length; i++) {
        const id = this.formatNodes[i]
        const d = 180 - i * (360 / this.formatNodes.length)
        const a = (d * Math.PI) / 180
        this.map.Nodes[id].X = Math.max(Math.trunc(r * Math.cos(a) + cx), 0)
        this.map.Nodes[id].Y = Math.max(Math.trunc(r * Math.sin(a) + cy), 0)
        list.push({
          ID: id,
          X: this.map.Nodes[id].X,
          Y: this.map.Nodes[id].Y,
        })
      }
      if (list.length > 0) {
        await this.$axios.post('/api/map/update', list)
      }
    },
  },
}
</script>

<style scoped>
/* コンテキストメニューの高さ調整 */
.v-list--dense .v-list-item .v-list-item__icon {
  height: 20px;
  margin-top: 4px;
  margin-bottom: 4px;
}
.v-list--dense .v-list-item {
  min-height: 30px;
}
</style>
