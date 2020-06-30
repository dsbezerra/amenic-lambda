package v2

import (
	"net/http"
	"testing"

	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares"
	"github.com/dsbezerra/amenic-lambda/src/lib/middlewares/rest"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSession(t *testing.T) {
	data := NewMockDataAccessLayer()

	r := NewMockRouter(data)
	r.Use(rest.Init(), middlewares.ValidObjectIDHex(), middlewares.BaseParseQuery())

	testSession := models.Session{
		ID: primitive.NewObjectID(),
	}
	err := data.InsertSession(testSession)
	assert.NoError(t, err)

	s := RESTService{data: data}
	s.ServeSessions(&r.RouterGroup)

	clientAuthToken := getClientAuthToken(t)
	adminAuthToken := getAdminAuthToken(t)

	HexID := testSession.ID.Hex()

	cases := []apiTestCase{
		apiTestCase{
			name:   "It should return Unauthorized",
			method: "GET",
			url:    "/sessions",
			status: http.StatusUnauthorized,
		},
		apiTestCase{
			name:      "It should return BadRequest since ID is not a valid ObjectId",
			method:    "GET",
			url:       "/sessions/session/invalid-session-id",
			status:    http.StatusBadRequest,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return NotFound since Session with ID 5c353e8cebd54428b4f25447 doesn't exist",
			method:    "GET",
			url:       "/sessions/session/5c353e8cebd54428b4f25447",
			status:    http.StatusNotFound,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return a Session with ID " + HexID,
			method:    "GET",
			url:       "/sessions/session/" + HexID,
			status:    http.StatusOK,
			authToken: clientAuthToken,
		},
		apiTestCase{
			name:      "It should return a list of Session models",
			method:    "GET",
			url:       "/sessions",
			status:    http.StatusOK,
			authToken: adminAuthToken,
		},
	}

	r.RunTests(t, cases)

	err = data.DeleteSession(testSession.ID.Hex())
	assert.NoError(t, err)
}
