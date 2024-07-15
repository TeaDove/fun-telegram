package analitics

import (
	"context"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/pkg/errors"
)

func (r *Service) KandinskyImageInsert(
	ctx context.Context,
	input *db_repository.KandinskyImageDenormalized,
) error {
	err := r.dbRepository.KandinskyImageInsert(ctx, input)
	if err != nil {
		return errors.Wrap(err, "failed to insert image")
	}

	return nil
}

func (r *Service) KandinskyImagePaginate(
	ctx context.Context,
	input *db_repository.KandinskyImagePaginateInput,
) ([]db_repository.KandinskyImageDenormalized, error) {
	res, err := r.dbRepository.KandinskyImagePaginate(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to paginate in mongodb")
	}

	return res, nil
}
