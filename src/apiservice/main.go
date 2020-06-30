package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	v1 "github.com/dsbezerra/amenic-lambda/src/apiservice/v1"
	v2 "github.com/dsbezerra/amenic-lambda/src/apiservice/v2"
	"github.com/dsbezerra/amenic-lambda/src/lib/config"
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares"
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/mongolayer"
	"github.com/gin-gonic/gin"
)

const (
	ServiceName = "API"
)

// Context ...
type Context struct {
	Service string
	Config  *config.ServiceConfig
	Data    persistence.DataAccessLayer
}

var appCtx *Context
var ginLambda *ginadapter.GinLambda

// StaticHandler is used to serve .json files from amenic-static bucket
func StaticHandler(ctx context.Context, proxy string) (events.APIGatewayProxyResponse, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("sa-east-1")})
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Internal Server Error", StatusCode: 500}, nil
	}
	svc := s3.New(sess)
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("STATIC_BUCKET_NAME")),
		Key:    aws.String(proxy),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Internal Server Error", StatusCode: 500}, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Internal Server Error", StatusCode: 500}, nil
	}

	headers := map[string]string{
		"Content-Type":   "application/json",
		"Last-Modified":  resp.LastModified.String(),
		"Etag":           *resp.ETag,
		"Content-Length": fmt.Sprintf("%d", *resp.ContentLength),
	}
	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Body:            string(body),
		Headers:         headers,
		IsBase64Encoded: false,
	}, nil
}

// Handler ...
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check for static urls here (for now only json files are treated like static files)
	if proxy := req.PathParameters["proxy"]; proxy == "" {
		return events.APIGatewayProxyResponse{Body: "Hello!", StatusCode: 200}, nil
	} else if strings.HasSuffix(proxy, ".json") {
		return StaticHandler(ctx, proxy)
	}

	if ginLambda == nil {
		router := appCtx.buildRouter()
		ginLambda = ginadapter.New(router)
	}

	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// Initialize app context
	if appCtx == nil {
		settings, err := config.LoadConfiguration()
		if err != nil {
			log.Fatal(err)
		}

		data, err := mongolayer.NewMongoDAL(settings.DBConnection)
		if err != nil {
			log.Fatal(err)
		}

		if setupDatabaseAtStart := os.Getenv("SETUP_DATABASE_AT_START"); setupDatabaseAtStart != "" {
			value, err := strconv.ParseBool(setupDatabaseAtStart)
			if err == nil && value {
				data.Setup()
			}
		}

		appCtx = &Context{
			Service: ServiceName,
			Config:  settings,
			Data:    data,
		}
	}

	if os.Getenv("AWS_LAMBDA") == "true" {
		lambda.Start(Handler)
	} else {
		defer appCtx.Data.Close()

		// Make sure we have our static files updated
		v1.CreateStatic(appCtx.Data, v1.StaticTypeHome)
		router := appCtx.buildRouter()
		router.Run(appCtx.Config.RESTEndpoint)
	}
}

func (ctx *Context) buildRouter() *gin.Engine {
	if ctx.Config.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.SecureJsonPrefix(")]}',\n")

	// Setup middlewares.
	if !ctx.Config.AWSLambda {
		router.Use(helmet.Default())
		router.Use(
			rest.Init(),
			cors.Default(),
			gzip.Gzip(gzip.DefaultCompression),
		)
		router.Use(middlewares.StaticWithCache("/", static.LocalFile("./static", false)))
	} else {
		router.Use(rest.Init())
	}

	// Setup route handlers.
	if !ctx.Config.AWSLambda {
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "Hello!")
		})
	}

	v1.AddRoutes(router, ctx.Data)
	v2.AddRoutes(router, ctx.Data)

	return router
}

// Mode ...
func (ctx *Context) Mode() string {
	return os.Getenv("AMENIC_MODE")
}
