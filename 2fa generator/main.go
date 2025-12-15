package main

import (
	"fmt"

	"github.com/pquerna/otp/totp"
)

func main() {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoExile",
		AccountName: "admin",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Your 2FA Secret (Base32): %s\n", key.Secret())
	fmt.Printf("Provisioning URL: %s\n", key.URL())
}
