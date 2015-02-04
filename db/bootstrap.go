package db

import (
	"github.com/rchargel/localiday/app"
	"github.com/robfig/cron"
)

// Defines the set of roles available to the application.
const (
	RoleUser         = "USER"
	RoleAdmin        = "ADMIN"
	RoleSystemUser   = "SYSTEM_USER"
	RoleOAuthUser    = "OPEN_AUTH_USER"
	RoleGoogleUser   = "GOOGLE_USER"
	RoleFacebookUser = "FACEBOOK_USER"
	RoleTwitterUser  = "TWITTER_USER"
)

// BootStrap bootstraps the application.
func BootStrap() error {
	err := initORM()
	err = initData()
	err = initCron()

	app.Log(app.Info, "There are %v users and %v active users in the system.", User{}.Count(), User{}.CountActive())
	return err
}

func initORM() error {
	DB.AddTableWithName(User{}, "users").SetKeys(true, "ID")
	DB.AddTableWithName(Role{}, "roles").SetKeys(true, "ID")
	DB.AddTableWithName(UserRole{}, "user_roles").SetKeys(true, "ID")
	DB.AddTableWithName(Session{}, "sessions").SetKeys(true, "ID")

	return nil
}

func initData() error {
	var err error
	count := User{}.Count()
	if count == 0 {
		admin, _ := CreateNewUser("admin", "admin", "", "admin", "admin@localiday.com")
		userRole := CreateAuthority(RoleUser)
		adminRole := CreateAuthority(RoleAdmin)
		systemUserRole := CreateAuthority(RoleSystemUser)
		CreateAuthority(RoleOAuthUser)
		CreateAuthority(RoleGoogleUser)
		CreateAuthority(RoleFacebookUser)
		CreateAuthority(RoleTwitterUser)

		AddAuthorityToUser(admin, userRole)
		AddAuthorityToUser(admin, adminRole)
		AddAuthorityToUser(admin, systemUserRole)

		app.Log(app.Debug, "Created user %v.", admin.Username)
	}

	return err
}

func initCron() error {
	var err error
	c := cron.New()

	err = c.AddFunc("0 */5 * * * *", func() { CleanSessions() })
	c.Start()

	return err
}

func insert(obj interface{}) error {
	return DB.Insert(obj)
}

func count(script string) uint32 {
	var v uint32
	i, err := DB.SelectInt(script)

	if err != nil {
		app.Log(app.Error, "Could not count items in table.", err)
	}
	v = uint32(i)
	return v
}
