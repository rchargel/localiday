package domain

import (
	"github.com/rchargel/localiday/db"
	"github.com/rchargel/localiday/util"
	"github.com/robfig/cron"
)

// BootStrap bootstraps the application.
func BootStrap() error {
	err := initORM()
	err = initData()
	err = initCron()

	util.Log(util.Info, "There are %v users and %v active users in the system.", User{}.Count(), User{}.CountActive())
	return err
}

func initORM() error {
	db.DB.AddTableWithName(User{}, "users").SetKeys(true, "ID")
	db.DB.AddTableWithName(Role{}, "roles").SetKeys(true, "ID")
	db.DB.AddTableWithName(UserRole{}, "user_roles").SetKeys(true, "ID")
	db.DB.AddTableWithName(Session{}, "sessions").SetKeys(true, "ID")

	return nil
}

func initData() error {
	var err error
	count := User{}.Count()
	if count == 0 {
		admin := CreateNewUser("admin", "admin", "", "admin", "admin@localiday.com")
		userRole := CreateAuthority("USER")
		adminRole := CreateAuthority("ADMIN")
		systemUserRole := CreateAuthority("SYSTEM_USER")
		CreateAuthority("OPEN_AUTH_USER")
		CreateAuthority("GOOGLE_USER")
		CreateAuthority("FACEBOOK_USER")
		CreateAuthority("TWITTER_USER")

		AddAuthorityToUser(admin, userRole)
		AddAuthorityToUser(admin, adminRole)
		AddAuthorityToUser(admin, systemUserRole)

		util.Log(util.Debug, "Created user %v.", admin.Username)
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

func insert(obj interface{}) {
	checkError(db.DB.Insert(obj))
}

func checkError(err error) {
	if err != nil {
		util.Log(util.Fatal, "Could not perform operation.", err)
	}
}

func count(script string) uint32 {
	var v uint32
	i, err := db.DB.SelectInt(script)

	if err != nil {
		util.Log(util.Error, "Could not count items in table.", err)
	}
	v = uint32(i)
	return v
}
