package dynamodb

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	aws_dynamodb "github.com/aws/aws-sdk-go/service/dynamodb"
)

func CreateAccountsTable(client *aws_dynamodb.DynamoDB, opts *DynamoDBAccountsDatabaseOptions) (bool, error) {

	has_table, err := hasTable(client, opts.TableName)

	if err != nil {
		return false, err
	}

	if has_table {
		return true, nil
	}

	req := &aws_dynamodb.CreateTableInput{
		AttributeDefinitions: []*aws_dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String("created"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*aws_dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("email"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("url"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*aws_dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("email"),
				KeySchema: []*aws_dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("email"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("url"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &aws_dynamodb.Projection{
					// maybe just address...?
					ProjectionType: aws.String("ALL"),
				},
			},
		},
		BillingMode: aws.String(opts.BillingMode),
		TableName:   aws.String(opts.TableName),
	}

	_, err = client.CreateTable(req)

	if err != nil {
		return false, err
	}

	return true, nil
}

func CreateAccessTokensTable(client *aws_dynamodb.DynamoDB, opts *DynamoDBAccessTokensDatabaseOptions) (bool, error) {

	return false, errors.New("Please write me")

	has_table, err := hasTable(client, opts.TableName)

	if err != nil {
		return false, err
	}

	if has_table {
		return true, nil
	}

	req := &aws_dynamodb.CreateTableInput{
		AttributeDefinitions: []*aws_dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("address"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("event"),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String("created"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*aws_dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("address"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("created"),
				KeyType:       aws.String("RANGE"),
			},
		},
		GlobalSecondaryIndexes: []*aws_dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("address"),
				KeySchema: []*aws_dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("address"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &aws_dynamodb.Projection{
					// maybe just address...?
					ProjectionType: aws.String("ALL"),
				},
			},
			{
				IndexName: aws.String("event"),
				KeySchema: []*aws_dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("event"),
						KeyType:       aws.String("HASH"),
					},
				},
				Projection: &aws_dynamodb.Projection{
					// maybe just address...?
					ProjectionType: aws.String("ALL"),
				},
			},
		},
		BillingMode: aws.String(opts.BillingMode),
		TableName:   aws.String(opts.TableName),
	}

	_, err = client.CreateTable(req)

	if err != nil {
		return false, err
	}

	return true, nil
}

func hasTable(client *aws_dynamodb.DynamoDB, table string) (bool, error) {

	tables, err := listTables(client)

	if err != nil {
		return false, err
	}

	has_table := false

	for _, name := range tables {

		if name == table {
			has_table = true
			break
		}
	}

	return has_table, nil
}

func listTables(client *aws_dynamodb.DynamoDB) ([]string, error) {

	tables := make([]string, 0)

	input := &aws_dynamodb.ListTablesInput{}

	for {

		rsp, err := client.ListTables(input)

		if err != nil {
			return nil, err
		}

		for _, n := range rsp.TableNames {
			tables = append(tables, *n)
		}

		input.ExclusiveStartTableName = rsp.LastEvaluatedTableName

		if rsp.LastEvaluatedTableName == nil {
			break
		}
	}

	return tables, nil
}
