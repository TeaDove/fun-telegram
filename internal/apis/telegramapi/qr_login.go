package telegramapi

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/cockroachdb/errors"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/rs/zerolog"
	"golang.org/x/term"
	"rsc.io/qr"
)

const (
	sessionPath = ".mtproto"
	qrQuietZone = 4
)

type qrTerminal struct {
	input  *os.File
	output *os.File
}

func LoginQR(ctx context.Context, input, output *os.File) error {
	if !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		return errors.New("run go run . login-qr in an interactive terminal to authorize this session")
	}

	settings := utils.AppSettings.Telegram
	if settings.AppID <= 0 || settings.AppHash == "" {
		return errors.New("set FUN_TELEGRAM__APP_ID and FUN_TELEGRAM__APP_HASH before QR login")
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	_, storage, err := sessionMaker.NewSessionStorage(ctx, sessionMaker.SqlSession(sqlite.Open(sessionPath)), false)
	if err != nil {
		return errors.Wrap(err, "open Telegram session")
	}

	dispatcher := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(dispatcher)
	client := telegram.NewClient(settings.AppID, settings.AppHash, telegram.Options{
		UpdateHandler:  dispatcher,
		SessionStorage: storage,
	})
	terminal := qrTerminal{input: input, output: output}

	err = client.Run(ctx, func(ctx context.Context) error {
		return authenticateQR(ctx, client, loggedIn, terminal)
	})
	if err != nil {
		return errors.Wrap(err, "Telegram QR login")
	}

	zerolog.Ctx(ctx).Info().Msg("telegram.session.authorized")

	return nil
}

func authenticateQR(
	ctx context.Context,
	client *telegram.Client,
	loggedIn qrlogin.LoggedIn,
	terminal qrTerminal,
) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return errors.Wrap(err, "check Telegram session")
	}

	if status.Authorized {
		return nil
	}

	_, err = terminal.output.WriteString("\x1b[?1049h")
	if err != nil {
		return errors.Wrap(err, "open QR screen")
	}

	defer func() {
		_, restoreErr := terminal.output.WriteString("\x1b[0m\x1b[?1049l")
		if restoreErr != nil {
			zerolog.Ctx(ctx).Warn().Err(restoreErr).Msg("restore terminal")
		}
	}()

	_, err = client.QR().Auth(ctx, loggedIn, terminal.show)
	if tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
		return terminal.enterPassword(ctx, client)
	}

	return errors.Wrap(err, "authorize via QR")
}

func (t qrTerminal) enterPassword(ctx context.Context, client *telegram.Client) error {
	_, err := t.output.WriteString("\x1b[2J\x1b[HTelegram two-step verification password: ")
	if err != nil {
		return errors.Wrap(err, "show password prompt")
	}

	password, err := term.ReadPassword(int(t.input.Fd()))
	if err != nil {
		return errors.Wrap(err, "read Telegram password")
	}

	defer clear(password)

	_, err = client.Auth().Password(ctx, string(password))

	return errors.Wrap(err, "verify Telegram password")
}

func (t qrTerminal) show(_ context.Context, token qrlogin.Token) error {
	code, err := qr.Encode(token.URL(), qr.M)
	if err != nil {
		return errors.Wrap(err, "encode Telegram QR")
	}

	columns, rows, err := term.GetSize(int(t.output.Fd()))
	if err != nil {
		return errors.Wrap(err, "read terminal size")
	}

	if columns < code.Size+2*qrQuietZone || rows < (code.Size+2*qrQuietZone+1)/2+2 {
		return errors.New("enlarge the terminal window for the Telegram QR code")
	}

	_, err = t.output.WriteString(qrScreen(code))

	return errors.Wrap(err, "display Telegram QR")
}

func qrScreen(code *qr.Code) string {
	var screen strings.Builder

	screen.WriteString("\x1b[2J\x1b[HIn Telegram, open Settings > Devices > Link Desktop Device.\r\n")

	for y := -qrQuietZone; y < code.Size+qrQuietZone; y += 2 {
		for x := -qrQuietZone; x < code.Size+qrQuietZone; x++ {
			if code.Black(x, y) {
				screen.WriteString("\x1b[30m")
			} else {
				screen.WriteString("\x1b[97m")
			}

			if code.Black(x, y+1) {
				screen.WriteString("\x1b[40m")
			} else {
				screen.WriteString("\x1b[107m")
			}

			screen.WriteRune('▀')
		}

		screen.WriteString("\x1b[0m\r\n")
	}

	return screen.String()
}
