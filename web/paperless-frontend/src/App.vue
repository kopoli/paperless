<template>

  <div id="app">

    <modal name="upload"
           :width="600"
           :height="800">
      <div class="container-fluid pap-scrollable">
        <h2>Upload images</h2>
        <div class="pap-dropbox">
          <form class="form-inline" enctype="multipart/form-data">
            <input type="file" name="image" multiple :disabled="isUploading" accept="image/*"
                   @change="startUpload($event.target.name, $event.target.files)"
                   class="pap-input-file" />
            <p v-if="!isUploading">
              Click or drag images here to upload
            </p>
            <p v-if="isUploading">
              Uploading and processing files...
            </p>
          </form>
        </div>
        <div class="alert alert-danger" v-for="err in upload.errors">
          {{err}}
        </div>
        <ul>
          <li v-for="image in upload.images">
            {{image}}
          </li>
        </ul>
        <!-- {{upload.images}} -->
      </div>
    </modal>

    <modal v-if="imageInfo" name="image-info" :width="900" :height="700">
      <div class="container-fluid">
        <div class="row">
          <div class="col-md-2">
            <div class="navbar navbar-default navbar-collapse">
              <ul class="nav navbar-nav">
                <li>
                  <a target="_blank" :href="imgbase + imageInfo.CleanImg">
                    <img class="img-rounded" :src="imgbase + imageInfo.ThumbImg"/>
                  </a>
                </li>
                <li><a target="_blank" :href="imgbase + imageInfo.OrigImg">Raw image</a></li>
                <li><a href="#" @click="showLog = !showLog">{{showLog ? "Show Text" : "Show Processing Log"}}</a></li>
              </ul>
            </div>
          </div>
          <div class="col-md-8 pap-scrollable">
            <pre class="pre-scrollable pap-text">{{showLog ? imageInfo.ProcessLog : imageInfo.Text}}</pre>

            <!-- Debugging -->
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
                <input type="text" class="form-control input-lg pap-search-input" v-model="query"
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
      </div>
    </div>
  </div>
        <!-- <a href="#" class="list-group-item pap-active pap-item">
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
             </div> -->

        <!-- {{images}}
           -->

        <!-- <div id="example">
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
           -->
</template>


<script>

 import {ImageApi} from './rest'

 function getImagesOK(obj, response) {
   return function(response) {
     obj.images = response.data.data;
     obj.matches = obj.images.length;
   }
 }

 function getImagesFail(obj, e) {
   return function(e) {
     console.log(e);
     obj.errors.push(e)
   }
 }

 const STATUS_INITIAL = 0, STATUS_UPLOADING = 1, STATUS_SUCCESS = 2, STATUS_FAILED = 3;

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

       // upload page
       upload: {
         status: STATUS_INITIAL,
         uploading: false,
         images: [],
         errors: [],
       },

       // modal image information page
       imageInfo: null,
       showLog: false,
     }
   },

   created () {
     ImageApi.get('', {params: {q: this.query}})
             .then(getImagesOK(this))
             .catch(getImagesFail(this))
   },

   computed: {
     isInitial() {
       return this.upload.status === STATUS_INITIAL;
     },
     isUploading() {
       return this.upload.status === STATUS_UPLOADING;
     },
     isSuccess() {
       return this.upload.status === STATUS_SUCCESS;
     },
     isFailed() {
       return this.upload.status === STATUS_FAILED;
     }
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
       console.log("query on talla")
       console.log(this.query)
       ImageApi.get('', {params: {q: this.query}})
              .then(getImagesOK(this))
              .catch(getImagesFail(this))
     },

     doUpload: function() {
       this.upload.uploading = false;
       this.$modal.show('upload')
     },
     startUpload: function(name, files) {
       console.log(name);
       console.log(files);

       if (!files.length) {
         return
       }


       for (var i = 0; i< files.length; i++) {
         const fdata = new FormData();
         fdata.append(name, files[i], files[i].name)
         ImageApi.post('', fdata, {
           filename: files[i].name
         })
              .then(response => {
                /* this.upload.images.push(response)
                 * this.upload.images.push(response.data)*/
                this.upload.images.push(response.config.filename + ': Uploaded successfully')
                this.upload.status = STATUS_SUCCESS;
              })
              .catch(e => {
                this.upload.images.push(e.config.filename + ': Error: ' + e.response.data.message)
                /* this.upload.errors.push(e)
                 * this.upload.errors.push(e.response.data.message)*/
                this.upload.status = STATUS_FAILED;
              })
       }
       this.upload.status = STATUS_UPLOADING;
     },

     showInfo: function(image) {
       this.imageInfo = image
       this.showLog = false
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

 .pap-search-input {
   width: 100% !important;
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

 .pap-scrollable {
   height:450px;
   overflow-y:scroll;
 }

 .pap-text {
   height: 450px;
 }

 /* file upload styles */
 .pap-dropbox {
   outline: 2px dashed grey;
   outline-offset: -10px;
   background: white;
   color: dimgray;
   padding: 10px 10px;
   min-height: 300px;
   position: relative;
   /* cursor: pointer; */
 }

 .pap-dropbox:hover {
   background: #42b983;
   color: white;
 }

 .pap-input-file {
   opacity: 0;
   width: 100%;
   height: 300px;
   position: absolute;
   /* cursor: pointer; */
 }
 .pap-dropbox p {
   font-size: 1.2em;
   text-align: center;
   padding: 50px 0;
 }

</style>
