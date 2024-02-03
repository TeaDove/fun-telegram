package presentation

import (
	"github.com/teadove/goteleout/internal/presentation/telegram"
	"net/http"
)

func healthServer(presentation *telegram.Presentation) {
	http.HandleFunc("/health", presentation.ApiHealth)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
}
