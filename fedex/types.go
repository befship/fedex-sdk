package fedex

import fedexapi "github.com/befship/fedex-sdk/internal/generated/fedexapi"

type (
	ClientWithResponses = fedexapi.ClientWithResponses
	RequestEditorFn     = fedexapi.RequestEditorFn

	XCustomerTransactionId = fedexapi.XCustomerTransactionId
	XLocale                = fedexapi.XLocale

	OauthTokenFormdataRequestBody = fedexapi.OauthTokenFormdataRequestBody
	OauthTokenFormdataBody        = fedexapi.OauthTokenFormdataBody

	ResolveAddressJSONRequestBody  = fedexapi.ResolveAddressJSONRequestBody
	ResolveAddressParams           = fedexapi.ResolveAddressParams
	RateQuotesJSONRequestBody      = fedexapi.RateQuotesJSONRequestBody
	RateQuotesParams               = fedexapi.RateQuotesParams
	CreateShipmentJSONRequestBody  = fedexapi.CreateShipmentJSONRequestBody
	CreateShipmentParams           = fedexapi.CreateShipmentParams
	CancelShipmentJSONRequestBody  = fedexapi.CancelShipmentJSONRequestBody
	CancelShipmentParams           = fedexapi.CancelShipmentParams
	ShipmentResultsJSONRequestBody = fedexapi.ShipmentResultsJSONRequestBody
	ShipmentResultsParams          = fedexapi.ShipmentResultsParams
	TrackByNumbersJSONRequestBody  = fedexapi.TrackByNumbersJSONRequestBody
	TrackByNumbersParams           = fedexapi.TrackByNumbersParams
)

const ClientCredentials = fedexapi.ClientCredentials
