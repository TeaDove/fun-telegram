package analitics

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/core/repository/ch_repository"
	"github.com/teadove/goteleout/core/repository/mongo_repository"
	"github.com/teadove/goteleout/core/shared"
)

func draw(t *testing.T, reportImage RepostImage) {
	img, _, err := image.Decode(bytes.NewReader(reportImage.Content))
	require.NoError(t, err)

	out, err := os.Create(reportImage.Filename())
	defer out.Close()

	err = jpeg.Encode(out, img, nil)
	require.NoError(t, err)
}

func getService(t *testing.T) *Service {
	ctx := shared.GetCtx()
	dbRepository, err := mongo_repository.New()
	require.NoError(t, err)

	chRepository, err := ch_repository.New(ctx)
	require.NoError(t, err)

	r, err := New(dbRepository, chRepository)
	require.NoError(t, err)

	return r
}

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, 1701683862, 3, "") //1779431332 1350141926 1178533048
	require.NoError(t, err)

	for _, reportImage := range report.Images {
		draw(t, reportImage)
	}
}

func TestIntegration_AnaliticsService_AnaliseChatForUser_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, 1701683862, 3, "abaturoff") //1779431332 1350141926 1178533048
	require.NoError(t, err)

	for _, reportImage := range report.Images {
		draw(t, reportImage)
	}
}
