package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dsbezerra/amenic-lambda/src/jobservice/jobs"
)

// Handler main worker handler
func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		job, ok := message.MessageAttributes["Job"]
		if !ok || job.StringValue == nil {
			fmt.Println("No job found")
			continue
		}

		fmt.Printf("Handling job %s...\n", *job.StringValue)

		switch *job.StringValue {
		case jobs.JobCreateStatic:
			cst, ok := message.MessageAttributes["Type"]
			if !ok || cst.StringValue == nil {
				fmt.Println("Missing type for create_static")
				continue
			}
			err := jobs.RunCreateStatic(*cst.StringValue)
			if err != nil {
				fmt.Println(err.Error())
			}

			// case jobs.JobStartScraper:
			// 	st, ok := message.MessageAttributes["Type"]
			// 	if !ok || st.StringValue == nil {
			// 		fmt.Println("Missing type for start_scraper")
			// 		continue
			// 	}

			// 	var ignoreLastRun bool
			// 	var err error
			// 	ignoreLastRunString, ok := message.MessageAttributes["IgnoreLastRun"]
			// 	if !ok || ignoreLastRunString.StringValue == nil {
			// 	} else {
			// 		ignoreLastRun, err = strconv.ParseBool(*ignoreLastRunString.StringValue)
			// 	}
			// 	err = jobs.StartScrapers(*st.StringValue, ignoreLastRun)
			// 	if err != nil {
			// 		fmt.Println(err.Error())
			// 	}
		}

		// TODO: Update task document status matching the job
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
