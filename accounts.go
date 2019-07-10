package dynamodb

import (
	_ "context"
	"errors"
	"github.com/aaronland/go-auth/account"
	"github.com/aaronland/go-auth/database"
	"github.com/aaronland/go-aws-session"
	aws "github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	aws_dynamodbattribute "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"strconv"
	"time"
)

const ACCOUNTS_DEFAULT_TABLENAME string = "accounts"

type DynamoDBAccountsDatabaseOptions struct {
	TableName   string
	BillingMode string
	CreateTable bool
}

type DynamoDBAccount struct {
	ID      int64            `json:"id"`
	Created int64            `json:"created"`
	Email   string           `json:"email"`
	URL     string           `json:"url"`
	Account *account.Account `json:"account"`
}

func DefaultDynamoDBAccountsDatabaseOptions() *DynamoDBAccountsDatabaseOptions {

	opts := DynamoDBAccountsDatabaseOptions{
		TableName:   ACCOUNTS_DEFAULT_TABLENAME,
		BillingMode: "PAY_PER_REQUEST",
		CreateTable: false,
	}

	return &opts
}

type DynamoDBAccountsDatabase struct {
	database.AccountsDatabase
	client  *aws_dynamodb.DynamoDB
	options *DynamoDBAccountsDatabaseOptions
}

func NewDynamoDBAccountsDatabaseWithDSN(dsn string, opts *DynamoDBAccountsDatabaseOptions) (database.AccountsDatabase, error) {

	sess, err := session.NewSessionWithDSN(dsn)

	if err != nil {
		return nil, err
	}

	return NewDynamoDBAccountsDatabaseWithSession(sess, opts)
}

func NewDynamoDBAccountsDatabaseWithSession(sess *aws_session.Session, opts *DynamoDBAccountsDatabaseOptions) (database.AccountsDatabase, error) {

	client := aws_dynamodb.New(sess)

	if opts.CreateTable {

		_, err := CreateAccountsTable(client, opts)

		if err != nil {
			return nil, err
		}
	}

	db := DynamoDBAccountsDatabase{
		client:  client,
		options: opts,
	}

	return &db, nil
}

func (db *DynamoDBAccountsDatabase) GetAccountByID(id int64) (*account.Account, error) {

	str_id := strconv.FormatInt(id, 10)

	req := &aws_dynamodb.GetItemInput{
		TableName: aws.String(db.options.TableName),
		Key: map[string]*aws_dynamodb.AttributeValue{
			"id": {
				N: aws.String(str_id),
			},
		},
	}

	rsp, err := db.client.GetItem(req)

	if err != nil {
		return nil, err
	}

	return itemToAccount(rsp.Item)
}

func (db *DynamoDBAccountsDatabase) GetAccountByEmailAddress(addr string) (*account.Account, error) {
	return db.getAccountByPointer("email", "email", addr)
}

func (db *DynamoDBAccountsDatabase) GetAccountByURL(url string) (*account.Account, error) {
	return db.getAccountByPointer("url", "url", url)
}

func (db *DynamoDBAccountsDatabase) getAccountByPointer(idx string, key string, value string) (*account.Account, error) {

	req := &aws_dynamodb.QueryInput{
		TableName: aws.String(db.options.TableName),
		KeyConditions: map[string]*aws_dynamodb.Condition{
			key: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*aws_dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
			},
		},
		ProjectionExpression: aws.String("id"),
		IndexName:            aws.String(idx),
	}

	// log.Println("GET", req)

	rsp, err := db.client.Query(req)

	if err != nil {
		return nil, err
	}

	count_items := len(rsp.Items)

	if count_items < 1 {
		return nil, new(database.ErrNoAccount)
	}

	if count_items > 1 {
		return nil, errors.New("Multiple results for key!")
	}

	rsp_id := rsp.Items[0]["id"]
	str_id := *rsp_id.N

	id, err := strconv.ParseInt(str_id, 10, 64)

	if err != nil {
		return nil, err
	}

	log.Printf("GET BY ID FOR %s=%s (%d)\n", key, value, id)
	return db.GetAccountByID(id)
}

func (db *DynamoDBAccountsDatabase) AddAccount(acct *account.Account) (*account.Account, error) {

	existing_acct, err := db.GetAccountByEmailAddress(acct.Address.URI)

	if err != nil && !database.IsNotExist(err) {
		return nil, err
	}

	if existing_acct != nil {
		return nil, errors.New("Account already exists")
	}

	existing_acct, err = db.GetAccountByURL(acct.Username.Safe)

	if err != nil && !database.IsNotExist(err) {
		return nil, err
	}

	if existing_acct != nil {
		return nil, errors.New("Account already exists")
	}

	id, err := database.NewID()

	if err != nil {
		return nil, err
	}

	acct.ID = id

	err = putAccount(db.client, db.options, acct)

	if err != nil {
		return nil, err
	}

	return acct, nil
}

func (db *DynamoDBAccountsDatabase) RemoveAccount(acct *account.Account) (*account.Account, error) {

	str_id := strconv.FormatInt(acct.ID, 10)

	req := &aws_dynamodb.DeleteItemInput{
		TableName: aws.String(db.options.TableName),
		Key: map[string]*aws_dynamodb.AttributeValue{
			"id": {
				N: aws.String(str_id),
			},
		},
	}

	_, err := db.client.DeleteItem(req)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *DynamoDBAccountsDatabase) UpdateAccount(acct *account.Account) (*account.Account, error) {

	now := time.Now()
	acct.LastModified = now.Unix()

	err := putAccount(db.client, db.options, acct)

	if err != nil {
		return acct, err
	}

	return acct, nil
}

func putAccount(client *aws_dynamodb.DynamoDB, opts *DynamoDBAccountsDatabaseOptions, acct *account.Account) error {

	dynamodb_acct := accountToDynamoDBAccount(acct)

	item, err := aws_dynamodbattribute.MarshalMap(dynamodb_acct)

	if err != nil {
		return err
	}

	req := &aws_dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(opts.TableName),
	}

	_, err = client.PutItem(req)

	if err != nil {
		return err
	}

	return nil
}

func itemToAccount(item map[string]*aws_dynamodb.AttributeValue) (*account.Account, error) {

	var dynamodb_acct *DynamoDBAccount

	err := aws_dynamodbattribute.UnmarshalMap(item, &dynamodb_acct)

	if err != nil {
		return nil, err
	}

	if dynamodb_acct.ID == 0 {
		return nil, new(database.ErrNoAccount)
	}

	acct := dynamodbAccountToAccount(dynamodb_acct)

	return acct, nil
}

func accountToDynamoDBAccount(acct *account.Account) *DynamoDBAccount {

	dynamodb_acct := DynamoDBAccount{
		ID:      acct.ID,
		Created: acct.Created,
		Email:   acct.Address.URI,
		URL:     acct.Username.Safe,
		Account: acct,
	}

	return &dynamodb_acct
}

func dynamodbAccountToAccount(dynamodb_acct *DynamoDBAccount) *account.Account {
	return dynamodb_acct.Account
}
