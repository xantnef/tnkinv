# \OrdersApi

All URIs are relative to *https://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OrdersCancelPost**](OrdersApi.md#OrdersCancelPost) | **Post** /orders/cancel | Отмена заявки
[**OrdersGet**](OrdersApi.md#OrdersGet) | **Get** /orders | Получение списка активных заявок
[**OrdersLimitOrderPost**](OrdersApi.md#OrdersLimitOrderPost) | **Post** /orders/limit-order | Создание лимитной заявки


# **OrdersCancelPost**
> OrdersCancelPost(ctx, orderId)
Отмена заявки

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **orderId** | [**interface{}**](.md)| ID заявки | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **OrdersGet**
> OrdersGet(ctx, )
Получение списка активных заявок

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

# **OrdersLimitOrderPost**
> OrdersLimitOrderPost(ctx, figi)
Создание лимитной заявки

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **figi** | [**interface{}**](.md)| FIGI инструмента | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

