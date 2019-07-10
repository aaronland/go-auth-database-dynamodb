package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/aaronland/go-auth-database-dynamodb"
	"github.com/aaronland/go-auth/account"
	"github.com/aaronland/go-password/cli"
	"log"
	"os"
)

func main() {

	email := flag.String("email", "", "...")
	username := flag.String("username", "", "...")
	password := flag.String("password", "", "...")

	accounts_dsn := flag.String("accounts-dsn", "", "...")
	accounts_table := flag.String("accounts-table", dynamodb.ACCOUNTS_DEFAULT_TABLENAME, "...")

	flag.Parse()

	accounts_opts := dynamodb.DefaultDynamoDBAccountsDatabaseOptions()
	accounts_opts.TableName = *accounts_table

	accounts_db, err := dynamodb.NewDynamoDBAccountsDatabaseWithDSN(*accounts_dsn, accounts_opts)

	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	if *email == "" {

		fmt.Print("Email address: ")

		addr, err := reader.ReadString('\n')

		if err != nil {
			log.Fatal(err)
		}

		*email = addr
	}

	if *username == "" {

		fmt.Print("Username: ")
		name, err := reader.ReadString('\n')

		if err != nil {
			log.Fatal(err)
		}

		*username = name
	}

	if *password == "" {

		pswd_opts := cli.DefaultGetPasswordOptions()
		pswd, err := cli.GetPassword(pswd_opts)

		if err != nil {
			log.Fatal(err)
		}

		*password = pswd
	}

	// scrub, validate and sanity check email, password, username here...

	acct, err := account.NewAccount(*email, *password, *username)

	if err != nil {
		log.Fatal(err)
	}

	acct, err = accounts_db.AddAccount(acct)

	if err != nil {
		log.Fatal(err)
	}

	secret, err := acct.GetMFASecret()

	if err != nil {
		log.Fatal(err)
	}

	log.Println(acct.ID, secret)
}
