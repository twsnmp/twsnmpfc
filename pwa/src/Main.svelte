<script>
  import * as echarts from "echarts";
  import { html, h } from "gridjs";
  import "./gridjs.css";
  import {getState} from "./common.js"
  import {setMAP, showMAP} from "./map"
  import { session, twsnmpApiGetJSON } from "./twsnmpapi.js";
  import { createEventDispatcher, onMount } from "svelte";
  import Grid from "gridjs-svelte";
  import jaJP from "./gridjsJaJP";
  const dispatch = createEventDispatcher();

  let map = {
    MapConf: {
      MapName: "",
    },
  };

  const formatLevel = (level) => {
    const e = getState(level);
    return html(`<div><span class="mdi ${e.icon}" style="color: ${e.color};"></span>${e.text}</div>`);
  };

  let logs = [];
  let columns = [
    {
      name: "状態",
      width: "8%",
      formatter: (cell) => formatLevel(cell),
    },
    {
      name: "発生日時",
      width: "12%",
      formatter: (cell) =>
        echarts.time.format(
          new Date(cell / (1000 * 1000)),
          "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
        ),
    },
    { name: "種別", width: "10%" },
    { name: "関連ノード", width: "15%" },
    { name: "イベント", width: "55%" },
  ];
  let errMsg = "";
  const logout = () => {
    $session.token = "";
    dispatch("done", {});
  };
  let pagination = {
    limit: 10,
    enable: true,
  };
  const refresh = async () => {
    const r = await twsnmpApiGetJSON("/api/map");
    if (!r) {
      errMsg = "マップ情報を取得できません！";
      return;
    }
    map = r;
    const tmp = [];
    map.Logs.forEach((l) => {
      tmp.push([l.Level, l.Time, l.Type, l.NodeName, l.Event]);
    });
    logs = tmp;
    setMAP(map);
  };
  onMount(() => {
    showMAP('map');
    refresh();
  });
</script>

<div class="Box">
  <div class="Box-header">
    <h3 class="Box-title">
      {map.MapConf.MapName}
    </h3>
  </div>
  {#if errMsg}
    <div class="flash flash-error">
      <span class="mdi mdi-alert-circle" />
      {errMsg}
    </div>
  {/if}
  <div class="Box-body">
    <div id="map" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid data={logs} sort resizable search {pagination} {columns} language={jaJP} />
  </div>
  <div class="Box-footer text-right">
    <button class="btn btn-danger" type="button" on:click={logout}>
      <span class="mdi mdi-logout" />
      ログアウト
    </button>
    <button class="btn" type="button" on:click={refresh}>
      <span class="mdi mdi-refresh" />
      更新
    </button>
  </div>
</div>

<style>
  #map {
    width: 100%;
    height: 600px;
    overflow: scroll;
  }
</style>
