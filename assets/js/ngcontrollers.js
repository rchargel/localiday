localidayApp.controller('ApplicationController', function($scope, USER_ROLES, UserService, AuthService) {
  $scope.currentUser = null;
  $scope.userRoles = USER_ROLES;
  $scope.isAuthenticated = AuthService.isAuthenticated;
  $scope.isAuthorized = AuthService.isAuthorized;

  $scope.setCurrentUser = function(user) {
    $scope.currentUser = user;
  };

  AuthService.init(function(user) {
    $scope.setCurrentUser(user);
  });
}).controller('LoginController', function($scope, $rootScope, $location, AUTH_EVENTS, AuthService) {
  $scope.credentials = {
    username : '',
    password : ''
  }
  $scope.openLogin = function() {
    $rootScope.$broadcast(AUTH_EVENTS.loginRequest);
  };

  $scope.oauth = function(provider) {
    location.href = '/oauth/authenticate/' + provider;
  };

  $scope.login = function(credentials) {
    AuthService.login(credentials).then(function(user) {
      $rootScope.$broadcast(AUTH_EVENTS.loginSuccess);
      $scope.setCurrentUser(user);
      $location.path('/');
    }, function() {
      $rootScope.$broadcast(AUTH_EVENTS.loginFailed);
      $scope.setCurrentUser(null);
    });
  };

  $scope.cancelLogin = function() {
    $rootScope.$broadcast(AUTH_EVENTS.cancelLogin);
  };

  $scope.logout = function() {
    AuthService.logout().then(function() {
      $scope.setCurrentUser(null);
      $rootScope.$broadcast(AUTH_EVENTS.logoutSuccess);
    });
  };
});
