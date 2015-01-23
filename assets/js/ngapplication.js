var localidayApp = angular.module('localidayApp', [ 'ngRoute', 'uiGmapgoogle-maps' ]).
config(function($locationProvider, $routeProvider, $httpProvider, $interpolateProvider, uiGmapGoogleMapApiProvider) {
  $locationProvider.html5Mode({enabled: true, requireBase: false}).hashPrefix("!");
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
  $httpProvider.interceptors.push('AuthInterceptor');
  /*
  $httpProvider.responseInterceptors.push(['$injector', function() {
    return $injector.get('ResponseObserver');
  }]);
  */
  uiGmapGoogleMapApiProvider.configure({
    //key: 'AIzaSyC3BFO3MQdsGQJ1j5BHZqU47StUZcRb0U4',
    v: '3.17',
    libraries: 'geometry,visualization'
  });
  $routeProvider.
    when('/page/:pageName', {
      templateUrl: function (elem, attrs) {
        return '/templates/' + elem.pageName + ".html";
      }
    }).
    when('/', {
      templateUrl: '/templates/index.html'
    }).
    otherwise({
      redirectTo: '/'
    });
}).run(function($rootScope, $http, $location, AUTH_EVENTS) {
  $rootScope.$on('$routeChangeStart', function(event, next, current) {
    var cookie = getTokenCookie($location),
    path = $location.path();

    if (typeof($http.defaults.headers.common.Authorization) === 'undefined' && path !== '/home') {
      if (cookie && cookie.token && cookie.token !== 'null') {
        $http.defaults.headers.common.Authorization = cookie.tokenType + ' ' + cookie.token;
      }
    }
  });
  $rootScope.$on(AUTH_EVENTS.logoutSuccess, function(event) {
    delete $http.defaults.headers.common.Authorization;
    $location.path('/home');
  });
}).constant('AUTH_EVENTS',  {
  loginRequest     : 'auth-login-request',
  loginSuccess     : 'auth-login-success',
  loginFailed      : 'auth-login-failed',
  logoutSuccess    : 'auth-logout-success',
  sessionTimeout   : 'auth-session-timeout',
  notAuthenticated : 'auth-not-authenticated',
  notAuthorized    : 'auth-not-authorized',
  cancelLogin      : 'cancel-login-request'
}).constant('USER_ROLES', {
  all          : '*',
  admin        : 'ROLE_ADMIN',
  urser        : 'ROLE_USER',
  editableUser : 'ROLE_USER_EDITABLE'
}).factory('AuthInterceptor', function($rootScope, $q, AUTH_EVENTS) {
  return {
    responseError: function(response) {
      $rootScope.$broadcast({
        401: AUTH_EVENTS.notAuthenticated,
        403: AUTH_EVENTS.notAuthorized,
        419: AUTH_EVENTS.sessionTimeout,
        440: AUTH_EVENTS.sessionTimeout
      }[response.status], response);
      return $q.reject(response);
    }
  }
  /*
}).factory('ResponseObserver', function($rootScope, $q, $window, AUTH_EVENTS) {
  return function(promise) {
    return promise.then(function(successResponse) {
      return successResponse;
    }, function (errorResponse) {
      var event = {
        401: AUTH_EVENTS.notAuthenticated,
        403: AUTH_EVENTS.notAuthorized,
        419: AUTH_EVENTS.sessionTimeout,
        440: AUTH_EVENTS.sessionTimeout
      }[errorResponse.status];
      $rootScope.$broadcast(event, errorResponse);
      return $q.reject(errorResponse);
    });
  };
  */
});
