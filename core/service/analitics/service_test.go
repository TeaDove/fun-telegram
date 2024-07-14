package analitics

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/teadove/fun_telegram/core/infrastructure/pg"
	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/teadove/fun_telegram/core/service/resource"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"github.com/teadove/fun_telegram/core/shared"
)

func draw(t *testing.T, reportImages []File) {
	files, err := filepath.Glob(".test-*")
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

		err = jpeg.Encode(out, img, nil)
		require.NoError(t, err)

		shared.CloseOrLog(shared.GetCtx(), out)
	}
}

func getService(t *testing.T) *Service {
	ctx := shared.GetCtx()
	mongoRepository, err := mongo_repository.New()
	require.NoError(t, err)

	dsSupplier, err := ds_supplier.New(ctx)
	require.NoError(t, err)

	resourceService, err := resource.New(ctx)
	require.NoError(t, err)

	db, err := pg.NewClientFromSettings()
	require.NoError(t, err)

	dbRepository, err := db_repository.NewRepository(ctx, db)
	require.NoError(t, err)

	r, err := New(mongoRepository, dsSupplier, resourceService, dbRepository)
	require.NoError(t, err)

	return r
}

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	report, err := r.AnaliseChat(ctx, &AnaliseChatInput{
		TgChatId:  1825059942,
		Tz:        3,
		Locale:    resource.Ru,
		Anonymize: true,
	}) // 1779431332 1350141926 1178533048
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
	}) // 1779431332 1350141926 1178533048
	require.NoError(t, err)

	draw(t, report.Images)
	shared.SendInterface(report.MessagesCount)
	shared.SendInterface(report.FirstMessageAt)
}
