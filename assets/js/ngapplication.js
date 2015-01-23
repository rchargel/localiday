var localidayApp = angular.module('localidayApp', [ 'ngRoute', 'uiGmapgoogle-maps' ]).
config(function($locationProvider, $routeProvider, $httpProvider, $interpolateProvider, uiGmapGoogleMapApiProvider) {
  $locationProvider.html5Mode({enabled: true, requireBase: false}).hashPrefix("!");
  $interpolateProvider.startSymbol('[[');
  $interpolateProvider.endSymbol(']]');
  uiGmapGoogleMapApiProvider.configure({
    //key: 'AIzaSyC3BFO3MQdsGQJ1j5BHZqU47StUZcRb0U4',
    v: '3.17',
    libraries: 'geometry,visualization'
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
});
