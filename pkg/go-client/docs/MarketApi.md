# \MarketApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**MarketBondsGet**](MarketApi.md#MarketBondsGet) | **Get** /market/bonds | Получение списка облигаций
[**MarketCandlesGet**](MarketApi.md#MarketCandlesGet) | **Get** /market/candles | Получение исторических свечей по FIGI
[**MarketCurrenciesGet**](MarketApi.md#MarketCurrenciesGet) | **Get** /market/currencies | Получение списка валютных пар
[**MarketEtfsGet**](MarketApi.md#MarketEtfsGet) | **Get** /market/etfs | Получение списка ETF
[**MarketOrderbookGet**](MarketApi.md#MarketOrderbookGet) | **Get** /market/orderbook | Получение исторических стакана по FIGI
[**MarketSearchByFigiGet**](MarketApi.md#MarketSearchByFigiGet) | **Get** /market/search/by-figi | Получение инструмента по FIGI
[**MarketSearchByTickerGet**](MarketApi.md#MarketSearchByTickerGet) | **Get** /market/search/by-ticker | Получение инструмента по тикеру
[**MarketStocksGet**](MarketApi.md#MarketStocksGet) | **Get** /market/stocks | Получение списка акций


# **MarketBondsGet**
> MarketBondsGet(ctx, )
Получение списка облигаций

### Required Parameters
This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketCandlesGet**
> MarketCandlesGet(ctx, figi, from, to, interval)
Получение исторических свечей по FIGI

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **figi** | [**interface{}**](.md)| FIGI | 
  **from** | [**interface{}**](.md)| Начало временного промежутка | 
  **to** | [**interface{}**](.md)| Конец временного промежутка | 
  **interval** | [**interface{}**](.md)| Интервал свечи | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketCurrenciesGet**
> MarketCurrenciesGet(ctx, )
Получение списка валютных пар

### Required Parameters
This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketEtfsGet**
> MarketEtfsGet(ctx, )
Получение списка ETF

### Required Parameters
This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketOrderbookGet**
> MarketOrderbookGet(ctx, figi, depth)
Получение исторических стакана по FIGI

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **figi** | [**interface{}**](.md)| FIGI | 
  **depth** | [**interface{}**](.md)| Глубина стакана [1..20] | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketSearchByFigiGet**
> MarketSearchByFigiGet(ctx, figi)
Получение инструмента по FIGI

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **figi** | [**interface{}**](.md)| FIGI | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketSearchByTickerGet**
> MarketSearchByTickerGet(ctx, ticker)
Получение инструмента по тикеру

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ticker** | [**interface{}**](.md)| Тикер инструмента | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **MarketStocksGet**
> MarketStocksGet(ctx, )
Получение списка акций

### Required Parameters
This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

