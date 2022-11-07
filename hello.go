package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
	ID       int    `dynamodbav:"id" json:"id"`
	Day      string `dynamodbav:"day" json:"day"`
	Order    string `dynamodbav:"order" json:"order"`
	Name     string `dynamodbav:"name" json:"name"`
	Setlist1 string `dynamodbav:"set_list1" json:"set_list1"`
	Setlist2 string `dynamodbav:"set_list2" json:"set_list2"`
	Setlist3 string `dynamodbav:"set_list3" json:"set_list3"`
	Setlist4 string `dynamodbav:"set_list4" json:"set_list4"`
}

type Response struct {
	Day1Items []Item `json:"1"`
	Day2Items []Item `json:"2"`
}

var tableName string = "homes_serverless_bands"

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// DB接続
	sess, err := session.NewSession()
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}
	db := dynamodb.New(sess)

	var items []Item = []Item{}

	scaned, err := db.Scan(&dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	for _, scanedItem := range scaned.Items {
		var item Item
		err = dynamodbattribute.UnmarshalMap(scanedItem, &item)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       err.Error(),
				StatusCode: 500,
			}, err
		}
		items = append(items, item)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	var day1items []Item 
	var day2items []Item 
	//day1のバンドを出演順に並び替え
	go func(items []Item) {
		defer wg.Done()
		for _, item := range items {
			if item.Day == "1" {
				day1items = append(day1items, item)
			}
		}
		sort.Slice(day1items, func(i, j int) bool { return day1items[i].Order < day1items[j].Order })
		return
	}(items)

	go func(items []Item) {
		defer wg.Done()
		for _, item := range items {
			if item.Day == "2" {
				day2items = append(day2items, item)
			}
		}
		sort.Slice(day2items, func(i, j int) bool { return day2items[i].Order < day2items[j].Order })
		return
	}(items)
	wg.Wait()
	res := Response{
		Day1Items: day1items,
		Day2Items: day2items,
	}
	jsonBytes, err := json.Marshal(res)
    if err != nil {
        fmt.Println(err)
    }
    headers := map[string]string{
        "Content-Type":                    "application/json",
        "Access-Control-Allow-Origin":     request.Headers["origin"], 
        "Access-Control-Allow-Methods":    "OPTIONS,POST,GET",
        "Access-Control-Allow-Headers":    "Origin,Authorization,Accept,X-Requested-With",
        "Access-Control-Allow-Credential": "true",
    }

	return events.APIGatewayProxyResponse{
        Headers:    headers,
		Body:       string(jsonBytes),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
