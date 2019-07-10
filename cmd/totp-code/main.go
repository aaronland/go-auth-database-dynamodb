package main

import (
	"flag"
	"fmt"
	"github.com/aaronland/go-auth-database-dynamodb"
	"log"
)

func main() {

	addr := flag.String("email", "", "...")
	accounts_dsn := flag.String("accounts-dsn", "", "...")

	accounts_table := flag.String("accounts-table", dynamodb.ACCOUNTS_DEFAULT_TABLENAME, "...")

	flag.Parse()

	accounts_opts := dynamodb.DefaultDynamoDBAccountsDatabaseOptions()
	accounts_opts.TableName = *accounts_table

	accounts_db, err := dynamodb.NewDynamoDBAccountsDatabaseWithDSN(*accounts_dsn, accounts_opts)

	if err != nil {
		log.Fatal(err)
	}

	acct, err := accounts_db.GetAccountByEmailAddress(*addr)

	if err != nil {
		log.Fatal(err)
	}

	mfa := acct.MFA

	if mfa == nil {
		log.Fatal("MFA not configured")
	}

	code, err := mfa.GetCode()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(code)
}
