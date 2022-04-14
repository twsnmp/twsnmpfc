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
  let logColumns = [
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
  let timer;

  const logPagination = {
    limit: 10,
    enable: true,
  };

  const listPagination = {
    limit: 20,
    enable: true,
  };

  let page = "map";
  let nodes = [];
  let nodeColumns = [
    {
      name: "状態",
      width: "10%",
      formatter: (cell) => formatLevel(cell),
    },
    { name: "名前", width: "20%" },
    { name: "IPアドレス", width: "15%" },
    { name: "MACアドレス", width: "25%" },
    { name: "説明", width: "35%" },
  ];


  const refreshMAP = async () => {
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
  }

  const refreshNodes = async () => {
    const r = await twsnmpApiGetJSON("/api/nodes");
    if (!r) {
      errMsg = "ノードリストを取得できません！";
      return;
    }
    const tmp = [];
    r.forEach((n) => {
      tmp.push([n.State, n.Name, n.IP, n.MAC, n.Descr]);
    });
    nodes = tmp;
  }

  const refresh = () => {
    switch (page) {
    case "map":
      refreshMAP();
      return;
    case "node":
      refreshNodes();
      return;
    }
  };

  const loop = () => {
    refresh();
    timer = setTimeout(loop,60000);
  }

  onMount(() => {
    showMAP('map');
    loop();
  });

  const showPage = () => {
    if (page =="map") {
      showMAP("map");
    }
    refresh();
  }

  const logout = () => {
    $session.token = "";
    clearTimeout(timer);
    dispatch("done", {});
  };

</script>

<div class="Box">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{map.MapConf.MapName}</h3>
    <select
      class="form-select mr-1"
      bind:value={page}
      on:change={showPage}
    >
      <option value="map">マップ</option>
      <option value="node">ノードリスト</option>
    </select>
  </div>
  {#if errMsg}
    <div class="flash flash-error">
      <span class="mdi mdi-alert-circle" />
      {errMsg}
    </div>
  {/if}
  {#if page == "map"}
    <div class="Box-body">
      <div id="map" />
    </div>
    <div class="Box-row markdown-body log">
      <Grid data={logs} sort resizable search pagination={logPagination} columns={logColumns} language={jaJP} />
    </div>
  {:else if page == "node"}
    <div class="Box-row markdown-body log">
      <Grid data={nodes} sort resizable search pagination={listPagination} columns={nodeColumns} language={jaJP} />
    </div>
  {:else if page == "polling"}
    <div class="Box-body">
    </div>
  {:else if page == "sensor"}
    <div class="Box-body">
    </div>
  {:else if page == "ai"}
    <div class="Box-body">
    </div>
  {/if}
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
