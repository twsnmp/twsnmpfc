<script>
  import { session } from "./twsnmpapi.js";
  import Login from "./Login.svelte";
  import Map from "./Map.svelte";
  let page = "login";

  const handleDone = () => {
    page = $session.token ? "map" : "login";
  };
</script>

<main>
  <div id="page">
    {#if page == "login"}
      <Login on:done={handleDone} />
    {:else if page == "map"}
      <Map on:done={handleDone} />
    {/if}
  </div>
</main>

<style>
  main {
    height: 100%;
    width: 100%;
  }
  #page {
    height: 100%;
    width: 98%;
    margin: 10px auto;
  }
</style>
