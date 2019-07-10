package main

import (
       "context"
	"flag"
	"github.com/aaronland/go-auth-database-dynamodb"
	"github.com/aaronland/go-auth/www"
	"log"
)

func main() {

	email := flag.String("email", "", "...")

	accounts_dsn := flag.String("accounts-dsn", "", "...")
	tokens_dsn := flag.String("tokens-dsn", "", "...")
	aws_dsn := flag.String("aws-dsn", "", "...")

	accounts_table := flag.String("accounts-table", dynamodb.ACCOUNTS_DEFAULT_TABLENAME, "...")
	tokens_table := flag.String("tokens-table", dynamodb.ACCESSTOKENS_DEFAULT_TABLENAME, "...")

	flag.Parse()

	if *aws_dsn != "" {

		if *accounts_dsn == "" {
			*accounts_dsn = *aws_dsn
		}

		if *tokens_dsn == "" {
			*tokens_dsn = *aws_dsn
		}
	}

	accounts_opts := dynamodb.DefaultDynamoDBAccountsDatabaseOptions()
	accounts_opts.TableName = *accounts_table

	accounts_db, err := dynamodb.NewDynamoDBAccountsDatabaseWithDSN(*accounts_dsn, accounts_opts)

	if err != nil {
		log.Fatal(err)
	}

	tokens_opts := dynamodb.DefaultDynamoDBAccessTokensDatabaseOptions()
	tokens_opts.TableName = *tokens_table

	tokens_db, err := dynamodb.NewDynamoDBAccessTokensDatabaseWithDSN(*tokens_dsn, tokens_opts)

	if err != nil {
		log.Fatal(err)
	}

	acct, err := accounts_db.GetAccountByEmailAddress(*email)

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tok, err := www.GetSiteTokenForAccount(ctx, tokens_db, acct)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(tok)
}
