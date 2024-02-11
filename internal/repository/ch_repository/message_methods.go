package ch_repository

import "context"

func (r *Repository) MessageCreate(ctx context.Context, message *Message) error {
	return nil
	//err := r.conn.AsyncInsert(ctx)
	//if err != nil {
	//	return errors.WithStack(err)
	//}

	return nil
}
