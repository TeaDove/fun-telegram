package analitics

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/teadove/teasutils/utils/test_utils"
	"gorm.io/gorm"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/teadove/fun_telegram/core/service/resource"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/stretchr/testify/require"
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

		shared.CloseOrLog(test_utils.GetLoggedContext(), out)
	}
}

func getService(t *testing.T) *Service {
	ctx := test_utils.GetLoggedContext()

	dsSupplier, err := ds_supplier.New(ctx)
	require.NoError(t, err)

	resourceService, err := resource.New(ctx)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	dbRepository, err := db_repository.NewRepository(ctx, db)
	require.NoError(t, err)

	r, err := New(dsSupplier, resourceService, dbRepository)
	require.NoError(t, err)

	return r
}

func TestIntegration_AnaliticsService_AnaliseChat_Ok(t *testing.T) {
	r := getService(t)
	ctx := test_utils.GetLoggedContext()

	report, err := r.AnaliseChat(ctx, &AnaliseChatInput{
		TgChatId:  1825059942,
		Tz:        3,
		Locale:    resource.Ru,
		Anonymize: true,
	}) // 1779431332 1350141926 1178533048
	require.NoError(t, err)

	draw(t, report.Images)
	test_utils.Pprint(report.MessagesCount)
	test_utils.Pprint(report.FirstMessageAt)
}

func TestIntegration_AnaliticsService_AnaliseChatForUser_Ok(t *testing.T) {
	r := getService(t)
	ctx := test_utils.GetLoggedContext()

	report, err := r.AnaliseChat(ctx, &AnaliseChatInput{
		TgChatId: 1178533048,
		Tz:       3,
		TgUserId: 418878871,
		Locale:   resource.En,
	}) // 1779431332 1350141926 1178533048
	require.NoError(t, err)

	draw(t, report.Images)
	test_utils.Pprint(report.MessagesCount)
	test_utils.Pprint(report.FirstMessageAt)
}
