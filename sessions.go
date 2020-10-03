package dynamodb

import (
	"context"
	"errors"
	"github.com/aaronland/go-auth/account"
	"github.com/aaronland/go-auth/database"
	"github.com/aaronland/go-aws-session"
	aws "github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
	aws_dynamodbattribute "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"net/url"
	"strconv"
	"time"
)

const SESSIONS_DEFAULT_TABLENAME string = "sessions"

type DynamoDBSessionsDatabaseOptions struct {
	TableName   string
	BillingMode string
	CreateTable bool
}

type DynamoDBSession struct {
	SessionId      string           `json:"session_id"`
	AccountId int64            `json:"account_id"`
	Expires   int64           `json:"expires"`
}

func init() {

	ctx := context.Background()
	err := database.RegisterSessionsDatabase(ctx, "session", NewDynamoDBSessionsDatabase)

	if err != nil {
		panic(err)
	}
}

func DefaultDynamoDBSessionsDatabaseOptions() *DynamoDBSessionsDatabaseOptions {

	opts := DynamoDBSessionsDatabaseOptions{
		TableName:   SESSIONS_DEFAULT_TABLENAME,
		BillingMode: "PAY_PER_REQUEST",
		CreateTable: false,
	}

	return &opts
}

type DynamoDBSessionsDatabase struct {
	database.SessionsDatabase
	client  *aws_dynamodb.DynamoDB
	options *DynamoDBSessionsDatabaseOptions
}

func NewDynamoDBSessionsDatabase(ctx context.Context, uri string) (database.SessionsDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}
	
	client := aws_dynamodb.New(sess)

	if opts.CreateTable {

		_, err := CreateSessionsTable(client, opts)

		if err != nil {
			return nil, err
		}
	}

	db := DynamoDBSessionsDatabase{
		client:  client,
		options: opts,
	}

	return &db, nil
}

func (db *DynamoDBSessionsDatabase) GetSessionById(str_id string) (*session.SessionRecord, error) {

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

	return itemToSession(rsp.Item)
}

func (db *DynamoDBSessionsDatabase) AddSession(sess *session.SessionRecord) error  {

	existing_sess, err := db.GetSessionByEmailAddress(sess.Address.URI)

	if err != nil && !database.IsNotExist(err) {
		return nil, err
	}

	if existing_sess != nil {
		return nil, errors.New("Session already exists")
	}

	existing_sess, err = db.GetSessionByURL(sess.Username.Safe)

	if err != nil && !database.IsNotExist(err) {
		return nil, err
	}

	if existing_sess != nil {
		return nil, errors.New("Session already exists")
	}

	id, err := database.NewID()

	if err != nil {
		return nil, err
	}

	sess.ID = id

	err = putSession(db.client, db.options, sess)

	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (db *DynamoDBSessionsDatabase) RemoveSession(sess *session.Session) (*session.Session, error) {

	str_id := strconv.FormatInt(sess.ID, 10)

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

func (db *DynamoDBSessionsDatabase) UpdateSession(sess *session.Session) (*session.Session, error) {

	now := time.Now()
	sess.LastModified = now.Unix()

	err := putSession(db.client, db.options, sess)

	if err != nil {
		return sess, err
	}

	return sess, nil
}

func putSession(client *aws_dynamodb.DynamoDB, opts *DynamoDBSessionsDatabaseOptions, sess *session.Session) error {

	dynamodb_sess := sessionToDynamoDBSession(sess)

	item, err := aws_dynamodbattribute.MarshalMap(dynamodb_sess)

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

func itemToSession(item map[string]*aws_dynamodb.AttributeValue) (*session.Session, error) {

	var dynamodb_sess *DynamoDBSession

	err := aws_dynamodbattribute.UnmarshalMap(item, &dynamodb_sess)

	if err != nil {
		return nil, err
	}

	if dynamodb_sess.ID == 0 {
		return nil, new(database.ErrNoSession)
	}

	sess := dynamodbSessionToSession(dynamodb_sess)

	return sess, nil
}

func sessionToDynamoDBSession(sess *session.Session) *DynamoDBSession {

	dynamodb_sess := DynamoDBSession{
		ID:      sess.ID,
		Created: sess.Created,
		Email:   sess.Address.URI,
		URL:     sess.Username.Safe,
		Session: sess,
	}

	return &dynamodb_sess
}

func dynamodbSessionToSession(dynamodb_sess *DynamoDBSession) *session.Session {
	return dynamodb_sess.Session
}
