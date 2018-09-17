package account

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

type valid struct {
	seed   string
	phrase string
}

var (
	testnetData = valid{
		"5XEECt18HGBGNET1PpxLhy5CsCLG9jnmM6Q8QGF4U2yGb1DABXZsVeD",
		"accident syrup inquiry you clutch liquid fame upset joke glow best school repeat birth library combine access camera organ trial crazy jeans lizard science",
	}

	livenetData = valid{
		"5XEECqWqA47qWg86DR5HJ29HhbVqwigHUAhgiBMqFSBycbiwnbY639s",
		"ability panel leave spike mixture token voice certain today market grief crater cruise smart camera palm wheat rib swamp labor bid rifle piano glass",
	}
)

func TestTestnetAccount(t *testing.T) {
	acct1, _ := FromSeed(testnetData.seed)
	if acct1.Network() != sdk.Testnet {
		t.Fail()
	}

	acct2, _ := FromRecoveryPhrase(strings.Split(testnetData.phrase, " "))
	if acct2.Network() != sdk.Testnet {
		t.Fail()
	}

	if strings.Join(acct1.RecoveryPhrase(), " ") != testnetData.phrase {
		t.Fail()
	}

	if acct2.Seed() != testnetData.seed {
		t.Fail()
	}

	if acct1.AccountNumber() != acct2.AccountNumber() {
		t.Fail()
	}
}

func TestLivenetAccount(t *testing.T) {
	acct1, err := FromSeed(livenetData.seed)
	if err != nil {
		t.Fatal(err)
	}
	if acct1.Network() != sdk.Livenet {
		t.Fail()
	}

	acct2, err := FromRecoveryPhrase(strings.Split(livenetData.phrase, " "))
	if err != nil {
		t.Fatal(err)
	}
	if acct2.Network() != sdk.Livenet {
		t.Fatal("wrong network")
	}

	if strings.Join(acct1.RecoveryPhrase(), " ") != livenetData.phrase {
		t.Fatalf("wrong recovery phrase:\nactual = %s\nexpected = %s", strings.Join(acct1.RecoveryPhrase(), " "), livenetData.phrase)
	}

	if acct2.Seed() != livenetData.seed {
		t.Fatal("wrong seed")
	}

	if acct1.AccountNumber() != acct2.AccountNumber() {
		t.Fatal("wrong account number")
	}
}

func TestParseAccountNumber(t *testing.T) {
	accountNumber := "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog"
	network, pubkey, err := ParseAccountNumber(accountNumber)
	if err != nil {
		t.Fatal(err)
	}

	if network != sdk.Testnet {
		t.Log(network)
		t.Fatal("wrong network")
	}

	fmt.Println(hex.EncodeToString(pubkey))
}
