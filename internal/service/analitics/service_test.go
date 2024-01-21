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

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	dbRepository, err := db_repository.New(shared.AppSettings.Storage.MongoDbUrl)
	require.NoError(t, err)
	r, err := New(dbRepository)
	require.NoError(t, err)
	ctx := utils.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, 1350141926) //1779431332
	require.NoError(t, err)

	draw(t, "PopularWordsImage", report.PopularWordsImage)
	draw(t, "ChatterBoxesImage", report.ChatterBoxesImage)
	draw(t, "ChatTimeDistribution", report.ChatTimeDistribution)
}
