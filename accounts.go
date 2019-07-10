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
	_ "log"
	"strconv"
	"time"
)

const ACCOUNTS_DEFAULT_TABLENAME string = "accounts"

type DynamoDBAccountsDatabaseOptions struct {
	TableName   string
	BillingMode string
	CreateTable bool
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
			"address": {
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

	req := &aws_dynamodb.GetItemInput{
		TableName: aws.String(db.options.TableName),
		Key: map[string]*aws_dynamodb.AttributeValue{
			"address": {
				S: aws.String(addr),
			},
		},
	}

	rsp, err := db.client.GetItem(req)

	if err != nil {
		return nil, err
	}

	return itemToAccount(rsp.Item)
}

func (db *DynamoDBAccountsDatabase) GetAccountByURL(url string) (*account.Account, error) {

	req := &aws_dynamodb.GetItemInput{
		TableName: aws.String(db.options.TableName),
		Key: map[string]*aws_dynamodb.AttributeValue{
			"url": {
				S: aws.String(url),
			},
		},
	}

	rsp, err := db.client.GetItem(req)

	if err != nil {
		return nil, err
	}

	return itemToAccount(rsp.Item)
}

func (db *DynamoDBAccountsDatabase) AddAccount(acct *account.Account) (*account.Account, error) {

	existing_acct, err := db.GetAccountByEmailAddress("FIXME")

	if err != nil && !database.IsNotExist(err) {
		return nil, err
	}

	if existing_acct != nil {
		return nil, errors.New("Account already exists")
	}

	existing_acct, err = db.GetAccountByURL("FIXME")

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

	item, err := aws_dynamodbattribute.MarshalMap(acct)

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

	var acct *account.Account

	err := aws_dynamodbattribute.UnmarshalMap(item, &acct)

	if err != nil {
		return nil, err
	}

	if acct.ID == 0 {
		return nil, new(database.ErrNoAccount)
	}

	return acct, nil
}
