<script>
  import {login,session} from './twsnmpapi.js';
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  let errMsg = "";
  let user = '';
  let password = '';
  const auth = async () => {
    const r = await login(user,password);
    if(r) {
      $session.token = r.token;
      dispatch('done', {});
      return
    }
    errMsg = "ユーザーIDまたは、パスワードが違います";
  }
</script>

<div class="Box" id="login">
  <div class="Box-header">
    <h3 class="Box-title">
      ログイン
    </h3>
  </div>
  {#if errMsg}
    <div class="flash flash-error">
      <span class="mdi mdi-alert-circle"></span>
      {errMsg}
    </div>
  {/if}
  <div class="Box-body">
    <form>
      <input class="form-control input-block mt-2" type="text" placeholder="ユーザーID" bind:value="{user}"/>
      <input class="form-control input-block mt-2" type="password" placeholder="パスワード"bind:value="{password}"/>
    </form>  
  </div>
  <div class="Box-footer text-right">
    <button class="btn btn-primary" type="button" on:click={auth}>
      <span class="mdi mdi-check"></span>
      ログイン
    </button>
  </div>
</div>

<style>
  #login {
    margin: auto;
    margin-top: 50px;
    width: 50%;
    min-width: 800px;
  }
</style>