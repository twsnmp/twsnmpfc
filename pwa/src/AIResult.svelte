<script>
  import * as echarts from "echarts";
  import { twsnmpApiGetJSON } from "./twsnmpapi.js";
  import { onMount } from "svelte";
  export let id;
  let chart;
  let title;
  const showAIHeatMap = (scores) => {
    if (chart) {
      chart.dispose();
    }
    chart = echarts.init(document.getElementById("chart"), "dark");
    const hours = [
      "0時",
      "1時",
      "2時",
      "3時",
      "4時",
      "5時",
      "6時",
      "7時",
      "8時",
      "9時",
      "10時",
      "11時",
      "12時",
      "13時",
      "14時",
      "15時",
      "16時",
      "17時",
      "18時",
      "19時",
      "20時",
      "21時",
      "22時",
      "23時",
    ];
    const option = {
      title: {
        show:true,
        text: title,
        textStyle: {
          fontSize: '14px',
        },
        left: 'center',
      },
      grid: {
        left: "10%",
        right: "5%",
        top: 30,
        buttom: 0,
      },
      toolbox: {
        feature: {
          dataZoom: {},
          saveAsImage: { name: "twsnmp_aiheatmap"},
        },
      },
      dataZoom: [{}],
      tooltip: {
        trigger: "item",
        formatter(params) {
          return (
            params.name +
            " " +
            params.data[1] +
            "時 : " +
            params.data[2].toFixed(1)
          );
        },
        axisPointer: {
          type: "shadow",
        },
      },
      xAxis: {
        type: "category",
        name: "日付",
        nameTextStyle: {
          fontSize: 10,
          margin: 2,
        },
        axisLabel: {
          fontSize: 10,
          margin: 2,
        },
        data: [],
      },
      yAxis: {
        type: "category",
        name: "時間帯",
        nameTextStyle: {
          fontSize: 10,
          margin: 2,
        },
        axisLabel: {
          fontSize: 10,
          margin: 2,
        },
        data: hours,
      },
      visualMap: {
        min: 40,
        max: 80,
        textStyle: {
          fontSize: 8,
        },
        calculable: true,
        realtime: false,
        inRange: {
          color: [
            "#313695",
            "#4575b4",
            "#74add1",
            "#abd9e9",
            "#e0f3f8",
            "#ffffbf",
            "#fee090",
            "#fdae61",
            "#f46d43",
            "#d73027",
            "#a50026",
          ],
        },
      },
      series: [
        {
          name: "Score",
          type: "heatmap",
          data: [],
          emphasis: {
            itemStyle: {
              borderWidth: 1,
            },
          },
          progressive: 1000,
          animation: false,
        },
      ],
    };
    if (!scores) {
      chart.setOption(option);
      chart.resize();
      return;
    }
    let nD = 0;
    let x = -1;
    scores.forEach((e) => {
      const t = new Date(e[0] * 1000);
      if (nD !== t.getDate()) {
        option.xAxis.data.push(echarts.time.format(t, "{yyyy}/{MM}/{dd}"));
        nD = t.getDate();
        x++;
      }
      option.series[0].data.push([x, t.getHours(), e[1]]);
    });
    chart.setOption(option);
    chart.resize();
  };

  onMount(async () => {
    const r = await twsnmpApiGetJSON("/api/report/ai/" + id);
    if (!r || !r.AIResult) {
      return;
    }
    title = r.NodeName + ":" + r.PollingName;
    showAIHeatMap(r.AIResult.ScoreData);
  });
</script>

<div id="chart" />

<style>
  #chart {
    width: 100%;
    height: 300px;
  }
</style>
