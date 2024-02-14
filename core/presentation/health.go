package presentation

import (
	"net/http"

	"github.com/teadove/goteleout/core/service/job"
)

func healthServer(job *job.Service) {
	http.HandleFunc("/health", job.ApiHealth)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
}
