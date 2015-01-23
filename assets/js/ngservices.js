localidayApp.service('Session', function($http) {
  this.create = function(sessionId, tokenType, userId, userRoles) {
    this.id = sessionId;
    this.tokenType = tokenType;
    this.userId = userId;
    this.userRoles = userRoles;
    $http.defaults.headers.common.Authorization = this.tokenType + ' ' + this.id;
  };
  this.destroy = function() {
    this.id = null;
    this.tokenType = null;
    this.userId = null;
    this.userRoles = null;
    $http.defaults.headers.common.Authorization = null;
    delete $http.defaults.headers.common.Authorization;
  };
  this.getHttpConfig = function(token, type) {
    return {
      headers : {
        'Authorization' : (type || this.tokenType) + ' ' + (token || this.id)
      }
    };
  };
  return this;
}).factory('GeolocationService', function(uiGmapGoogleMapApi, $http, $q) {
  var geolocationService = { };

  geolocationService.lookupStates = function() {
    var url = 'http://zilch.zcarioca.net/countries.js?callback=JSON_CALLBACK';
    var deferred = $q.defer();

    $http.jsonp(url).success(function(countries) {
      var array = [];
      array.push({code : '', name: '---'});
      for (var i = 0; i < countries.length; i++) {
        if (countries[i].CountryName === 'United States') {
          for (var j = 0; j < countries[i].States.length; j++) {
            array.push({
              code: countries[i].States[j].State,
              name: countries[i].States[j].StateName
            });
          }
        }
      }
      deferred.resolve(array);
    });

    return deferred.promise;
  };

  geolocationService.lookupAddress = function(address) {
    var addressString = (address.address1 ? address.address1 : '') +
    (address.city ? ', ' + address.city : ',') +
    (address.state ? ' ' + address.state : '') +
    (address.zip ? ' ' + address.zip : '');

    var geocoder = new google.maps.Geocoder();
    var deferred = $q.defer();

    var getAddrPart = function(parts, type) {
      for (var i = 0; i < parts.length; i++) {
        for (var j = 0; j < parts[i].types.length; j++) {
          if (parts[i].types[j] === type) {
            return parts[i].short_name;
          }
        }
      }
    };

    geocoder.geocode({'address':addressString}, function(results, status) {
      if (results.length === 1 && status == google.maps.GeocoderStatus.OK) {
        uiGmapGoogleMapApi.then(function() {
          var coords = {
            latitude : results[0].geometry.location.k,
            longitude : results[0].geometry.location.C
          };
          var streetName = results[0].formatted_address.substring(0, results[0].formatted_address.indexOf(','));
          var zip = getAddrPart(results[0].address_components, 'postal_code');
          var state = getAddrPart(results[0].address_components, 'administrative_area_level_1');
          var city = getAddrPart(results[0].address_components, 'locality')
          var result = {
            result : {
              'street' : streetName,
              'city' : city,
              'state' : state,
              'zip' : zip
            },
            location: coords
          }
          deferred.resolve(result);
        });
      }
    });
    return deferred.promise;
  };

  return geolocationService;
}).factory('UserService', function($http, Session) {
  var userService = {
    endpoint : '/user'
  };

  userService.loadUserInfo = function() {
    return $http.post(userService.endpoint + '/loadInfo', {'username' : Session.userId }).then(function(response) {
      return response.data;
    });
  };

  return userService;
}).factory('AuthService', function($q, $http, $location, Session) {
  var authService = {
    loginUrl : '/api/login',
    logoutUrl : '/api/logout',
    validateUrl : '/api/validate'
  };

  authService.init = function(callback, errorCallback) {
    authService.validate().then(callback, errorCallback);
  };

  authService.validate = function() {
    var deferred = $q.defer();

    (function() {
      var token = getTokenCookie($location);
      if (token.token && token.token !== null && token.token !== 'null') {
        $http.post(authService.validateUrl, null, Session.getHttpConfig(token.token, token.tokenType)).success(function(user) {
          Session.create(user.access_token, user.token_type, user.username, user.roles);
          createTokenCookie(user.access_token, user.token_type);
          deferred.resolve(user);
        }).error(function(data, status) {
          removeTokenCookie();
          deferred.reject('Token rejected with status: ' + status);
        });
      } else {
        deferred.reject('No token found');
      }
    }());

    return deferred.promise;
  };

  authService.login = function(credentials) {
    return $http.post(authService.loginUrl, credentials).then(function(res) {
      var user = res.data;
      Session.create(user.access_token, user.token_type, user.username, user.roles);
      createTokenCookie(user.access_token, user.token_type);
      return user;
    });
  };

  authService.logout = function() {
    return $http.post(authService.logoutUrl, null).then(function(res) {
      Session.destroy();
      removeTokenCookie();
      return null;
    });
  };

  authService.isAuthenticated = function() {
    return !!Session.userId;
  };

  authService.isNotAuthorized = function(authorizedRoles) {
    return !authService.isAuthorized(authorizedRoles);
  };

  authService.isAuthorized = function(authorizedRoles) {
    if (!angular.isArray(authorizedRoles)) {
      authorizedRoles = [authorizedRoles];
    }
    if (authService.isAuthenticated()) {
      for (var i = 0; i < Session.userRoles.length; i++) {
        if (authorizedRoles.indexOf(Session.userRoles[i]) !== -1) {
          return true;
        }
      }
    }
    return false;
  };

  return authService;
});
