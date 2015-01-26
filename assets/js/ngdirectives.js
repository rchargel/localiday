localidayApp.directive('loginDialog', function(AUTH_EVENTS) {
  return {
    restrict: 'A',
    template: '<div ng-if="visible"><div class="login-background"></div><div class="login-screen" ng-include="\'/templates/auth.html\'"></div></div>',
    link: function(scope) {
      var showDialog = function() {
        scope.visible = true;
      };
      var hideDialog = function() {
        scope.visible = false;
      };
      scope.visible = false;
      scope.$on(AUTH_EVENTS.loginRequest, showDialog);
      scope.$on(AUTH_EVENTS.notAuthenticated, showDialog);
      scope.$on(AUTH_EVENTS.notAuthorized, showDialog);
      scope.$on(AUTH_EVENTS.sessionTimeout, showDialog)
      scope.$on(AUTH_EVENTS.loginSuccess, hideDialog);
      scope.$on(AUTH_EVENTS.cancelLogin, hideDialog);
    }
  };
});
