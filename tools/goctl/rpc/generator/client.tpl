{{.head}}

package client

import (
    "github.com/zeromicro/go-zero/zrpc"
    {{.imports}}
)


type Client struct {
    {{.cli}}
}

func New{{.service}}(cli zrpc.Client) Client {
    return Client{
        {{.newCli}}
    }
}

