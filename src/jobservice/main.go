package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/dsbezerra/amenic-lambda/src/jobservice/jobs"
)

var (
	// ErrHandlerNotFound is returned if we don't have a handler for the current lambda function
	ErrHandlerNotFound = errors.New("no handler found for this function")
)

// Handler main handler
func Handler(ctx context.Context, event jobs.Input) {
	handler, ok := jobs.Handlers[event.Name]
	if !ok {
		fmt.Println(fmt.Sprintf("[%s] - Failed. Error: %s", lambdacontext.FunctionName, ErrHandlerNotFound.Error()))
		return
	}

	data, err := jobs.GetDataAccessLayer()
	if err != nil {
		fmt.Println(fmt.Sprintf("[%s] - Failed. Error: %s", lambdacontext.FunctionName, err.Error()))
	} else {
		err = handler(&event, data)
		if err == nil {
			fmt.Println(fmt.Sprintf("[%s] - Executed successfully.", lambdacontext.FunctionName))
		} else {
			fmt.Println(fmt.Sprintf("[%s] - Failed. Error: %s", lambdacontext.FunctionName, err.Error()))
		}
	}
}

func main() {
	lambda.Start(Handler)
}
