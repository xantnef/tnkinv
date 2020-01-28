# \PortfolioApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**PortfolioCurrenciesGet**](PortfolioApi.md#PortfolioCurrenciesGet) | **Get** /portfolio/currencies | Получение валютных активов клиента
[**PortfolioGet**](PortfolioApi.md#PortfolioGet) | **Get** /portfolio | Получение портфеля клиента


# **PortfolioCurrenciesGet**
> PortfolioCurrenciesGet(ctx, optional)
Получение валютных активов клиента

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***PortfolioCurrenciesGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PortfolioCurrenciesGetOpts struct

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **brokerAccountId** | [**optional.Interface of interface{}**](.md)| Номер счета (по умолчанию - Тинькофф) | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PortfolioGet**
> PortfolioGet(ctx, optional)
Получение портфеля клиента

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***PortfolioGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PortfolioGetOpts struct

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **brokerAccountId** | [**optional.Interface of interface{}**](.md)| Номер счета (по умолчанию - Тинькофф) | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

