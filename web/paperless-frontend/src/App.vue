<template>

  <div id="app">

    <modal name="upload"
           :width="600"
           :height="700">
      upload-dialogi!!
    </modal>

    <modal v-if="imageInfo" name="image-info" :width="900" :height="700">
      <div class="container-fluid">
        <div class="row">
          <div class="col-md-2">
            <div class="navbar navbar-default navbar-collapse">
              <ul class="nav navbar-nav">
                <li>
                  <a :href="imgbase + imageInfo.CleanImg">
                    <img class="img-rounded" :src="imgbase + imageInfo.ThumbImg"/>
                  </a>
                </li>
                <li><a href="#">Raw image</a></li>
                <li><a href="#">Processing log</a></li>
              </ul>
            </div>
          </div>
          <div class="col-md-8">
            <pre class="pre-scrollable">{{imageInfo.Text}}</pre>
            <pre class="pre-scrollable">{{imageInfo}} </pre>
          </div>
        </div>
      </div>
    </modal>
    
    <!-- Search bar -->
    <form class="form-inline" @submit="doSearch">
      <div class="container pap-header">
        <div class="form-group">
          <div class="row">
            <div class="col col-md-12">
              <div class="input-group">
                <input type="text" class="form-control input-lg" v-model="query"
                       @keyup.enter="doSearch()" placeholder="Paperless | Search documents ..." />
                <span class="input-group-btn">
                  <button class="btn pap-search-btn btn-lg" type="button" @click="doSearch()">
                    <i class="glyphicon glyphicon-search"></i>
                  </button>
                </span>
              </div>
            </div>
            <div class="col col-md-2">
              <button class="btn pap-search-btn btn-lg" type="button" @click="doUpload()">
                +
              </button>
            </div>
          </div>
        </div>
      </div>
    </form>

    <!-- Information bar -->
    <div class="container pap-info panel">
        matches: {{matches}}
      <div class="alert alert-danger" v-for="err in errors">
        {{err}}
      </div>

    </div>

    <div class="container pap-body">
      <div class="list-group">
        <!-- active -->
        <a :href="'#' + image.Id" @click="showInfo(image)" v-for="image in images" class="list-group-item pap-item">
          <div class="media">
            <div class="media-left">
              <img class="media-object img-rounded" :src="imgbase + image.ThumbImg"
                   :alt="image.Filename" width="150" height="150" >
            </div>
            <div class="media-body">
              <!-- <h4 class="media-heading"> -->
                <span v-for="tag in image.Tags" class="badge">{{tag}}</span>
                <!-- dirps -->
                <!-- </h4> -->
                <p>
                  {{image.Text| truncate}}
                </p>
            </div>
          </div>
        </a>

        <a href="#" class="list-group-item pap-active pap-item">
          <div class="media">
            <div class="media-left">
              <img class="media-object img-rounded"  src="http://placehold.it/350x250" alt="" >
            </div>
            <div class="media-body">
              <h4 class="media-heading">
                <small><span class="tag is-small is-info">jep</span> jotain</small>
              </h4>
              qui diam libris ei, vidisse incorrupte at mel. his euismod salutandi dissentiunt eu. habeo offendit ea mea. nostro blandit sea ea, viris timeam molestiae an has. at nisl platonem eum. 
              vel et nonumy gubergren, ad has tota facilis probatus. ea legere legimus tibique cum, sale tantas vim ea, eu vivendo expetendis vim. voluptua vituperatoribus et mel, ius no elitr deserunt mediocrem. mea facilisi torquatos ad.
            </div>
          </div>
        </a>

      </div>
    </div>
      {{images}}


    <div id="example">
      <img src="./assets/logo.png" class="">
      <h1>{{msg}}</h1>
      <h2>Essential Links</h2>
      <ul>
	<li><a href="https://vuejs.org" target="_blank">Core Docs</a></li>
	<li><a href="https://forum.vuejs.org" target="_blank">Forum</a></li>
	<li><a href="https://gitter.im/vuejs/vue" target="_blank">Gitter Chat</a></li>
	<li><a href="https://twitter.com/vuejs" target="_blank">Twitter</a></li>
      </ul>
      <h2>Ecosystem</h2>
      <ul>
	<li><a href="http://router.vuejs.org/" target="_blank">vue-router</a></li>
	<li><a href="http://vuex.vuejs.org/" target="_blank">vuex</a></li>
	<li><a href="http://vue-loader.vuejs.org/" target="_blank">vue-loader</a></li>
	<li><a href="https://github.com/vuejs/awesome-vue" target="_blank">awesome-vue</a></li>
      </ul>
    </div>
  </div>
</template>


<script>

 import {ImageApi} from './rest'

 export default {
   name: 'app',
   data () {
     return {
       msg: 'Welcome to Your Vue.js App',
       imgbase: 'http://localhost:8078',
       errors: [],
       images: [],
       query: '',
       matches: 0,
       imageInfo: null
     }
   },

   created () {
       ImageApi.get('').then(response => {
         this.images = response.data.data;
         this.matches = this.images.length;
       }).catch( e => {
         console.log(e);
         this.errors.push(e)
       })
   },

   filters: {
     truncate: function(value) {
       value = value.toString()
       if (value.length > 300) {
         return value.slice(1,300) + " [...]"
       }
       return value
     }
   },

   methods: {

     doSearch: function() {
       console.log("query on")
       console.log(this.query)

     },
     doUpload: function() {
       this.$modal.show('upload')
     },
     showInfo: function(image) {
       this.imageInfo = image
       this.$modal.show('image-info')
     }
   }
 }
</script>

<style src="bootstrap/dist/css/bootstrap.css" />

 <style>

 #app {
   margin-top: 20px;
   margin-left: 10px;
   margin-right: 10px;
 }

 #example {
   font-family: 'Avenir', Helvetica, Arial, sans-serif;
   -webkit-font-smoothing: antialiased;
   -moz-osx-font-smoothing: grayscale;
   text-align: center;
   color: #2c3e50;
   margin-top: 30px;
 }

 h1, h2 {
   font-weight: normal;
 }

 ul {
   list-style-type: none;
   padding: 0;
 }

 li {
   display: inline-block;
   margin: 0 10px;
 }

 a {
   color: #42b983;
 }

 /* Paperless styles */
 a.list-group-item {
   height:auto;
   min-height:220px;
 }

 h4.list-group-item-heading {
   padding-top: 10px;
 }

.pap-info {
  margin-top: 10px;
}

.pap-body {
  padding-top: 10px;
}

.pap-item {
  padding-top: 10px;
  /* margin-top: 20px; */
 }

 button.pap-search-btn {
   background-color: #42b983;
   color: white;
 }

.pap-active {
  background-color: #42b983;
  color: #fff !important;
 }

 /* Modal handling */

</style>
