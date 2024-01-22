package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
)

func (r *Presentation) locationCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	//if len(update.EffectiveMessage.Message.Message) < 10 {
	//	_, err := ctx.Reply(update, "Err: need too pass ip v4/v6 address or domain", nil)
	//	if err != nil {
	//		return errors.WithStack(err)
	//	}
	//}

	//ipAddress := update.EffectiveMessage.Message.Message[10:]

	location, err := r.ipLocator.GetLocation(ctx, input.Text)
	if err != nil {
		_, err = ctx.Reply(update, fmt.Sprintf("Err: %s", location.Message), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	log.Info().Str("status", "ip.requested").Interface("location", location).Send()

	geoLocationString := fmt.Sprintf("%f%%2C%f", location.Lon, location.Lat)
	geoLocationReversedString := fmt.Sprintf("%f%%2C%f", location.Lat, location.Lon)

	stylingOptions := []styling.StyledTextOption{
		styling.Plain(fmt.Sprintf("Requested name: %s\n\n", location.Query)),
		styling.Plain("Country: "), styling.Bold(location.Country), styling.Plain("\n"),
		styling.Plain(fmt.Sprintf("Region: %s\n", location.RegionName)),
		styling.Plain(fmt.Sprintf("City: %s\n", location.City)),
		styling.Plain(fmt.Sprintf("Timezone: %s\n", location.Timezone)),
		styling.Plain(fmt.Sprintf("Zip: %s\n", location.Zip)),
		styling.Plain("Location: "), styling.Code(fmt.Sprintf("(%f, %f)", location.Lat, location.Lon)), styling.Plain(" "),
		styling.TextURL("yandex", fmt.Sprintf("https://yandex.ru/maps/213/moscow/?ll=%s&mode=whatshere&whatshere%%5Bpoint%%5D=%s&whatshere%%5Bzoom%%5D=15.41&z=15.41", geoLocationString, geoLocationString)), styling.Plain(", "),
		styling.TextURL("google", fmt.Sprintf("https://www.google.com/maps/place/%s", geoLocationReversedString)), styling.Plain("\n\n"),
		styling.Plain(fmt.Sprintf("ISP: %s\n", location.Isp)),
		styling.Plain(fmt.Sprintf("ORG: %s\n", location.Org)),
		styling.Plain(fmt.Sprintf("AS: %s\n", location.As)),
	}

	_, err = ctx.Reply(update, stylingOptions, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
