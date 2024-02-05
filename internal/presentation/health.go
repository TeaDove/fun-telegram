package presentation

import (
	"github.com/teadove/goteleout/internal/service/job"
	"net/http"
)

func healthServer(job *job.Service) {
	http.HandleFunc("/health", job.ApiHealth)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
}
