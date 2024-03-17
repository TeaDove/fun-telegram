package analitics

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"
	"os"
	"testing"
)

func saveFile(t *testing.T, file *File) {
	out, err := os.Create(fmt.Sprintf(".test-%s", file.Filename()))

	_, err = out.Write(file.Content)
	require.NoError(t, err)

	err = out.Close()
	require.NoError(t, err)
}

func TestIntegration_AnaliticsService_DumpChannels_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	files, err := r.DumpChannels(ctx, "CITeam", 10, 10)
	assert.NoError(t, err)
	assert.Len(t, files, 3)

	for _, file := range files {
		saveFile(t, &file)
	}
}

func TestIntegration_AnaliticsService_AnaliseChannel_Ok(t *testing.T) {
	r := getService(t)
	ctx := shared.GetModuleCtx("tests")

	file, err := r.AnaliseChannel(ctx, &AnaliseChannelInput{
		TgUsername: "svtvnews",
		Depth:      4,
		MaxOrder:   4,
		Locale:     resource.En,
	})
	assert.NoError(t, err)
	saveFile(t, &file)
}
