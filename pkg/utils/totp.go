package utils

import (
	"bytes"
	"encoding/base64"
	"image/png"

	"github.com/pquerna/otp/totp"
)

func GenerateTOTPSecret(issuer, accountEmail string) (secret string, qrBase64 string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountEmail,
	})
	if err != nil {
		return "", "", err
	}

	img, err := key.Image(256, 256)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}

	return key.Secret(), base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func VerifyTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
