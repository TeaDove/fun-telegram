package memory

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
)

func (r *Storage) loadFlushed() error {
	r.flushMu.Lock()
	defer r.flushMu.Unlock()

	content, err := os.ReadFile(r.filename)
	if err != nil {
		log.Warn().Str("status", "err.while.reading.file").Stack().Err(err).Send()
		return nil
	}

	newMap := make(map[string][]byte, 10)

	err = json.Unmarshal(content, &newMap)
	if err != nil {
		log.Warn().
			Str("status", "err.while.unmarshalling.json.recreating.file").
			Stack().
			Err(err).
			Send()

		err = os.Remove(r.filename)

		return errors.WithStack(err)
	}

	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()
	r.mapping = newMap
	log.Info().Str("status", "load_flushed.end").Int("len", len(r.mapping)).Send()

	return errors.WithStack(err)
}

func (r *Storage) flush() error {
	if !r.needFlush {
		log.Debug().Str("status", "flush.no.need").Send()
		return nil
	}

	log.Info().Str("status", "flush.begin").Send()

	defer func() { r.needFlush = false }()
	r.flushMu.Lock()
	defer r.flushMu.Unlock()

	jsonMap, err := json.Marshal(r.mapping)
	if err != nil {
		return errors.WithStack(err)
	}

	f, err := os.OpenFile(r.filename, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = f.Write(jsonMap)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Info().Str("status", "flush.end").Int("len", len(r.mapping)).Send()

	return nil
}
