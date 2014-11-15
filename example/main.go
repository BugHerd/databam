package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/kr/pretty"
	_ "github.com/mattn/go-sqlite3"

	"github.com/BugHerd/databam"
)

func main() {
	db, err := sql.Open("sqlite3", os.Args[1])
	if err != nil {
		panic(err)
	}
	defer db.Close()

	d := databam.New(db)

	d.SetDebugLogger(func(format string, a ...interface{}) {
		fmt.Printf("DEBUG: "+format+"\n", a...)
	})

	tenants := d.MustRepository(Tenant{})

	where := Tenant{
		Creator: &Person{
			Id: "9f4ad422-24e2-4a3f-9ccb-c173550ec69a",
		},
		Memberships: []Membership{
			Membership{
				Type: "tenant",
				Person: &Person{
					Id: "9f4ad422-24e2-4a3f-9ccb-c173550ec69a",
				},
			},
		},
	}

	var t []Tenant
	if err := tenants.Find(&t, &where); err != nil {
		panic(err)
	}

	pretty.Println(t)
}
