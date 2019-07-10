package main

import (
	"flag"
	"github.com/aaronland/go-auth-database-dynamodb"
	"github.com/aaronland/go-password/cli"
	"log"
)

func main() {

	email := flag.String("email", "", "...")

	accounts_dsn := flag.String("accounts-dsn", "", "...")
	accounts_table := flag.String("accounts-table", dynamodb.ACCOUNTS_DEFAULT_TABLENAME, "...")

	flag.Parse()

	accounts_opts := dynamodb.DefaultDynamoDBAccountsDatabaseOptions()
	accounts_opts.TableName = *accounts_table

	accounts_db, err := dynamodb.NewDynamoDBAccountsDatabaseWithDSN(*accounts_dsn, accounts_opts)

	if err != nil {
		log.Fatal(err)
	}

	acct, err := accounts_db.GetAccountByEmailAddress(*email)

	if err != nil {
		log.Fatal(err)
	}

	pswd_opts := cli.DefaultGetPasswordOptions()
	pswd, err := cli.GetPassword(pswd_opts)

	if err != nil {
		log.Fatal(err)
	}

	acct, err = acct.UpdatePassword(pswd)

	if err != nil {
		log.Fatal(err)
	}
 
	acct, err = accounts_db.UpdateAccount(acct)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("OK")
}

