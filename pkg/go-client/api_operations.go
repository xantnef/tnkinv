/*
 * OpenAPI
 *
 * tinkoff.ru/invest OpenAPI.
 *
 * API version: 1.0.0
 * Contact: n.v.melnikov@tinkoff.ru
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

import (
	"context"
	"io/ioutil"
	"net/url"
	"strings"
)

// Linger please
var (
	_ context.Context
)

type OperationsApiService service

/*
OperationsApiService Получение списка операций
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param from Начало временного промежутка
 * @param to Конец временного промежутка
 * @param optional nil or *OperationsGetOpts - Optional Parameters:
     * @param "Figi" (optional.Interface of interface{}) -  Figi инструмента для фильтрации
     * @param "BrokerAccountId" (optional.Interface of interface{}) -  Номер счета (по умолчанию - Тинькофф)


*/

type OptionalInterface interface {
	IsSet() bool
	Value() interface{}
}

type OperationsGetOpts struct {
	//Figi optional.Interface
	Figi, BrokerAccountId OptionalInterface
}

func (a *OperationsApiService) OperationsGet(ctx context.Context, from interface{}, to interface{}, localVarOptionals *OperationsGetOpts) ([]byte, error) {
	var (
		localVarHttpMethod = strings.ToUpper("Get")
		localVarPostBody   interface{}
		localVarFileName   string
		localVarFileBytes  []byte
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/operations"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	localVarQueryParams.Add("from", parameterToString(from, ""))
	localVarQueryParams.Add("to", parameterToString(to, ""))
	if localVarOptionals != nil && localVarOptionals.Figi.IsSet() {
		localVarQueryParams.Add("figi", parameterToString(localVarOptionals.Figi.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.BrokerAccountId.IsSet() {
		localVarQueryParams.Add("brokerAccountId", parameterToString(localVarOptionals.BrokerAccountId.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return nil, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return nil, err
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}

		return nil, newErr
	}

	return localVarBody, nil
}
