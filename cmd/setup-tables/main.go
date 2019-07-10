package main

import (
	"flag"
	"github.com/aaronland/go-auth-database-dynamodb"
	"log"
)

func main() {

	accounts_table := flag.String("accounts-table", dynamodb.ACCOUNTS_DEFAULT_TABLENAME, "...")
	tokens_table := flag.String("access-tokens-table", dynamodb.ACCESSTOKENS_DEFAULT_TABLENAME, "...")

	dsn := flag.String("dsn", "", "...")

	flag.Parse()

	accounts_opts := dynamodb.DefaultDynamoDBAccountsDatabaseOptions()
	tokens_opts := dynamodb.DefaultDynamoDBAccessTokensDatabaseOptions()

	accounts_opts.TableName = *accounts_table
	accounts_opts.CreateTable = true

	tokens_opts.TableName = *tokens_table
	tokens_opts.CreateTable = true

	var err error

	_, err = dynamodb.NewDynamoDBAccountsDatabaseWithDSN(*dsn, accounts_opts)

	if err != nil {
		log.Printf("Failed to set up %s table, %s\n", accounts_opts.TableName, err)
	}

	_, err = dynamodb.NewDynamoDBAccessTokensDatabaseWithDSN(*dsn, tokens_opts)

	if err != nil {
		log.Printf("Failed to set up %s table, %s\n", tokens_opts.TableName, err)
	}

}
