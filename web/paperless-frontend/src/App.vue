<template>

  <div id="app">

    <!-- Upload modal page -->
    <!-- https://www.npmjs.com/package/vue-js-modal -->
    <modal name="upload"
           @closed="modalClose"
           :width="600"
           :height="800">
      <div class="container-fluid pap-scrollable">
        <h2>Upload images</h2>
        <input type="text" class="form-control" name="tags" v-model="upload.tags" />
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
      </div>
    </modal>

    <!-- Image information page -->
    <modal name="image-info"
           @closed="modalClose"
           :width="900"
           :height="700">
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
                <li>
                  <a href="javascript:void(0)" @click="showLog = !showLog">
                    {{showLog ? "Show Text" : "Show Processing Log"}}
                  </a>
                </li>
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
    <div class="container pap-header">
      <form class="form-inline" v-on:submit.prevent>
        <div class="row">
          <div class="col col-xs-11">
            <div class="input-group pap-search-input">
              <input type="text" class="form-control input-lg" v-model="query"
                     @keyup.enter="doSearch()" placeholder="Paperless | Search documents ..." />
              <span class="input-group-btn" style="width: 1px">
                <button class="btn pap-search-btn btn-lg" type="button" @click="doSearch()">
                  <i class="glyphicon glyphicon-search"></i>
                </button>
              </span>
            </div>
          </div>
            <button class="btn pap-search-btn btn-lg" type="button" @click="doUpload()">
              +
            </button>
        </div>
      </form>
    </div>

    <!-- Information bar -->
    <div class="container pap-info panel">
      <div class="container">
        matches: {{matches}} Results per page: {{images.length}}
        <div class="alert alert-danger" v-for="err in errors">
          {{err}}
        </div>
      </div>

      <!-- https://github.com/lokyoung/vuejs-paginate -->
      <paginate
        :page-count="paging.pages"
        :page-range="3"
        :margin-pages="2"
        :force-page="paging.current"
        :click-handler="paginateHandler"
        :prev-text="'Prev'"
        :next-text="'Next'"
        :container-class="'pagination'"
        :page-class="'page-item'">
      </paginate>
    </div>

    <div class="container pap-body">
      <div class="list-group">
        <!-- active -->
        <a href="javascript:void(0)" @click.prevent.stop="showInfo(image)" v-for="image in images" class="list-group-item pap-item">
          <div class="media">
            <div class="media-left">
              <img class="media-object img-rounded" :src="imgbase + image.ThumbImg"
                   :alt="image.Filename" width="150" height="150" >
            </div>
            <div class="media-body">
              <!-- <h4 class="media-heading"> -->
                <span v-for="tag in image.Tags" class="badge">{{tag.Name}}</span>
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

 import Url from 'domurl'

 const STATUS_INITIAL = 0, STATUS_UPLOADING = 1, STATUS_SUCCESS = 2, STATUS_FAILED = 3;

 export default {
   name: 'app',
   data () {
     return {
       msg: 'Welcome to Your Vue.js App',
       imgbase: '',
       errors: [],
       images: [],
       query: '',
       matches: 0,

       paging: {
         starts: [],
         current: 0,
         pages: 0,
         perpage: 0,
       },

       // upload page
       upload: {
         visible: false,
         status: STATUS_INITIAL,
         uploading: false,
         images: [],
         errors: [],
         tags: "",
       },

       info: {
         visible: false,
       },

       // modal image information page
       imageInfo: { },
       showLog: false,
     }
   },

   mounted () {
     this.applyURL()

     var vm = this;

     window.onpopstate = function(event) {
       console.log("OnPopState funktiossa ja polku on " + window.location)
       console.log(event)
       vm.applyURL()
     }
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
     // Switch the browser url without refreshing
     switchURL: function(path) {
       var url = new Url()

       // Only remember previous url if it is the search view
       var resource = '/'
       if (url.path === '/' && url.query.toString() !== '') {
         resource += '?' + url.query
       }

       history.pushState({paperless_path: resource}, "paperless office", path)
       console.log("Changing to URL: " + window.location)
       this.applyURL()
     },

     resetURL: function() {
       this.switchURL('/')
     },

     previousURL: function() {
       console.log("State value when going back")
       console.log(history.state)
       if (history.state && 'paperless_path' in history.state) {
         this.switchURL(history.state.paperless_path)
       } else {
         this.resetURL()
       }
     },

     modalClose: function() {
       console.log("Closing a modal!");
       this.previousURL()
     },

     // interpret URL when changing a view
     applyURL: function() {
       // interpret URL when changing a view
       // - Load models from backend

       var url = new Url()
       console.log("JEJE!")
       console.log(url)
       console.log(this)
       /* console.log(this)
        */

       /* Main UI with image searching */
       if (url.path === '/') {
         this.$modal.hide('image-info')
         this.$modal.hide('upload')

         this.query = url.query.q

         var since = url.query.since
         var vm = this
         ImageApi.get('', {params: {
           q: url.query.q,
           since: url.query.since,
           count: url.query.count,
         }})
                 .then(function(response) {
                   console.log("RESPONSE")
                   console.log(response)
                   vm.images = response.data.data.Images
                   vm.matches = response.data.data.ResultCount
                   vm.paging.starts = response.data.data.SinceIDs
                   vm.paging.pages = vm.paging.starts.length
                   vm.paging.perpage = response.data.data.Count
                   for (var i=0; i<vm.paging.starts.length; i++) {
                     if (since == vm.paging.starts[i]) {
                       vm.paging.current = i
                       break
                     }
                   }
                 })
                 .catch(function(e) {
                   vm.errors.push(e)
                 })
       } else if (url.path === '/info/') {
         console.log("Päästiin modaaliseksi!")
         console.log(this.imageInfo)

         this.showLog = false
         if (this.imageInfo !== {}) {
           if (url.query.id === null) {
             this.resetURL()
             return
           }
           ImageApi.get('/'+ parseInt(url.query.id))
                   .then(response => {
                     console.log("Single image query")
                     console.log(response)
                     this.imageInfo = response.data.data
                     this.$modal.show('image-info')
                   })
                   .catch(e => {
                     this.errors.push(e)
                   })
         } else {
           console.log("Näytetään modaalinen ikkuna!!")
           console.log(this.imageInfo)
           this.$modal.show('image-info')
         }
       } else if (url.path === '/upload') {
         console.log("Upload päälle")
         console.log(this.$modal)
         this.upload.uploading = false;
         this.$modal.show('upload')
       } else {
         console.log("Unknown path. restarting to default")
         this.resetURL()
       }
     },

     doSearch: function() {
       console.log("QUERYING: " + this.query)
       this.switchURL('?q=' + encodeURIComponent(this.query))
     },

     paginateHandler: function(page) {
       console.log(page);
       this.switchURL('?since=' + encodeURIComponent(this.paging.starts[page - 1]) + '&count=' +
                      encodeURIComponent(this.paging.perpage))
     },

     showInfo: function(image) {
       /* this.imageInfo = image
        * this.showLog = false
        * this.$modal.show('image-info')
        */
       this.imageInfo = image

       console.log("Showing info on image: ")
       console.log(image)
       this.switchURL('info/?id=' + encodeURIComponent(image.Id))
     },

     /* Uploading functionality*/
     doUpload: function() {
       this.switchURL('upload')
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
         fdata.append('tags', this.upload.tags)
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

 /* .pagination:hover {
    color: #42b983 !important;
    }

    .page-item:hover a {
    color: #42b983;
    }

    .page-item.active a {
    background-color: #42b983;
    }

    .page-item.active a {
    background-color: #42b983;
    }
  */
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
