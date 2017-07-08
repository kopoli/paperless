<template>

  <div id="app">

    <div class="container pap-header">
      <div class="row">
        <div class="col-md-12">
          <div id="custom-search-input">
            <div class="input-group col-md-12">
              <input type="text" class="form-control input-lg" placeholder="Search Paperless Office" />
              <span class="input-group-btn">
                <button class="btn btn-info btn-lg" type="button">
                  <i class="glyphicon glyphicon-search"></i>
                </button>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="container pap-body">
      <div class="list-group">
        <!-- active -->
        <a href="#" class="list-group-item pap-item">
          <div class="media">
            <div class="media-left">
              <img class="media-object img-rounded"  src="http://placehold.it/350x250" alt="" >
            </div>
            <div class="media-body">
              <h4 class="media-heading"> dirps </h4>
              qui diam libris ei, vidisse incorrupte at mel. his euismod salutandi dissentiunt eu. habeo offendit ea mea. nostro blandit sea ea, viris timeam molestiae an has. at nisl platonem eum. 
              vel et nonumy gubergren, ad has tota facilis probatus. ea legere legimus tibique cum, sale tantas vim ea, eu vivendo expetendis vim. voluptua vituperatoribus et mel, ius no elitr deserunt mediocrem. mea facilisi torquatos ad.
            </div>
          </div>
        </a>

        <a href="#" class="list-group-item active pap-item">
          <div class="media">
            <div class="media-left">
              <img class="media-object img-rounded"  src="http://placehold.it/350x250" alt="" >
            </div>
            <div class="media-body">
              <h4 class="media-heading"> dirps </h4>
              qui diam libris ei, vidisse incorrupte at mel. his euismod salutandi dissentiunt eu. habeo offendit ea mea. nostro blandit sea ea, viris timeam molestiae an has. at nisl platonem eum. 
              vel et nonumy gubergren, ad has tota facilis probatus. ea legere legimus tibique cum, sale tantas vim ea, eu vivendo expetendis vim. voluptua vituperatoribus et mel, ius no elitr deserunt mediocrem. mea facilisi torquatos ad.
            </div>
          </div>
        </a>

      </div>
    </div>
    <!-- <div class="container bg-faded my-1 w-100 py-2">
         <form class="form-inline">
         <div class="row">
         <div class="col-3">
         Paperless
         </div>
         <div class="col-9">
         <!-- <div class="input-group"> -->
    <!-- <input class="form-control form-control-success" style="width:100%" type="text" placeholder="Search"> -->
    <!-- </div> -->
    <!-- </div> -->
    <!-- <div class="col-3">
         <button class="btn btn-outline-success mr-2 btn-block" type="submit">Search</button>
         </div> -->
    <!-- </div>
         </form> -->
    <!-- </div>
       -->
    <!-- <nav class="navbar navbar-toggleable-sm sticky-top navbar-light bg-faded">
         <a class="navbar-brand my-1" href="/">Paperless</a>

         <form class="form-inline my-2">
         <input class="form-control mr-2 w-50" type="text" placeholder="Search">
         <button class="btn btn-outline-success mr-2" type="submit">Search</button>
         </form>

         </nav>
       -->

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
       msg: 'Welcome to Your Vue.js App'
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

 a.list-group-item {
   height:auto;
   min-height:220px;
 }

 h4.list-group-item-heading {
   padding-top: 10px;
 }

.pap-body {
  padding-top: 10px;
}

.pap-item {
  padding-top: 10px;
  /* margin-top: 20px; */
 }

</style>
