package analitics

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func draw(t *testing.T, name string, imageBytes []byte) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	require.NoError(t, err)

	out, err := os.Create(fmt.Sprintf("./%s.jpeg", name))
	defer out.Close()

	err = jpeg.Encode(out, img, nil)
	require.NoError(t, err)
}

func getService(t *testing.T) *Service {
	dbRepository, err := db_repository.New(shared.AppSettings.Storage.MongoDbUrl)
	require.NoError(t, err)
	r, err := New(dbRepository)
	require.NoError(t, err)

	return r
}

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	r := getService(t)
	ctx := utils.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, 1825059942, 3, "") //1779431332 1350141926 1178533048
	require.NoError(t, err)

	for idx, reportImage := range report.Images {
		draw(t, fmt.Sprintf("%d", idx), reportImage)
	}
}

func TestIntegration_AnaliticsService_AnaliseChatForUser_Ok(t *testing.T) {
	r := getService(t)
	ctx := utils.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, 1825059942, 3, "TeaDove") //1779431332 1350141926 1178533048
	require.NoError(t, err)

	for idx, reportImage := range report.Images {
		draw(t, fmt.Sprintf("%d", idx), reportImage)
	}
}
