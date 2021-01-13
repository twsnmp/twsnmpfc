<template>
  <v-row>
    <v-col class="text-center">
      <logo />
      <blockquote class="blockquote">
        &#8220;ここにマップを表示する予定！！&#8221;
        <footer>
          <small>
            <em>&mdash;TWSNMP FC</em>
          </small>
        </footer>
      </blockquote>
    </v-col>
  </v-row>
</template>

<script>
import Logo from '~/components/Logo.vue'

export default {
  components: {
    Logo,
  },
}
</script>
