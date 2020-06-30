package jobs

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dsbezerra/amenic-lambda/src/imageservice/cloudinary"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
)

type image struct {
	url       string
	imageType string
}
type uploadResult struct {
	movie  models.Movie
	images []models.Image // NOTE(diego): These should be added to database
}
type uploadJob struct {
	movie  models.Movie
	images []image
}

// UploadImagesToCloudinary ...
func UploadImagesToCloudinary(input *Input, data persistence.DataAccessLayer) error {
	mtable := make(map[string]models.Movie, 0)
	movies, err := data.GetNowPlayingMovies(data.DefaultQuery())
	if err != nil {
		return err
	}
	for _, m := range movies {
		_, ok := mtable[m.ID.Hex()]
		if !ok {
			mtable[m.ID.Hex()] = m
		}
	}

	movies, err = data.GetUpcomingMovies(data.DefaultQuery())
	if err != nil {
		return err
	}
	for _, m := range movies {
		_, ok := mtable[m.ID.Hex()]
		if !ok {
			mtable[m.ID.Hex()] = m
		}
	}

	uploadJobs := make([]uploadJob, 0)
	for _, m := range mtable {
		j := uploadJob{movie: m}
		if !matchesCloudinary(m.PosterURL) && m.PosterURL != "" {
			j.images = append(j.images, image{
				url:       m.PosterURL,
				imageType: "poster",
			})
		}
		if !matchesCloudinary(m.BackdropURL) && m.BackdropURL != "" {
			j.images = append(j.images, image{
				url:       m.BackdropURL,
				imageType: "backdrop",
			})
		}
		if len(j.images) > 0 {
			uploadJobs = append(uploadJobs, j)
		}
	}

	if len(uploadJobs) == 0 {
		return nil
	}

	jobs := make(chan uploadJob, 10)
	results := make(chan uploadResult, 10)

	done := make(chan bool)
	go allocateUploadJobs(jobs, uploadJobs)
	go resultUploadJobs(done, results)

	cloudinary.InitService(os.Getenv("CLOUDINARY_URL"))

	uploadPool(data, jobs, results, 10)
	<-done

	return nil
}

func allocateUploadJobs(jobs chan uploadJob, uploads []uploadJob) {
	for _, j := range uploads {
		jobs <- j
	}
	close(jobs)
}

func resultUploadJobs(done chan bool, results chan uploadResult) {
	for result := range results {
		fmt.Printf("Upload for movie %s finished.\n", result.movie.Title)
	}
	done <- true
}

func uploadPool(data persistence.DataAccessLayer, jobs chan uploadJob, results chan uploadResult, n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go uploadWorker(data, &wg, jobs, results)
	}
	wg.Wait()
	close(results)
}

func uploadWorker(data persistence.DataAccessLayer, wg *sync.WaitGroup, jobs chan uploadJob, results chan uploadResult) {
	for job := range jobs {
		upResult := uploadResult{movie: job.movie}
		for _, image := range job.images {
			result, err := cloudinary.UploadWebImage(image.url, image.imageType)
			if err != nil {
				// Ignore error for now.
				continue
			}
			upResult.images = append(upResult.images, *result)
		}

		// Update images
		for _, i := range upResult.images {
			i.MovieID = upResult.movie.ID
			i.Main = true
			switch i.Type {
			case "poster":
				upResult.movie.PosterURL = i.SecureURL
			case "backdrop":
				upResult.movie.BackdropURL = i.SecureURL
			}
			_, err := data.UpdateImage(i.ID.Hex(), i)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		// Update movie
		_, err := data.UpdateMovie(upResult.movie.ID.Hex(), upResult.movie)
		if err != nil {
			fmt.Println(err.Error())
		}

		results <- upResult
	}
	wg.Done()
}

func matchesCloudinary(str string) bool {
	return strings.Contains(str, "cloudinary")
}
