package jobs

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/scraperservice/task"
	"github.com/pkg/errors"
)

// TriggerStartScrapers sends a message to our AmenicJobQueue
// to execute a start_scraper job with type 'schedule'.
// It should execute start_scraper schedule for both theaters.
// func TriggerStartScrapers() error {
// 	attrs := map[string]*sqs.MessageAttributeValue{
// 		"Job": &sqs.MessageAttributeValue{
// 			DataType:    aws.String("String"),
// 			StringValue: aws.String(JobStartScraper),
// 		},
// 		"Type": &sqs.MessageAttributeValue{
// 			DataType:    aws.String("String"),
// 			StringValue: aws.String("schedule"),
// 		},
// 		"IgnoreLastRun": &sqs.MessageAttributeValue{
// 			DataType:    aws.String("String"),
// 			StringValue: aws.String("1"),
// 		},
// 	}
// 	messageId, err := SendMessageToQueue(&sqs.SendMessageInput{
// 		MessageAttributes: attrs,
// 		MessageBody:       aws.String("Information about which scraper to run."),
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Success", messageId)
// 	return nil
// }

// StartScrapers ...
func StartScrapers(input *Input, data persistence.DataAccessLayer) error {
	args := parseArgs(input)
	if args == nil {
		return errors.New("missing args for start_scrapers")
	}

	var err error
	var ignoreLastRun bool

	st := args["type"]
	if ilr := args["ignore_last_run"]; ilr != "" {
		ignoreLastRun, err = strconv.ParseBool(ilr)
		if err != nil {
			return errors.New("found ignore_last_run argument, but the passed value is not a valid bool string")
		}
	}

	cinemais, _ := data.FindTheater(data.DefaultQuery().AddCondition("shortName", "Cinemais").AddCondition("internalId", "34"))
	ibicinemas, _ := data.FindTheater(data.DefaultQuery().AddCondition("internalId", "ibicinemas"))

	// NOTE(diego): Always run nowplaying first
	nowplaying := []task.ScraperOptions{}
	schedule := []task.ScraperOptions{}
	upcoming := []task.ScraperOptions{}
	prices := []task.ScraperOptions{}

	types := strings.Split(st, ",")
	for _, tt := range types {

		var s *[]task.ScraperOptions
		if tt == "now_playing" {
			s = &nowplaying
		} else if tt == "schedule" {
			s = &schedule
		} else if tt == "upcoming" {
			s = &upcoming
		} else if tt == "prices" {
			s = &prices
		}

		tt = strings.TrimSpace(tt)

		if cinemais != nil {
			*s = append(*s, task.ScraperOptions{
				TheaterID:     cinemais.ID.Hex(),
				Type:          tt,
				Provider:      "cinemais",
				IgnoreLastRun: ignoreLastRun,
			})
		}

		if ibicinemas != nil {
			*s = append(*s, task.ScraperOptions{
				TheaterID:     ibicinemas.ID.Hex(),
				Type:          tt,
				Provider:      "ibicinemas",
				IgnoreLastRun: ignoreLastRun,
			})
		}
	}

	var wg sync.WaitGroup
	var errs []error

	execute := func(wg *sync.WaitGroup, opt task.ScraperOptions) {
		defer wg.Done()
		_, err := task.StartScraper(data, opt)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, opt := range nowplaying {
		wg.Add(1)
		go execute(&wg, opt)
	}

	wg.Wait()

	for _, opt := range upcoming {
		wg.Add(1)
		go execute(&wg, opt)
	}

	wg.Wait()

	for _, opt := range schedule {
		wg.Add(1)
		go execute(&wg, opt)
	}

	wg.Wait()

	for _, opt := range prices {
		wg.Add(1)
		go execute(&wg, opt)
	}

	wg.Wait()

	// Make sure our static json are updated.
	if len(schedule) > 0 || len(nowplaying) > 0 || len(upcoming) > 0 {
		err := RunCreateStatic("home")
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
	}

	if len(errs) > 0 {
		var retError error
		for _, e := range errs {
			retError = errors.Wrap(retError, e.Error())
		}
		return retError
	}

	return nil
}
