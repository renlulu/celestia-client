package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/namespace"
	"github.com/celestiaorg/celestia-app/pkg/user"
	blobtypes "github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
)

func getCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func main() {
	gasLimit := blobtypes.DefaultEstimateGas([]uint32{1})
	fmt.Println(gasLimit)

	keyringDir := os.ExpandEnv("/Users/xiaohuo/.celestia-light-mocha-4/keys/")
	// Create or load the keyring
	kr, err := keyring.New("xiaohuo", "file", keyringDir, strings.NewReader(""), getCodec())
	if err != nil {
		fmt.Println("Error creating or loading keyring:", err)
		return
	}
	err = DemoSubmitData("grpc-celestia-mocha.architectnodes.com:1443", kr)
	if err != nil {
		fmt.Println("submit data error: ", err.Error())
	}

}

// SubmitData is a demo function that shows how to use the signer to submit data
// to the blockchain directly via a celestia node. We can manage this keyring
// using the `celestia-appd keys` or `celestia keys` sub commands and load this
// keyring from a file and use it to programmatically sign transactions.
func DemoSubmitData(grpcAddr string, kr keyring.Keyring) error {
	// create an encoding config that can decode and encode all celestia-app
	// data structures.
	ecfg := encoding.MakeConfig(app.ModuleEncodingRegisters...)

	// create a connection to the grpc server on the consensus node.
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	// get the address of the account we want to use to sign transactions.
	rec, err := kr.Key("xiaohuo")
	if err != nil {
		return err
	}

	addr, err := rec.GetAddress()
	if err != nil {
		return err
	}

	// Setup the signer. This function will automatically query the relevant
	// account information such as sequence (nonce) and account number.
	signer, err := user.SetupSigner(context.TODO(), kr, conn, addr, ecfg)
	if err != nil {
		return err
	}

	ns := namespace.MustNewV0([]byte("1234567890"))
	fmt.Println("ns: ", ns)

	fmt.Println("namepace: ", len(ns.Bytes()))

	blob, err := blobtypes.NewBlob(ns, []byte("some data"), appconsts.ShareVersionZero)
	if err != nil {
		return err
	}

	gasLimit := blobtypes.DefaultEstimateGas([]uint32{uint32(len(blob.Data))})

	options := []user.TxOption{
		// here we're setting estimating the gas limit from the above estimated
		// function, and then setting the gas price to 0.1utia per unit of gas.
		user.SetGasLimitAndFee(gasLimit, 0.1),
	}

	// this function will submit the transaction and block until a timeout is
	// reached or the transaction is committed.
	resp, err := signer.SubmitPayForBlob(context.TODO(), []*tmproto.Blob{blob}, options...)
	if err != nil {
		return err
	}

	// check the response code to see if the transaction was successful.
	if resp.Code != 0 {
		// handle code
		fmt.Println(resp.Code, resp.Codespace, resp.RawLog)
	}

	// if we don't want to wait for the transaction to be confirmed, we can
	// manually sign and submit the transaction using the same package.
	blobTx, err := signer.CreatePayForBlob([]*tmproto.Blob{blob}, options...)
	if err != nil {
		return err
	}

	resp, err = signer.BroadcastTx(context.TODO(), blobTx)
	if err != nil {
		return err
	}

	// check the response code to see if the transaction was successful. Note
	// that this time we're not waiting for the transaction to be committed.
	// Therefore the code here is only from the consensus node's mempool.
	if resp.Code != 0 {
		// handle code
		fmt.Println(resp.Code, resp.Codespace, resp.RawLog)
	}

	return err
}
