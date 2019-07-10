package dynamodb

import (
	_ "context"
	_ "errors"
	_ "github.com/aaronland/go-auth/token"
	"github.com/aaronland/go-auth/database"
	"github.com/aaronland/go-aws-session"
	// aws "github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	// aws_dynamodbattribute "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	_ "log"
	_ "strconv"
	_ "time"
)

const ACCESSTOKENS_DEFAULT_TABLENAME string = "tokens"

type DynamoDBAccessTokensDatabaseOptions struct {
	TableName   string
	BillingMode string
	CreateTable bool
}

func DefaultDynamoDBAccessTokensDatabaseOptions() *DynamoDBAccessTokensDatabaseOptions {

	opts := DynamoDBAccessTokensDatabaseOptions{
		TableName:   ACCESSTOKENS_DEFAULT_TABLENAME,
		BillingMode: "PAY_PER_REQUEST",
		CreateTable: false,
	}

	return &opts
}

type DynamoDBAccessTokensDatabase struct {
	database.AccessTokensDatabase
	client                   *aws_dynamodb.DynamoDB
	options                  *DynamoDBAccessTokensDatabaseOptions
}

func NewDynamoDBAccessTokensDatabaseWithDSN(dsn string, opts *DynamoDBAccessTokensDatabaseOptions) (database.AccessTokensDatabase, error) {

	sess, err := session.NewSessionWithDSN(dsn)

	if err != nil {
		return nil, err
	}

	return NewDynamoDBAccessTokensDatabaseWithSession(sess, opts)
}

func NewDynamoDBAccessTokensDatabaseWithSession(sess *aws_session.Session, opts *DynamoDBAccessTokensDatabaseOptions) (database.AccessTokensDatabase, error) {

	client := aws_dynamodb.New(sess)

	if opts.CreateTable {
		_, err := CreateAccessTokensTable(client, opts)

		if err != nil {
			return nil, err
		}
	}

	db := DynamoDBAccessTokensDatabase{
		client:  client,
		options: opts,
	}

	return &db, nil
}
