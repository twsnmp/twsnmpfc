<script>
  import {session} from './twsnmpapi.js';
	import Login from "./Login.svelte";
	import Main from "./Main.svelte";
  let page = "login";

	const handleLogin = () => {
		if ($session.token) {
			page = "main";
		}
  }

</script>

<main>
  <div id="page">
    {#if page == "login" }
     <Login on:done={handleLogin}/>
    {:else if page == "main"}
     <Main/>
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
    margin:10px auto;
  }
</style>
