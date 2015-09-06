var paperlessApp = angular.module('PaperlessApp',["ngResource"]);

paperlessApp.config(function($locationProvider) {
    $locationProvider.html5Mode(true)
})

paperlessApp.controller("PaperlessCtrl", ["$scope", "$location", "$resource", function($scope, $location, $resource) {
    var Image = $resource("/images/:id", {id: '@id'}, {})

    var showError = function(error) {
	$scope.error = error.data
    }

    $scope.list = function(searchparam) {
	Image.get({q: searchparam},function(data) {
	    $scope.images = data['Images'];
	    $scope.img_offset = data['Offset'];
	    $scope.img_count = data['Count'];
	    $scope.img_limit = data['Limit'];
	    $scope.error = null

	}, showError);
    };

    $scope.search = function() {
	$scope.images = null
	$scope.list($scope.searchstring)
	$location.search('q',$scope.searchstring)
    }

    $scope.ifTrue = function(cond, appendix) {
	return cond ? appendix : '';
    }

    $scope.toDate = function(dateint) {
	if(dateint <= 0) {
	    return "";
	}
	return new Date(dateint * 1000);
    }

    $scope.toggleProcessed = function(img) {
	img.showImg = !img.showImg;
	img.procURL = "static/" + img.Fileid + "processed.jpg";
    }

    // Initial load
    $scope.view = 'list';
    $scope.error = null;
    $scope.list();
    // $scope.error ="Testing error message"

    var args = $location.search();
    if(args['q'] != undefined) {
	$scope.searchstring = args['q'];
	$scope.search();
    }

}]);
