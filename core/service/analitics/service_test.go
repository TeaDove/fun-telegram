package analitics

import (
	"bytes"
	"fmt"
	"github.com/teadove/goteleout/core/service/resource"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/teadove/goteleout/core/supplier/ds_supplier"

	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/core/repository/ch_repository"
	"github.com/teadove/goteleout/core/repository/mongo_repository"
	"github.com/teadove/goteleout/core/shared"
)

func draw(t *testing.T, reportImages []RepostImage) {
	files, err := filepath.Glob(".test-*.jpeg")
	if err != nil {
		require.NoError(t, err)
	}

	for _, f := range files {
		if err = os.Remove(f); err != nil {
			require.NoError(t, err)
		}
	}

	for _, reportImage := range reportImages {
		img, _, err := image.Decode(bytes.NewReader(reportImage.Content))
		require.NoError(t, err)

		out, err := os.Create(fmt.Sprintf(".test-%s", reportImage.Filename()))
		defer out.Close()

		err = jpeg.Encode(out, img, nil)
		require.NoError(t, err)
	}
}

func getService(t *testing.T) *Service {
	ctx := shared.GetCtx()
	dbRepository, err := mongo_repository.New()
	require.NoError(t, err)

	chRepository, err := ch_repository.New(ctx)
	require.NoError(t, err)

	dsSupplier, err := ds_supplier.New(ctx)
	require.NoError(t, err)

	resourceService, err := resource.New(ctx)
	require.NoError(t, err)

	r, err := New(dbRepository, chRepository, dsSupplier, resourceService)
	require.NoError(t, err)

	return r
}

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, &AnaliseChatInput{
		TgChatId: 1701683862,
		Tz:       3,
		Locale:   resource.Ru,
	}) //1779431332 1350141926 1178533048
	require.NoError(t, err)

	draw(t, report.Images)
	shared.SendInterface(report.MessagesCount)
	shared.SendInterface(report.FirstMessageAt)
}

func TestIntegration_AnaliticsService_AnaliseChatForUser_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, &AnaliseChatInput{
		TgChatId: 1701683862,
		Tz:       3,
		TgUserId: 418878871,
		Locale:   resource.En,
	}) //1779431332 1350141926 1178533048
	require.NoError(t, err)

	draw(t, report.Images)
	shared.SendInterface(report.MessagesCount)
	shared.SendInterface(report.FirstMessageAt)
}
