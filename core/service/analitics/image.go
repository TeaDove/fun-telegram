package analitics

import (
	"context"

	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
)

func (r *Service) KandinskyImageInsert(
	ctx context.Context,
	input *mongo_repository.KandinskyImageDenormalized,
) error {
	err := r.mongoRepository.KandinskyImageInsert(ctx, input)
	if err != nil {
		return errors.Wrap(err, "failed to insert image")
	}

	return nil
}

func (r *Service) KandinskyImagePaginate(
	ctx context.Context,
	input *mongo_repository.KandinskyImagePaginateInput,
) ([]mongo_repository.KandinskyImageDenormalized, error) {
	res, err := r.mongoRepository.KandinskyImagePaginate(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to paginate in mongodb")
	}

	return res, nil
}
