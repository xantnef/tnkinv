# tnkinv

[Tinkoff OpenAPI](https://tinkoffcreditsystems.github.io/invest-openapi/) client.

Currently safe and readonly; and only prints some portfolio information.

## Running
```
tnkinv --token {file_with_token} [--sandtoken {file_with_sandbox_token}]
```

## Info

[Online Swagger Generator](https://generator.swagger.io/) is used for basic client generation (pkg/go-client).

```
echo '{"options":{}, "spec": ' $(cat api/swagger.json ) '}' | curl -X POST -H "content-type:application/json" -d @- https://generator.swagger.io/api/gen/clients/go
```

then get and unzip the result
