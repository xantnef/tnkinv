# \OperationsApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OperationsGet**](OperationsApi.md#OperationsGet) | **Get** /operations | Получение списка операций


# **OperationsGet**
> OperationsGet(ctx, from, to, optional)
Получение списка операций

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **from** | [**interface{}**](.md)| Начало временного промежутка | 
  **to** | [**interface{}**](.md)| Конец временного промежутка | 
 **optional** | ***OperationsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a OperationsGetOpts struct

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **figi** | [**optional.Interface of interface{}**](.md)| Figi инструмента для фильтрации | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

