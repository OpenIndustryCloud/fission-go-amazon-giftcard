package main

/*
This API will connect to Amazon Gift Card Ondemand Service
to generate GIFT Cards

INPUT
{"amount" = 20}

OUT
{"cardInfo":{"cardNumber":null,"cardStatus":"Fulfilled","expirationDate":null,"value":{"amount":10.0,"currencyCode":"GBP"}},"creationRequestId":"LegenB5ta6171061710","gcClaimCode":"UL7G-6VNS6V-USPU","gcExpirationDate":"Tue Nov 30 23:59:59 UTC 2027","gcId":"A300FFD9R1809J","status":"SUCCESS"}

*/
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rubyphunk/goamz/aws"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ACCEPT_HEADER        = "accept"
	CONTENT_HEADER       = "content-type"
	HOST_HEADER          = "host"
	XAMZDATE_HEADER      = "x-amz-date"
	XAMZTARGET_HEADER    = "x-amz-target"
	AUTHORIZATION_HEADER = "Authorization"

	//Static format parameters
	DATE_FORMAT   = "yyyyMMdd'T'HHmmss'Z'"
	DATE_TIMEZONE = "UTC"
	UTF8_CHARSET  = "UTF-8"

	//Signature calculation related parameters
	HMAC_SHA256_ALGORITHM = "HmacSHA256"
	HASH_SHA256_ALGORITHM = "SHA-256"
	AWS_SHA256_ALGORITHM  = "AWS4-HMAC-SHA256"
	KEY_QUALIFIER         = "AWS4"
	TERMINATION_STRING    = "aws4_request"

	//User and instance parameters
	awsKeyID     = "" // Your KeyID
	awsSecretKey = "" // Your Key
	namespace    = "default"
	secretName   = "agcod-secret"

	//Service and target (API) parameters
	regionName  = aws.EUWest.Name //"eu-west-1" !
	region      = aws.EUWest
	serviceName = "AGCODService"

	//Payload parameters
	partnerID  = "Legen"
	requestID  = partnerID + "B5ta"
	cardNumber = ""

	//Additional payload parameters for ActivateGiftCard
	currencyCode = "GBP"
	amount       = "" //default

	//Additional payload parameters for CancelGiftCard
	gcId = ""

	//Additional payload parameters for GetGiftCardActivityPage
	pageIndex    = "0"
	pageSize     = "1"
	utcStartDate = "" //"yyyy-MM-ddTHH:mm:ss eg. 2013-06-01T23:10:10"
	utcEndDate   = "" //"yyyy-MM-ddTHH:mm:ss eg. 2013-06-01T23:15:10"
	showNoOps    = "true"

	//Parameters that specify what format the payload should be in and what fields will be in the payload
	//msgPayloadType = "XML"
	msgPayloadType = "JSON"

	//serviceOperation = ActivationStatusCheck;
	//serviceOperation = ActivateGiftCard;
	//serviceOperation = DeactivateGiftCard;
	serviceOperation = CreateGiftCard
	//serviceOperation = CancelGiftCard;
	//serviceOperation = GetGiftCardActivityPage;

	//Parameters used in the message header
	dateTimeString = "" // use null for service call. Code below will add current x-amz-date time stamp
	host           = "agcod-v2-eu-gamma.amazon.com"
	protocol       = "https"
	queryString    = ""

	contentType   = ""
	requestURI    = "/" + serviceOperation
	serviceTarget = "com.amazonaws.agcod.AGCODService." + serviceOperation
	hostName      = protocol + "://" + host + requestURI
)

/**
* Types of API this sample code supports
and supported formats for the payload
*/
const (
	ActivateGiftCard        = "ActivateGiftCard"
	DeactivateGiftCard      = "DeactivateGiftCard"
	ActivationStatusCheck   = "ActivationStatusCheck"
	CreateGiftCard          = "CreateGiftCard"
	CancelGiftCard          = "CancelGiftCard"
	GetGiftCardActivityPage = "GetGiftCardActivityPage"
	JSON                    = "JSON"
	XML                     = "XML"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	//Initialize whole payload in the specified format for the given operation and set additional headers based on these settings
	var inputData InputData
	err := json.NewDecoder(r.Body).Decode(&inputData)
	if err == io.EOF || err != nil {
		createErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	//get API keys
	getAPIKeys(w)

	println("Create Gift Card for - ", inputData.Amount)
	strAmount := strconv.Itoa(inputData.Amount)

	//Dynamic request ID -
	rGen := rand.New(rand.NewSource(99))
	randomNum := rGen.Intn(1000)

	requestID = requestID + strconv.Itoa(randomNum) + strAmount
	payLoadString := populatePayload(strAmount)

	// Get auth from env vars, pass explictly
	println(awsKeyID + " || " + awsSecretKey)
	auth, err := aws.GetAuth(awsKeyID, awsSecretKey, "", time.Time{})
	if err != nil {
		fmt.Println(err)
	}

	// Create a signer with the auth, name of the service, and aws region
	signer := aws.NewV4Signer(auth, serviceName, region)

	// Create a request
	println("requesting host : " + hostName)
	req, err := http.NewRequest("POST", hostName, strings.NewReader(payLoadString))
	if err != nil {
		fmt.Println(err)
	}

	// Date or x-amz-date header is required to sign a request //http.TimeFormat
	req.Header.Add("Date", time.Now().UTC().Format("yyyyMMdd'T'HHmmss'Z'"))
	req.Header.Add(ACCEPT_HEADER, contentType)
	req.Header.Add(CONTENT_HEADER, contentType)
	req.Header.Add(HOST_HEADER, host)
	req.Header.Add(XAMZDATE_HEADER, dateTimeString)
	req.Header.Add(XAMZTARGET_HEADER, serviceTarget)

	// Sign the request
	signer.Sign(req)

	// Issue signed request
	//res, _ := http.DefaultClient.Do(req)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

	w.Header().Set("content-type", "application/json")
	w.Write([]byte(string(body)))

}

func getAPIKeys(w http.ResponseWriter) {
	println("[CONFIG] Reading Env variables")

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secret, err := clientset.Core().Secrets(namespace).Get(secretName, meta_v1.GetOptions{})

	awsKeyID = string(secret.Data["awsKeyID"])
	awsKeyID = strings.Replace(awsKeyID, "\n", "", -1)

	awsSecretKey = string(secret.Data["awsSecretKey"])
	awsSecretKey = strings.Replace(awsSecretKey, "\n", "", -1)
	//awsKeyID = "AKIAILRAHUK2LV3FRKRA" // Your KeyID
	//awsSecretKey = "lR9q9le9/h69t/9ByrbrWdmcjYKTCPryXurWyJHC" // Your Key

	println(awsKeyID + " || " + awsSecretKey)

	if len(awsKeyID) == 0 {
		createErrorResponse(w, "Missing API Key", http.StatusBadRequest)
	}
	if len(awsSecretKey) == 0 {
		createErrorResponse(w, "Missing API Password", http.StatusBadRequest)
	}

}

type InputData struct {
	Amount int `json:"amount"`
}

/**
 * Sets the payload to be the requested encoding and creates the payload based on the static parameters.
 *
 * @param payload - The payload to be set that is sent to the AGCOD service
 *
 */
func populatePayload(amount string) string {

	if amount == "" {
		println("generating gift card for 1 GBP")
		amount = "1.00" //TODO
	}

	payload := new(bytes.Buffer)

	//Set payload based on operation and format
	if msgPayloadType == XML {
		contentType = "application/x-www-form-urlencoded;charset=UTF-8"
		switch serviceOperation {
		case ActivateGiftCard:
			payload.WriteString("<ActivateGiftCardRequest><activationRequestId>")
			payload.WriteString(requestID)
			payload.WriteString("</activationRequestId> <partnerId>")
			payload.WriteString(partnerID)
			payload.WriteString("</partnerId><cardNumber>")
			payload.WriteString(cardNumber)
			payload.WriteString("</cardNumber><value><currencyCode>")
			payload.WriteString(currencyCode)
			payload.WriteString("</currencyCode><amount>")
			payload.WriteString(amount)
			payload.WriteString("</amount></value></ActivateGiftCardRequest>")

		case DeactivateGiftCard:
			payload.WriteString("<DeactivateGiftCardRequest><activationRequestId>")
			payload.WriteString(requestID)
			payload.WriteString("</activationRequestId> <partnerId>")
			payload.WriteString(partnerID)
			payload.WriteString("</partnerId><cardNumber>")
			payload.WriteString(cardNumber)
			payload.WriteString("</cardNumber></DeactivateGiftCardRequest>")

		case ActivationStatusCheck:
			payload.WriteString("<ActivationStatusCheckRequest><statusCheckRequestId>")
			payload.WriteString(requestID)
			payload.WriteString("</statusCheckRequestId> <partnerId>")
			payload.WriteString(partnerID)
			payload.WriteString("</partnerId><cardNumber>")
			payload.WriteString(cardNumber)
			payload.WriteString("</cardNumber></ActivationStatusCheckRequest>")

		case CreateGiftCard:
			payload.WriteString("<CreateGiftCardRequest><creationRequestId>")
			payload.WriteString(requestID)
			payload.WriteString("</creationRequestId><partnerId>")
			payload.WriteString(partnerID)
			payload.WriteString("</partnerId><value><currencyCode>")
			payload.WriteString(currencyCode)
			payload.WriteString("</currencyCode><amount>")
			payload.WriteString(amount)
			payload.WriteString("</amount></value></CreateGiftCardRequest>")

		case CancelGiftCard:
			payload.WriteString("<CancelGiftCardRequest><creationRequestId>")
			payload.WriteString(requestID)
			payload.WriteString("</creationRequestId><partnerId>")
			payload.WriteString(partnerID)
			payload.WriteString("</partnerId><gcId>")
			payload.WriteString(gcId)
			payload.WriteString("</gcId></CancelGiftCardRequest>")

		case GetGiftCardActivityPage:
			payload.WriteString("<GetGiftCardActivityPageRequest><requestId>")
			payload.WriteString(requestID)
			payload.WriteString("</requestId> <partnerId>")
			payload.WriteString(partnerID)
			payload.WriteString("</partnerId><utcStartDate>")
			payload.WriteString(utcStartDate)
			payload.WriteString("</utcStartDate><utcEndDate>")
			payload.WriteString(utcEndDate)
			payload.WriteString("</utcEndDate><pageIndex>")
			payload.WriteString(pageIndex)
			payload.WriteString("</pageIndex><pageSize>")
			payload.WriteString(pageSize)
			payload.WriteString("</pageSize><showNoOps>")
			payload.WriteString(showNoOps)
			payload.WriteString("</showNoOps></GetGiftCardActivityPageRequest>")

		default:
			// throw new IllegalArgumentException();
		}
	} else if msgPayloadType == JSON {
		contentType = "application/json"
		switch serviceOperation {
		case ActivateGiftCard:
			payload.WriteString("{\"activationRequestId\": \"")
			payload.WriteString(requestID)
			payload.WriteString("\", \"partnerId\": \"")
			payload.WriteString(partnerID)
			payload.WriteString("\", \"cardNumber\": \"")
			payload.WriteString(cardNumber)
			payload.WriteString("\", \"value\": {\"currencyCode\": \"")
			payload.WriteString(currencyCode)
			payload.WriteString("\", \"amount\": ")
			payload.WriteString(amount)
			payload.WriteString("}}")
			break
		case DeactivateGiftCard:
			payload.WriteString("{\"activationRequestId\": \"")
			payload.WriteString(requestID)
			payload.WriteString("\", \"partnerId\": \"")
			payload.WriteString(partnerID)
			payload.WriteString("\", \"cardNumber\": \"")
			payload.WriteString(cardNumber)
			payload.WriteString("\"}")
			break
		case ActivationStatusCheck:
			payload.WriteString("{\"statusCheckRequestId\": \"")
			payload.WriteString(requestID)
			payload.WriteString("\", \"partnerId\": \"")
			payload.WriteString(partnerID)
			payload.WriteString("\", \"cardNumber\": \"")
			payload.WriteString(cardNumber)
			payload.WriteString("\"}")
			break
		case CreateGiftCard:
			payload.WriteString("{\"creationRequestId\": \"")
			payload.WriteString(requestID)
			payload.WriteString("\", \"partnerId\": \"")
			payload.WriteString(partnerID)
			payload.WriteString("\", \"value\": {\"currencyCode\": \"")
			payload.WriteString(currencyCode)
			payload.WriteString("\", \"amount\": ")
			payload.WriteString(amount)
			payload.WriteString("}}")
			break
		case CancelGiftCard:
			payload.WriteString("{\"creationRequestId\": \"")
			payload.WriteString(requestID)
			payload.WriteString("\", \"partnerId\": \"")
			payload.WriteString(partnerID)
			payload.WriteString("\", \"gcId\": \"")
			payload.WriteString(gcId)
			payload.WriteString("\"}")
			break
		case GetGiftCardActivityPage:
			payload.WriteString("{\"requestId\": \"")
			payload.WriteString(requestID)
			payload.WriteString("\", \"partnerId\": \"")
			payload.WriteString(partnerID)
			payload.WriteString("\", \"utcStartDate\": \"")
			payload.WriteString(utcStartDate)
			payload.WriteString("\", \"utcEndDate\": \"")
			payload.WriteString(utcEndDate)
			payload.WriteString("\", \"pageIndex\": ")
			payload.WriteString(pageIndex)
			payload.WriteString(", \"pageSize\": ")
			payload.WriteString(pageSize)
			payload.WriteString(", \"showNoOps\": \"")
			payload.WriteString(showNoOps)
			payload.WriteString("\"}")
			break
		default:
			//throw new IllegalArgumentException();
		}
	} else {
		println("hiii")
		// throw new IllegalArgumentException();
	}

	reqPayload, _ := ioutil.ReadAll(payload)
	println(string(reqPayload))

	return string(reqPayload)
}

func main() {
	println("staritng app..")
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8084", nil)
}

func createErrorResponse(w http.ResponseWriter, message string, status int) {
	errorJSON, _ := json.Marshal(&Error{
		Status:  status,
		Message: message})
	//Send custom error message to caller
	w.WriteHeader(status)
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(errorJSON))
}

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
