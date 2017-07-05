var paperlessApp = angular.module('PaperlessApp',["ngResource"]);

paperlessApp.config(function($locationProvider) {
    $locationProvider.html5Mode(true);
});

paperlessApp.controller("PaperlessCtrl", ["$scope", "$location", "$resource", function($scope, $location, $resource) {
    var Image = $resource("/api/v1/image/:id", {id: '@id'}, {});

    var showError = function(error) {
	$scope.error = error.data.error;
    };

    $scope.list = function(searchparam) {
	Image.get({q: searchparam},function(data) {
          console.log(data);
          $scope.images = data.data;
            // $scope.img_offset = data['Offset'];
	    // // $scope.img_count = data['Count'];
	    // $scope.img_count = data.length;
	    // $scope.img_limit = data['Limit'];
          if (data.status != 'success') {
	    $scope.error = data.status;
          }
	}, showError);
    };

    $scope.search = function() {
	$scope.images = null;
	$scope.list($scope.searchstring);
	$location.search('q',$scope.searchstring);
    };

    $scope.ifTrue = function(cond, appendix) {
	return cond ? appendix : '';
    };

    $scope.toDate = function(dateint) {
	if(dateint <= 0) {
	    return "";
	}
	return new Date(dateint * 1000);
    };

    $scope.toggleProcessed = function(img) {
	img.showImg = !img.showImg;
    };

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
