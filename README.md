# AWS GIFTCARD API

API for Google Cloud Storage, to upload objects to your bucket, download objects from your bucket, and remove objects from your bucket

## API reference

 - [Amazon GiftCard API](https://www.amazon.co.uk/gift-card-api/b?ie=UTF8&node=11938308031)

## Authentication

Authentication is implemented using API token, you can either configure a [Secret in Kubernetes](https://kubernetes.io/docs/concepts/configuration/secret/) 

you can generate your API [Amazon GiftCard API](https://www.amazon.co.uk/gift-card-api/b?ie=UTF8&node=11938308031)


## Error hanlding

Empty Payload or malformed JSON would result in error reponse.

- Technical error :  `{"status":400,"message":"EOF"}`
- Malformed JSON `{"status":400,"message":"invalid character 'a' looking for beginning of object key string"}`

## Sample Input/Output

- Request payload

```
{
	"amount" : 51
}

```
- Response

```
{
    "cardInfo": {
        "cardNumber": null,
        "cardStatus": "Fulfilled",
        "expirationDate": null,
        "value": {
            "amount": 51,
            "currencyCode": "GBP"
        }
    },
    "creationRequestId": "XXXXXXXX",
    "gcClaimCode": "XXXX-XXXXXX-XXXX",
    "gcExpirationDate": "Tue Dec 07 23:59:59 UTC 2027",
    "gcId": "A3NIVER1GCA0E2",
    "status": "SUCCESS"
}

-----
if amount is not valid

{
    "agcodResponse": {
        "__type": "CreateGiftCardResponse:http://internal.amazon.com/coral/com.amazonaws.agcod/",
        "cardInfo": null,
        "creationRequestId": "XXXXXXXX",
        "gcClaimCode": null,
        "gcExpirationDate": null,
        "gcId": null,
        "status": "FAILURE"
    },
    "errorCode": "F200",
    "errorType": "InvalidAmountValue",
    "message": "Amount must be larger than 0"
}
----------

```


## Example Usage

## 1.  Deploy as Fission Functions

First, set up your fission deployment with the go environment.

```
fission env create --name go-env --image fission/go-env:1.8.1
```

To ensure that you build functions using the same version as the
runtime, fission provides a docker image and helper script for
building functions.


- Download the build helper script

```
$ curl https://raw.githubusercontent.com/fission/fission/master/environments/go/builder/go-function-build > go-function-build
$ chmod +x go-function-build
```

- Build the function as a plugin. Outputs result to 'function.so'

`$ go-function-build amazon-giftcard.go`

- Upload the function to fission

`$ fission function create --name amazon-giftcard --env go-env --package function.so`

- Map /amazon-giftcard to the amazon-giftcard function

`$ fission route create --method POST --url /amazon-giftcard --function amazon-giftcard`

- Run the function

```$ curl -d `sample request` -H "Content-Type: application/json" -X POST http://$FISSION_ROUTER/amazon-giftcard```

## 2. Deploy as AWS Lambda

> to be updated
