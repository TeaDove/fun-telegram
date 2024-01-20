package analitics

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	dbRepository, err := db_repository.New(shared.AppSettings.Storage.MongoDbUrl)
	require.NoError(t, err)
	r, err := New(dbRepository)
	require.NoError(t, err)
	ctx := utils.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, 1178533048) //1779431332
	require.NoError(t, err)

	img, _, err := image.Decode(bytes.NewReader(report.PopularWordsImage))
	require.NoError(t, err)

	out, err := os.Create("./PopularWordsImage.jpeg")
	defer out.Close()

	err = jpeg.Encode(out, img, nil)
	require.NoError(t, err)

	imgChatter, _, err := image.Decode(bytes.NewReader(report.ChatterBoxesImage))
	require.NoError(t, err)

	outChatter, err := os.Create("./ChatterBoxesImage.jpeg")
	defer outChatter.Close()

	err = jpeg.Encode(outChatter, imgChatter, nil)
	require.NoError(t, err)

	utils.SendInterface(report.FirstMessageAt)
}
