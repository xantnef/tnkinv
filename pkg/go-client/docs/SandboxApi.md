# \SandboxApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SandboxClearPost**](SandboxApi.md#SandboxClearPost) | **Post** /sandbox/clear | Удаление всех позиций
[**SandboxCurrenciesBalancePost**](SandboxApi.md#SandboxCurrenciesBalancePost) | **Post** /sandbox/currencies/balance | Выставление баланса по валютным позициям
[**SandboxPositionsBalancePost**](SandboxApi.md#SandboxPositionsBalancePost) | **Post** /sandbox/positions/balance | Выставление баланса по инструментным позициям
[**SandboxRegisterPost**](SandboxApi.md#SandboxRegisterPost) | **Post** /sandbox/register | Регистрация клиента в sandbox
[**SandboxRemovePost**](SandboxApi.md#SandboxRemovePost) | **Post** /sandbox/remove | Удаление счета


# **SandboxClearPost**
> SandboxClearPost(ctx, optional)
Удаление всех позиций

Удаление всех позиций клиента

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SandboxClearPostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SandboxClearPostOpts struct

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

# **SandboxCurrenciesBalancePost**
> SandboxCurrenciesBalancePost(ctx, optional)
Выставление баланса по валютным позициям

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SandboxCurrenciesBalancePostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SandboxCurrenciesBalancePostOpts struct

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

# **SandboxPositionsBalancePost**
> SandboxPositionsBalancePost(ctx, optional)
Выставление баланса по инструментным позициям

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SandboxPositionsBalancePostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SandboxPositionsBalancePostOpts struct

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

# **SandboxRegisterPost**
> SandboxRegisterPost(ctx, )
Регистрация клиента в sandbox

Создание счета и валютных позиций для клиента

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

# **SandboxRemovePost**
> SandboxRemovePost(ctx, optional)
Удаление счета

Удаление счета клиента

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SandboxRemovePostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SandboxRemovePostOpts struct

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

