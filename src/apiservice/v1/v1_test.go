/** DEPRECATED */

package v1

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/mongolayer"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/apiutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	testConnection = "mongodb://localhost/amenic-test"
)

type apiTestCase struct {
	name         string
	method       string
	url          string
	body         string
	status       int
	appendAPIKey bool
	onResponse   func(r *httptest.ResponseRecorder)
}

type mockRouter struct {
	*gin.Engine
	data persistence.DataAccessLayer
}

func (r *mockRouter) Call(method, URL, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, URL, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	return res
}

func (r *mockRouter) RunTests(t *testing.T, tests []apiTestCase) {
	// Make sure we have a api-key for authenticated requests...
	apikey := models.APIKey{
		ID:       primitive.NewObjectID(),
		Key:      "test-api-key",
		UserType: "admin",
		Owner:    "test",
	}

	r.data.InsertAPIKey(apikey)
	defer r.data.DeleteAPIKey(apikey.ID.Hex())

	for _, test := range tests {
		if test.appendAPIKey {
			if strings.Index(test.url, "?") > -1 {
				test.url += "&"
			} else {
				test.url += "?"
			}
			test.url += "api_key=" + apikey.Key
		}
		res := r.Call(test.method, test.url, test.body)
		assert.Equal(t, test.status, res.Code, test.name)
		if res != nil && test.onResponse != nil {
			test.onResponse(res)
		}
	}
}

func newAPITestCase(name, method, url, body string, status int, appendAPIKey bool, onResponse func(*httptest.ResponseRecorder)) apiTestCase {
	return apiTestCase{
		name,
		method,
		url,
		body,
		status,
		appendAPIKey,
		onResponse,
	}
}

func NewMockRouter(data persistence.DataAccessLayer) *mockRouter {
	return &mockRouter{Engine: gin.New(), data: data}
}

func NewMockDataAccessLayer() persistence.DataAccessLayer {
	data, err := mongolayer.NewMongoDAL(testConnection)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func ConvertAPIResponse(r *httptest.ResponseRecorder, result interface{}) {
	var response apiutil.APIResponse
	err := json.Unmarshal(r.Body.Bytes(), &response)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(response.Data)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Fatal(err)
	}
}
