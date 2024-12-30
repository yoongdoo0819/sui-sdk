package main

import (
	"context"
	"fmt"

	"encoding/json"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

type Config struct {
	MNEMONIC string `toml:"mnemonic"`
}

type RunInferenceReq struct {
	// Arguments []interface{} `json:"arguments"`
	In1 []string `json:"in1"`
	In2 []string `json:"in2"`
	In3 string   `json:"in3"`
}

type ResponseBody struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Response 구조체 (JSON 응답을 위한 구조체)
type Response struct {
	Message string `json:"message"`
}

func runInference(mnemonic string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// POST 요청만 허용
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 요청 바디 파싱
		var reqBody RunInferenceReq
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
			return
		}

		// Sui 클라이언트 초기화
		ctx := context.Background()
		cli := sui.NewSuiClient("https://sui-devnet-endpoint.blockvision.org")

		// 서명자 생성
		signerAccount, err := signer.NewSignertWithMnemonic(mnemonic)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create signer: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		priKey := signerAccount.PriKey
		gasObj := "0xfa6e9bf9f256f322330c56b8ad2b128c051f95d21e78482a64e8fd72eeea6bc2"

		req := []interface{}{}
		req = append(req, reqBody.In1)
		req = append(req, reqBody.In2)
		req = append(req, reqBody.In3)

		// MoveCall 실행
		rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
			Signer:          signerAccount.Address,
			PackageObjectId: "0x1a17fdd92c9d989f5200900302df4901d66fd04062a34eccbb83f085230838d7",
			Module:          "Inference",
			Function:        "run",
			TypeArguments:   []interface{}{},
			Arguments:       req,
			Gas:             &gasObj,
			GasBudget:       "2000000000",
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("MoveCall error: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// 트랜잭션 서명 및 실행
		rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
			TxnMetaData: rsp,
			PriKey:      priKey,
			Options: models.SuiTransactionBlockOptions{
				ShowInput:    true,
				ShowRawInput: true,
				ShowEffects:  true,
			},
			RequestType: "WaitForLocalExecution",
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Transaction execution error: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		utils.PrettyPrint(rsp2)

		// JSON 응답 반환
		response := ResponseBody{
			Message: "Transaction executed successfully",
			Data:    rsp2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	}
}

// Hello 핸들러 (GET 요청 처리)
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := Response{Message: "Hello, World!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {

	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// /run 엔드포인트 등록
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/run", runInference(config.MNEMONIC))

	// 서버 시작
	port := "8080"
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	// var ctx = context.Background()
	// var cli = sui.NewSuiClient("https://sui-devnet-endpoint.blockvision.org")

	// fmt.Println("cli", cli)
	// signerAccount, err := signer.NewSignertWithMnemonic("")

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// priKey := signerAccount.PriKey
	// fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

	// gasObj := "0xfa6e9bf9f256f322330c56b8ad2b128c051f95d21e78482a64e8fd72eeea6bc2"

	// a := []interface{}{}
	// b := []string{
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
	// 	"44", "30", "37", "0", "0", "0", "50", "3", "89", "0",
	// 	"0", "0", "0", "0", "99", "5", "0", "0", "0", "0",
	// 	"63", "5", "46", "0", "0", "0", "0", "97", "85", "0",
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0",
	// }
	// c := []string{
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
	// 	"0", "0", "0", "0", "0", "0", "0", "0", "0",
	// }
	// d := "2"
	// a = append(a, b)
	// a = append(a, c)
	// a = append(a, d)
	// // in1 := []uint64{3}
	// rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
	// 	Signer:          signerAccount.Address,
	// 	PackageObjectId: "0x1a17fdd92c9d989f5200900302df4901d66fd04062a34eccbb83f085230838d7", // "0x1c11747b58ed2c3817cadafff0d66138daf5ad734d051516df17012fb596b253", //"0x8d77eaf0b5503337f113fae8878e5a15ce1af0b13ea3e38327d2e280d4fa9679",
	// 	Module:          "Inference",                                                          // "FuncTest",                                                           //"model",
	// 	Function:        "run",                                                                // "test_vector",                                                        //"add2",
	// 	TypeArguments:   []interface{}{},
	// 	Arguments:       a,
	// 	Gas:             &gasObj,
	// 	GasBudget:       "2000000000",
	// })

	// if err != nil {
	// 	fmt.Println(">>", err.Error())
	// 	return
	// }

	// // see the successful transaction url: https://explorer.sui.io/txblock/CD5hFB4bWFThhb6FtvKq3xAxRri72vsYLJAVd7p9t2sR?network=testnet
	// rsp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
	// 	TxnMetaData: rsp,
	// 	PriKey:      priKey,
	// 	// only fetch the effects field
	// 	Options: models.SuiTransactionBlockOptions{
	// 		ShowInput:    true,
	// 		ShowRawInput: true,
	// 		ShowEffects:  true,
	// 	},
	// 	RequestType: "WaitForLocalExecution",
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// utils.PrettyPrint(rsp2)
}
