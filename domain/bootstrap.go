package domain

import (
	"log"

	"github.com/rchargel/localiday/db"
)

// BootStrap bootstraps the application.
func BootStrap() error {
	err := initORM()

	err = initData()

	log.Printf("There %v users and %v active users in the system.", User{}.Count(), User{}.CountActive())
	return err
}

func initORM() error {
	db.DB.AddTableWithName(User{}, "users").SetKeys(true, "ID")
	db.DB.AddTableWithName(Role{}, "roles").SetKeys(true, "ID")
	db.DB.AddTableWithName(UserRole{}, "user_roles").SetKeys(true, "ID")

	return nil
}

func initData() error {
	var err error
	count := User{}.Count()
	if count == 0 {
		admin := CreateNewUser("admin", "admin", "", "admin", "admin@localiday.com")
		userRole := CreateAuthority("ROLE_USER")
		adminRole := CreateAuthority("ROLE_ADMIN")

		AddAuthorityToUser(admin, userRole)
		AddAuthorityToUser(admin, adminRole)

		log.Println("Created user " + admin.Username)
	}

	return err
}

func insert(obj interface{}) {
	checkError(db.DB.Insert(obj))
}

func checkError(err error) {
	if err != nil {
		log.Panicln("Could not perform operation", err)
	}
}

func count(script string) uint32 {
	var v uint32
	i, err := db.DB.SelectInt(script)

	if err != nil {
		log.Fatalln("Could not count items in table.", err)
	}
	v = uint32(i)
	return v
}