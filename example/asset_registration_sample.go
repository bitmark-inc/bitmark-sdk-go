package main

import (
	"fmt"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"io/ioutil"
)

func registerAsset(registrant account.Account, assetName string, assetFilePath string, metadata map[string]string) (string, error) {
	assetFileContent, err := ioutil.ReadFile(assetFilePath)

	if err != nil {
		fmt.Println("Can not read file!")
		return "", err
	}

	registrationParams, _ := asset.NewRegistrationParams(assetName, metadata)
	registrationParams.SetFingerprint([]byte(assetFileContent))
	registrationParams.Sign(registrant)

	assetId, err := asset.Register(registrationParams)

	return assetId, err
}
