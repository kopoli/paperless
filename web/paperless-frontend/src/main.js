import Vue from 'vue'
import App from './App.vue'

import vmodal from 'vue-js-modal'

import Paginate from 'vuejs-paginate'
Vue.component('paginate', Paginate)

Vue.use(vmodal)

new Vue({
  el: '#app',
  render: h => h(App)
})
