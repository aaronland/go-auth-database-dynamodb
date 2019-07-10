package dynamodb

import (
	"context"
	"errors"
	"github.com/aaronland/go-auth/account"
	"github.com/aaronland/go-auth/database"
	"github.com/aaronland/go-auth/token"
	"github.com/aaronland/go-aws-session"
	aws "github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	aws_dynamodbattribute "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strconv"
	"time"
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
	client  *aws_dynamodb.DynamoDB
	options *DynamoDBAccessTokensDatabaseOptions
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

func (db *DynamoDBAccessTokensDatabase) GetTokenByID(id int64) (*token.Token, error) {

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

	return itemToToken(rsp.Item)
}

func (db *DynamoDBAccessTokensDatabase) GetTokenByAccessToken(access_token string) (*token.Token, error) {

	return db.getAccountByPointer("access_token", "access_token", access_token)
}

func (db *DynamoDBAccessTokensDatabase) getAccountByPointer(idx string, key string, value string) (*token.Token, error) {

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

	rsp, err := db.client.Query(req)

	if err != nil {
		return nil, err
	}

	count_items := len(rsp.Items)

	if count_items < 1 {
		return nil, new(database.ErrNoToken)
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

	return db.GetTokenByID(id)
}

func (db *DynamoDBAccessTokensDatabase) AddToken(tok *token.Token) (*token.Token, error) {

	id, err := database.NewID()

	if err != nil {
		return nil, err
	}

	tok.ID = id

	err = putToken(db.client, db.options, tok)

	if err != nil {
		return nil, err
	}

	return tok, nil
}

func (db *DynamoDBAccessTokensDatabase) UpdateToken(tok *token.Token) (*token.Token, error) {

	now := time.Now()
	tok.LastModified = now.Unix()

	err := putToken(db.client, db.options, tok)

	if err != nil {
		return tok, err
	}

	return tok, nil
}

func (db *DynamoDBAccessTokensDatabase) RemoveToken(tok *token.Token) (*token.Token, error) {

	str_id := strconv.FormatInt(tok.ID, 10)

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

func (db *DynamoDBAccessTokensDatabase) ListAccessTokens(ctx context.Context, callback database.ListAccessTokensFunc) error {

	req := &aws_dynamodb.ScanInput{
		TableName: aws.String(db.options.TableName),
	}

	return scanTokens(ctx, db.client, req, callback)
}

func (db *DynamoDBAccessTokensDatabase) ListAccessTokensForAccount(ctx context.Context, acct *account.Account, callback database.ListAccessTokensFunc) error {

	str_id := strconv.FormatInt(acct.ID, 10)

	req := &aws_dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#account_id": aws.String("account_id"),
		},
		ExpressionAttributeValues: map[string]*aws_dynamodb.AttributeValue{
			":account_id": {
				N: aws.String(str_id),
			},
		},
		FilterExpression: aws.String("#account_id = :account_id"),
		// ProjectionExpression: aws.String("#status, address"),
		TableName: aws.String(db.options.TableName),
	}

	return scanTokens(ctx, db.client, req, callback)
}

func putToken(client *aws_dynamodb.DynamoDB, opts *DynamoDBAccessTokensDatabaseOptions, tok *token.Token) error {

	item, err := aws_dynamodbattribute.MarshalMap(tok)

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

func itemToToken(item map[string]*aws_dynamodb.AttributeValue) (*token.Token, error) {

	var tok *token.Token

	err := aws_dynamodbattribute.UnmarshalMap(item, &tok)

	if err != nil {
		return nil, err
	}

	return tok, nil
}

func scanTokens(ctx context.Context, client *aws_dynamodb.DynamoDB, req *aws_dynamodb.ScanInput, callback database.ListAccessTokensFunc) error {

	for {

		rsp, err := client.Scan(req)

		if err != nil {
			return err
		}

		for _, item := range rsp.Items {

			tok, err := itemToToken(item)

			if err != nil {
				return err
			}

			err = callback(tok)

			if err != nil {
				return err
			}
		}

		req.ExclusiveStartKey = rsp.LastEvaluatedKey

		if rsp.LastEvaluatedKey == nil {
			break
		}
	}

	return nil
}
