package analitics

import (
	"context"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
	"github.com/xitongsys/parquet-go-source/mem"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"
	"io"
)

func getMemFileWrite(file *File) (source.ParquetFile, error) {
	var err error
	fw, err := mem.NewMemFileWriter(file.Name, func(name string, r io.Reader) error {
		file.Content, err = io.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, "failed to read all")
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mem file write")
	}

	return fw, nil
}

func pwWriteSlice(pw *writer.ParquetWriter, slice []any) error {
	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.PageSize = 8 * 1024              //8K
	pw.CompressionType = parquet.CompressionCodec_SNAPPY
	for _, item := range slice {
		err := pw.Write(item)
		if err != nil {
			return errors.Wrap(err, "failed to write")
		}
	}

	err := pw.WriteStop()
	if err != nil {
		return errors.Wrap(err, "failed to write stop")
	}

	return nil
}

func (r *Service) dumpChannelsParquet(ctx context.Context, tgChatIds []int64) (File, error) {
	channels, err := r.chRepository.ChannelSelectByIds(ctx, tgChatIds)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to select channel")
	}

	slice := make([]any, 0, len(channels))
	for _, v := range channels {
		slice = append(slice, v)
	}

	file := File{Name: "channels", Extension: "pqt"}

	fw, err := getMemFileWrite(&file)

	pw, err := writer.NewParquetWriter(fw, new(ch_repository.Channel), 4)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to create parquet writer")
	}

	err = pwWriteSlice(pw, slice)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to dump slice to parquet")
	}

	err = fw.Close()
	if err != nil {
		return File{}, errors.Wrap(err, "failed to close fw")
	}

	return file, nil
}

func (r *Service) dumpChannelsEdgeParquet(channels ch_repository.ChannelsEdges) (File, error) {
	slice := make([]any, 0, len(channels))
	for _, v := range channels {
		slice = append(slice, v)
	}

	file := File{Name: "channels_edges", Extension: "pqt"}

	fw, err := getMemFileWrite(&file)

	pw, err := writer.NewParquetWriter(fw, new(ch_repository.ChannelEdge), 4)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to create parquet writer")
	}

	err = pwWriteSlice(pw, slice)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to dump slice to parquet")
	}

	err = fw.Close()
	if err != nil {
		return File{}, errors.Wrap(err, "failed to close fw")
	}

	return file, nil
}

func (r *Service) dumpMessagesParquet(ctx context.Context, tgChatIds []int64) (File, error) {
	channels, err := r.chRepository.MessagesGetByChatIds(ctx, tgChatIds)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to select items")
	}

	slice := make([]any, 0, len(channels))
	for _, v := range channels {
		slice = append(slice, v.ToParquet())
	}

	file := File{Name: "messages", Extension: "pqt"}

	fw, err := getMemFileWrite(&file)

	pw, err := writer.NewParquetWriter(fw, new(ch_repository.MessageParquet), 4)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to create parquet writer")
	}

	err = pwWriteSlice(pw, slice)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to dump slice to parquet")
	}

	err = fw.Close()
	if err != nil {
		return File{}, errors.Wrap(err, "failed to close fw")
	}

	return file, nil
}
