package jobs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadImagesToCloudinary(t *testing.T) {
	// NOTE(diego): Needed to make upload goes to a path preffixed with dev/
	os.Setenv("AMENIC_MODE", "debug")
	os.Setenv("CLOUDINARY_URL", "cloudinary://649342348956991:KunmyL9eIQH_HU-2BqhByPR3nYs@dyrib46is")

	data, err := mockDataAccessLayer()
	assert.NoError(t, err)
	assert.NotNil(t, data)
	err = UploadImagesToCloudinary(nil, data)
	assert.NoError(t, err)
}
