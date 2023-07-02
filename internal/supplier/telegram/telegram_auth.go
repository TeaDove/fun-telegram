package telegram

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
)

type terminalAuth struct{}

func (terminalAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (terminalAuth) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (terminalAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := terminal.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(code)), nil
}

func (terminalAuth) Phone(_ context.Context) (string, error) {
	fmt.Print("Enter phone: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	bytePwd, err := terminal.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePwd)), nil
}
