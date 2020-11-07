# tnkinv

[Tinkoff OpenAPI](https://tinkoffcreditsystems.github.io/invest-openapi/) client.

Currently safe and readonly; and only prints some portfolio information.

## Running
```
 tnkinv {subcmd} [params] --token file_with_token
   common params:
     --account broker|iis|all
     --operations filename
     --fictives filename
     --loglevel {debug|all}
   subcmds:
     show   [--at 1922/12/28 (default: today)]
     story  [--start 1901/01/01 (default: year ago)]
            [--period day|week|month (default: month)]
            [--format human|table (default: human)]
     deals  [--start 1901/01/01 (default: none)]
            [--end 1902/02/02 (default: now)]
            [--period day|week|month|all (default: month)]
     price  --tickers ticker1,ticker2,..
            [--start 1901/01/01 (default: year ago)]
            [--end 1902/02/02 (default: now)]
     sandbox
```

## Info

[Online Swagger Generator](https://generator.swagger.io/) is used for basic client generation (pkg/go-client).

```
echo '{"options":{}, "spec": ' $(cat api/swagger.json ) '}' | curl -X POST -H "content-type:application/json" -d @- https://generator.swagger.io/api/gen/clients/go
```

then get and unzip the result
