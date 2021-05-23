import Vue from 'vue'
import JsonExcel from 'vue-json-excel'
import VueClipboard from 'vue-clipboard2'

Vue.component('downloadExcel', JsonExcel)
Vue.use(VueClipboard)
