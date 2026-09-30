<template>
  <v-app>
    <v-main>
      <nuxt />
    </v-main>
  </v-app>
</template>

<script>
export default {
  name: 'DashboardLayout',
}
</script>

<style>
#public-map {
  width: 100%;
  height: calc(100vh - 64px);
  overflow: scroll;
}
</style>
