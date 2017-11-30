# fission-go-amazon-giftcard
This API provide Implementation if on demand Generation Amazon Gift card


## Export Fission Variables
export FISSION_URL=http://$(kubectl --namespace fission get svc controller -o=jsonpath='{..ip}')
export FISSION_ROUTER=$(kubectl --namespace fission get svc router -o=jsonpath='{..ip}')

## Create Funtions and Route
fission function create --name development-amazon-giftcard --env go-env --deploy function.so 
fission route create --method POST --url /development/amazon-giftcard --function   development-amazon-giftcard

## Test Amazon Gift Card API
curl -d `{"amount":2}` -H "Content-Type: application/json" -X POST http://fission.landg.madeden.net/development/amazon-giftcard