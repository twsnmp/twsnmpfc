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

  let errMsg = "";
  let timer;
  let page = "map";

  const pagination = {
    limit: 10,
    enable: true,
  };

  const formatLevel = (level) => {
    const e = getState(level);
    if(!e || !e.text){
      console.log(level);
    }
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

  let pollings = [];
  let pollingColumns = [
    {
      name: "状態",
      width: "10%",
      formatter: (cell) => formatLevel(cell),
    },
    { name: "ノード名", width: "20%" },
    { name: "名前", width: "35%" },
    {
      name: "レベル",
      width: "10%",
      formatter: (cell) => formatLevel(cell),
    },
    { name: "種別", width: "10%" },
    {
      name: "最終実施",
      width: "15%",
      formatter: (cell) =>
        echarts.time.format(
          new Date(cell / (1000 * 1000)),
          "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
        ),
    },
  ];

  const refreshPollings = async () => {
    const r = await twsnmpApiGetJSON("/api/pollings");
    if (!r) {
      errMsg = "ポーリングリストを取得できません！";
      return;
    }
    const nodeNameMap = {};
    r.NodeList.forEach((n)=>{
      nodeNameMap[n.value] = n.text;
    });
    const tmp = [];
    r.Pollings.forEach((p) => {
      tmp.push([p.State,nodeNameMap[p.NodeID] || "", p.Name, p.Level, p.Type,p.LastTime]);
    });
    pollings = tmp;
  }

  const refresh = () => {
    switch (page) {
    case "map":
      refreshMAP();
      return;
    case "node":
      refreshNodes();
      return;
    case "polling":
      refreshPollings();
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
      pagination.limit = 10;
      showMAP("map");
    } else {
      pagination.limit = 20;
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
      <option value="polling">ポーリングリスト</option>
      <option value="sensor">センサーリスト</option>
      <option value="ai">AI分析リスト</option>
      <option value="device">LANデバイスリスト</option>
      <option value="ip">IPアドレスリスト</option>
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
      <Grid data={logs} sort resizable search {pagination} columns={logColumns} language={jaJP} />
    </div>
  {:else if page == "node"}
    <div class="Box-row markdown-body log">
      <Grid data={nodes} sort resizable search {pagination} columns={nodeColumns} language={jaJP} />
    </div>
  {:else if page == "polling"}
    <div class="Box-row markdown-body log">
      <Grid data={pollings} sort resizable search {pagination} columns={pollingColumns} language={jaJP} />
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
