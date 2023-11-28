package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/container"
	"github.com/teadove/goteleout/internal/utils"
	"net/http"
)

var combatContainer container.Container

func init() {
	combatContainer = container.MustNewCombatContainer()
}

func StartApp() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.POST("tg-files", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, errors.WithStack(err))
			return
		}
		if file == nil {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("file is nil"))
			return
		}
		err = combatContainer.Presentation.Upload(c.Request.Context(), file)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, errors.WithStack(err))
			return
		}

		c.String(http.StatusOK, "ok")
	})

	log.Info().Str("status", "web.server.starting").Send()

	err := r.Run("localhost:8000")
	utils.Check(err)
}

func main() {
	StartApp()
}
