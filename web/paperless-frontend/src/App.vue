<template>

  <div id="app">

    <!-- Upload modal page -->
    <!-- https://www.npmjs.com/package/vue-js-modal -->
    <modal name="upload"
           @closed="modalClose"
           :adaptive="true"
           :min-width="300"
           :min-height="600"
           width="80%"
           height="90%">
      <div class="container-fluid pap-scrollable">
        <h2>Upload images</h2>
        <div class="form-inline">
          <div class="form-group">
            <label for="tags">Tags for the images</label>
            <input type="text" class="form-control" name="tags" v-model="upload.tags" />
          </div>
        </div>
        <div class="pap-dropbox">
          <form class="form-inline" enctype="multipart/form-data">
            <input type="file" name="image" multiple :disabled="isUploading" accept="image/*"
                   @change="doStartUpload($event.target.name, $event.target.files)"
                   class="pap-input-file" />
            <p v-if="!isUploading">
              Click or drag images here to upload
            </p>
            <p v-if="isUploading">
              Uploading and processing files...
            </p>
          </form>
        </div>
        <div class="alert alert-danger" v-for="err in upload.errors" :key="err">
          {{err}}
        </div>
        <ul style="margin-top: 10px">
          <li v-for="image in upload.images" :key="image">
            {{image}}
          </li>
        </ul>
      </div>
    </modal>

    <!-- Image information page -->
    <modal name="image-info"
           @closed="modalClose"
           :min-width="300"
           :min-height="600"
           width="80%"
           height="auto">
      <div class="container-fluid">
        <h3>Image information</h3>
        <div class="row">
          <div class="col-md-2">
            <div class="navbar navbar-default navbar-collapse" style="margin-top: 10px">
              <ul class="nav navbar-nav">
                <li>
                  <a target="_blank" :href="imgbase + info.image.CleanImg">
                    <img class="img-rounded pap-thumb" :src="imgbase + info.image.ThumbImg" />
                  </a>
                </li>
                <li><a target="_blank" :href="imgbase + info.image.OrigImg">Raw image</a></li>
                <li>
                  <a href="javascript:void(0)" @click="info.showLog = !info.showLog">
                    {{info.showLog ? "Show Text" : "Show Processing Log"}}
                  </a>
                </li>
                <li>
                  <a href="javascript:void(0)" @click="doDeleteImage()">
                    {{info.confirmDelete ? "Confirm delete image?" : "Delete image"}}
                  </a>
                </li>
              </ul>
            </div>
          </div>
          <div class="col-md-10">
            <span v-for="tag in info.image.Tags" :key="tag.Id" class="badge">{{tag.Name}}</span>
            <pre class="pre-scrollable pap-text">{{info.showLog ? info.image.ProcessLog : info.image.Text}}</pre>

            <!-- Debugging -->
            <!-- <pre class="pre-scrollable">{{info.image}} </pre> -->
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
              <input type="text" class="form-control input-lg" v-model="query" autofocus
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
      <div class="text-center">
        matches: {{matches}} Results per page: {{images.length}} <br>
        <button type="button" v-for="tag in tags" :key="tag.Id" class="badge" style="margin-right:2pt" @click="doTagSearch(tag.Name)">
          {{tag.Name}}
        </button>
        <div class="alert alert-danger" v-for="err in errors" :key="err">
          {{err}}
        </div>
      </div>
    </div>

    <!-- https://github.com/lokyoung/vuejs-paginate -->
    <div class="container panel">
      <div class="text-center">
        <paginate
          :page-count="paging.pages"
          :page-range="3"
          :margin-pages="2"
          :value="paging.current"
          :click-handler="doPaginate"
          ref="paginate"
          :prev-text="'Prev'"
          :next-text="'Next'"
          :container-class="'pagination'"
          :page-class="'page-item'">
        </paginate>
      </div>
    </div>

    <!-- Search results -->
    <div class="container pap-body">
      <div class="list-group">
        <a href="javascript:void(0)" @click.prevent.stop="doShowInfo(image)"
           v-for="image in images" :key="image.Id" class="list-group-item pap-item">
          <div class="media">
            <div class="media-left">
              <img class="media-object img-rounded" :src="imgbase + image.ThumbImg"
                   :alt="image.Filename" width="150" height="150" >
            </div>
            <div class="media-body">
              <!-- <h4 class="media-heading"> -->
                <span v-for="tag in image.Tags" :key="tag.Id" class="badge">{{tag.Name}}</span>
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

    <!-- https://github.com/lokyoung/vuejs-paginate -->
    <div class="container panel">
      <div class="text-center">
        <paginate
          :page-count="paging.pages"
          :page-range="3"
          :margin-pages="2"
          :value="paging.current"
          :click-handler="doPaginate"
          ref="paginate"
          :prev-text="'Prev'"
          :next-text="'Next'"
          :container-class="'pagination'"
          :page-class="'page-item'">
        </paginate>
      </div>
    </div>
  </div>
</template>


<script>

 import {ImageApi, TagApi} from './rest'

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
       tags: [],
       query: '',
       tag: '',
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

       // modal image information page
       info: {
         image: {},
         showLog: false,
         confirmDelete: false,
       },
     }
   },

   mounted () {
     this.applyURL()

     var vm = this;

     window.onpopstate = function() {
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
         return value.slice(0,300) + " [...]"
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
       this.applyURL()
     },

     // reset URL back to default
     resetURL: function() {
       this.switchURL('/')
     },

     // reset URL back to previous
     previousURL: function() {
       if (history.state && 'paperless_path' in history.state) {
         this.switchURL(history.state.paperless_path)
       } else {
         this.resetURL()
       }
     },

     // close the current modal pane
     modalClose: function() {
       this.previousURL()
     },

     // interpret URL when changing a view
     applyURL: function() {
       var url = new Url()

       /* Main UI with image searching */
       if (url.path === '/') {
         this.$modal.hide('image-info')
         this.$modal.hide('upload')

         this.query = 'q' in url.query ? url.query.q : '';

         var since = url.query.since
         var vm = this
         ImageApi.get('', {params: {
           q: url.query.q,
           t: url.query.t,
           since: url.query.since,
           count: url.query.count,
         }})
                 .then(function(response) {
                   vm.images = response.data.data.Images
                   vm.matches = response.data.data.ResultCount
                   vm.paging.starts = response.data.data.SinceIDs
                   vm.paging.pages = vm.paging.starts.length
                   vm.paging.perpage = response.data.data.Count
                   var page = -1
                   for (var i=0; i<vm.paging.starts.length; i++) {
                     if (since == vm.paging.starts[i]) {
                       page = i
                       break
                     }
                   }
                   if (page == -1) {
                     page = 0
                   }
                   vm.paging.current = page
                   vm.$refs.paginate.selected = page
                 })
                 .catch(function(e) {
                   vm.errors.push(e)
                 })
         TagApi.get('', {params: {}})
                 .then(function(response) {
                   vm.tags = response.data.data
                 })
                 .catch(function(e) {
                   vm.errors.push(e)
                 })
       } else if (url.path === '/info/') {
         this.info.showLog = false
         this.info.confirmDelete = false;
         if (this.info.image !== {}) {
           if (url.query.id === null) {
             this.resetURL()
             return
           }
           ImageApi.get('/'+ parseInt(url.query.id))
                   .then(response => {
                     this.info.image = response.data.data
                     this.$modal.show('image-info')
                   })
                   .catch(e => {
                     this.errors.push(e)
                   })
         } else {
           this.$modal.show('image-info')
         }
       } else if (url.path === '/upload') {
         this.upload.uploading = false;
         this.$modal.show('upload')
       } else {
         this.resetURL()
       }
     },

     doSearch: function() {
       this.switchURL('?q=' + encodeURIComponent(this.query))
       this.tag = ''
     },

     doTagSearch: function(tag) {
       this.switchURL('?t=' + encodeURIComponent(tag))
       this.tag = tag
     },

     doPaginate: function(page) {
       var url = 'since=' + encodeURIComponent(this.paging.starts[page - 1]) +
                 '&count=' + encodeURIComponent(this.paging.perpage);
       if(this.query && this.query != '') {
         url = 'q=' + this.query + '&' + url
       }
       if (this.tag && this.tag !== '') {
         url = 't=' + this.tag + '&' + url
       }

       this.switchURL('?' + url)
     },

     // open the info modal pane
     doShowInfo: function(image) {
       this.info.image = image
       this.switchURL('info/?id=' + encodeURIComponent(image.Id))
     },

     // open the upload modal pane
     doUpload: function() {
       this.switchURL('upload')
     },

     // start uploading files
     doStartUpload: function(name, files) {
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
                this.upload.images.push(response.config.filename + ': Uploaded successfully')
                this.upload.status = STATUS_SUCCESS;
              })
              .catch(e => {
                this.upload.images.push(e.config.filename + ': Error: ' + e.response.data.message)
                this.upload.status = STATUS_FAILED;
              })
       }
       this.upload.status = STATUS_UPLOADING;
     },

     // delete image. First call pops up a confirmation, second sends the DELETE query
     doDeleteImage: function() {
       if (this.info.confirmDelete) {
         ImageApi.delete('/'+ parseInt(this.info.image.Id))
                 .then(() => {
                   this.info.image = {}
                 })
                 .catch(e => {
                   this.errors.push(e)
                 })
         this.$modal.hide('image-info')
       }
       this.info.confirmDelete = !this.info.confirmDelete;
     }
   }
 }
</script>

<style src="bootstrap/dist/css/bootstrap.css" />

 <style>

 /* Paperless styles */
 a.list-group-item {
   height:auto;
   min-height:220px;
 }

 h4.list-group-item-heading {
   padding-top: 10px;
 }

 .pap-header {
   margin-top: 10px
 }

 .pap-search-input {
   width: 100% !important;
 }

 .pap-info {
   margin-top: 10px;
   margin-bottom: 0px;
 }

 .pap-body {
 }

 .pap-item {
   padding-top: 10px;
 }

 button.pap-search-btn {
   background-color: #42b983;
   color: white;
 }

 button.pap-search-btn:hover {
   color: #26a76d;
 }

 .pap-active {
   background-color: #42b983;
   color: #fff !important;
 }

 .pagination {
   margin-top: 5px;
   margin-bottom: 5px;
 }

 .pagination li a {
   color: #42b983;
 }

 .pagination li a:hover {
   color: #26a76d;
 }

 .pagination li a:focus {
   color: #26a76d;
 }

 .page-item {
 }

 .page-item.active a {
   background-color: #42b983;
   border-color: #42b983;
 }

 .pagination > .active > a:hover {
   background-color: #26a76d;
   border-color: #26a76d;
 }

 .pagination > .active > a:focus {
   background-color: #0d8e53;
   border-color: #0d8e53;
 }

 .pap-scrollable {
   height: 100%;
   overflow-y:scroll;
 }

 .pap-text {
   max-width: 100%;
   max-height: 400px;
   margin-top: 10px;
 }

 .pap-thumb {
   max-width: 100% !important;
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
